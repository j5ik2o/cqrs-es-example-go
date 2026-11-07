package rmu

import (
	"context"
	"cqrs-es-example-go/pkg/command/domain/events"
	"cqrs-es-example-go/pkg/command/domain/models"
	"cqrs-es-example-go/pkg/command/interfaceAdaptor/repository"
	"cqrs-es-example-go/test"
	"fmt"
	"strconv"
	"testing"
	"time"

	dynamodbevents "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func TestUpdateReadModel_AppliesHeadInsertAndModify(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	fixture := newReadModelFixture(t)
	journal := &fakeJournalReader{events: map[string][]JournalEvent{
		fixture.aid(): journalEvents(fixture.aid(), fixture.create, fixture.addMember, fixture.rename),
	}}
	updater := NewReadModelUpdater(NewGroupChatDaoImpl(db), journal)

	event := dynamodbevents.DynamoDBEvent{Records: []dynamodbevents.DynamoDBEventRecord{
		headRecord(t, "INSERT", fixture.aid(), fixture.create),
		headRecord(t, "MODIFY", fixture.aid(), fixture.addMember),
		headRecord(t, "MODIFY", fixture.aid(), fixture.rename),
	}}
	require.NoError(t, updater.UpdateReadModel(ctx, event))

	require.Equal(t, int64(3), readModelPosition(t, db, fixture.aid()))
	require.Equal(t, "renamed", readModelName(t, db, fixture.aid()))
	require.Equal(t, 1, countRows(t, db, `SELECT COUNT(*) FROM group_chats`))
	require.Equal(t, 2, countRows(t, db, `SELECT COUNT(*) FROM members`))
}

func TestUpdateReadModel_DropsDuplicateAndStaleEvents(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	fixture := newReadModelFixture(t)
	journal := &fakeJournalReader{events: map[string][]JournalEvent{
		fixture.aid(): journalEvents(fixture.aid(), fixture.create, fixture.addMember, fixture.rename),
	}}
	updater := NewReadModelUpdater(NewGroupChatDaoImpl(db), journal)

	first := dynamodbevents.DynamoDBEvent{Records: []dynamodbevents.DynamoDBEventRecord{
		headRecord(t, "INSERT", fixture.aid(), fixture.create),
		headRecord(t, "MODIFY", fixture.aid(), fixture.addMember),
		headRecord(t, "MODIFY", fixture.aid(), fixture.rename),
	}}
	require.NoError(t, updater.UpdateReadModel(ctx, first))

	// Replaying the same records simulates a duplicate delivery.
	require.NoError(t, updater.UpdateReadModel(ctx, first))

	// A stale record that arrives after a newer one must be dropped.
	stale := dynamodbevents.DynamoDBEvent{Records: []dynamodbevents.DynamoDBEventRecord{
		headRecord(t, "MODIFY", fixture.aid(), fixture.create),
	}}
	require.NoError(t, updater.UpdateReadModel(ctx, stale))

	require.Equal(t, int64(3), readModelPosition(t, db, fixture.aid()))
	require.Equal(t, "renamed", readModelName(t, db, fixture.aid()))
	require.Equal(t, 1, countRows(t, db, `SELECT COUNT(*) FROM group_chats`))
	require.Equal(t, 2, countRows(t, db, `SELECT COUNT(*) FROM members`))
}

func TestUpdateReadModel_FillsGapFromJournal(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	fixture := newReadModelFixture(t)
	journal := &fakeJournalReader{events: map[string][]JournalEvent{
		fixture.aid(): journalEvents(fixture.aid(), fixture.create, fixture.addMember, fixture.rename),
	}}
	updater := NewReadModelUpdater(NewGroupChatDaoImpl(db), journal)

	// The head record for seq_nr 3 arrives without 1 and 2.
	require.NoError(t, updater.UpdateReadModel(ctx, dynamodbevents.DynamoDBEvent{
		Records: []dynamodbevents.DynamoDBEventRecord{
			headRecord(t, "MODIFY", fixture.aid(), fixture.rename),
		},
	}))

	require.Equal(t, int64(3), readModelPosition(t, db, fixture.aid()))
	require.Equal(t, "renamed", readModelName(t, db, fixture.aid()))
	require.Equal(t, 2, countRows(t, db, `SELECT COUNT(*) FROM members`))
	require.Equal(t, []journalCall{{aggregateID: fixture.aid(), fromSeqNr: 1, toSeqNr: 2}}, journal.calls)
}

func TestUpdateReadModel_AppliesOutOfOrderHeadEvents(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	fixture := newReadModelFixture(t)
	journal := &fakeJournalReader{events: map[string][]JournalEvent{
		fixture.aid(): journalEvents(fixture.aid(), fixture.create, fixture.addMember, fixture.rename),
	}}
	updater := NewReadModelUpdater(NewGroupChatDaoImpl(db), journal)

	// seq_nr 3 arrives before 1 and 2. The journal supplies the missing events.
	require.NoError(t, updater.UpdateReadModel(ctx, dynamodbevents.DynamoDBEvent{
		Records: []dynamodbevents.DynamoDBEventRecord{
			headRecord(t, "MODIFY", fixture.aid(), fixture.rename),
		},
	}))
	require.NoError(t, updater.UpdateReadModel(ctx, dynamodbevents.DynamoDBEvent{
		Records: []dynamodbevents.DynamoDBEventRecord{
			headRecord(t, "INSERT", fixture.aid(), fixture.create),
		},
	}))

	require.Equal(t, int64(3), readModelPosition(t, db, fixture.aid()))
	require.Equal(t, "renamed", readModelName(t, db, fixture.aid()))
	require.Equal(t, 1, countRows(t, db, `SELECT COUNT(*) FROM group_chats`))
	require.Equal(t, 2, countRows(t, db, `SELECT COUNT(*) FROM members`))
}

func TestUpdateReadModel_RollsBackFailingRecordInBatch(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	fixture := newReadModelFixture(t)
	journal := &fakeJournalReader{events: map[string][]JournalEvent{
		fixture.aid(): journalEvents(fixture.aid(), fixture.create, fixture.duplicateMember),
	}}
	updater := NewReadModelUpdater(NewGroupChatDaoImpl(db), journal)

	// The second record re-adds the administrator, which violates the unique
	// key on (group_chat_id, user_account_id).
	batch := dynamodbevents.DynamoDBEvent{Records: []dynamodbevents.DynamoDBEventRecord{
		headRecord(t, "INSERT", fixture.aid(), fixture.create),
		headRecord(t, "MODIFY", fixture.aid(), fixture.duplicateMember),
	}}
	require.Error(t, updater.UpdateReadModel(ctx, batch))

	// The first record committed; the second rolled back with its position.
	require.Equal(t, int64(1), readModelPosition(t, db, fixture.aid()))
	require.Equal(t, 1, countRows(t, db, `SELECT COUNT(*) FROM group_chats`))
	require.Equal(t, 1, countRows(t, db, `SELECT COUNT(*) FROM members`))

	// Retrying the failed position with a valid event succeeds.
	require.NoError(t, updater.UpdateReadModel(ctx, dynamodbevents.DynamoDBEvent{
		Records: []dynamodbevents.DynamoDBEventRecord{
			headRecord(t, "MODIFY", fixture.aid(), fixture.addMember),
		},
	}))
	require.Equal(t, int64(2), readModelPosition(t, db, fixture.aid()))
	require.Equal(t, 2, countRows(t, db, `SELECT COUNT(*) FROM members`))
}

func TestUpdateReadModel_SkipsConfigForeignAndRemoveRecords(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	fixture := newReadModelFixture(t)
	journal := &fakeJournalReader{}
	updater := NewReadModelUpdater(NewGroupChatDaoImpl(db), journal)

	record := headRecord(t, "INSERT", fixture.aid(), fixture.create)
	configRecord := headRecord(t, "INSERT", configAggregateID, fixture.create)
	foreignRecord := headRecord(t, "INSERT", "UserAccount-01H42K4ABWQ5V2XQEP3A48VE0Z", fixture.create)
	removeRecord := record
	removeRecord.EventName = "REMOVE"
	removeRecord.Change.NewImage = nil

	require.NoError(t, updater.UpdateReadModel(ctx, dynamodbevents.DynamoDBEvent{Records: []dynamodbevents.DynamoDBEventRecord{
		configRecord, foreignRecord, removeRecord,
	}}))

	require.Equal(t, 0, countRows(t, db, `SELECT COUNT(*) FROM group_chats`))
	require.Equal(t, 0, countRows(t, db, `SELECT COUNT(*) FROM read_model_positions`))
}

func TestUpdateReadModel_RestartResumesFromCommittedPosition(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	fixture := newReadModelFixture(t)
	journal := &fakeJournalReader{events: map[string][]JournalEvent{
		fixture.aid(): journalEvents(fixture.aid(), fixture.create, fixture.addMember),
	}}

	first := NewReadModelUpdater(NewGroupChatDaoImpl(db), journal)
	require.NoError(t, first.UpdateReadModel(ctx, dynamodbevents.DynamoDBEvent{
		Records: []dynamodbevents.DynamoDBEventRecord{
			headRecord(t, "INSERT", fixture.aid(), fixture.create),
		},
	}))

	// A new process starts from the beginning of the retained stream.
	restarted := NewReadModelUpdater(NewGroupChatDaoImpl(db), journal)
	require.NoError(t, restarted.UpdateReadModel(ctx, dynamodbevents.DynamoDBEvent{
		Records: []dynamodbevents.DynamoDBEventRecord{
			headRecord(t, "INSERT", fixture.aid(), fixture.create),
			headRecord(t, "MODIFY", fixture.aid(), fixture.addMember),
		},
	}))

	require.Equal(t, int64(2), readModelPosition(t, db, fixture.aid()))
	require.Equal(t, 1, countRows(t, db, `SELECT COUNT(*) FROM group_chats`))
	require.Equal(t, 2, countRows(t, db, `SELECT COUNT(*) FROM members`))
}

func TestDynamoDBJournalReader_ReadEvents(t *testing.T) {
	ctx := context.Background()
	container, err := test.CreateLocalStackContainer(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, container.Terminate(ctx)) })

	client, err := test.CreateDynamoDBClient(t, ctx, container)
	require.NoError(t, err)
	test.WaitForDynamoDB(t, ctx, client)
	require.NoError(t, test.CreateJournalTable(t, ctx, client, "journal"))

	aid := "GroupChat-01H42K4ABWQ5V2XQEP3A48VE0Z"
	for seqNr := int64(1); seqNr <= 3; seqNr++ {
		_, err := client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String("journal"),
			Item: map[string]types.AttributeValue{
				"aid":         &types.AttributeValueMemberS{Value: aid},
				"seq_nr":      &types.AttributeValueMemberN{Value: strconv.FormatInt(seqNr, 10)},
				"occurred_at": &types.AttributeValueMemberN{Value: strconv.FormatInt(time.Unix(1700000000, 0).Add(time.Duration(seqNr)*time.Second).UnixNano(), 10)},
				"manifest":    &types.AttributeValueMemberS{Value: fmt.Sprintf("GroupChatEvent%d@1", seqNr)},
				"payload":     &types.AttributeValueMemberB{Value: []byte(fmt.Sprintf(`{"seq":%d}`, seqNr))},
			},
		})
		require.NoError(t, err)
	}

	reader := NewDynamoDBJournalReader(client, "journal")
	events, err := reader.ReadEvents(ctx, aid, 2, 3)
	require.NoError(t, err)
	require.Len(t, events, 2)
	require.Equal(t, int64(2), events[0].SeqNr)
	require.Equal(t, int64(3), events[1].SeqNr)
	require.Equal(t, "GroupChatEvent2@1", events[0].Manifest)
	require.JSONEq(t, `{"seq":2}`, string(events[0].Payload))
}

// fakeJournalReader returns the seeded journal events for the requested range.
type fakeJournalReader struct {
	events map[string][]JournalEvent
	calls  []journalCall
	errs   []error
}

type journalCall struct {
	aggregateID string
	fromSeqNr   int64
	toSeqNr     int64
}

func (f *fakeJournalReader) ReadEvents(_ context.Context, aggregateID string, fromSeqNr, toSeqNr int64) ([]JournalEvent, error) {
	f.calls = append(f.calls, journalCall{aggregateID: aggregateID, fromSeqNr: fromSeqNr, toSeqNr: toSeqNr})
	if len(f.errs) > 0 {
		err := f.errs[0]
		f.errs = f.errs[1:]
		return nil, err
	}
	var result []JournalEvent
	for _, event := range f.events[aggregateID] {
		if event.SeqNr >= fromSeqNr && event.SeqNr <= toSeqNr {
			result = append(result, event)
		}
	}
	return result, nil
}

type readModelFixture struct {
	groupChatID     models.GroupChatId
	adminID         models.UserAccountId
	memberID        models.UserAccountId
	create          headEvent
	addMember       headEvent
	rename          headEvent
	duplicateMember headEvent
}

func newReadModelFixture(t *testing.T) readModelFixture {
	t.Helper()
	groupChatID := models.NewGroupChatId()
	adminID := models.NewUserAccountId()
	memberID := models.NewUserAccountId()
	name := models.NewGroupChatName("group-chat").MustGet()
	renamed := models.NewGroupChatName("renamed").MustGet()

	created := events.NewGroupChatCreated(name, models.NewMembers(adminID), adminID)
	added := events.NewGroupChatMemberAdded(models.NewMember(models.NewMemberId(), memberID, models.MemberRole), adminID)
	renamedEvent := events.NewGroupChatRenamed(renamed, adminID)
	duplicated := events.NewGroupChatMemberAdded(models.NewMember(models.NewMemberId(), adminID, models.MemberRole), adminID)

	return readModelFixture{
		groupChatID:     groupChatID,
		adminID:         adminID,
		memberID:        memberID,
		create:          newHeadEvent(t, 1, &created),
		addMember:       newHeadEvent(t, 2, &added),
		rename:          newHeadEvent(t, 3, &renamedEvent),
		duplicateMember: newHeadEvent(t, 2, &duplicated),
	}
}

func (f readModelFixture) aid() string {
	return f.groupChatID.AsString()
}

func newHeadEvent(t *testing.T, seqNr int64, event events.GroupChatEvent) headEvent {
	t.Helper()
	payload, err := repository.NewEventSerializer().Serialize(event)
	require.NoError(t, err)
	return headEvent{
		seqNr:      seqNr,
		occurredAt: time.Unix(1700000000, 0).Add(time.Duration(seqNr) * time.Second),
		manifest:   events.Manifest(event.GetTypeName()),
		payload:    payload,
	}
}

func headRecord(t *testing.T, eventName, aggregateID string, headEvents ...headEvent) dynamodbevents.DynamoDBEventRecord {
	t.Helper()
	require.NotEmpty(t, headEvents)
	attributes := make([]dynamodbevents.DynamoDBAttributeValue, 0, len(headEvents))
	for _, headEvent := range headEvents {
		attributes = append(attributes, dynamodbevents.NewMapAttribute(map[string]dynamodbevents.DynamoDBAttributeValue{
			"seq_nr":      dynamodbevents.NewNumberAttribute(strconv.FormatInt(headEvent.seqNr, 10)),
			"occurred_at": dynamodbevents.NewNumberAttribute(strconv.FormatInt(headEvent.occurredAt.UnixNano(), 10)),
			"manifest":    dynamodbevents.NewStringAttribute(headEvent.manifest),
			"payload":     dynamodbevents.NewBinaryAttribute(headEvent.payload),
		}))
	}
	last := headEvents[len(headEvents)-1]
	return dynamodbevents.DynamoDBEventRecord{
		EventID:   fmt.Sprintf("%s-%d", aggregateID, last.seqNr),
		EventName: eventName,
		Change: dynamodbevents.DynamoDBStreamRecord{
			NewImage: map[string]dynamodbevents.DynamoDBAttributeValue{
				"aid":       dynamodbevents.NewStringAttribute(aggregateID),
				"type_name": dynamodbevents.NewStringAttribute(models.GroupChatIdPrefix),
				"seq_nr":    dynamodbevents.NewNumberAttribute(strconv.FormatInt(last.seqNr, 10)),
				"events":    dynamodbevents.NewListAttribute(attributes),
			},
		},
	}
}

func journalEvents(aggregateID string, headEvents ...headEvent) []JournalEvent {
	result := make([]JournalEvent, len(headEvents))
	for i, headEvent := range headEvents {
		result[i] = JournalEvent{
			AggregateID: aggregateID,
			SeqNr:       headEvent.seqNr,
			OccurredAt:  headEvent.occurredAt,
			Manifest:    headEvent.manifest,
			Payload:     headEvent.payload,
		}
	}
	return result
}

func newReadModelTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	ctx := context.Background()
	container, err := test.CreateMySQLContainer(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, container.Terminate(ctx)) })
	port, err := container.MappedPort(ctx, "3306")
	require.NoError(t, err)
	db, err := sqlx.Connect("mysql", test.GetDataSourceName(port))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	require.NoError(t, test.MigrateDB(t, nil, db, "../../"))
	return db
}

func readModelPosition(t *testing.T, db *sqlx.DB, aggregateID string) int64 {
	t.Helper()
	var seqNr int64
	require.NoError(t, db.Get(&seqNr, `SELECT seq_nr FROM read_model_positions WHERE aggregate_id = ?`, aggregateID))
	return seqNr
}

func readModelName(t *testing.T, db *sqlx.DB, aggregateID string) string {
	t.Helper()
	var name string
	require.NoError(t, db.Get(&name, `SELECT name FROM group_chats WHERE id = ?`, aggregateID))
	return name
}

func countRows(t *testing.T, db *sqlx.DB, query string, args ...interface{}) int {
	t.Helper()
	var count int
	require.NoError(t, db.Get(&count, query, args...))
	return count
}
