package action

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestTimestampMarshalJSON(t *testing.T) {
	var tm = TimestampNow()
	bs, err := json.Marshal(tm)
	if err != nil {
		t.Fatal(err)
	}

	strip := strings.Trim(string(bs), "\"")
	parsed, err := time.Parse(SupportedTimeFormat, strip)
	if err != nil {
		t.Fatalf("can not parse with SupportedTimeFormat, expect format: %v, got: %v", SupportedTimeFormat, strip)
	}
	if !parsed.Equal(tm.Time()) {
		t.Errorf("different time after parsing from json string")
	}

	var newTm Timestamp
	if err := json.Unmarshal(bs, &newTm); err != nil {
		t.Fatalf("can not Unmarshal: %v", err)
	}
	if !newTm.Time().Equal(tm.Time()) {
		t.Errorf("different time after Unmarshaling from json")
	}
}

func TestTimestampUnmarshalText(t *testing.T) {
	var testStrings = make([]string, 0, 2)
	var tm = Timestamp(time.Now())

	// json text
	bs, err := json.Marshal(tm)
	if err != nil {
		t.Fatal(err)
	}
	testStrings = append(testStrings, strings.Trim(string(bs), "\""))

	// format text
	bs, err = tm.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	testStrings = append(testStrings, string(bs))

	for _, text := range testStrings {
		var newTm Timestamp
		if err := newTm.UnmarshalParam(text); err != nil {
			t.Fatalf("can not UnmarshalParam: %v", err)
		}
		if !newTm.Time().Equal(tm.Time()) {
			t.Errorf("different time after UnmarshalParam, expect: %v, got: %v", newTm.Time(), tm.Time())
		}
	}
	t.Logf("src text are: %#v", testStrings)
}
