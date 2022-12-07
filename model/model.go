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
	ImageId          string     `json:"imageId"`
	Timestamp        time.Time  `json:"timestamp"`
	Retweet          bool       `json:"retweet"`
	OriginalPostedBy string     `json:"originalPostedBy"`
	Ad               bool       `json:"ad"`
}

type TweetDTO struct {
	ID               gocql.UUID `json:"id"`
	PostedBy         string     `json:"postedBy"`
	Text             string     `json:"text"`
	ImageId          string     `json:"-"` //only for backend
	Image            []byte     `json:"image"`
	Timestamp        time.Time  `json:"timestamp"`
	Retweet          bool       `json:"retweet"`
	OriginalPostedBy string     `json:"originalPostedBy"`
	LikesCount       int16      `json:"likesCount"`
	LikedByMe        bool       `json:"likedByMe"`
	Ad               bool       `json:"ad"`
}

type Like struct {
	Username string     `json:"username"`
	TweetId  gocql.UUID `json:"tweetId"`
}

// Ad proof of concept structs
type Ad struct {
	Tweet       Tweet       `json:"tweet"`
	TargetGroup TargetGroup `json:"targetGroup"`
}

type TargetGroup struct {
	City    string `json:"city"`
	Gender  string `json:"gender"`
	AgeFrom int    `json:"ageFrom"`
	AgeTo   int    `json:"ageTo"`
}
