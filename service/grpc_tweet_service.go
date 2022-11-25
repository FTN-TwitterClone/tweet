package service

import (
	"context"
	"github.com/FTN-TwitterClone/grpc-stubs/proto/tweet"
	"github.com/golang/protobuf/ptypes/empty"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"tweet/repository"
)

type gRPCTweetService struct {
	tweet.UnimplementedTweetServiceServer
	tracer          trace.Tracer
	tweetRepository repository.TweetRepository
}

func NewgRPCTweetService(tracer trace.Tracer, tweetRepository repository.TweetRepository) *gRPCTweetService {
	return &gRPCTweetService{
		tracer:          tracer,
		tweetRepository: tweetRepository,
	}
}

func (s gRPCTweetService) UpdateFeed(ctx context.Context, user *tweet.User) (*empty.Empty, error) {
	serviceCtx, span := s.tracer.Start(ctx, "gRPCTweetService.UpdateFeed")
	defer span.End()

	err := s.tweetRepository.UpdateFeed(serviceCtx, user.Username)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return new(empty.Empty), err
	}

	return new(empty.Empty), nil
}
