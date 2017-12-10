package inmemory

import (
	"context"
	"testing"

	"github.com/shirasudon/go-chat/domain"
)

var (
	userRepository = &UserRepository{}
)

func TestUsersStore(t *testing.T) {
	// case1: success
	newUser := domain.User{Name: "stored-user"}
	id, err := userRepository.Store(context.Background(), newUser)
	if err != nil {
		t.Fatal(err)
	}
	if id == 0 {
		t.Errorf("created id is invalid (0)")
	}

	// case2: duplicated name error
	_, err = userRepository.Store(context.Background(), newUser)
	if err == nil {
		t.Errorf("store duplicated name user, but no error")
	}
}

func TestUsersFind(t *testing.T) {
	// case1: found
	newUser := domain.User{Name: "find-user"}
	id, err := userRepository.Store(context.Background(), newUser)
	if err != nil {
		t.Fatal(err)
	}

	res, err := userRepository.Find(context.Background(), id)
	if err != nil {
		t.Fatalf("can not find user: %v", err)
	}
	if res.Name != newUser.Name {
		t.Errorf("different user name, expect: %v, got: %v", newUser.Name, res.Name)
	}

	// case2: not found
	const NotFoundUserID = 999999
	_, err = userRepository.Find(context.Background(), NotFoundUserID)
	if err == nil {
		t.Errorf("given not found user id, but no error")
	}
}

func TestUsersFindUserRelation(t *testing.T) {
	repo := userRepository

	// case success
	const TestUserID = uint64(2)

	user, err := repo.Find(context.Background(), TestUserID)
	if err != nil {
		t.Fatal(err)
	}

	relation, err := repo.FindUserRelation(context.Background(), TestUserID)
	if err != nil {
		t.Fatal(err)
	}

	if expect, got := TestUserID, relation.UserID; expect != got {
		t.Errorf("different user id, expect: %v, got: %v", expect, got)
	}
	if expect, got := user.Name, relation.UserName; expect != got {
		t.Errorf("different user name, expect: %v, got: %v", expect, got)
	}
	if expect, got := user.FirstName, relation.FirstName; expect != got {
		t.Errorf("different user first name, expect: %v, got: %v", expect, got)
	}
	if expect, got := user.LastName, relation.LastName; expect != got {
		t.Errorf("different user last name, expect: %v, got: %v", expect, got)
	}

	if expect, got := 1, len(relation.Friends); expect != got {
		t.Errorf("different number of friends, expect: %v, got: %v", expect, got)
	}
	if expect, got := 2, len(relation.Rooms); expect != got {
		t.Errorf("different number of rooms, expect: %v, got: %v", expect, got)
	}

	// case fail
	const NotExistUserID = uint64(99)
	if _, err := repo.FindUserRelation(context.Background(), NotExistUserID); err == nil {
		t.Fatalf("query no exist user ID (%v) but no error", NotExistUserID)
	}
}

func TestUsersFindByNameAndPassword(t *testing.T) {
	newUser := domain.User{
		Name:     "new user",
		Password: "password",
	}

	id, err := userRepository.Store(context.Background(), newUser)
	if err != nil {
		t.Fatal(err)
	}

	// case1: found
	res, err := userRepository.FindByNameAndPassword(context.Background(), newUser.Name, newUser.Password)
	if err != nil {
		t.Fatalf("can not find user with name and password: %v", err)
	}

	if res.ID != id {
		t.Errorf("different user id, expect: %v, got: %v", id, res.ID)
	}
	if res.Name != newUser.Name {
		t.Errorf("different user name, expect: %v, got: %v", newUser.Name, res.Name)
	}
	if res.Password != newUser.Password {
		t.Errorf("different user password, expect: %v, got: %v", newUser.Password, res.Password)
	}

	// case2: not found
	if _, err := userRepository.FindByNameAndPassword(context.Background(), "not found", newUser.Password); err == nil {
		t.Errorf("not found user name and password specified but non erorr")
	}
}
