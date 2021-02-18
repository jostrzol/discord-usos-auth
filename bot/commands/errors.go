package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
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

// ErrHandler represtents failure in
type ErrHandler struct {
	error
	RedirectToCallerChannel bool
}

// NewErrHandler returns a pointer to a new instance of ErrCommandHandlerError
func NewErrHandler(err error, RedirectToCallerChannel bool) *ErrHandler {
	return &ErrHandler{
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

// ErrUnprivilaged represtents failure in executing a command due to insufficient privilages of the user
type ErrUnprivilaged struct {
	Message *discordgo.MessageCreate
	Command *DiscordCommand
}

func newErrUnprivilaged(Message *discordgo.MessageCreate, Command *DiscordCommand) *ErrUnprivilaged {
	return &ErrUnprivilaged{
		Message: Message,
		Command: Command,
	}
}
func (e *ErrUnprivilaged) Error() string {
	return "You are unprivilaged for this command"
}
