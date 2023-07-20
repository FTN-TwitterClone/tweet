package repository

import (
	"context"
)

type RedisRepository interface {
	PostImage(ctx context.Context, imageId string, image []byte) error
	GetImage(ctx context.Context, imageId string) ([]byte, error)
	ImageExists(ctx context.Context, imageId string) bool
	PostToken(ctx context.Context, username string, token string) error
	GetToken(ctx context.Context, username string) ([]byte, error)
	TokenExists(ctx context.Context, username string) bool
}
