package service

import (
	"context"
	"github.com/FTN-TwitterClone/grpc-stubs/proto/social_graph"
	"github.com/gocql/gocql"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"io"
	"net/http"
	"os"
	"tweet/app_errors"
	"tweet/model"
	"tweet/repository"
	"tweet/service/circuit_breaker"
)

type TweetService struct {
	cassandraRepository repository.CassandraRepository
	cache               repository.RedisRepository
	tracer              trace.Tracer
	socialGraphCB       *circuit_breaker.SocialGraphCircuitBreaker
}

func NewTweetService(cassandraRepository repository.CassandraRepository, redisRepository repository.RedisRepository, tracer trace.Tracer, socialGraphCB *circuit_breaker.SocialGraphCircuitBreaker) *TweetService {
	return &TweetService{
		cassandraRepository,
		redisRepository,
		tracer,
		socialGraphCB,
	}
}

func (s *TweetService) CreateTweet(ctx context.Context, tweet model.Tweet) (*model.TweetDTO, *app_errors.AppError) {
	serviceCtx, span := s.tracer.Start(ctx, "TweetService.CreateTweet")
	defer span.End()

	authUser := serviceCtx.Value("authUser").(model.AuthUser)
	id := gocql.TimeUUID()

	t := model.TweetDTO{
		ID:               id,
		PostedBy:         authUser.Username,
		Text:             tweet.Text,
		ImageId:          tweet.ImageId,
		Timestamp:        id.Time(),
		LikesCount:       0,
		LikedByMe:        false,
		Retweet:          false,
		OriginalPostedBy: "",
		Ad:               false,
	}
	if len(tweet.ImageId) > 0 {
		t.Image, _ = s.GetImage(serviceCtx, tweet.ImageId)
	}

	followers, err := s.socialGraphCB.GetMyFollowers(serviceCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}

	repoErr := s.cassandraRepository.SaveTweet(serviceCtx, &t, followers)

	if repoErr != nil {
		span.SetStatus(codes.Error, repoErr.Error())
		return nil, &app_errors.AppError{Code: 500, Message: repoErr.Error()}
	}

	return &t, nil
}

func (s *TweetService) CreateAd(ctx context.Context, ad model.Ad, authUser model.AuthUser) (*model.TweetDTO, *app_errors.AppError) {
	serviceCtx, span := s.tracer.Start(ctx, "TweetService.CreateAd")
	defer span.End()
	//TODO call ads service via grpc
	id := gocql.TimeUUID()

	t := model.TweetDTO{
		ID:               id,
		PostedBy:         authUser.Username,
		Text:             ad.Tweet.Text,
		ImageId:          ad.Tweet.ImageId,
		Timestamp:        id.Time(),
		LikesCount:       0,
		LikedByMe:        false,
		Retweet:          false,
		OriginalPostedBy: "",
		Ad:               true,
	}
	if len(ad.Tweet.ImageId) > 0 {
		t.Image, _ = s.GetImage(serviceCtx, ad.Tweet.ImageId)
	}

	followers, err := s.socialGraphCB.GetMyFollowers(serviceCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}

	targetGroupUsers, err := s.socialGraphCB.GetTargetGroupUsers(serviceCtx, ad.TargetGroup)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}

	targetGroupUsers = append(targetGroupUsers, followers...)

	repoErr := s.cassandraRepository.SaveTweet(serviceCtx, &t, targetGroupUsers)

	if repoErr != nil {
		span.SetStatus(codes.Error, repoErr.Error())
		return nil, &app_errors.AppError{Code: 500, Message: repoErr.Error()}
	}

	return &t, nil
}

func (s *TweetService) CreateLike(ctx context.Context, id string) (*model.Like, *app_errors.AppError) {
	serviceCtx, span := s.tracer.Start(ctx, "TweetService.CreateLike")
	defer span.End()

	authUser := serviceCtx.Value("authUser").(model.AuthUser)
	tweetId, err := gocql.ParseUUID(id)

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, &app_errors.AppError{Code: 500, Message: err.Error()}
	}

	l := model.Like{
		Username: authUser.Username,
		TweetId:  tweetId,
	}

	//TODO call ads service via grpc if tweet is ad

	err = s.cassandraRepository.SaveLike(serviceCtx, &l)

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, &app_errors.AppError{Code: 500, Message: err.Error()}
	}

	return &l, nil
}

func (s *TweetService) DeleteLike(ctx context.Context, id string) (string, *app_errors.AppError) {
	serviceCtx, span := s.tracer.Start(ctx, "TweetService.DeleteLike")
	defer span.End()

	authUser := serviceCtx.Value("authUser").(model.AuthUser)

	//TODO call ads service via grpc if tweet is ad

	err := s.cassandraRepository.DeleteLike(serviceCtx, id, authUser.Username)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", &app_errors.AppError{Code: 500, Message: err.Error()}
	}

	return id, nil
}

func (s *TweetService) GetTimelineTweets(ctx context.Context, username string, lastTweetId string) (*[]model.TweetDTO, *app_errors.AppError) {
	serviceCtx, span := s.tracer.Start(ctx, "TweetService.GetProfileTweets")
	defer span.End()

	targetUser := social_graph.SocialGraphUsername{
		Username: username,
	}

	visibility, err := s.socialGraphCB.CheckVisibility(serviceCtx, &targetUser)
	if err != nil && err.Code == 503 {
		span.SetStatus(codes.Error, err.Error())
		return nil, &app_errors.AppError{Code: 503, Message: "Service unavailable"}
	}

	if !visibility {
		return nil, &app_errors.AppError{Code: 403}
	}

	tweets, repoErr := s.cassandraRepository.GetTimelineTweets(serviceCtx, username, lastTweetId)
	if repoErr != nil {
		span.SetStatus(codes.Error, repoErr.Error())
		return nil, &app_errors.AppError{Code: 500, Message: repoErr.Error()}
	}

	var responseTweets []model.TweetDTO
	for _, tweet := range tweets {
		if tweet.Retweet {
			targetUser.Username = tweet.OriginalPostedBy
			visibility, err = s.socialGraphCB.CheckVisibility(serviceCtx, &targetUser)

			if err != nil && err.Code == 503 {
				continue
			}

			if !visibility {
				tweet.Text = ""
			} else if len(tweet.ImageId) > 0 {
				tweet.Image, _ = s.GetImage(serviceCtx, tweet.ImageId)
			}

		} else if len(tweet.ImageId) > 0 { // if photo is present
			tweet.Image, _ = s.GetImage(serviceCtx, tweet.ImageId)
		}
		responseTweets = append(responseTweets, tweet)
	}

	return &responseTweets, nil
}

func (s *TweetService) GetLikesByTweet(ctx context.Context, tweetId string) *[]model.Like {
	serviceCtx, span := s.tracer.Start(ctx, "TweetService.GetLikesByTweet")
	defer span.End()

	likes := s.cassandraRepository.GetLikesByTweet(serviceCtx, tweetId)

	return likes
}

func (s *TweetService) GetHomeFeed(ctx context.Context, lastTweetId string) (*[]model.TweetDTO, *app_errors.AppError) {
	serviceCtx, span := s.tracer.Start(ctx, "TweetService.GetHomeFeed")
	defer span.End()

	targetUser := social_graph.SocialGraphUsername{}
	authUser := serviceCtx.Value("authUser").(model.AuthUser)

	tweets, err := s.cassandraRepository.GetFeedTweets(serviceCtx, authUser.Username, lastTweetId)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, &app_errors.AppError{Code: 500, Message: err.Error()}
	}

	var responseTweets []model.TweetDTO
	for _, tweet := range tweets {
		if tweet.Retweet {
			targetUser.Username = tweet.OriginalPostedBy
			visible, err := s.socialGraphCB.CheckVisibility(serviceCtx, &targetUser)

			if err != nil && err.Code == 503 {
				continue
			}

			if !visible {
				tweet.Text = ""
			} else if len(tweet.ImageId) > 0 {
				tweet.Image, _ = s.GetImage(serviceCtx, tweet.ImageId)
			}

		} else if len(tweet.ImageId) > 0 { // if photo is present
			tweet.Image, _ = s.GetImage(serviceCtx, tweet.ImageId)
		}
		responseTweets = append(responseTweets, tweet)
	}

	return &responseTweets, nil
}

func (s *TweetService) Retweet(ctx context.Context, tweetId string) (*model.TweetDTO, *app_errors.AppError) {
	serviceCtx, span := s.tracer.Start(ctx, "TweetService.Retweet")
	defer span.End()

	tweet, err := s.cassandraRepository.FindTweet(serviceCtx, tweetId)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, &app_errors.AppError{Code: 500, Message: "Tweet not found"}
	}

	if tweet.Retweet {
		return nil, &app_errors.AppError{Code: 406, Message: "You cant retweet a retweet"}
	}

	targetUser := social_graph.SocialGraphUsername{
		Username: tweet.PostedBy,
	}

	visibility, sbErr := s.socialGraphCB.CheckVisibility(serviceCtx, &targetUser)

	if sbErr != nil && sbErr.Code == 503 {
		span.SetStatus(codes.Error, sbErr.Error())
		return nil, &app_errors.AppError{Code: 503, Message: "Service unavailable"}
	}

	if !visibility {
		return nil, &app_errors.AppError{Code: 403}
	}

	authUser := serviceCtx.Value("authUser").(model.AuthUser)
	id := gocql.TimeUUID()
	t := model.TweetDTO{
		ID:               id,
		PostedBy:         authUser.Username,
		Text:             tweet.Text,
		ImageId:          tweet.ImageId,
		Timestamp:        id.Time(),
		Retweet:          true,
		OriginalPostedBy: tweet.PostedBy,
		LikedByMe:        false,
		LikesCount:       0,
		Ad:               tweet.Ad,
	}

	if len(tweet.ImageId) > 0 {
		t.Image, _ = s.GetImage(serviceCtx, tweet.ImageId)
	}

	followers, sbErr := s.socialGraphCB.GetMyFollowers(serviceCtx)

	if sbErr != nil && sbErr.Code == 503 {
		span.SetStatus(codes.Error, sbErr.Error())
		return nil, &app_errors.AppError{Code: 503, Message: "Service unavailable"}
	}

	err = s.cassandraRepository.SaveTweet(serviceCtx, &t, followers)

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, &app_errors.AppError{Code: 500, Message: err.Error()}
	}

	return &t, nil
}

func (s *TweetService) SaveImage(ctx context.Context, req *http.Request) (*string, *app_errors.AppError) {
	_, span := s.tracer.Start(ctx, "TweetService.SaveImage")
	defer span.End()

	// left shift 32 << 20 which results in 32*2^20 = 33554432
	// x << y, results in x*2^y
	err := req.ParseMultipartForm(32 << 20)
	if err != nil {
		return nil, &app_errors.AppError{Code: 500, Message: err.Error()}
	}
	// Retrieve the file from form data
	f, _, err := req.FormFile("image")
	if err != nil {
		return nil, &app_errors.AppError{Code: 500, Message: err.Error()}
	}
	defer f.Close()
	imageName := gocql.TimeUUID().String()
	fullPath := os.Getenv("IMAGES") + "/" + imageName
	file, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, &app_errors.AppError{Code: 500, Message: err.Error()}
	}
	defer file.Close()
	// Copy the file to the destination path
	_, err = io.Copy(file, f)
	if err != nil {
		return nil, &app_errors.AppError{Code: 500, Message: err.Error()}
	}
	return &imageName, nil
}

func (s *TweetService) GetImage(ctx context.Context, imageId string) ([]byte, error) {
	serviceCtx, span := s.tracer.Start(ctx, "TweetService.GetImage")
	defer span.End()

	image, err := s.cache.Get(serviceCtx, imageId)
	if err != nil {
		//time.Sleep(10 * time.Second) // proof of concept

		fullPath := os.Getenv("IMAGES") + "/" + imageId
		image, err = os.ReadFile(fullPath)

		if err != nil {
			span.SetStatus(500, err.Error())
			return nil, err
		}

		err = s.cache.Post(serviceCtx, imageId, image)
		if err != nil {
			span.SetStatus(500, err.Error())
		}
	}
	return image, nil
}
