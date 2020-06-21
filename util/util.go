package util

import (
	"fmt"
	"time"
)

func TimeToMS(t time.Time) *int {
	if t.IsZero() {
		return nil
	}

	i := int(t.UnixNano() / int64(time.Millisecond))
	return &i
}

func ErrToString(e error) *string {
	if e == nil {
		return nil
	}

	s := e.Error()
	return &s
}

// TODO: this is cumbersome; ponder a better approach

type displayableErr struct{ e error }

func Err(e error) error {
	return &displayableErr{e}
}

func (de *displayableErr) MarshalJSON() ([]byte, error) {
	if de.e == nil {
		return nil, nil
	}
	return []byte(fmt.Sprintf("%q", de.e.Error())), nil
}

func (de *displayableErr) Error() string {
	return de.e.Error()
}
