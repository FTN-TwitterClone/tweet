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

func (r *RedisTweetRepository) Post(ctx context.Context, imageId string, image []byte) error {
	_, span := r.tracer.Start(ctx, "RedisTweetRepository.Post")
	defer span.End()

	err := r.cli.Set(constructKey(imageId), image, 30*time.Second).Err()

	return err
}

func (r *RedisTweetRepository) Get(ctx context.Context, imageId string) ([]byte, error) {
	_, span := r.tracer.Start(ctx, "RedisTweetRepository.Get")
	defer span.End()

	value, err := r.cli.Get(constructKey(imageId)).Bytes()
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (r *RedisTweetRepository) Exists(ctx context.Context, imageId string) bool {
	_, span := r.tracer.Start(ctx, "RedisTweetRepository.Exists")
	defer span.End()

	cnt, err := r.cli.Exists(constructKey(imageId)).Result()
	if err != nil {
		return false
	}
	return cnt == 1
}

const (
	cacheImage = "images:%s"
	cacheAll   = "images"
)

func constructKey(id string) string {
	return fmt.Sprintf(cacheImage, id)
}
