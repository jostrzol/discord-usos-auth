package bot

import (
	"github.com/Ogurczak/discord-usos-auth/usos"
)

// ErrAlreadyUnregisteredUser represtents failure in
type ErrAlreadyUnregisteredUser struct {
	UserID string
}

func newErrAlreadyUnregisteredUser(UserID string) *ErrAlreadyUnregisteredUser {
	return &ErrAlreadyUnregisteredUser{
		UserID: UserID,
	}
}
func (e *ErrAlreadyUnregisteredUser) Error() string {
	return "You were not registered for authorization."
}

// ErrUnregisteredUnauthorizedUser represents failure in authorization of unregistered user
type ErrUnregisteredUnauthorizedUser struct {
	UserID string
}

func newErrUnregisteredUnauthorizedUser(UserID string) *ErrUnregisteredUnauthorizedUser {
	return &ErrUnregisteredUnauthorizedUser{
		UserID: UserID,
	}
}
func (e *ErrUnregisteredUnauthorizedUser) Error() string {
	return "You must first register for authorization by adding reaction to the bot's message on a server."
}

// ErrFilteredOut represtents failure in passing the filter
type ErrFilteredOut struct {
	UserID string
}

func newErrFilteredOut(UserID string) *ErrFilteredOut {
	return &ErrFilteredOut{
		UserID: UserID,
	}
}
func (e *ErrFilteredOut) Error() string {
	return "You do not meet the requirements. Consult server administrators for details."
}

// ErrWrongVerifier represtents failure in verifying the user with usos caused by wrong verifier
type ErrWrongVerifier struct {
	*usos.ErrUnableToCall
	UserID                string
	RequestTokenGuildPair *requestTokenGuildPair
	verifier              string
}

func newErrWrongVerifier(cause *usos.ErrUnableToCall, UserID string, RequestTokenGuildPair *requestTokenGuildPair, verifier string) *ErrWrongVerifier {
	return &ErrWrongVerifier{
		ErrUnableToCall:       cause,
		UserID:                UserID,
		RequestTokenGuildPair: RequestTokenGuildPair,
		verifier:              verifier,
	}
}
func (e *ErrWrongVerifier) Error() string {
	return "Cannot make call for required information to usos-api. Probably the verifier is wrong"
}

// ErrAlreadyRegistered represtents failure in registering a new user for authorization because he was already registered
type ErrAlreadyRegistered struct {
	UserID                string
	RequestTokenGuildPair *requestTokenGuildPair
}

func newErrAlreadyRegistered(UserID string, RequestTokenGuildPair *requestTokenGuildPair) *ErrAlreadyRegistered {
	return &ErrAlreadyRegistered{
		UserID:                UserID,
		RequestTokenGuildPair: RequestTokenGuildPair,
	}
}
func (e *ErrAlreadyRegistered) Error() string {
	return "User was already registered for authorization"
}

// ErrLogChannelAlreadyAdded represtents failure in adding a new log channel, because it was added before
type ErrLogChannelAlreadyAdded struct {
	ChannelID string
}

func newErrLogChannelAlreadyAdded(ChannelID string) *ErrLogChannelAlreadyAdded {
	return &ErrLogChannelAlreadyAdded{
		ChannelID: ChannelID,
	}
}
func (e *ErrLogChannelAlreadyAdded) Error() string {
	return "This log channel is already added"
}

// ErrLogChannelNotAdded represtents failure in removing a log channel, because it was not in the log channel list anyway
type ErrLogChannelNotAdded struct {
	ChannelID string
}

func newErrLogChannelNotAdded(ChannelID string) *ErrLogChannelNotAdded {
	return &ErrLogChannelNotAdded{
		ChannelID: ChannelID,
	}
}
func (e *ErrLogChannelNotAdded) Error() string {
	return "This log channel is not on the log channel list"
}

// ErrLogChannelNotInGuild represtents failure in adding a new log channel due to it not belonging to the server
type ErrLogChannelNotInGuild struct {
	ChannelID string
	GuildID   string
}

func newErrLogChannelNotInGuild(ChannelID string, GuildID string) *ErrLogChannelNotInGuild {
	return &ErrLogChannelNotInGuild{
		ChannelID: ChannelID,
		GuildID:   GuildID,
	}
}
func (e *ErrLogChannelNotInGuild) Error() string {
	return "Ths log channel does not belog to the server"
}

// ErrRoleNotInGuild represtents failure in setting a role to one outside of guild
type ErrRoleNotInGuild struct {
	RoleID  string
	GuildID string
}

func newErrRoleNotInGuild(RoleID string, GuildID string) *ErrRoleNotInGuild {
	return &ErrRoleNotInGuild{
		RoleID:  RoleID,
		GuildID: GuildID,
	}
}
func (e *ErrRoleNotInGuild) Error() string {
	return "This role does not belong to this server"
}
