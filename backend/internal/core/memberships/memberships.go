package memberships

import "github.com/google/uuid"

// Membership represents the association between a user and a group.
type Membership struct {
	ID      uuid.UUID `json:"id"`
	GroupID uuid.UUID `json:"group_id"`
	UserID  uuid.UUID `json:"user_id"`
}
