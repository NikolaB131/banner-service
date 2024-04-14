package entity

import "time"

type User struct {
	ID           string    `db:"id"`
	Username     string    `db:"username"`
	PasswordHash []byte    `db:"password_hash"`
	Role         string    `db:"role"`
	CreatedAt    time.Time `db:"created_at"`
}
