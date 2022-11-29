package service

import (
	"context"
	"github.com/FTN-TwitterClone/grpc-stubs/proto/social_graph"
	"github.com/gocql/gocql"
	"github.com/golang/protobuf/ptypes/empty"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"log"
	"tweet/app_errors"
	"tweet/model"
	"tweet/repository"
	"tweet/tls"
)

type TweetService struct {
	tweetRepository repository.TweetRepository
	tracer          trace.Tracer
}

func NewTweetService(tweetRepository repository.TweetRepository, tracer trace.Tracer) *TweetService {
	return &TweetService{
		tweetRepository,
		tracer,
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

	conn, err := getgRPCConnection("social-graph:9001")
	defer conn.Close()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, &app_errors.AppError{500, ""}
	}

	socialGraphService := social_graph.NewSocialGraphServiceClient(conn)
	serviceCtx = metadata.AppendToOutgoingContext(serviceCtx, "authUsername", authUser.Username)

	response, _ := socialGraphService.GetMyFollowers(serviceCtx, new(empty.Empty))

	err = s.tweetRepository.SaveTweet(serviceCtx, &t, response.Usernames)

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, &app_errors.AppError{500, ""}
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
		return nil, &app_errors.AppError{500, ""}
	}

	l := model.Like{
		Username: authUser.Username,
		TweetId:  tweetId,
	}

	err = s.tweetRepository.SaveLike(serviceCtx, &l)

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, &app_errors.AppError{500, ""}
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
		return "", &app_errors.AppError{500, ""}
	}

	return id, nil
}

func (s *TweetService) GetTimelineTweets(ctx context.Context, username string, lastTweetId string) (*[]model.TweetDTO, *app_errors.AppError) {
	serviceCtx, span := s.tracer.Start(ctx, "TweetService.GetProfileTweets")
	defer span.End()

	conn, err := getgRPCConnection("social-graph:9001")
	defer conn.Close()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, &app_errors.AppError{500, ""}
	}

	targetUser := social_graph.SocialGraphUsername{
		Username: username,
	}

	authUser := serviceCtx.Value("authUser").(model.AuthUser)

	socialGraphService := social_graph.NewSocialGraphServiceClient(conn)
	serviceCtx = metadata.AppendToOutgoingContext(serviceCtx, "authUsername", authUser.Username)

	response, err := socialGraphService.CheckVisibility(serviceCtx, &targetUser)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, &app_errors.AppError{500, ""}
	}

	if !response.Visibility {
		return nil, &app_errors.AppError{403, ""}
	}

	tweets, err := s.tweetRepository.GetTimelineTweets(serviceCtx, username, lastTweetId)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, &app_errors.AppError{500, ""}
	}

	var responseTweets []model.TweetDTO
	for _, tweet := range tweets {
		if tweet.Retweet {
			targetUser.Username = tweet.OriginalPostedBy
			response, err = socialGraphService.CheckVisibility(serviceCtx, &targetUser)

			if err != nil || !response.Visibility {
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

	conn, err := getgRPCConnection("social-graph:9001")
	defer conn.Close()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, &app_errors.AppError{500, ""}
	}

	targetUser := social_graph.SocialGraphUsername{}
	authUser := serviceCtx.Value("authUser").(model.AuthUser)

	socialGraphService := social_graph.NewSocialGraphServiceClient(conn)
	serviceCtx = metadata.AppendToOutgoingContext(serviceCtx, "authUsername", authUser.Username)

	tweets, err := s.tweetRepository.GetFeedTweets(serviceCtx, authUser.Username, lastTweetId)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, &app_errors.AppError{500, ""}
	}

	var responseTweets []model.TweetDTO
	for _, tweet := range tweets {
		if tweet.Retweet {
			targetUser.Username = tweet.OriginalPostedBy
			response, err := socialGraphService.CheckVisibility(serviceCtx, &targetUser)

			if err != nil || !response.Visibility {
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
		return nil, &app_errors.AppError{500, "Tweet not found"}
	}

	if tweet.Retweet {
		return nil, &app_errors.AppError{406, "You cant retweet a retweet"}
	}

	conn, err := getgRPCConnection("social-graph:9001")
	defer conn.Close()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, &app_errors.AppError{500, ""}
	}

	targetUser := social_graph.SocialGraphUsername{
		Username: tweet.PostedBy,
	}

	authUser := serviceCtx.Value("authUser").(model.AuthUser)

	socialGraphService := social_graph.NewSocialGraphServiceClient(conn)
	serviceCtx = metadata.AppendToOutgoingContext(serviceCtx, "authUsername", authUser.Username)

	response, err := socialGraphService.CheckVisibility(serviceCtx, &targetUser)

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, &app_errors.AppError{500, ""}
	}

	if !response.Visibility {
		return nil, &app_errors.AppError{403, ""}
	}

	id := gocql.TimeUUID()
	t := model.Tweet{
		ID:               id,
		PostedBy:         authUser.Username,
		Text:             tweet.Text,
		TimeStamp:        id.Time(),
		Retweet:          true,
		OriginalPostedBy: tweet.PostedBy,
	}

	followers, _ := socialGraphService.GetMyFollowers(serviceCtx, new(empty.Empty))

	err = s.tweetRepository.SaveTweet(serviceCtx, &t, followers.Usernames)

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, &app_errors.AppError{500, ""}
	}

	return &t, nil
}

func getgRPCConnection(address string) (*grpc.ClientConn, error) {
	creds := credentials.NewTLS(tls.GetgRPCClientTLSConfig())

	conn, err := grpc.DialContext(
		context.Background(),
		address,
		grpc.WithTransportCredentials(creds),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
	)

	if err != nil {
		log.Fatalf("Failed to start gRPC connection: %v", err)
	}

	return conn, err
}
