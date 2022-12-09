package repository

import (
	"context"
)

type RedisRepository interface {
	Post(ctx context.Context, imageId string, image []byte) error
	Get(ctx context.Context, imageId string) ([]byte, error)
	Exists(ctx context.Context, imageId string) bool
}
