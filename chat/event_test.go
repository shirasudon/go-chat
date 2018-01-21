package chat

import (
	"testing"

	"github.com/shirasudon/go-chat/domain/event"
)

func TestNewExternalEvents(t *testing.T) {
	for _, testcase := range []struct {
		Ev     event.TypeStringer
		Expect string
	}{
		{eventUserLoggedIn{}, "type_user_logged_in"},
		{eventUserLoggedOut{}, "type_user_logged_out"},
	} {
		if got := testcase.Ev.TypeString(); got != testcase.Expect {
			t.Errorf("different type string, expect: %v, got: %v", testcase.Expect, got)
		}
	}
}
