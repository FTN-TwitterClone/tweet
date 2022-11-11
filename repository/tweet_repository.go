package repository

import (
	"context"
	"tweet/model"
)

type TweetRepository interface {
	SaveTweet(ctx context.Context, tweet *model.Tweet) error
	SaveLike(ctx context.Context, like *model.Like) error
}
