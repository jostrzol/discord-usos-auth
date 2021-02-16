package commands

import (
	"fmt"
)

// ErrIncorrectCommandScope represtents failure in setting a command scope to one not defined
type ErrIncorrectCommandScope struct {
	Correct []CommandScope
	Scope   CommandScope
}

func newErrIncorrectCommandScope(Correct []CommandScope, Scope CommandScope) *ErrIncorrectCommandScope {
	return &ErrIncorrectCommandScope{
		Correct: Correct,
		Scope:   Scope,
	}
}
func (e *ErrIncorrectCommandScope) Error() string {
	return fmt.Sprintf("Tried to set incorrect command scope.")
}

// ErrCommandInWrongScope represtents failure in executing a command due to it beeing used in the wrong scope
type ErrCommandInWrongScope struct {
	Scope   CommandScope
	Command *DiscordCommand
}

func newErrCommandInWrongScope(Scope CommandScope, Command *DiscordCommand) *ErrCommandInWrongScope {
	return &ErrCommandInWrongScope{
		Scope:   Scope,
		Command: Command,
	}
}
func (e *ErrCommandInWrongScope) Error() string {
	switch e.Command.scope {
	case ScopeGuild:
		return "This command may only be used in server chats"
	case ScopePrivate:
		return "This command may only be used in private chats"
	default:
		return "Command used in wrong scope"
	}
}

// ErrInCommandHandler represtents failure in
type ErrInCommandHandler struct {
	error
	RedirectToCallerChannel bool
}

// NewErrInCommandHandler returns a pointer to a new instance of ErrCommandHandlerError
func NewErrInCommandHandler(err error, RedirectToCallerChannel bool) *ErrInCommandHandler {
	return &ErrInCommandHandler{
		error:                   err,
		RedirectToCallerChannel: RedirectToCallerChannel,
	}
}

// ErrParse represtents failure in parsing commands
type ErrParse struct {
	error
}

func newErrParse(err error) *ErrParse {
	return &ErrParse{
		error: err,
	}
}