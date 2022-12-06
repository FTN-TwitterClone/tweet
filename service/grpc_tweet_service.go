package service

import (
	"context"
	"github.com/FTN-TwitterClone/grpc-stubs/proto/tweet"
	"github.com/golang/protobuf/ptypes/empty"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
	"tweet/repository"
)

type gRPCTweetService struct {
	tweet.UnimplementedTweetServiceServer
	tracer              trace.Tracer
	cassandraRepository repository.CassandraRepository
}

func NewgRPCTweetService(tracer trace.Tracer, cassandraRepository repository.CassandraRepository) *gRPCTweetService {
	return &gRPCTweetService{
		tracer:              tracer,
		cassandraRepository: cassandraRepository,
	}
}

func (s gRPCTweetService) UpdateFeed(ctx context.Context, user *tweet.User) (*empty.Empty, error) {
	serviceCtx, span := s.tracer.Start(ctx, "gRPCTweetService.UpdateFeed")
	defer span.End()

	md, _ := metadata.FromIncomingContext(ctx)
	authUsername := md.Get("authUsername")[0]

	err := s.cassandraRepository.UpdateFeed(serviceCtx, authUsername, user.Username)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return new(empty.Empty), err
	}

	return new(empty.Empty), nil
}
