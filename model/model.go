package model

import (
	"github.com/gocql/gocql"
	"time"
)

// Info from JWT token
type AuthUser struct {
	Username string
	Role     string
	Exp      time.Time
}

type Tweet struct {
	ID               gocql.UUID `json:"id"`
	PostedBy         string     `json:"postedBy"`
	Text             string     `json:"text"`
	TimeStamp        time.Time  `json:"timestamp"`
	Retweet          bool       `json:"retweet"`
	OriginalPostedBy string     `json:"originalPostedBy"`
	//Photo  string `json:"photo"` TODO save photo to db
}

type TweetDTO struct {
	ID               gocql.UUID `json:"id"`
	PostedBy         string     `json:"postedBy"`
	Text             string     `json:"text"`
	Timestamp        time.Time  `json:"timestamp"`
	Retweet          bool       `json:"retweet"`
	OriginalPostedBy string     `json:"originalPostedBy"`
	LikesCount       int16      `json:"likesCount"`
	LikedByMe        bool       `json:"likedByMe"`
}

type Like struct {
	Username string     `json:"username"`
	TweetId  gocql.UUID `json:"tweetId"`
}
