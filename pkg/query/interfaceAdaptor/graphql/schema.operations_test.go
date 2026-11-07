package querygraphql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"cqrs-es-example-go/pkg/command/domain/models"
	"cqrs-es-example-go/pkg/rmu"
	"cqrs-es-example-go/test"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

type graphQLError struct {
	Message    string         `json:"message"`
	Extensions map[string]any `json:"extensions"`
}

type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []graphQLError  `json:"errors"`
}

type groupChatOutput struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	OwnerID   string `json:"ownerId"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type memberOutput struct {
	ID            string `json:"id"`
	GroupChatID   string `json:"groupChatId"`
	UserAccountID string `json:"userAccountId"`
	Role          string `json:"role"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type messageOutput struct {
	ID            string `json:"id"`
	GroupChatID   string `json:"groupChatId"`
	UserAccountID string `json:"userAccountId"`
	Text          string `json:"text"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type readModelFixture struct {
	adminID          string
	memberID         string
	outsiderID       string
	groupChatID      string
	otherGroupChatID string
	messageID        string
	otherMessageID   string
}

func newGraphQLTestServer(t *testing.T) (*handler.Server, *sqlx.DB) {
	t.Helper()
	ctx := context.Background()
	container, err := test.CreateMySQLContainer(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx))
	})
	port, err := container.MappedPort(ctx, "3306")
	require.NoError(t, err)
	db, err := sqlx.Connect("mysql", test.GetDataSourceName(port))
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})
	require.NoError(t, test.MigrateDB(t, nil, db, "../../../../"))
	return handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: NewResolver(db)})), db
}

func seedReadModel(t *testing.T, db *sqlx.DB) readModelFixture {
	t.Helper()
	dao := rmu.NewGroupChatDaoImpl(db)
	now := time.Now().Truncate(time.Second)

	adminID := models.NewUserAccountId()
	memberID := models.NewUserAccountId()
	outsiderID := models.NewUserAccountId()

	groupChatID := models.NewGroupChatId()
	groupChatName := models.NewGroupChatName("group-chat").MustGet()
	require.NoError(t, dao.InsertGroupChat(&groupChatID, &groupChatName, &adminID, now))
	admin := models.NewMember(models.NewMemberId(), adminID, models.AdminRole)
	require.NoError(t, dao.InsertMember(&groupChatID, &admin, now))
	member := models.NewMember(models.NewMemberId(), memberID, models.MemberRole)
	require.NoError(t, dao.InsertMember(&groupChatID, &member, now))
	messageID := models.NewMessageId()
	require.NoError(t, dao.InsertMessage(&messageID, &groupChatID, &memberID, "hello", now))

	otherGroupChatID := models.NewGroupChatId()
	otherGroupChatName := models.NewGroupChatName("other-group-chat").MustGet()
	require.NoError(t, dao.InsertGroupChat(&otherGroupChatID, &otherGroupChatName, &outsiderID, now))
	outsider := models.NewMember(models.NewMemberId(), outsiderID, models.AdminRole)
	require.NoError(t, dao.InsertMember(&otherGroupChatID, &outsider, now))
	otherMessageID := models.NewMessageId()
	require.NoError(t, dao.InsertMessage(&otherMessageID, &otherGroupChatID, &outsiderID, "secret", now))

	return readModelFixture{
		adminID:          adminID.AsString(),
		memberID:         memberID.AsString(),
		outsiderID:       outsiderID.AsString(),
		groupChatID:      groupChatID.AsString(),
		otherGroupChatID: otherGroupChatID.AsString(),
		messageID:        messageID.String(),
		otherMessageID:   otherMessageID.String(),
	}
}

func executeGraphQL(t *testing.T, server *handler.Server, query string, target any) graphQLResponse {
	t.Helper()
	body, err := json.Marshal(map[string]string{"query": query})
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodPost, "/query", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)

	var response graphQLResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	if target != nil && len(response.Errors) == 0 {
		require.NoError(t, json.Unmarshal(response.Data, target))
	}
	return response
}

var naiveDateTimePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?$`)

func TestReadAPIQueries(t *testing.T) {
	server, db := newGraphQLTestServer(t)
	fixture := seedReadModel(t, db)

	t.Run("getGroupChat returns the group chat to a member", func(t *testing.T) {
		var data struct {
			GetGroupChat groupChatOutput `json:"getGroupChat"`
		}
		response := executeGraphQL(t, server, fmt.Sprintf(
			`{ getGroupChat(groupChatId: %q, userAccountId: %q) { id name ownerId createdAt updatedAt } }`,
			fixture.groupChatID, fixture.memberID), &data)
		require.Empty(t, response.Errors)
		require.Equal(t, fixture.groupChatID, data.GetGroupChat.ID)
		require.Equal(t, "group-chat", data.GetGroupChat.Name)
		require.Equal(t, fixture.adminID, data.GetGroupChat.OwnerID)
		require.Regexp(t, naiveDateTimePattern, data.GetGroupChat.CreatedAt)
		require.Regexp(t, naiveDateTimePattern, data.GetGroupChat.UpdatedAt)
	})

	t.Run("getGroupChat rejects a non-member", func(t *testing.T) {
		response := executeGraphQL(t, server, fmt.Sprintf(
			`{ getGroupChat(groupChatId: %q, userAccountId: %q) { id } }`,
			fixture.groupChatID, fixture.outsiderID), nil)
		require.Len(t, response.Errors, 1)
		require.Equal(t, "404", response.Errors[0].Extensions["code"])
	})

	t.Run("getGroupChats only lists chats the account belongs to", func(t *testing.T) {
		var memberData struct {
			GetGroupChats []groupChatOutput `json:"getGroupChats"`
		}
		response := executeGraphQL(t, server, fmt.Sprintf(
			`{ getGroupChats(userAccountId: %q) { id } }`, fixture.memberID), &memberData)
		require.Empty(t, response.Errors)
		require.Len(t, memberData.GetGroupChats, 1)
		require.Equal(t, fixture.groupChatID, memberData.GetGroupChats[0].ID)

		var outsiderData struct {
			GetGroupChats []groupChatOutput `json:"getGroupChats"`
		}
		response = executeGraphQL(t, server, fmt.Sprintf(
			`{ getGroupChats(userAccountId: %q) { id } }`, fixture.outsiderID), &outsiderData)
		require.Empty(t, response.Errors)
		require.Len(t, outsiderData.GetGroupChats, 1)
		require.Equal(t, fixture.otherGroupChatID, outsiderData.GetGroupChats[0].ID)
	})

	t.Run("getMember returns the requested member", func(t *testing.T) {
		var data struct {
			GetMember memberOutput `json:"getMember"`
		}
		response := executeGraphQL(t, server, fmt.Sprintf(
			`{ getMember(groupChatId: %q, userAccountId: %q) { id groupChatId userAccountId role createdAt updatedAt } }`,
			fixture.groupChatID, fixture.memberID), &data)
		require.Empty(t, response.Errors)
		require.Equal(t, fixture.groupChatID, data.GetMember.GroupChatID)
		require.Equal(t, fixture.memberID, data.GetMember.UserAccountID)
		require.Equal(t, "member", data.GetMember.Role)
		require.Regexp(t, naiveDateTimePattern, data.GetMember.CreatedAt)
	})

	t.Run("getMembers returns all members to a member", func(t *testing.T) {
		var data struct {
			GetMembers []memberOutput `json:"getMembers"`
		}
		response := executeGraphQL(t, server, fmt.Sprintf(
			`{ getMembers(groupChatId: %q, userAccountId: %q) { userAccountId role } }`,
			fixture.groupChatID, fixture.memberID), &data)
		require.Empty(t, response.Errors)
		require.Len(t, data.GetMembers, 2)
		accounts := make([]string, 0, len(data.GetMembers))
		for _, member := range data.GetMembers {
			accounts = append(accounts, member.UserAccountID)
		}
		require.ElementsMatch(t, []string{fixture.adminID, fixture.memberID}, accounts)
	})

	t.Run("getMembers hides members from a non-member", func(t *testing.T) {
		var data struct {
			GetMembers []memberOutput `json:"getMembers"`
		}
		response := executeGraphQL(t, server, fmt.Sprintf(
			`{ getMembers(groupChatId: %q, userAccountId: %q) { userAccountId } }`,
			fixture.groupChatID, fixture.outsiderID), &data)
		require.Empty(t, response.Errors)
		require.Empty(t, data.GetMembers)
	})

	t.Run("getMessage returns the message to a member", func(t *testing.T) {
		var data struct {
			GetMessage messageOutput `json:"getMessage"`
		}
		response := executeGraphQL(t, server, fmt.Sprintf(
			`{ getMessage(messageId: %q, userAccountId: %q) { id groupChatId userAccountId text createdAt updatedAt } }`,
			fixture.messageID, fixture.memberID), &data)
		require.Empty(t, response.Errors)
		require.Equal(t, fixture.messageID, data.GetMessage.ID)
		require.Equal(t, fixture.groupChatID, data.GetMessage.GroupChatID)
		require.Equal(t, fixture.memberID, data.GetMessage.UserAccountID)
		require.Equal(t, "hello", data.GetMessage.Text)
		require.Regexp(t, naiveDateTimePattern, data.GetMessage.CreatedAt)
	})

	t.Run("getMessage hides the message from a non-member", func(t *testing.T) {
		response := executeGraphQL(t, server, fmt.Sprintf(
			`{ getMessage(messageId: %q, userAccountId: %q) { id } }`,
			fixture.messageID, fixture.outsiderID), nil)
		require.Len(t, response.Errors, 1)
		require.Equal(t, "404", response.Errors[0].Extensions["code"])
	})

	t.Run("getMessage hides messages of another chat", func(t *testing.T) {
		response := executeGraphQL(t, server, fmt.Sprintf(
			`{ getMessage(messageId: %q, userAccountId: %q) { id } }`,
			fixture.otherMessageID, fixture.memberID), nil)
		require.Len(t, response.Errors, 1)
		require.Equal(t, "404", response.Errors[0].Extensions["code"])
	})

	t.Run("getMessages returns the chat messages to a member", func(t *testing.T) {
		var data struct {
			GetMessages []messageOutput `json:"getMessages"`
		}
		response := executeGraphQL(t, server, fmt.Sprintf(
			`{ getMessages(groupChatId: %q, userAccountId: %q) { id text } }`,
			fixture.groupChatID, fixture.memberID), &data)
		require.Empty(t, response.Errors)
		require.Len(t, data.GetMessages, 1)
		require.Equal(t, fixture.messageID, data.GetMessages[0].ID)
		require.Equal(t, "hello", data.GetMessages[0].Text)
	})

	t.Run("getMessages hides messages from a non-member", func(t *testing.T) {
		var data struct {
			GetMessages []messageOutput `json:"getMessages"`
		}
		response := executeGraphQL(t, server, fmt.Sprintf(
			`{ getMessages(groupChatId: %q, userAccountId: %q) { id } }`,
			fixture.groupChatID, fixture.outsiderID), &data)
		require.Empty(t, response.Errors)
		require.Empty(t, data.GetMessages)
	})
}
