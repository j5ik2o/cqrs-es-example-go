package api

//import (
//	"cqrs-es-example-go/pkg/command/interfaceAdaptor/repository"
//	"github.com/gin-gonic/gin"
//	event_store_adapter_go "github.com/j5ik2o/event-store-adapter-go"
//	"testing"
//)
//
//func TestGroupChat_AddMember(t *testing.T) {
//	groupChatRepository := repository.NewGroupChatRepository(event_store_adapter_go.NewEventStoreOnMemory())
//	groupChatController := NewGroupChatController(groupChatRepository)
//
//	engine := gin.Default()
//	groupChat := engine.Group("/group-chats")
//	{
//		groupChat.POST("/add-member", AddMember)
//	}
//}
