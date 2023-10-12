package dto

import (
	"github.com/drmaples/starter-app/app/repo"
)

// User represents a user in db
type User struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// Model converts a dto object to model object
func (u *User) Model() repo.User {
	return repo.User{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
	}
}

// CreateUser is for dto for creating new user
type CreateUser struct {
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
}

// Model converts a dto object to model object
func (u *CreateUser) Model() repo.User {
	return repo.User{
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
	}
}