package bot

import "github.com/bwmarrin/discordgo"

// ErrUnregisteredUserNotFound represtents failure in aborting an authorization of non-registered unauthorized user
type ErrUnregisteredUserNotFound struct {
	UserID string
}

func newErrUnregisteredUserNotFound(UserID string) *ErrUnregisteredUserNotFound {
	return &ErrUnregisteredUserNotFound{
		UserID: UserID,
	}
}
func (e *ErrUnregisteredUserNotFound) Error() string {
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
	error
	UserID                string
	RequestTokenGuildPair *requestTokenGuildPair
	verifier              string
}

func newErrWrongVerifier(cause error, UserID string, RequestTokenGuildPair *requestTokenGuildPair, verifier string) *ErrWrongVerifier {
	return &ErrWrongVerifier{
		error:                 cause,
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

// ErrLogChannelPresent represtents failure in adding a new log channel, because it was added before
type ErrLogChannelPresent struct {
	ChannelID string
	GuildID   string
}

func newErrLogChannelPresent(ChannelID string, GuildID string) *ErrLogChannelPresent {
	return &ErrLogChannelPresent{
		ChannelID: ChannelID,
		GuildID:   GuildID,
	}
}
func (e *ErrLogChannelPresent) Error() string {
	return "This log channel is already present on this server"
}

// ErrLogChannelNotFound represtents an error in attempt to get a log channel which does not belong to the server
type ErrLogChannelNotFound struct {
	ChannelID string
	GuildID   string
}

func newErrLogChannelNotFound(ChannelID string, GuildID string) *ErrLogChannelNotFound {
	return &ErrLogChannelNotFound{
		ChannelID: ChannelID,
		GuildID:   GuildID,
	}
}
func (e *ErrLogChannelNotFound) Error() string {
	return "This log channel does not belog to the server"
}

// ErrChannelNotFound represtents an error in attempt to get a non-existant channel
type ErrChannelNotFound struct {
	error
	ChannelID string
}

func newErrChannelNotFound(cause error, ChannelID string) *ErrChannelNotFound {
	return &ErrChannelNotFound{
		error:     cause,
		ChannelID: ChannelID,
	}
}
func (e *ErrChannelNotFound) Error() string {
	return "This channel does not exit"
}

// ErrAuthorizeRoleNotFound represtents failure in an attempt to get an authorize role
// that does not have one registered
type ErrAuthorizeRoleNotFound struct {
	GuildID string
}

func newErrAuthorizeRoleNotFound(GuildID string) *ErrAuthorizeRoleNotFound {
	return &ErrAuthorizeRoleNotFound{
		GuildID: GuildID,
	}
}
func (e *ErrAuthorizeRoleNotFound) Error() string {
	return "This server does not have an authorize role registered"
}

// ErrRoleNotFound represtents failure in attempt to get a role that does not exist on a server
type ErrRoleNotFound struct {
	RoleID  string
	GuildID string
}

func newErrRoleNotFound(RoleID string, GuildID string) *ErrRoleNotFound {
	return &ErrRoleNotFound{
		RoleID:  RoleID,
		GuildID: GuildID,
	}
}
func (e *ErrRoleNotFound) Error() string {
	return "This role does not belong to this server"
}

// ErrFilterEmpty represtents failure in setting a filter to an empty one
type ErrFilterEmpty struct{}

func newErrFilterEmpty() *ErrFilterEmpty {
	return &ErrFilterEmpty{}
}

func (e *ErrFilterEmpty) Error() string {
	return "This filter is empty"
}

// ErrFilterNotFound represtents failure in removing a non-existant filter
type ErrFilterNotFound struct {
	ID int
}

func newErrFilterNotFound(ID int) *ErrFilterNotFound {
	return &ErrFilterNotFound{
		ID: ID,
	}
}
func (e *ErrFilterNotFound) Error() string {
	return "No filter with such ID specified."
}

// IsNotFound checks if given error is a not found error (on discordgo package and this package)
func IsNotFound(err error) bool {
	switch err.(type) {
	case *ErrChannelNotFound, *ErrLogChannelNotFound, *ErrRoleNotFound, *ErrAuthorizeRoleNotFound:
		return true
	case *discordgo.RESTError:
		code := err.(*discordgo.RESTError).Response.StatusCode
		return code == 404
	default:
		return false
	}
}
