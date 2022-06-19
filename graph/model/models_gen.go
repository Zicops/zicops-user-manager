// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"github.com/99designs/gqlgen/graphql"
)

type User struct {
	ID         *string `json:"id"`
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	Status     string  `json:"status"`
	Role       string  `json:"role"`
	IsVerified bool    `json:"is_verified"`
	IsActive   bool    `json:"is_active"`
	Gender     string  `json:"gender"`
	CreatedBy  string  `json:"created_by"`
	UpdatedBy  string  `json:"updated_by"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
	Email      string  `json:"email"`
	Phone      string  `json:"phone"`
	PhotoURL   *string `json:"photo_url"`
}

type UserInput struct {
	ID         *string         `json:"id"`
	FirstName  string          `json:"first_name"`
	LastName   string          `json:"last_name"`
	Status     string          `json:"status"`
	Role       string          `json:"role"`
	IsVerified bool            `json:"is_verified"`
	IsActive   bool            `json:"is_active"`
	Gender     string          `json:"gender"`
	CreatedBy  string          `json:"created_by"`
	UpdatedBy  string          `json:"updated_by"`
	Email      string          `json:"email"`
	Phone      string          `json:"phone"`
	Photo      *graphql.Upload `json:"Photo"`
	PhotoURL   *string         `json:"photo_url"`
}

type UserLoginContext struct {
	User        *User  `json:"user"`
	AccessToken string `json:"access_token"`
}
