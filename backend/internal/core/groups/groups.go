package groups

import "github.com/google/uuid"

// Group represents a directory group synchronised from Entra.
type Group struct {
	ID          uuid.UUID `json:"id"           validate:"required"`
	DisplayName string    `json:"display_name" validate:"required"`
	Description string    `json:"description"`
	MemberCount int32     `json:"member_count"`
}
