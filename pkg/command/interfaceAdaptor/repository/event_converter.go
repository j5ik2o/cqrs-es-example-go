package repository

import (
	"cqrs-es-example-go/pkg/command/domain/events"
	"cqrs-es-example-go/pkg/command/domain/models"
	"fmt"
)

// EventConverter restores a domain event from its JSON payload. Unknown fields
// are ignored; a missing required field is reported as an error.
func EventConverter(m map[string]interface{}) (events.GroupChatEvent, error) {
	typeName, err := jsonString(m, "type")
	if err != nil {
		return nil, err
	}
	eventId, err := jsonString(m, "id")
	if err != nil {
		return nil, err
	}
	executorIdValue, err := jsonValueString(m, "executor_id")
	if err != nil {
		return nil, err
	}
	executorId, err := models.NewUserAccountIdFromString(executorIdValue).Get()
	if err != nil {
		return nil, err
	}

	switch typeName {
	case events.EventTypeGroupChatCreated:
		name, err := convertGroupChatName(m)
		if err != nil {
			return nil, err
		}
		membersObject, err := jsonObject(m, "members")
		if err != nil {
			return nil, err
		}
		members, err := convertMembers(membersObject)
		if err != nil {
			return nil, err
		}
		event := events.NewGroupChatCreatedFrom(eventId, name, members, executorId)
		return &event, nil
	case events.EventTypeGroupChatDeleted:
		event := events.NewGroupChatDeletedFrom(eventId, executorId)
		return &event, nil
	case events.EventTypeGroupChatRenamed:
		name, err := convertGroupChatName(m)
		if err != nil {
			return nil, err
		}
		event := events.NewGroupChatRenamedFrom(eventId, name, executorId)
		return &event, nil
	case events.EventTypeGroupChatMemberAdded:
		memberObject, err := jsonObject(m, "member")
		if err != nil {
			return nil, err
		}
		member, err := convertMember(memberObject)
		if err != nil {
			return nil, err
		}
		event := events.NewGroupChatMemberAddedFrom(eventId, member, executorId)
		return &event, nil
	case events.EventTypeGroupChatMemberRemoved:
		userAccountIdValue, err := jsonValueString(m, "user_account_id")
		if err != nil {
			return nil, err
		}
		userAccountId, err := models.NewUserAccountIdFromString(userAccountIdValue).Get()
		if err != nil {
			return nil, err
		}
		event := events.NewGroupChatMemberRemovedFrom(eventId, userAccountId, executorId)
		return &event, nil
	case events.EventTypeGroupChatMessagePosted:
		messageObject, err := jsonObject(m, "message")
		if err != nil {
			return nil, err
		}
		message, err := convertMessage(messageObject)
		if err != nil {
			return nil, err
		}
		event := events.NewGroupChatMessagePostedFrom(eventId, message, executorId)
		return &event, nil
	case events.EventTypeGroupChatMessageEdited:
		messageObject, err := jsonObject(m, "message")
		if err != nil {
			return nil, err
		}
		message, err := convertMessage(messageObject)
		if err != nil {
			return nil, err
		}
		event := events.NewGroupChatMessageEditedFrom(eventId, message, executorId)
		return &event, nil
	case events.EventTypeGroupChatMessageDeleted:
		messageIdValue, err := jsonValueString(m, "message_id")
		if err != nil {
			return nil, err
		}
		messageId, err := models.NewMessageIdFromString(messageIdValue).Get()
		if err != nil {
			return nil, err
		}
		event := events.NewGroupChatMessageDeletedFrom(eventId, messageId, executorId)
		return &event, nil
	default:
		return nil, fmt.Errorf("unknown event type: %s", typeName)
	}
}

func convertGroupChatName(m map[string]interface{}) (models.GroupChatName, error) {
	value, err := jsonValueString(m, "name")
	if err != nil {
		return models.GroupChatName{}, err
	}
	return models.NewGroupChatName(value).Get()
}

func convertMember(object map[string]interface{}) (models.Member, error) {
	memberIdValue, err := jsonValueString(object, "id")
	if err != nil {
		return models.Member{}, err
	}
	memberId, err := models.NewMemberIdFromString(memberIdValue).Get()
	if err != nil {
		return models.Member{}, err
	}
	userAccountIdValue, err := jsonValueString(object, "user_account_id")
	if err != nil {
		return models.Member{}, err
	}
	userAccountId, err := models.NewUserAccountIdFromString(userAccountIdValue).Get()
	if err != nil {
		return models.Member{}, err
	}
	roleValue, err := jsonInteger(object, "role")
	if err != nil {
		return models.Member{}, err
	}
	if roleValue != int64(models.MemberRole) && roleValue != int64(models.AdminRole) {
		return models.Member{}, fmt.Errorf("unknown member role: %d", roleValue)
	}
	return models.NewMember(memberId, userAccountId, models.Role(roleValue)), nil
}

func convertMembers(object map[string]interface{}) (models.Members, error) {
	values, err := jsonArray(object, "values")
	if err != nil {
		return models.Members{}, err
	}
	members := make([]models.Member, len(values))
	for i, value := range values {
		memberObject, ok := value.(map[string]interface{})
		if !ok {
			return models.Members{}, fmt.Errorf("member at index %d is not an object", i)
		}
		member, err := convertMember(memberObject)
		if err != nil {
			return models.Members{}, err
		}
		members[i] = member
	}
	return models.NewMembersFrom(members), nil
}

func convertMessage(object map[string]interface{}) (models.Message, error) {
	messageIdValue, err := jsonValueString(object, "id")
	if err != nil {
		return models.Message{}, err
	}
	messageId, err := models.NewMessageIdFromString(messageIdValue).Get()
	if err != nil {
		return models.Message{}, err
	}
	text, err := jsonString(object, "text")
	if err != nil {
		return models.Message{}, err
	}
	senderIdValue, err := jsonValueString(object, "sender_id")
	if err != nil {
		return models.Message{}, err
	}
	senderId, err := models.NewUserAccountIdFromString(senderIdValue).Get()
	if err != nil {
		return models.Message{}, err
	}
	return models.NewMessage(messageId, text, senderId).Get()
}

func convertMessages(object map[string]interface{}) (models.Messages, error) {
	values, err := jsonArray(object, "values")
	if err != nil {
		return models.Messages{}, err
	}
	messages := make(map[string]models.Message, len(values))
	for i, value := range values {
		messageObject, ok := value.(map[string]interface{})
		if !ok {
			return models.Messages{}, fmt.Errorf("message at index %d is not an object", i)
		}
		message, err := convertMessage(messageObject)
		if err != nil {
			return models.Messages{}, err
		}
		messages[message.GetId().GetValue()] = message
	}
	return models.NewMessagesFromMap(messages), nil
}
