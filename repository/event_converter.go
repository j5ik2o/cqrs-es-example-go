package repository

import (
	"cqrs-es-example-go/domain/events"
	"cqrs-es-example-go/domain/models"
	"fmt"
	esa "github.com/j5ik2o/event-store-adapter-go"
)

func EventConverter(m map[string]interface{}) (esa.Event, error) {
	eventId := m["Id"].(string)
	groupChatId := models.ConvertGroupChatIdFromJSON(m["AggregateId"].(map[string]interface{}))
	groupChatName, err := models.ConvertGroupChatNameFromJSON(m["Name"].(map[string]interface{}))
	if err != nil {
		return nil, err
	}
	members := models.ConvertMembersFromJSON(m["Members"].(map[string]interface{}))
	executorId, err := models.ConvertUserAccountIdFromJSON(m["ExecutorId"].(map[string]interface{}))
	if err != nil {
		return nil, err
	}
	seqNr := uint64(m["SeqNr"].(float64))
	occurredAt := uint64(m["OccurredAt"].(float64))
	switch m["TypeName"].(string) {
	case "GroupChatCreated":
		return events.NewGroupChatCreatedFrom(
			eventId,
			groupChatId,
			groupChatName,
			members,
			seqNr,
			executorId,
			occurredAt,
		), nil
	case "GroupChatDeleted":
		return events.NewGroupChatDeletedFrom(
			eventId,
			groupChatId,
			seqNr,
			executorId,
			occurredAt,
		), nil
	case "GroupChatRenamed":
		name, err := models.NewGroupChatName(m["Name"].(string))
		if err != nil {
			return nil, err
		}
		return events.NewGroupChatRenamedFrom(
			eventId,
			groupChatId,
			name,
			seqNr,
			executorId,
			occurredAt,
		), nil
	case "GroupChatMemberAdded":
		memberObj := m["Member"].(map[string]interface{})
		memberId := models.ConvertMemberIdFromJSON(memberObj["MemberId"].(map[string]interface{}))
		userAccountId, err := models.ConvertUserAccountIdFromJSON(memberObj["UserAccountId"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		role := models.Role(memberObj["Role"].(int))
		member := models.NewMember(memberId, userAccountId, role)
		return events.NewGroupChatMemberAddedFrom(
			eventId,
			groupChatId,
			member,
			seqNr,
			executorId,
			occurredAt,
		), nil
	case "GroupChatMemberRemoved":
		userAccountId, err := models.ConvertUserAccountIdFromJSON(m["UserAccountId"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		return events.NewGroupChatMemberRemovedFrom(
			eventId,
			groupChatId,
			userAccountId,
			seqNr,
			executorId,
			occurredAt,
		), nil
	case "GroupChatMessagePosted":
		message := models.ConvertMessageFromJSON(m["Message"].(map[string]interface{}))
		return events.NewGroupChatMessagePostedFrom(
			eventId,
			groupChatId,
			message,
			seqNr,
			executorId,
			occurredAt,
		), nil
	case "GroupChatMessageDeleted":
		messageId := models.ConvertMessageIdFromJSON(m["MessageId"].(map[string]interface{}))
		return events.NewGroupChatMessageDeletedFrom(
			eventId,
			groupChatId,
			messageId,
			seqNr,
			executorId,
			occurredAt,
		), nil
	default:
		return nil, fmt.Errorf("unknown event type")
	}
}
