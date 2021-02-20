package utils

import (
	"fmt"
	"reflect"
)

// ErrUnmatchedTypes represtents failure in
type ErrUnmatchedTypes struct {
	type1 reflect.Type
	type2 reflect.Type
}

func newErrUnmatchedTypes(type1 reflect.Type, type2 reflect.Type) *ErrUnmatchedTypes {
	return &ErrUnmatchedTypes{
		type1: type1,
		type2: type2,
	}
}
func (e *ErrUnmatchedTypes) Error() string {
	return fmt.Sprintf("Types unmached: %s and %s.", e.type1, e.type2)
}
