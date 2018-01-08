package chat

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
