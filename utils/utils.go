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

// FilterRec recursively checks if the value struct complies with the filter struct
func FilterRec(filter interface{}, value interface{}) bool {
	f := reflect.ValueOf(filter)
	v := reflect.ValueOf(value)

	for i := 0; i < f.NumField(); i++ {
		switch f.Field(i).Kind() {
		case reflect.Slice, reflect.Array:
			return FilterRec(f.Field(i).Interface(), v.Field(i).Interface())
		default:
			if !f.IsZero() && f.Field(i) != v.Field(i) {
				return false
			}
		}
	}
	return true
}
