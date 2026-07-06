package queries

import "errors"

type Admin struct {
	ID           int
	Email        string
	PasswordHash string
}

var ErrAdminNotFound = errors.New("admin not found")