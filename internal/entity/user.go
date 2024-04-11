package entity

type User struct {
	ID           string
	Username     string
	PasswordHash []byte
	Role         string
}
