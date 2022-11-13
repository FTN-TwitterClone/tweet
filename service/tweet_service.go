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

	authUser := serviceCtx.Value("authUser").(model.AuthUser)

	t := model.Tweet{
		ID:        gocql.TimeUUID(),
		Username:  authUser.Username,
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
