// Package model contains data structures used throughout the service.
package model

// User represents service user.
type User struct {
	ID    string `json:"id" bun:",pk,type:uuid,default:gen_random_uuid()"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
