package entity

import "time"

type User struct {
	ID           string
	Username     string
	PasswordHash []byte
	Role         string
	CreatedAt    time.Time
}
