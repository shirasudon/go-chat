package action

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestAnyMessage(t *testing.T) {
	var (
		primitives = AnyMessage{
			"number":  float64(2),
			"uint64":  uint64(1),
			"string":  "test",
			"time":    time.Now(),
			"object":  map[string]interface{}{"A": 3, "B": "test2"},
			"array":   []interface{}{1, 2, 3},
			"uint64s": []uint64{4, 5, 6},
		}
	)

	b, err := json.Marshal(primitives)
	if err != nil {
		t.Fatal(err)
	}

	var got AnyMessage
	err = json.Unmarshal(b, &got)
	if err != nil {
		t.Fatal(err)
	}

	for _, tcase := range []struct {
		Key string
		LHS interface{}
		RHS interface{}
	}{
		{"string", got.String("string"), primitives["string"]},
		{"no-string", got.String("no-string"), ""},

		{"uint64", got.UInt64("uint64"), primitives["uint64"]},
		{"no^uint64", got.UInt64("no-uint64"), uint64(0)},

		{"number", got.Number("number"), primitives["number"]},
		{"no-number", got.Number("no-number"), float64(0)},

		{"time", got.Time("time"), primitives["time"]},
		{"no-time", got.Time("no-time"), time.Time{}},

		{"object", got.Object("object"), primitives["object"]},
		{"no-object", got.Time("no-object"), map[string]interface{}{}},

		{"array", got.Array("array"), primitives["array"]},
		{"no-array", got.Array("no-array"), []interface{}{}},

		{"uint64s", got.UInt64s("uint64s"), primitives["uint64s"]},
		{"no-uint64s", got.UInt64s("no-uint64s"), []uint64{}},
	} {
		lhs, rhs := tcase.LHS, tcase.RHS
		if !reflect.DeepEqual(rhs, rhs) {
			t.Errorf("different data for key(%v), expect: %v, got: %v", tcase.Key, lhs, rhs)
		}
	}
}

// TODO add test for ParseXXX

func TestParseAddRoomMember(t *testing.T) {
	const (
		SenderID  = uint64(1)
		RoomID    = uint64(2)
		AddUserID = uint64(3)
	)
	origin := AddRoomMember{
		SenderID:  SenderID,
		RoomID:    RoomID,
		AddUserID: AddUserID,
	}
	bs, err := json.Marshal(origin)
	if err != nil {
		t.Fatal(err)
	}

	var any AnyMessage
	if err := json.Unmarshal(bs, &any); err != nil {
		t.Fatal(err)
	}

	got, err := ParseAddRoomMember(any, ActionAddRoomMember)
	if err != nil {
		t.Fatal(err)
	}
	if got.RoomID != RoomID {
		t.Errorf("different room id")
	}
	if got.AddUserID != AddUserID {
		t.Errorf("different add user id")
	}
	if got.SenderID != SenderID {
		t.Errorf("different sender id")
	}
}
