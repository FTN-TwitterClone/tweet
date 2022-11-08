package model

import (
	"context"
	"encoding/json"
	"time"
)

// Info from JWT token
type AuthUser struct {
	Username string
	Role     string
	Exp      time.Time
}

type Tweet struct {
	ID        string    `json:"id"`
	Author    string    `json:"author"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
	//Photo  string `json:"photo"` TODO save photo to db
}

func AuthFromCtx(ctx context.Context) (*AuthUser, error) {
	jsonData, err := json.Marshal(ctx.Value("authUser"))
	if err != nil {
		return nil, err
	}

	var authUser AuthUser
	unmarshalErr := json.Unmarshal(jsonData, &authUser)

	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return &authUser, nil
}
