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
	GetTimelineTweets(ctx context.Context, username string, lastTweetId string) (*[]model.TweetDTO, error)
	GetFeedTweets(ctx context.Context, username string, lastTweetId string) (*[]model.TweetDTO, error)
	GetLikesByTweet(ctx context.Context, tweetId string) *[]model.Like
	CountLikes(ctx context.Context, tweetId *gocql.UUID) (int16, error)
	FindTweet(ctx context.Context, tweetId string) (model.Tweet, error)
	LikedByMe(ctx context.Context, tweetId *gocql.UUID) (bool, error)
}
