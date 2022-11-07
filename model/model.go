package model

import "time"

// Info from JWT token
type AuthUser struct {
	Username string
	Role     string
	Exp      time.Time
}
