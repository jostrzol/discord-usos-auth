package commands

// CommandScope represents a scope in which a command is active
type CommandScope int

const (
	// ScopeGuild indicates that the command is usable in guilds chats
	ScopeGuild CommandScope = 1 << iota
	// ScopePrivate indicates that the command is usable in private chats
	ScopePrivate
)

var correctScopes = [...]CommandScope{ScopeGuild, ScopePrivate}
var maxScope CommandScope = 0

// init maxScope
func init() {
	for _, correctScope := range correctScopes {
		maxScope += correctScope
	}
}
