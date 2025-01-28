package storage

import "errors"

var (
	ErrURLNotFound = errors.New("url not found")
	ErrURLExists = errors.New("url exists")

	ErrUserExists = errors.New("user exists")

	ErrInvalidUsername = errors.New("invalid username")
	ErrUsernamelExists = errors.New("username email")
	ErrUserNotFound = errors.New("user not found")

	ErrInvalidEmail = errors.New("invalid email format")
	ErrEmailExists = errors.New("exists email")
	
	ErrInvalidPassword = errors.New("password does not meet security requirements")
)