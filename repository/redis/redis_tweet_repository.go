package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"go.opentelemetry.io/otel/trace"
	"os"
	"time"
)

type RedisTweetRepository struct {
	tracer trace.Tracer
	cli    *redis.Client
}

func NewRedisTweetRepository(tracer trace.Tracer) *RedisTweetRepository {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisAddress := fmt.Sprintf("%s:%s", redisHost, redisPort)

	client := redis.NewClient(&redis.Options{
		Addr: redisAddress,
	})

	return &RedisTweetRepository{
		cli:    client,
		tracer: tracer,
	}
}

func (r *RedisTweetRepository) PostImage(ctx context.Context, imageId string, image []byte) error {
	_, span := r.tracer.Start(ctx, "RedisTweetRepository.PostImage")
	defer span.End()

	err := r.cli.Set(constructImageKey(imageId), image, 30*time.Second).Err()

	return err
}

func (r *RedisTweetRepository) GetImage(ctx context.Context, imageId string) ([]byte, error) {
	_, span := r.tracer.Start(ctx, "RedisTweetRepository.GetImage")
	defer span.End()

	value, err := r.cli.Get(constructImageKey(imageId)).Bytes()
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (r *RedisTweetRepository) ImageExists(ctx context.Context, imageId string) bool {
	_, span := r.tracer.Start(ctx, "RedisTweetRepository.ImageExists")
	defer span.End()

	cnt, err := r.cli.Exists(constructImageKey(imageId)).Result()
	if err != nil {
		return false
	}
	return cnt == 1
}

func (r *RedisTweetRepository) PostToken(ctx context.Context, username string, token string) error {
	_, span := r.tracer.Start(ctx, "RedisTweetRepository.PostToken")
	defer span.End()

	err := r.cli.Set(constructTokenKey(username), token, 0).Err()

	return err
}

func (r *RedisTweetRepository) GetToken(ctx context.Context, username string) (string, error) {
	_, span := r.tracer.Start(ctx, "RedisTweetRepository.GetToken")
	defer span.End()

	value, err := r.cli.Get(constructTokenKey(username)).Bytes()
	if err != nil {
		return "", err
	}

	return string(value), nil
}

func (r *RedisTweetRepository) TokenExists(ctx context.Context, username string) bool {
	_, span := r.tracer.Start(ctx, "RedisTweetRepository.TokenExists")
	defer span.End()

	cnt, err := r.cli.Exists(constructTokenKey(username)).Result()
	if err != nil {
		return false
	}
	return cnt == 1
}

const (
	cacheImage = "images:%s"
	cacheToken = "token:%s"
	cacheAll   = "images"
)

func constructImageKey(id string) string {
	return fmt.Sprintf(cacheImage, id)
}

func constructTokenKey(id string) string {
	return fmt.Sprintf(cacheToken, id)
}
