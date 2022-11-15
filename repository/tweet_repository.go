package repository

import (
	"context"
	"github.com/gocql/gocql"
	"tweet/model"
)

type TweetRepository interface {
	SaveTweet(ctx context.Context, tweet *model.Tweet) error
	SaveLike(ctx context.Context, like *model.Like) error
	DeleteLike(ctx context.Context, id string, username string) error
	GetProfileTweets(ctx context.Context, username string, lastTweetId string) (*[]model.TweetDTO, error)
	CountLikes(ctx context.Context, tweetId *gocql.UUID) (int16, error)
	LikedByMe(ctx context.Context, tweetId *gocql.UUID) (bool, error)
}
