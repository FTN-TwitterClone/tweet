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
	Text             string     `json:"text" validate:"required_without=ImageId"`
	ImageId          string     `json:"imageId" validate:"required_without=Text"`
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

type CodeDTO struct {
	Code string `json:"code"`
}

type TokenResponseDTO struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type Like struct {
	Username string     `json:"username"`
	TweetId  gocql.UUID `json:"tweetId"`
}

// Ad proof of concept structs
type Ad struct {
	Tweet       Tweet       `json:"tweet" validate:"required"`
	TargetGroup TargetGroup `json:"targetGroup" validate:"required"`
}

type TargetGroup struct {
	Town   string `json:"town"   validate:"required"`
	Gender string `json:"gender" validate:"required"`
	MinAge int32  `json:"minAge" validate:"required,min=0,ltfield=MaxAge"`
	MaxAge int32  `json:"maxAge" validate:"required,min=0,gtfield=MinAge"`
}

type RedditCommunityDTO struct {
	Id              int64    `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Rules           []string `json:"rules"`
	Suspended       bool     `json:"suspended"`
	SuspendedReason string   `json:"suspendedReason"`
	ImageId         int64    `json:"imageId"`
}

type RedditCreatePostDTO struct {
	Title       string `json:"title"`
	Text        string `json:"text"`
	CommunityId int64  `json:"communityId"`
}

type ShareTweetDTO struct {
	Text        string `json:"text"`
	CommunityId int64  `json:"communityId"`
}
