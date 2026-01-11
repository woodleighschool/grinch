// Package users contains user domain models and behavior.
package users

import "github.com/google/uuid"

// User represents a user synchronised from Entra.
type User struct {
	ID          uuid.UUID `json:"id"           validate:"required"`
	UPN         string    `json:"upn"          validate:"required"`
	DisplayName string    `json:"display_name"`
}
