package models

import (
	"errors"
	"time"
)

// Snippet holds data from one record of the "snippets" table of the snippetbox database
type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

var (
	// Errors relating to the user table (logins)
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	ErrDuplicateEmail     = errors.New("models: duplicate email")
)

// User holds data from one record of the "users" table of the snippetbox database
type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
}
