package utils

import (
	"fmt"
	"reflect"
)

// DiscordCodeBlock returns the message formated as code block in discord
func DiscordCodeBlock(msg interface{}, lang string) string {
	return fmt.Sprintf("```%s\n%s\n```", lang, msg)
}

// DiscordCodeSpan returns the message formated as code span in discord
func DiscordCodeSpan(msg interface{}) string {
	return fmt.Sprintf("`%s`", msg)
}

// DiscordBold returns the message formatted to be bold in discord
func DiscordBold(msg interface{}) string {
	return fmt.Sprintf("**%s**", msg)
}

// BitmaskCheck checks if the given value contains bits in the given mask
func BitmaskCheck(value int64, mask int64) bool {
	return value&mask == mask
}

// FilterRec recursively checks if the value struct complies with the filter struct.
// Structs have to be of the same type
func FilterRec(filter interface{}, value interface{}) (bool, error) {
	f := reflect.Indirect(reflect.ValueOf(filter))
	v := reflect.Indirect(reflect.ValueOf(value))

	if f.Type() != v.Type() {
		return false, newErrUnmatchedTypes(f.Type(), v.Type())
	}

	for i := 0; i < f.NumField(); i++ {
		switch f.Field(i).Kind() {
		case reflect.Slice, reflect.Array:
			for j := 0; j < f.Field(i).Len(); j++ {
				fEl := reflect.Indirect(f.Field(i).Index(j))
				if fEl.IsZero() {
					continue
				}
				passes := false
				for k := 0; k < v.Field(i).Len(); k++ {
					vEl := reflect.Indirect(v.Field(i).Index(k))
					match, err := FilterRec(fEl.Interface(), vEl.Interface())
					if err != nil {
						return false, err
					}
					if match {
						passes = true
						break
					}
				}
				if !passes {
					return false, nil
				}
			}
		case reflect.Map:
			for iter := f.Field(i).MapRange(); iter.Next(); {
				fVal := iter.Value()
				vVal := v.Field(i).MapIndex(iter.Key())

				if fVal.IsZero() {
					continue
				}
				if vVal == reflect.ValueOf(nil) {
					return false, nil
				}
				match, err := FilterRec(fVal.Interface(), vVal.Interface())
				if err != nil {
					return false, err
				}
				if !match {
					return false, nil
				}
			}
		default:
			if !f.Field(i).IsZero() && f.Field(i).Interface() != v.Field(i).Interface() {
				return false, nil
			}
		}
	}
	return true, nil
}
