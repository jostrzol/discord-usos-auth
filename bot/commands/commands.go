package commands

import (
	"log"
	"strings"

	"github.com/Ogurczak/discord-usos-auth/utils"
	"github.com/akamensky/argparse"
	"github.com/bwmarrin/discordgo"
)

// DiscordParser represents a parser for commands send through discord
type DiscordParser struct {
	*argparse.Parser
	*DiscordCommand
	helpSName  string
	helpLName  string
	parsedHelp bool
}

// DiscordCommand represents a command send through discord
type DiscordCommand struct {
	*argparse.Command

	// Handler is the function executed during parsing this command
	Handler func(*DiscordCommand, *discordgo.MessageCreate) *ErrHandler

	session            *discordgo.Session
	commands           []*DiscordCommand
	scope              CommandScope
	PrivilagesRequired bool
	parent             *DiscordCommand
}

// NewDiscordParser returns a new instance of DiscordParser Class
func NewDiscordParser(name string, description string, session *discordgo.Session) *DiscordParser {
	parser := new(DiscordParser)

	parser.Parser = argparse.NewParser(name, description)

	parser.DiscordCommand = newDiscordCommand(&parser.Parser.Command, session)

	parser.helpSName = "h"
	parser.helpLName = "help"

	return parser
}

// Parse parses given string arguments and sends parsing errors to user with the given user ID.
// See github.com/akamensky/argparse.Parser.Parse
func (parser *DiscordParser) Parse(e *discordgo.MessageCreate) error {
	args := strings.Fields(e.Content)

	// replace help func
	parser.setDiscordHelp(e.ChannelID)

	// original parse method
	parseErr := parser.Parser.Parse(args)
	if parseErr != nil {
		parseErr = newErrParse(parseErr)
		message := parser.youngestCommandHappened().Usage(parseErr.Error())
		_, errSend := parser.session.ChannelMessageSend(e.ChannelID, message)
		if errSend != nil {
			return errSend
		}
		return parseErr
	}

	// set parsedHelp
	for _, arg := range args {
		if arg == "-"+parser.helpSName || "--"+arg == parser.helpLName {
			parser.parsedHelp = true
			break
		}
	}

	return nil
}

// Handle calls appropriate command handlers
func (parser *DiscordParser) Handle(e *discordgo.MessageCreate) error {
	// set youngest command
	cmd := parser.youngestCommandHappened()

	// validate scopes
	var scopeErr error
	if e.GuildID == "" {
		scopeErr = cmd.validateScope(ScopePrivate)
	} else {
		scopeErr = cmd.validateScope(ScopeGuild)
	}
	if scopeErr != nil {
		message := cmd.Usage(scopeErr.Error())
		_, err := parser.session.ChannelMessageSend(e.ChannelID, message)
		if err != nil {
			return err
		}
		return scopeErr
	}

	// validate privilages
	if cmd.PrivilagesRequired {
		isPrivilaged, err := cmd.IsPrivilaged(e)
		if err != nil {
			return err
		}
		if !isPrivilaged {
			privilageErr := newErrUnprivilaged(e, cmd)
			message := cmd.Usage(privilageErr.Error())
			_, err := parser.session.ChannelMessageSend(e.ChannelID, message)
			if err != nil {
				return err
			}
			return privilageErr
		}
	}

	// execute handlers
	handleErr := cmd.Handler(cmd, e)
	if handleErr != nil {
		if handleErr.RedirectToCallerChannel {
			message := cmd.Usage(handleErr.Error())
			_, err := parser.session.ChannelMessageSend(e.ChannelID, message)
			if err != nil {
				return err
			}
		}
		return handleErr
	}

	return nil
}

// SetHelp removes the previous help argument, and creates a new one with the desired sname/lname
func (parser *DiscordParser) SetHelp(sname string, lname string) {
	parser.helpSName = sname
	parser.helpLName = lname
	parser.Parser.SetHelp(sname, lname)
}

// ParsedHelp indicates if the help argument was parsed or not
func (parser *DiscordParser) ParsedHelp() bool {
	return parser.parsedHelp
}

// NewCommand will create a sub-command and propagate all necessary fields.
// See github.com/akamensky/argparse.Command.NewCommand
func (command *DiscordCommand) NewCommand(name string, description string) *DiscordCommand {
	argparseCommand := command.Command.NewCommand(name, description)
	newCommand := newDiscordCommand(argparseCommand, command.session)

	newCommand.PrivilagesRequired = command.PrivilagesRequired
	newCommand.scope = command.scope

	// set relations
	newCommand.parent = command
	command.commands = append(command.commands, newCommand)

	return newCommand
}

// newDiscordCommand returns a new discord command from a regular command
func newDiscordCommand(command *argparse.Command, session *discordgo.Session) *DiscordCommand {
	discordCommand := &DiscordCommand{
		Command:  command,
		commands: make([]*DiscordCommand, 0),
		session:  session,
		scope:    maxScope,
		Handler:  func(cmd *DiscordCommand, e *discordgo.MessageCreate) *ErrHandler { return nil },
	}
	discordCommand.ExitOnHelp(false)
	return discordCommand
}

// func (command *DiscordCommand) setUserIDRecursive(message *discordgo.MessageCreate) {
// 	command.message = message
// 	for _, command := range command.commands {
// 		command.setUserIDRecursive(message)
// 	}
// }

func (command *DiscordCommand) validateScope(demandedScope CommandScope) error {
	if command.scope&demandedScope != demandedScope {
		return newErrCommandInWrongScope(demandedScope, command)
	}
	return nil
}

// GetScope returns the command's operating scope
func (command *DiscordCommand) GetScope() CommandScope {
	return command.scope
}

// SetScope sets the command's operation scope
func (command *DiscordCommand) SetScope(scope CommandScope) error {
	if scope < 0 || scope > maxScope {
		return newErrIncorrectCommandScope(correctScopes[:], scope)
	}
	command.scope = scope
	return nil
}

// Session returns the commands discordgo.Session context
func (command *DiscordCommand) Session() *discordgo.Session {
	return command.session
}

func (command *DiscordCommand) youngestCommandHappened() *DiscordCommand {
	for _, cmd := range command.commands {
		if cmd.Happened() {
			return cmd.youngestCommandHappened()
		}
	}
	return command
}

func (command *DiscordCommand) setDiscordHelp(channelID string) {
	oldHelpFunc := command.HelpFunc
	command.HelpFunc = func(c *argparse.Command, msg interface{}) string {
		message := oldHelpFunc(c, msg)
		_, err := command.session.ChannelMessageSend(channelID, utils.DiscordCodeBlock(message, ""))
		if err != nil {
			log.Println(err)
			return ""
		}
		return ""
	}
	for _, cmd := range command.commands {
		cmd.setDiscordHelp(channelID)
	}
}

// Usage returns discord-formatted Usage version of github.com/akamensky/argparse.Command.Usage
func (command *DiscordCommand) Usage(msg interface{}) string {
	var prefix string
	if msg != nil {
		prefix = utils.DiscordCodeBlock(msg, "fix")
	} else {
		prefix = ""
	}
	return prefix + utils.DiscordCodeBlock(command.Command.Usage(nil), "")
}

// GetParent exposes Command's parent field
func (command *DiscordCommand) GetParent() *DiscordCommand {
	return command.parent
}

// PrivilagesRequired indicates if privilages are required to execute this command
// func (command *DiscordCommand) PrivilagesRequired() bool {
// 	if command.privilagesRequired {
// 		return true
// 	}
// 	if command.GetParent() != nil {
// 		return command.GetParent().PrivilagesRequired()
// 	}
// 	return false
// }

// SetPrivilagesRequired sets the preivilage level for this command
// func (command *DiscordCommand) SetPrivilagesRequired(b bool) {
// 	command.privilagesRequired = b
// }

// IsPrivilaged checks if a given message is privilaged for this command
func (command *DiscordCommand) IsPrivilaged(e *discordgo.MessageCreate) (bool, error) {
	// private always privilaged
	if e.GuildID == "" {
		return true, nil
	}

	// is owner
	guild, err := command.session.Guild(e.GuildID)
	if err != nil {
		return false, err
	}
	if guild.OwnerID == e.Author.ID {
		return true, nil
	}

	// get roles
	member, err := command.session.GuildMember(e.GuildID, e.Author.ID)
	if err != nil {
		return false, err
	}
	roles, err := command.session.GuildRoles(e.GuildID)
	if err != nil {
		return false, err
	}

	rolesMap := make(map[string]*discordgo.Role)
	for _, role := range roles {
		rolesMap[role.ID] = role
	}

	// is administrator or has manage server permissions
	for _, roleID := range member.Roles {
		perms := rolesMap[roleID].Permissions
		if utils.BitmaskCheck(perms, discordgo.PermissionManageServer) ||
			utils.BitmaskCheck(perms, discordgo.PermissionAdministrator) {
			return true, nil
		}
	}
	return false, nil
}
