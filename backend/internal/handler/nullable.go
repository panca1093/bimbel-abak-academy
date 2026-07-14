package handler

import "encoding/json"

// Nullable distinguishes an absent JSON field from an explicit null, which a
// plain *T cannot: encoding/json leaves a *T nil in both cases, so a PATCH
// overlay written as `if req.X != nil` can never tell "the client wants this
// cleared" apart from "the client didn't send this field."
//
// Set is true iff the key was present in the request body at all, regardless
// of its value. Valid is true iff the key was present AND its value was not
// JSON null. A PATCH handler should overlay only when Set is true, and clear
// the target field when Set is true but Valid is false.
type Nullable[T any] struct {
	Value T
	Valid bool
	Set   bool
}

func (n *Nullable[T]) UnmarshalJSON(data []byte) error {
	n.Set = true
	if string(data) == "null" {
		n.Valid = false
		return nil
	}
	if err := json.Unmarshal(data, &n.Value); err != nil {
		return err
	}
	n.Valid = true
	return nil
}
