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
	var tm = Timestamp(time.Now())
	bs, err := json.Marshal(tm)
	if err != nil {
		t.Fatal(err)
	}

	strip := strings.Trim(string(bs), "\"")
	var newTm Timestamp
	if err := newTm.UnmarshalParam(strip); err != nil {
		t.Fatalf("can not UnmarshalParam: %v", err)
	}
	if !newTm.Time().Equal(tm.Time()) {
		t.Errorf("different time after UnmarshalParam, expect: %v, got: %v", newTm.Time(), tm.Time())
	}
}
