package events

import (
	"cqrs-es-example-go/pkg/command/domain/models"
	"fmt"
)

const (
	// SchemaVersion is the payload schema version carried in the envelope manifest.
	SchemaVersion = 1

	EventTypeGroupChatCreated        = "GroupChatCreated"
	EventTypeGroupChatDeleted        = "GroupChatDeleted"
	EventTypeGroupChatRenamed        = "GroupChatRenamed"
	EventTypeGroupChatMemberAdded    = "GroupChatMemberAdded"
	EventTypeGroupChatMemberRemoved  = "GroupChatMemberRemoved"
	EventTypeGroupChatMessagePosted  = "GroupChatMessagePosted"
	EventTypeGroupChatMessageEdited  = "GroupChatMessageEdited"
	EventTypeGroupChatMessageDeleted = "GroupChatMessageDeleted"
)

// GroupChatEvent is the payload of a domain event. Envelope metadata such as
// the aggregate id, sequence number and occurrence time is not part of the payload.
type GroupChatEvent interface {
	GetId() string
	GetTypeName() string
	GetExecutorId() *models.UserAccountId
	ToJSON() map[string]interface{}
}

// Manifest returns the envelope manifest for the event type.
func Manifest(eventType string) string {
	return fmt.Sprintf("%s@%d", eventType, SchemaVersion)
}
