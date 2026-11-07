package repository

import (
	"cqrs-es-example-go/pkg/command/domain"
	"cqrs-es-example-go/pkg/command/domain/events"
	"cqrs-es-example-go/pkg/command/domain/models"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func pointerTo[T any](value T) *T {
	return &value
}

func sampleEvents() []events.GroupChatEvent {
	adminId := models.NewUserAccountId()
	name := models.NewGroupChatName("test").MustGet()
	member := models.NewMember(models.NewMemberId(), models.NewUserAccountId(), models.AdminRole)
	messageId := models.NewMessageId()
	message := models.NewMessage(messageId, "text", adminId).MustGet()
	return []events.GroupChatEvent{
		pointerTo(events.NewGroupChatCreated(name, models.NewMembers(adminId), adminId)),
		pointerTo(events.NewGroupChatDeleted(adminId)),
		pointerTo(events.NewGroupChatRenamed(name, adminId)),
		pointerTo(events.NewGroupChatMemberAdded(member, adminId)),
		pointerTo(events.NewGroupChatMemberRemoved(adminId, adminId)),
		pointerTo(events.NewGroupChatMessagePosted(message, adminId)),
		pointerTo(events.NewGroupChatMessageEdited(message, adminId)),
		pointerTo(events.NewGroupChatMessageDeleted(messageId, adminId)),
	}
}

func TestEventManifest(t *testing.T) {
	assert.Equal(t, "GroupChatCreated@1", events.Manifest(events.EventTypeGroupChatCreated))
	assert.Equal(t, "GroupChatMessageEdited@1", events.Manifest(events.EventTypeGroupChatMessageEdited))
}

func TestEventSerializer_PayloadHasNoEnvelopeMetadata(t *testing.T) {
	serializer := NewEventSerializer()
	for _, event := range sampleEvents() {
		data, err := serializer.Serialize(event)
		require.NoError(t, err)

		var payload map[string]interface{}
		require.NoError(t, json.Unmarshal(data, &payload))
		assert.Equal(t, event.GetTypeName(), payload["type"])
		for _, key := range []string{"type_name", "aggregate_id", "seq_nr", "occurred_at"} {
			assert.NotContains(t, payload, key)
		}

		decoded, err := serializer.Deserialize(data)
		require.NoError(t, err)
		assert.Equal(t, event.GetTypeName(), decoded.GetTypeName())
		assert.Equal(t, event.GetId(), decoded.GetId())
	}
}

func TestEventSerializer_IgnoresUnknownFields(t *testing.T) {
	data := []byte(`{"type":"GroupChatDeleted","id":"event-1","executor_id":{"value":"user-1"},"unknown":{"nested":1}}`)
	event, err := NewEventSerializer().Deserialize(data)
	require.NoError(t, err)
	deleted, ok := event.(*events.GroupChatDeleted)
	require.True(t, ok)
	assert.Equal(t, "event-1", deleted.GetId())
	assert.Equal(t, "user-1", deleted.GetExecutorId().GetValue())
}

func TestEventSerializer_RejectsMissingRequiredField(t *testing.T) {
	data := []byte(`{"type":"GroupChatRenamed","id":"event-1","executor_id":{"value":"user-1"}}`)
	_, err := NewEventSerializer().Deserialize(data)
	require.Error(t, err)

	_, err = NewEventSerializer().Deserialize([]byte(`{"id":"event-1","executor_id":{"value":"user-1"}}`))
	require.Error(t, err)
}

func TestSnapshotSerializer_PayloadHasNoVersion(t *testing.T) {
	adminId := models.NewUserAccountId()
	name := models.NewGroupChatName("test").MustGet()
	groupChat, _ := domain.NewGroupChat(name, adminId)

	serializer := NewSnapshotSerializer()
	data, err := serializer.Serialize(groupChat)
	require.NoError(t, err)

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &payload))
	for _, key := range []string{"version", "manifest"} {
		assert.NotContains(t, payload, key)
	}
	assert.Contains(t, payload, "seq_nr")

	decoded, err := serializer.Deserialize(data)
	require.NoError(t, err)
	assert.True(t, groupChat.Equals(&decoded))
}

func TestSnapshotSerializer_RejectsMissingRequiredField(t *testing.T) {
	data := []byte(`{"id":{"value":"group-chat-1"},"name":{"value":"test"},"members":{"values":[]},"messages":{"values":[]},"deleted":false}`)
	_, err := NewSnapshotSerializer().Deserialize(data)
	require.Error(t, err)
}
