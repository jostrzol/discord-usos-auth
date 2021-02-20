package bot

import (
	"encoding/json"
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
	authMsgCmd.PrivilagesRequired = true
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
		case *ErrUnregisteredUnauthorizedUser, *ErrFilteredOut, *usos.ErrUnableToCall, *ErrRoleNotInGuild, *ErrWrongVerifier:
			return commands.NewErrHandler(err, true)
		case nil:
			return nil
		default:
			return commands.NewErrHandler(err, false)
		}
	}

	logChannelCmd := parser.NewCommand("log", "manage discord log channels")
	logChannelCmd.PrivilagesRequired = true
	err = logChannelCmd.SetScope(commands.ScopeGuild)
	if err != nil {
		return nil, err
	}

	addLogChannelCmd := logChannelCmd.NewCommand("add", "Add a new log channel to this server")
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
	roleCmd.PrivilagesRequired = true
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

	filterCmd := parser.NewCommand("filter", "manage usos filters")
	filterCmd.PrivilagesRequired = true
	err = filterCmd.SetScope(commands.ScopeGuild)
	if err != nil {
		return nil, err
	}

	addFilterCmd := filterCmd.NewCommand("add", "add usos filter; user has to pass at least one of the filters to get past the authorization successfully")
	programmes := addFilterCmd.StringList("p", "programme", &argparse.Options{Required: false,
		Help: "Programme names which the user is required too have (all) to pass."})
	addFilterCmd.Handler = func(cmd *commands.DiscordCommand, e *discordgo.MessageCreate) *commands.ErrHandler {
		if len(*programmes) == 0 {
			return commands.NewErrHandler(newErrEmptyFilter(), true)
		}

		usosProrammes := make([]usos.Programme, len(*programmes))
		for i, programme := range *programmes {
			usosProrammes[i] = usos.Programme{Name: programme}
		}

		filter := &usos.User{
			Programmes: usosProrammes,
		}

		guildInfo := bot.getGuildUsosInfo(e.GuildID)
		guildInfo.filters = append(guildInfo.filters, filter)

		_, err := bot.ChannelMessageSend(e.ChannelID, "Filter added successfully")
		if err != nil {
			return commands.NewErrHandler(err, false)
		}
		return nil
	}

	removeFilterCmd := filterCmd.NewCommand("remove", "remove an existing filter")
	removeFilterID := removeFilterCmd.Int("i", "id", &argparse.Options{Required: true,
		Help: fmt.Sprintf("Filter's id, can be obtained using the %s command", utils.DiscordCodeSpan("!usos filter list"))})
	removeFilterCmd.Handler = func(cmd *commands.DiscordCommand, e *discordgo.MessageCreate) *commands.ErrHandler {
		guildInfo := bot.getGuildUsosInfo(e.GuildID)
		if *removeFilterID < 1 || *removeFilterID > len(guildInfo.filters) {
			commands.NewErrHandler(newErrNoSuchFilter(*removeFilterID), true)
		}
		guildInfo.filters = append(guildInfo.filters[:*removeFilterID-1], guildInfo.filters[*removeFilterID:]...)

		_, err := bot.ChannelMessageSend(e.ChannelID, "Filter removed successfully")
		if err != nil {
			return commands.NewErrHandler(err, false)
		}
		return nil

	}

	listFilterCmd := filterCmd.NewCommand("list", "list usos filters")
	listFilterCmd.Handler = func(cmd *commands.DiscordCommand, e *discordgo.MessageCreate) *commands.ErrHandler {
		guildInfo := bot.getGuildUsosInfo(e.GuildID)
		if len(guildInfo.filters) == 0 {
			_, err := bot.ChannelMessageSend(e.ChannelID, "No filters set yet, all usos-verified users are let through.")
			if err != nil {
				return commands.NewErrHandler(err, false)
			}
			return nil
		}

		guild, err := bot.Guild(e.GuildID)
		if err != nil {
			return commands.NewErrHandler(err, false)
		}

		msg := fmt.Sprintf("%s's Filters:", guild.Name)
		for i, filter := range guildInfo.filters {
			msg += fmt.Sprintf("\n%d.\n", i+1)
			body, err := json.MarshalIndent(filter, "", "    ")
			if err != nil {
				return commands.NewErrHandler(err, false)
			}
			msg += string(body)
		}
		_, err = bot.ChannelMessageSend(e.ChannelID, utils.DiscordCodeBlock(msg, ""))
		if err != nil {
			return commands.NewErrHandler(err, false)
		}
		return nil
	}

	return parser, nil
}
