package standartjson

import (
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrMarshaling   = errors.New("standartjson: failed to marshal")
	ErrUnmarshaling = errors.New("standartjson: failed to unmarshal")
)

type JSON struct{}

func New() *JSON {
	return &JSON{}
}

func (j *JSON) Marshal(data any) ([]byte, error) {
	d, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMarshaling, err)
	}
	return d, nil
}

func (j *JSON) Unmarshal(data []byte, v any) error {
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("%w: %v", ErrUnmarshaling, err)
	}
	return nil
}
