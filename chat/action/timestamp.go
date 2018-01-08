package action

import "time"

// SupportedTimeFormat is time representation format to be acceptable for the application.
// all of the time representation format should be in manner of this format.
const SupportedTimeFormat = time.RFC3339

// Timestamp is Time data which implements some custom encoders and decoders.
type Timestamp time.Time

// TimestampNow is shorthand for Timestamp(time.Now()).
func TimestampNow() Timestamp {
	return Timestamp(time.Now())
}

// implements json.Marshaler interface.
func (t Timestamp) MarshalJSON() ([]byte, error) {
	return t.Time().MarshalJSON() // The format of default marshaler is RFC3339.
}

// implements json.Unmarshaler interface.
func (t *Timestamp) UnmarshalJSON(data []byte) error {
	var ts = time.Time{}
	err := ts.UnmarshalJSON(data)
	*t = Timestamp(ts)
	return err
}

// implements github.com/labstack/echo.BindUnmarshaler interface.
func (t *Timestamp) UnmarshalParam(src string) error {
	ts, err := time.Parse(SupportedTimeFormat, src)
	*t = Timestamp(ts)
	return err
}

// Time returns its internal representation as time.Time.
func (t *Timestamp) Time() time.Time {
	return time.Time(*t)
}
