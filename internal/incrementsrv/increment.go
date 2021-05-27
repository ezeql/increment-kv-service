package incrementsrv

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/google/uuid"
)

var (
	ErrInput = errors.New("invalid input")
	ErrKey   = errors.New("missing or invalid 'key'")
	ErrValue = errors.New("missing or invalid 'value'")
)

type Input struct {
	Key   uuid.UUID `json:"key"`
	Value uint64    `json:"value"`
}

func InputFromJSONReader(r io.Reader) (*Input, error) {
	var inc Input

	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&inc); err != nil {
		// is the UUID invalid?
		if uuid.IsInvalidLengthError(err) {
			return nil, ErrKey
		}
		//try to find out which field failed
		if t, ok := err.(*json.UnmarshalTypeError); ok {
			switch t.Field {
			case "key":
				return nil, ErrKey
			case "value":
				return nil, ErrValue
			}
		}
		// could't narrow problem to field level
		return nil, ErrInput
	}

	if err := inc.Valid(); err != nil {
		return nil, err
	}

	return &inc, nil
}

//Valid performs data validation
func (inc *Input) Valid() error {
	if err := ValueValid(inc.Value); err != nil {
		return err
	}
	return KeyValid(inc.Key.String())
}

//KeyValid validates if a string is an UUID
func KeyValid(k string) error {
	if k == "00000000-0000-0000-0000-000000000000" {
		return ErrKey
	}

	if _, err := uuid.Parse(k); err != nil {
		return ErrKey
	}
	return nil
}

//ValueValid validates if the argument is a valid increment value
func ValueValid(v uint64) error {
	if v < 1 {
		return ErrValue
	}
	return nil
}
