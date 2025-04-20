package domain

import (
	"encoding/json"
	"fmt"
)

type Null[T any] struct {
	Value T
	Valid bool
}

func NewNull[T any](v T) Null[T] {
	return Null[T]{
		Value: v,
		Valid: true,
	}
}

func (n Null[T]) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}

	return json.Marshal(n.Value)
}

func (n *Null[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		var v T

		n.Value = v
		n.Valid = false

		return nil
	}

	var v T

	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("error unmarshaling Null: %w", err)
	}

	n.Value = v
	n.Valid = true

	return nil
}
