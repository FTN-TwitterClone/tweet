package repository

import (
	"context"
	"github.com/gocql/gocql"
	"tweet/model"
)

type TweetRepository interface {
	SaveTweet(ctx context.Context, tweet *model.Tweet) error
	SaveLike(ctx context.Context, like *model.Like) error
	LikeExists(ctx context.Context, username string, tweetId gocql.UUID) (bool, error)
}
