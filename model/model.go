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
	ID        gocql.UUID `json:"id"`
	Username  string     `json:"username"`
	Text      string     `json:"text"`
	Timestamp time.Time  `json:"timestamp"`
	//Photo  string `json:"photo"` TODO save photo to db
}

type Like struct {
	Username string     `json:"username"`
	TweetId  gocql.UUID `json:"tweet_id" validate:"required"`
}
