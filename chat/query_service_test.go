package chat

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/shirasudon/go-chat/chat/action"
	"github.com/shirasudon/go-chat/domain"
	"github.com/shirasudon/go-chat/internal/mocks"
)

func TestQueryServiceFindRoomMessagesSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var (
		user = domain.User{ID: 1}
		room = domain.Room{
			ID:          1,
			MemberIDSet: domain.NewUserIDSet(user.ID),
		}
	)

	queryRMsgs := action.QueryRoomMessages{
		RoomID: room.ID,
		Before: time.Now(),
		Limit:  10,
	}

	roomQr := mocks.NewMockRoomQueryer(mockCtrl)
	roomQr.EXPECT().
		Find(gomock.Any(), room.ID).
		Return(room, nil).
		Times(1)

	userQr := mocks.NewMockUserQueryer(mockCtrl)
	userQr.EXPECT().
		Find(gomock.Any(), user.ID).
		Return(user, nil).
		Times(1)

	messages := []domain.Message{
		{
			ID:        1,
			Content:   "hello",
			CreatedAt: queryRMsgs.Before.Add(-100 * time.Millisecond),
		},
	}
	msgQr := mocks.NewMockMessageQueryer(mockCtrl)
	msgQr.EXPECT().
		FindRoomMessagesOrderByLatest(
			gomock.Any(),
			room.ID,
			queryRMsgs.Before,
			queryRMsgs.Limit,
		).
		Return(messages, nil).
		Times(1)

	queryers := &Queryers{
		MessageQueryer: msgQr,
		RoomQueryer:    roomQr,
		UserQueryer:    userQr,
	}

	qservice := NewQueryService(queryers)

	msgs, err := qservice.FindRoomMessages(context.Background(), user.ID, queryRMsgs)
	if err != nil {
		t.Fatal(err)
	}

	if msgs.RoomID != queryRMsgs.RoomID {
		t.Errorf("different room id, expect: %v, got: %v", queryRMsgs.RoomID, msgs.RoomID)
	}
	if got, expect := msgs.Cursor.Current, queryRMsgs.Before; expect != got {
		t.Errorf("different current cursor, expect: %v, got: %v", expect, got)
	}
	if got, expect := msgs.Cursor.Next, messages[0].CreatedAt; expect != got {
		t.Errorf("different next cursor, expect: %v, got: %v", expect, got)
	}
	if got, expect := msgs.Msgs[0].Content, messages[0].Content; expect != got {
		t.Errorf("different messages content, expect: %v, got: %v", expect, got)
	}
}

func TestQueryServiceFindRoomMessagesFail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var (
		user = domain.User{ID: 1}
		room = domain.Room{
			ID:          1,
			MemberIDSet: domain.NewUserIDSet(), // no members
		}
	)

	queryRMsgs := action.QueryRoomMessages{
		RoomID: room.ID,
		Before: time.Now(),
		Limit:  10,
	}

	roomQr := mocks.NewMockRoomQueryer(mockCtrl)
	roomQr.EXPECT().
		Find(gomock.Any(), room.ID).
		Return(room, nil).
		Times(1)

	userQr := mocks.NewMockUserQueryer(mockCtrl)
	userQr.EXPECT().
		Find(gomock.Any(), user.ID).
		Return(user, nil).
		Times(1)

	msgQr := mocks.NewMockMessageQueryer(mockCtrl)
	msgQr.EXPECT().
		FindRoomMessagesOrderByLatest(
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
		).
		Times(0)

	queryers := &Queryers{
		MessageQueryer: msgQr,
		RoomQueryer:    roomQr,
		UserQueryer:    userQr,
	}

	qservice := NewQueryService(queryers)

	_, err := qservice.FindRoomMessages(context.Background(), user.ID, queryRMsgs)
	if err == nil {
		t.Fatal("The room does not have specified member(user) but can be returned messages")
	}
}
