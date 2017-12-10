package chat

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/shirasudon/go-chat/chat/action"
	"github.com/shirasudon/go-chat/chat/queried"
	"github.com/shirasudon/go-chat/domain"
	"github.com/shirasudon/go-chat/internal/mocks"
)

func TestQueryServiceFindUserByNameAndPassword(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var (
		user = &queried.AuthUser{
			ID:       1,
			Name:     "user",
			Password: "password",
		}
	)

	userQr := mocks.NewMockUserQueryer(mockCtrl)
	userQr.EXPECT().
		FindByNameAndPassword(gomock.Any(), user.Name, user.Password).
		Return(user, nil).
		Times(1)

	qservice := NewQueryService(&Queryers{
		UserQueryer: userQr,
	})

	res, err := qservice.FindUserByNameAndPassword(context.Background(), user.Name, user.Password)
	if err != nil {
		t.Fatal(err)
	}
	if user != res {
		t.Errorf("different queried result, expect: %v, got: %v", user, res)
	}
}

func TestQueryServiceFindRoomInfo(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var (
		roomInfo = &queried.RoomInfo{
			RoomName:  "room_name",
			RoomID:    1,
			CreatorID: 2,
			Members: []queried.UserProfile{
				{UserName: "user", UserID: 2},
			},
			MembersSize: 1,
		}
	)

	roomQr := mocks.NewMockRoomQueryer(mockCtrl)
	roomQr.EXPECT().
		FindRoomInfo(gomock.Any(), roomInfo.CreatorID, roomInfo.RoomID).
		Return(roomInfo, nil).
		Times(1)

	queryers := &Queryers{
		RoomQueryer: roomQr,
	}

	qservice := NewQueryService(queryers)

	_, err := qservice.FindRoomInfo(context.Background(), roomInfo.CreatorID, roomInfo.RoomID)
	if err != nil {
		t.Fatal(err)
	}
}

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

func TestQueryServiceFindUnreadRoomMessages(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var (
		user = domain.User{ID: 1}
		room = domain.Room{
			ID:          1,
			MemberIDSet: domain.NewUserIDSet(), // no members
		}
	)

	roomQr := mocks.NewMockRoomQueryer(mockCtrl)
	roomQr.EXPECT().
		Find(gomock.Any(), gomock.Any()).
		Return(domain.Room{}, nil).
		Times(1)

	userQr := mocks.NewMockUserQueryer(mockCtrl)
	userQr.EXPECT().
		Find(gomock.Any(), gomock.Any()).
		Return(domain.User{}, nil).
		Times(1)

	msgQr := mocks.NewMockMessageQueryer(mockCtrl)

	query := action.QueryUnreadRoomMessages{
		RoomID: room.ID,
		Limit:  10,
	}
	msgQr.EXPECT().
		FindUnreadRoomMessages(gomock.Any(), user.ID, query.RoomID, query.Limit).
		Return(&queried.UnreadRoomMessages{
			RoomID: room.ID,
			Msgs: []queried.Message{
				{Content: "hello0"},
				{Content: "hello1"},
			},
			MsgsSize: 2,
		}, nil).
		Times(1)

	queryers := &Queryers{
		MessageQueryer: msgQr,
		RoomQueryer:    roomQr,
		UserQueryer:    userQr,
	}

	qservice := NewQueryService(queryers)

	got, err := qservice.FindUnreadRoomMessages(context.Background(), user.ID, query)
	if err != nil {
		t.Fatalf("can not get RoomUnreadMessages: %v", err)
	}
	if got.MsgsSize != 2 {
		t.Errorf("different message size, expect: %v, got: %v", 2, got.MsgsSize)
	}
}
