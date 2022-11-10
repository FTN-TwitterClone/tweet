package service

import (
	"context"
	"github.com/gocql/gocql"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"time"
	"tweet/app_errors"
	"tweet/model"
	"tweet/repository"
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

	//authUser := serviceCtx.Value("authUser").(model.AuthUser)

	t := model.Tweet{
		ID: gocql.TimeUUID(),
		//Username:      authUser.Username,
		Username:  "usernameTest",
		Text:      tweet.Text,
		Timestamp: time.Now(),
	}

	err := s.tweetRepository.SaveTweet(serviceCtx, &t)

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, &app_errors.AppError{500, ""}
	}

	return &t, nil
}

func (s *TweetService) CreateLike(ctx context.Context, like model.Like) (*model.Like, *app_errors.AppError) {
	serviceCtx, span := s.tracer.Start(ctx, "TweetService.CreateLike")
	defer span.End()

	//authUser := serviceCtx.Value("authUser").(model.AuthUser)

	likeExists, err := s.tweetRepository.LikeExists(serviceCtx, "usernameTest", like.TweetId)

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, &app_errors.AppError{500, ""}
	}

	if likeExists {
		return nil, &app_errors.AppError{500, "You are already liked this tweet."}
	}

	l := model.Like{
		//Username:      authUser.Username,
		Username: "usernameTest",
		TweetId:  like.TweetId,
	}

	err = s.tweetRepository.SaveLike(serviceCtx, &l)

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, &app_errors.AppError{500, ""}
	}

	return &l, nil
}
