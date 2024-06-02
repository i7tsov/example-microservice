package model

// This file contains HTTP response structures that correspond to JSON
// responses of the service.

type ReasonResponse struct {
	Reason string `json:"reason"`
}

type InternalError struct {
	Message string `json:"message"`
}

type IDResponse struct {
	ID string `json:"id"`
}

type ListUsersResponse struct {
	Users []User `json:"users,omitempty"`
}
