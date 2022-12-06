package repository

import (
	"context"
	"github.com/FTN-TwitterClone/grpc-stubs/proto/social_graph"
	"github.com/gocql/gocql"
	"tweet/model"
)

type CassandraRepository interface {
	SaveTweet(ctx context.Context, tweet *model.TweetDTO, usernames []*social_graph.SocialGraphUsername) error
	SaveLike(ctx context.Context, like *model.Like) error
	DeleteLike(ctx context.Context, id string, username string) error
	GetTimelineTweets(ctx context.Context, username string, lastTweetId string) ([]model.TweetDTO, error)
	GetFeedTweets(ctx context.Context, username string, lastTweetId string) ([]model.TweetDTO, error)
	GetLikesByTweet(ctx context.Context, tweetId string) *[]model.Like
	CountLikes(ctx context.Context, tweetId *gocql.UUID) (int16, error)
	FindTweet(ctx context.Context, tweetId string) (model.Tweet, error)
	FindUserTweets(ctx context.Context, username string) []model.Tweet
	LikedByMe(ctx context.Context, tweetId *gocql.UUID) (bool, error)
	UpdateFeed(ctx context.Context, authUsername string, username string) error
}
