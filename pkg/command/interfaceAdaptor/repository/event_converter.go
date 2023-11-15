package repository

import (
	events2 "cqrs-es-example-go/pkg/command/domain/events"
	models2 "cqrs-es-example-go/pkg/command/domain/models"
	"fmt"
	esa "github.com/j5ik2o/event-store-adapter-go"
)

func EventConverter(m map[string]interface{}) (esa.Event, error) {
	eventId := m["Id"].(string)
	groupChatId := models2.ConvertGroupChatIdFromJSON(m["AggregateId"].(map[string]interface{}))
	groupChatName, err := models2.ConvertGroupChatNameFromJSON(m["Name"].(map[string]interface{})).Get()
	if err != nil {
		return nil, err
	}
	members := models2.ConvertMembersFromJSON(m["Members"].(map[string]interface{}))
	executorId, err := models2.ConvertUserAccountIdFromJSON(m["ExecutorId"].(map[string]interface{})).Get()
	if err != nil {
		return nil, err
	}
	seqNr := uint64(m["SeqNr"].(float64))
	occurredAt := uint64(m["OccurredAt"].(float64))
	switch m["TypeName"].(string) {
	case "GroupChatCreated":
		return events2.NewGroupChatCreatedFrom(
			eventId,
			groupChatId,
			groupChatName,
			members,
			seqNr,
			executorId,
			occurredAt,
		), nil
	case "GroupChatDeleted":
		return events2.NewGroupChatDeletedFrom(
			eventId,
			groupChatId,
			seqNr,
			executorId,
			occurredAt,
		), nil
	case "GroupChatRenamed":
		name, err := models2.NewGroupChatName(m["Name"].(string)).Get()
		if err != nil {
			return nil, err
		}
		return events2.NewGroupChatRenamedFrom(
			eventId,
			groupChatId,
			name,
			seqNr,
			executorId,
			occurredAt,
		), nil
	case "GroupChatMemberAdded":
		memberObj := m["Member"].(map[string]interface{})
		memberId, err := models2.ConvertMemberIdFromJSON(memberObj["MemberId"].(map[string]interface{})).Get()
		if err != nil {
			return nil, err
		}
		userAccountId, err := models2.ConvertUserAccountIdFromJSON(memberObj["UserAccountId"].(map[string]interface{})).Get()
		if err != nil {
			return nil, err
		}
		role := models2.Role(memberObj["Role"].(int))
		member := models2.NewMember(memberId, userAccountId, role)
		return events2.NewGroupChatMemberAddedFrom(
			eventId,
			groupChatId,
			member,
			seqNr,
			executorId,
			occurredAt,
		), nil
	case "GroupChatMemberRemoved":
		userAccountId, err := models2.ConvertUserAccountIdFromJSON(m["UserAccountId"].(map[string]interface{})).Get()
		if err != nil {
			return nil, err
		}
		return events2.NewGroupChatMemberRemovedFrom(
			eventId,
			groupChatId,
			userAccountId,
			seqNr,
			executorId,
			occurredAt,
		), nil
	case "GroupChatMessagePosted":
		message, err := models2.ConvertMessageFromJSON(m["Message"].(map[string]interface{})).Get()
		if err != nil {
			return nil, err
		}
		return events2.NewGroupChatMessagePostedFrom(
			eventId,
			groupChatId,
			message,
			seqNr,
			executorId,
			occurredAt,
		), nil
	case "GroupChatMessageDeleted":
		messageId := models2.ConvertMessageIdFromJSON(m["MessageId"].(map[string]interface{}))
		return events2.NewGroupChatMessageDeletedFrom(
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
