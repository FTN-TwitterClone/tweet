package service

import (
	"context"
	"github.com/FTN-TwitterClone/grpc-stubs/proto/social_graph"
	"github.com/gocql/gocql"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"tweet/app_errors"
	"tweet/model"
	"tweet/repository"
	"tweet/service/circuit_breaker"
)

type TweetService struct {
	tweetRepository repository.TweetRepository
	tracer          trace.Tracer
	socialGraphCB   *circuit_breaker.SocialGraphCircuitBreaker
}

func NewTweetService(tweetRepository repository.TweetRepository, tracer trace.Tracer, socialGraphCB *circuit_breaker.SocialGraphCircuitBreaker) *TweetService {
	return &TweetService{
		tweetRepository,
		tracer,
		socialGraphCB,
	}
}

func (s *TweetService) CreateTweet(ctx context.Context, tweet model.Tweet) (*model.Tweet, *app_errors.AppError) {
	serviceCtx, span := s.tracer.Start(ctx, "TweetService.CreateTweet")
	defer span.End()

	authUser := serviceCtx.Value("authUser").(model.AuthUser)
	id := gocql.TimeUUID()

	t := model.Tweet{
		ID:        id,
		PostedBy:  authUser.Username,
		Text:      tweet.Text,
		TimeStamp: id.Time(),
	}

	followers, err := s.socialGraphCB.GetMyFollowers(serviceCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}

	repoErr := s.tweetRepository.SaveTweet(serviceCtx, &t, followers)

	if repoErr != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, &app_errors.AppError{Code: 500, Message: err.Error()}
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

	err = s.tweetRepository.SaveLike(serviceCtx, &l)

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

	err := s.tweetRepository.DeleteLike(serviceCtx, id, authUser.Username)
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

	tweets, repoErr := s.tweetRepository.GetTimelineTweets(serviceCtx, username, lastTweetId)
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

			if err != nil || !visibility {
				tweet.Text = ""
			}
		}
		responseTweets = append(responseTweets, tweet)
	}

	return &responseTweets, nil
}

func (s *TweetService) GetLikesByTweet(ctx context.Context, tweetId string) *[]model.Like {
	serviceCtx, span := s.tracer.Start(ctx, "TweetService.GetLikesByTweet")
	defer span.End()

	likes := s.tweetRepository.GetLikesByTweet(serviceCtx, tweetId)

	return likes
}

func (s *TweetService) GetHomeFeed(ctx context.Context, lastTweetId string) (*[]model.TweetDTO, *app_errors.AppError) {
	serviceCtx, span := s.tracer.Start(ctx, "TweetService.GetHomeFeed")
	defer span.End()

	targetUser := social_graph.SocialGraphUsername{}
	authUser := serviceCtx.Value("authUser").(model.AuthUser)

	tweets, err := s.tweetRepository.GetFeedTweets(serviceCtx, authUser.Username, lastTweetId)
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

			if err != nil || !visible {
				tweet.Text = ""
			}
		}
		responseTweets = append(responseTweets, tweet)
	}

	return &responseTweets, nil
}

func (s *TweetService) Retweet(ctx context.Context, tweetId string) (*model.Tweet, *app_errors.AppError) {
	serviceCtx, span := s.tracer.Start(ctx, "TweetService.Retweet")
	defer span.End()

	tweet, err := s.tweetRepository.FindTweet(serviceCtx, tweetId)
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
	t := model.Tweet{
		ID:               id,
		PostedBy:         authUser.Username,
		Text:             tweet.Text,
		TimeStamp:        id.Time(),
		Retweet:          true,
		OriginalPostedBy: tweet.PostedBy,
	}

	followers, sbErr := s.socialGraphCB.GetMyFollowers(serviceCtx)

	if sbErr != nil && sbErr.Code == 503 {
		span.SetStatus(codes.Error, sbErr.Error())
		return nil, &app_errors.AppError{Code: 503, Message: "Service unavailable"}
	}

	err = s.tweetRepository.SaveTweet(serviceCtx, &t, followers)

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, &app_errors.AppError{Code: 500, Message: err.Error()}
	}

	return &t, nil
}
