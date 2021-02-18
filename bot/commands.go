package bot

import (
	"errors"
	"fmt"

	"github.com/Ogurczak/discord-usos-auth/bot/commands"
	"github.com/Ogurczak/discord-usos-auth/usos"
	"github.com/Ogurczak/discord-usos-auth/utils"
	"github.com/akamensky/argparse"
	"github.com/bwmarrin/discordgo"
)

func (bot *UsosBot) setupCommandParser() (*commands.DiscordParser, error) {
	parser := commands.NewDiscordParser("!usos", "Execute the Usos Authrization Bot's discord commands", bot.Session)

	authMsgCmd := parser.NewCommand("auth-msg", "Spawn a message which upon being reacted to begins the process of usos authentication")
	authMsgCmd.SetPrivilagesRequired(true)
	err := authMsgCmd.SetScope(commands.ScopeGuild)
	if err != nil {
		return nil, err
	}
	prompt := authMsgCmd.String("m", "msg",
		&argparse.Options{Required: false,
			Help:    "custom prompt on the message",
			Default: "React to this message to get authorized!"})
	authMsgCmd.Handler = func(cmd *commands.DiscordCommand, e *discordgo.MessageCreate) *commands.ErrHandler {
		err := bot.spawnAuthorizeMessage(e.ChannelID, *prompt)
		if err != nil {
			return commands.NewErrHandler(err, true)
		}
		return nil
	}

	verifyCmd := parser.NewCommand("verify", "Verify the user's authentication using the provided verification code")
	err = verifyCmd.SetScope(commands.ScopePrivate)
	if err != nil {
		return nil, err
	}
	verifier := verifyCmd.String("c", "code",
		&argparse.Options{Required: false,
			Help: "verification code"})
	abort := verifyCmd.Flag("a", "abort",
		&argparse.Options{Required: false,
			Help: "abort current verification process"})
	verifyCmd.Handler = func(cmd *commands.DiscordCommand, e *discordgo.MessageCreate) *commands.ErrHandler {
		if *abort {
			err := bot.removeUnauthorizedUser(e.Author.ID)
			switch err.(type) {
			case *ErrAlreadyUnregisteredUser:
				return commands.NewErrHandler(err, true)
			case nil:
				_, err = bot.ChannelMessageSend(e.ChannelID, "Authorization abort successful")
				if err != nil {
					return commands.NewErrHandler(err, false)
				}
				return nil
			default:
				return commands.NewErrHandler(err, false)
			}

		}
		if *verifier == "" {
			// ugly but i don't see a better approach with this argparse library
			return commands.NewErrHandler(errors.New("[-c|--code] or [-a|--abort] is required"), true)
		}
		err := bot.finalizeAuthorization(e.Author, *verifier)
		switch err.(type) {
		case *ErrUnregisteredUnauthorizedUser, *ErrFilteredOut, *usos.ErrUnableToCall, *ErrRoleNotInGuild:
			return commands.NewErrHandler(err, true)
		case nil:
			return nil
		default:
			return commands.NewErrHandler(err, false)
		}
	}

	logChannelCmd := parser.NewCommand("log", "manage discord log channels")
	logChannelCmd.SetPrivilagesRequired(true)

	addLogChannelCmd := logChannelCmd.NewCommand("add", "Add a new log channel to this server")
	err = addLogChannelCmd.SetScope(commands.ScopeGuild)
	if err != nil {
		return nil, err
	}
	logChannelID := addLogChannelCmd.String("i", "id",
		&argparse.Options{Required: false,
			Help: "channel id, defaults to channel in which the command was called"})
	addLogChannelCmd.Handler = func(cmd *commands.DiscordCommand, e *discordgo.MessageCreate) *commands.ErrHandler {
		if *logChannelID == "" {
			*logChannelID = e.ChannelID
		}
		err := bot.addLogChannel(e.GuildID, *logChannelID)
		if err != nil {
			return commands.NewErrHandler(err, true)
		}
		bot.ChannelMessageSend(e.ChannelID, "Successfully added a new log channel")
		return nil
	}

	removeLogChannelCmd := logChannelCmd.NewCommand("remove", "Remove a log channel from this server")
	err = removeLogChannelCmd.SetScope(commands.ScopeGuild)
	if err != nil {
		return nil, err
	}
	logChannelIDToRemove := removeLogChannelCmd.String("i", "id",
		&argparse.Options{Required: false,
			Help: "channel id, defaults to channel in which the command was called"})
	removeLogChannelCmd.Handler = func(cmd *commands.DiscordCommand, e *discordgo.MessageCreate) *commands.ErrHandler {
		if *logChannelIDToRemove == "" {
			*logChannelIDToRemove = e.ChannelID
		}
		err := bot.removeLogChannel(e.GuildID, *logChannelIDToRemove)
		if err != nil {
			return commands.NewErrHandler(err, true)
		}
		bot.ChannelMessageSend(e.ChannelID, "Successfully removed a log channel")
		return nil
	}

	listLogChannelCmd := logChannelCmd.NewCommand("list", "List log channels bound to this server")
	err = listLogChannelCmd.SetScope(commands.ScopeGuild)
	if err != nil {
		return nil, err
	}
	listLogChannelCmd.Handler = func(cmd *commands.DiscordCommand, e *discordgo.MessageCreate) *commands.ErrHandler {
		guildInfo := bot.getGuildUsosInfo(e.GuildID)
		if len(guildInfo.logChannelIDs) == 0 {
			msg := "No log channels on this servers set yet."
			_, err := bot.ChannelMessageSend(e.ChannelID, msg)
			if err != nil {
				return commands.NewErrHandler(err, false)
			}
			return nil
		}
		channels := make([]*discordgo.Channel, 0, len(guildInfo.logChannelIDs))
		for channelID := range guildInfo.logChannelIDs {
			channel, err := bot.Channel(channelID)
			if err != nil {
				return commands.NewErrHandler(err, false)
			}
			channels = append(channels, channel)
		}
		guild, err := bot.Guild(e.GuildID)
		if err != nil {
			return commands.NewErrHandler(err, false)
		}
		msg := fmt.Sprintf("%s's log channels:\n", guild.Name)
		for i, channel := range channels {
			msg += fmt.Sprintf("\t%d. %s#%s\n", i+1, channel.Name, channel.ID)
		}
		_, err = bot.ChannelMessageSend(e.ChannelID, utils.DiscordCodeBlock(msg, ""))
		if err != nil {
			return commands.NewErrHandler(err, false)
		}
		return nil
	}

	roleCmd := parser.NewCommand("role", "change authorization role")
	err = roleCmd.SetScope(commands.ScopeGuild)
	if err != nil {
		return nil, err
	}
	roleID := roleCmd.String("i", "id", &argparse.Options{Required: true,
		Help: "set the server's authorization role"})
	roleCmd.Handler = func(cmd *commands.DiscordCommand, e *discordgo.MessageCreate) *commands.ErrHandler {
		guildInfo := bot.getGuildUsosInfo(e.GuildID)

		roles, err := bot.GuildRoles(e.GuildID)
		if err != nil {
			return commands.NewErrHandler(err, false)
		}

		for _, role := range roles {
			if role.ID == *roleID {
				guildInfo.authorizeRoleID = *roleID
				_, err = bot.ChannelMessageSend(e.ChannelID, "Authorization role ID set successfully")
				if err != nil {
					return commands.NewErrHandler(err, false)
				}
				return nil
			}
		}

		err = newErrRoleNotInGuild(*roleID, e.GuildID)
		return commands.NewErrHandler(err, true)
	}

	return parser, nil
}
