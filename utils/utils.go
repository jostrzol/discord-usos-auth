package utils

import "fmt"

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
