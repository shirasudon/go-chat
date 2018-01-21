package domain

import (
	"testing"
	"time"

	"github.com/shirasudon/go-chat/domain/event"
)

type ConnImpl struct {
	userID     uint64
	receviedEv []event.Event
}

func (c ConnImpl) UserID() uint64 {
	return c.userID
}

func (c ConnImpl) Close() error {
	return nil
}

func (c *ConnImpl) Send(ev event.Event) {
	if c.receviedEv == nil {
		c.receviedEv = make([]event.Event, 0, 4)
	}
	c.receviedEv = append(c.receviedEv, ev)
}

func TestNewActiveClientSuccess(t *testing.T) {
	repo := NewActiveClientRepository(10)
	user := User{ID: 1}
	conn := &ConnImpl{userID: user.ID}

	ac, created, err := NewActiveClient(repo, conn, user)
	if err != nil {
		t.Fatal(err)
	}

	ac2, err := repo.Find(user.ID)
	if err != nil {
		t.Fatalf("ActiveClient is created but not in the repository")
	}
	if ac != ac2 {
		t.Fatalf("created ActiveClient is not same as that in the repository")
	}

	if ac.userID != user.ID {
		t.Errorf("created ActiveClient has different user id, expect: %d, got: %d", ac.userID, user.ID)
	}

	// check event fields
	if created.UserID != user.ID {
		t.Errorf("ActiveClientActivated has different user id, expect: %d, got: %d", created.UserID, user.ID)
	}
	if created.UserName != user.Name {
		t.Errorf("ActiveClientActivated has different user name, expect: %s, got: %s", created.UserName, user.Name)
	}
	if got := created.Timestamp(); got == (time.Time{}) {
		t.Error("ActiveClientActivated has no timestamp")
	}
}

func TestNewActiveClientFail(t *testing.T) {
	repo := NewActiveClientRepository(10)
	user := User{ID: 0}
	conn := &ConnImpl{userID: user.ID}

	_, _, err := NewActiveClient(repo, conn, user)
	if err == nil {
		t.Fatal("given non-exist user, but no-error")
	}
	_, _, err = NewActiveClient(repo, conn, User{ID: user.ID + 1})
	if err == nil {
		t.Fatal("different user IDs for conn's and given user, but no-error")
	}
}

func TestActiveClientAddConn(t *testing.T) {
	repo := NewActiveClientRepository(10)
	user := User{ID: 1}
	conn := &ConnImpl{userID: user.ID}

	ac, _, _ := NewActiveClient(repo, conn, user)

	// success case
	n, err := ac.AddConn(&ConnImpl{userID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("added conn but the number of conns is different, expect: %d, got: %d", 2, n)
	}

	// fail case
	n, err = ac.AddConn(&ConnImpl{userID: user.ID + 1})
	if err == nil {
		t.Fatalf("different userIDs for ActiveClient and added conn, but AddConn is succeed")
	}
	if n != 2 {
		t.Errorf("add conn is fail, but increase the number of conn, expect: %d, got: %d", 2, n)
	}
}

func TestActiveClientRemoveConn(t *testing.T) {
	repo := NewActiveClientRepository(10)
	user := User{ID: 1}
	conn := &ConnImpl{userID: user.ID}
	conn2 := &ConnImpl{userID: user.ID}

	ac, _, _ := NewActiveClient(repo, conn, user)
	ac.AddConn(conn2)

	for i, tcase := range []struct {
		*ConnImpl
		expectN int
	}{
		{conn, 1},
		{conn2, 0},
	} {
		t.Logf("currently inspecting for conn[%d]", i)
		n, err := ac.RemoveConn(tcase.ConnImpl)
		if err != nil {
			t.Fatal("can not remove already added conn.")
		}
		if n != tcase.expectN {
			t.Fatalf("removed conn but not decrease the number of conns, expect: %d, got: %d", 1, tcase.expectN)
		}
	}

	// currently, conns are zero.

	_, err := ac.RemoveConn(conn)
	if err == nil {
		t.Error("remove not added conn but succeed")
	}
}

func TestActiveClientSend(t *testing.T) {
	repo := NewActiveClientRepository(10)
	user := User{ID: 1}
	conn := &ConnImpl{userID: user.ID}

	ac, _, _ := NewActiveClient(repo, conn, user)

	conn2 := &ConnImpl{userID: user.ID}
	ac.AddConn(conn2)

	ac.Send(event.MessageCreated{})
	for i, c := range []*ConnImpl{conn, conn2} {
		t.Logf("currently inspecting for conn[%d]", i)

		if got := len(c.receviedEv); got != 1 {
			t.Fatalf("send one event to the conns but the number of event the conns are received is different, expect: %d, got: %d", 1, got)
		}

		if got, ok := c.receviedEv[0].(event.MessageCreated); !ok {
			t.Errorf("different event type for the conns received and ActiveClien sent, expect: %#v, got: %#v", event.MessageCreated{}, got)
		}
	}
}

func TestActiveClientHasConn(t *testing.T) {
	repo := NewActiveClientRepository(10)
	user := User{ID: 1}
	conn := &ConnImpl{userID: user.ID}

	ac, _, _ := NewActiveClient(repo, conn, user)

	if !ac.HasConn(conn) {
		t.Errorf("ActiveClient has given conn, but HasConn return false")
	}
	if ac.HasConn(&ConnImpl{}) {
		t.Errorf("ActiveClient does not have given conn, but HasConn return ture")
	}
}

func TestActiveClientDelete(t *testing.T) {
	repo := NewActiveClientRepository(10)
	user := User{ID: 1}
	conn := &ConnImpl{userID: user.ID}

	ac, _, _ := NewActiveClient(repo, conn, user)

	_, err := ac.Delete(repo)
	if err == nil {
		t.Fatalf("ActiveClient still have connection but deleted it without any error")
	}

	// any conns are removed
	if _, err := ac.RemoveConn(conn); err != nil {
		t.Fatal(err)
	}

	// then delete ActiveClient
	ev, err := ac.Delete(repo)
	if err != nil {
		t.Fatal(err)
	}
	if ev.UserID != user.ID {
		t.Errorf("ActiveClient deleted event has different user id, expect: %d, got: %d", user.ID, ev.UserID)
	}
	if got := ev.Timestamp(); got == (time.Time{}) {
		t.Error("ActiveClientInactivated has no timestamp")
	}

	invalidAC := &ActiveClient{}

	_, err = invalidAC.Delete(repo)
	if err == nil {
		t.Error("delete ActiveClient which not exist in the repository, but no error")
	}
}

func TestActiveClientForceDelete(t *testing.T) {
	t.Parallel()

	repo := NewActiveClientRepository(10)
	user := User{ID: 1}
	conn := &ConnImpl{userID: user.ID}

	ac, _, _ := NewActiveClient(repo, conn, user)

	ev, err := ac.ForceDelete(repo)
	if err != nil {
		t.Fatal(err)
	}
	if ev.UserID != user.ID {
		t.Errorf("ActiveClient deleted event has different user id, expect: %d, got: %d", user.ID, ev.UserID)
	}
	if got := ev.Timestamp(); got == (time.Time{}) {
		t.Error("ActiveClientInactivated has no timestamp")
	}

	invalidAC := &ActiveClient{}

	_, err = invalidAC.ForceDelete(repo)
	if err == nil {
		t.Error("delete ActiveClient which not exist in the repository, but no error")
	}
}

func TestACRepoExistByConn(t *testing.T) {
	repo := NewActiveClientRepository(10)
	user := User{ID: 1}
	conn := &ConnImpl{userID: user.ID}

	_, _, _ = NewActiveClient(repo, conn, user)

	if !repo.ExistByConn(conn) {
		t.Error("the repository has ActiveClient with conn, but ExistByConn returns false")
	}
	if repo.ExistByConn(&ConnImpl{}) {
		t.Error("the repository does not have ActiveClient with given conn, but ExistByConn returns true")
	}
}

func TestACRepoFindAllByUserIDs(t *testing.T) {
	repo := NewActiveClientRepository(10)
	user := User{ID: 1}
	conn := &ConnImpl{userID: user.ID}
	user2 := User{ID: 2}
	conn2 := &ConnImpl{userID: user2.ID}

	_, _, _ = NewActiveClient(repo, conn, user)
	_, _, _ = NewActiveClient(repo, conn2, user2)

	acs, err := repo.FindAllByUserIDs([]uint64{user.ID, user2.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(acs) != 2 {
		t.Fatalf("given exist two user ids, but got different number of ActiveClients, expect: %d, got: %d", 2, len(acs))
	}

	acs2, err := repo.FindAllByUserIDs([]uint64{user.ID, user2.ID + 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(acs2) != 1 {
		t.Fatalf("given user ids include one exist user's, but got different number of ActiveClients, expect: %d, got: %d", 1, len(acs2))
	}

	_, err = repo.FindAllByUserIDs([]uint64{user.ID + 10, user2.ID + 10})
	if err == nil {
		t.Fatal("given user ids exclude exist user's, but return no error")
	}
}
