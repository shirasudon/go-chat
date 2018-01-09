package server

import (
	"github.com/shirasudon/go-chat/chat/action"
)

// NormTimestampNow returns normalized action.Timestamp.
// Normalization is performed by dump it with MarshalText() and then
// re-construct with UnmarshalParam().
func NormTimestampNow() action.Timestamp {
	now := action.TimestampNow()
	bs, _ := now.MarshalText()
	var newNow action.Timestamp
	newNow.UnmarshalParam(string(bs))
	return newNow
}

// MustMarshal returns byte slice which ensure
// the error is nil.
func MustMarshal(bs []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return bs
}
