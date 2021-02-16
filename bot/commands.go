package bot

import (
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
	err := authMsgCmd.SetScope(commands.ScopeGuild)
	if err != nil {
		return nil, err
	}
	prompt := authMsgCmd.String("m", "msg",
		&argparse.Options{Required: false,
			Help:    "custom prompt on the message",
			Default: "React to this message to get authorized!"})
	authMsgCmd.Handler = func(cmd *commands.DiscordCommand, e *discordgo.MessageCreate) *commands.ErrInCommandHandler {
		err := bot.spawnAuthorizeMessage(e.ChannelID, *prompt)
		if err != nil {
			return commands.NewErrInCommandHandler(err, true)
		}
		return nil
	}

	verifyCmd := parser.NewCommand("verify", "Verify the user's authentication using the provided verification code")
	err = verifyCmd.SetScope(commands.ScopePrivate)
	if err != nil {
		return nil, err
	}
	verifier := verifyCmd.String("c", "code",
		&argparse.Options{Required: true,
			Help: "verification code"})
	verifyCmd.Handler = func(cmd *commands.DiscordCommand, e *discordgo.MessageCreate) *commands.ErrInCommandHandler {
		err := bot.finalizeAuthorization(e.Author, *verifier)
		if err != nil {
			switch err.(type) {
			case *ErrUnregisteredUnauthorizedUser, *ErrFilteredOut, *usos.ErrUnableToCall:
				return commands.NewErrInCommandHandler(err, true)
			default:
				return commands.NewErrInCommandHandler(err, false)
			}
		}
		return nil
	}

	logChannelCmd := parser.NewCommand("log", "manage discord log channels")

	addLogChannelCmd := logChannelCmd.NewCommand("add", "Add a new log channel to this server")
	err = addLogChannelCmd.SetScope(commands.ScopeGuild)
	if err != nil {
		return nil, err
	}
	logChannelID := addLogChannelCmd.String("", "id",
		&argparse.Options{Required: false,
			Help:    "channel id, defaults to channel in which the command was called",
			Default: ""})
	addLogChannelCmd.Handler = func(cmd *commands.DiscordCommand, e *discordgo.MessageCreate) *commands.ErrInCommandHandler {
		if *logChannelID == "" {
			*logChannelID = e.ChannelID
		}
		err := bot.addLogChannel(e.GuildID, *logChannelID)
		if err != nil {
			return commands.NewErrInCommandHandler(err, true)
		}
		bot.ChannelMessageSend(e.ChannelID, "Successfully added a new log channel")
		return nil
	}

	removeLogChannelCmd := logChannelCmd.NewCommand("remove", "Remove a log channel from this server")
	err = removeLogChannelCmd.SetScope(commands.ScopeGuild)
	if err != nil {
		return nil, err
	}
	logChannelIDToRemove := removeLogChannelCmd.String("", "id",
		&argparse.Options{Required: false,
			Help:    "channel id, defaults to channel in which the command was called",
			Default: ""})
	removeLogChannelCmd.Handler = func(cmd *commands.DiscordCommand, e *discordgo.MessageCreate) *commands.ErrInCommandHandler {
		if *logChannelIDToRemove == "" {
			*logChannelIDToRemove = e.ChannelID
		}
		err := bot.removeLogChannel(e.GuildID, *logChannelIDToRemove)
		if err != nil {
			return commands.NewErrInCommandHandler(err, true)
		}
		bot.ChannelMessageSend(e.ChannelID, "Successfully removed a log channel")
		return nil
	}

	listLogChannelCmd := logChannelCmd.NewCommand("list", "List log channels bound to this server")
	err = listLogChannelCmd.SetScope(commands.ScopeGuild)
	if err != nil {
		return nil, err
	}
	listLogChannelCmd.Handler = func(cmd *commands.DiscordCommand, e *discordgo.MessageCreate) *commands.ErrInCommandHandler {
		if len(bot.logChannelIDMap[e.GuildID]) == 0 {
			msg := "No log channels on this servers set yet."
			_, err := bot.ChannelMessageSend(e.ChannelID, msg)
			if err != nil {
				return commands.NewErrInCommandHandler(err, false)
			}
			return nil
		}
		channels := make([]*discordgo.Channel, 0, len(bot.logChannelIDMap[e.GuildID]))
		for channelID := range bot.logChannelIDMap[e.GuildID] {
			channel, err := bot.Channel(channelID)
			if err != nil {
				return commands.NewErrInCommandHandler(err, false)
			}
			channels = append(channels, channel)
		}
		guild, err := bot.Guild(e.GuildID)
		if err != nil {
			return commands.NewErrInCommandHandler(err, false)
		}
		msg := fmt.Sprintf("%s's log channels:\n", guild.Name)
		for i, channel := range channels {
			msg += fmt.Sprintf("\t%d. %s#%s\n", i+1, channel.Name, channel.ID)
		}
		_, err = bot.ChannelMessageSend(e.ChannelID, utils.DiscordCodeBlock(msg, ""))
		if err != nil {
			return commands.NewErrInCommandHandler(err, false)
		}
		return nil
	}

	return parser, nil
}
