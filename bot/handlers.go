package bot

import (
	"log"
	"strings"

	"github.com/Ogurczak/discord-usos-auth/bot/commands"
	"github.com/bwmarrin/discordgo"
)

// handlerMessageCreate handles messeges received by the bot
func (bot *UsosBot) handlerMessageCreate(session *discordgo.Session, e *discordgo.MessageCreate) {
	// ignore self
	if e.Author.ID == bot.State.User.ID {
		return
	}

	if strings.Fields(e.Content)[0] != "!usos" {
		return
	}
	log.Println("Command received")

	parser, err := bot.setupCommandParser()
	if err != nil {
		log.Println(err)
		return
	}

	err = parser.Parse(e)
	switch err.(type) {
	case *commands.ErrParse:
		return
	case nil:
		// no-op
	default:
		log.Println(err)
		return
	}

	if parser.ParsedHelp() {
		return
	}

	err = parser.Handle(e)
	switch err.(type) {
	case *commands.ErrHandler, *commands.ErrCommandInWrongScope, *commands.ErrUnprivilaged, nil:
		// no-op
	default:
		log.Println(err)
	}
}

// handlerReady indicates that the bot is ready
func (bot *UsosBot) handlerReady(session *discordgo.Session, e *discordgo.Ready) {
	log.Println("Ready")
}

// handlerReactionAdd handles reactions added to bot's messages
func (bot *UsosBot) handlerReactionAdd(session *discordgo.Session, e *discordgo.MessageReactionAdd) {
	log.Println("Reaction add")
	guildInfo := bot.getGuildUsosInfo(e.GuildID)
	if guildInfo.AuthorizeMessegeIDs[e.ChannelID][e.MessageID] {
		_, err := bot.ChannelMessage(e.ChannelID, e.MessageID)
		if err != nil {
			if IsNotFound(err) {
				// message was deleted, forget it
				delete(guildInfo.AuthorizeMessegeIDs[e.ChannelID], e.MessageID)
				return
			}
			log.Println(err)
			return
		}
		member, err := bot.GuildMember(e.GuildID, e.UserID)
		if err != nil {
			log.Println(err)
			return
		}
		member.GuildID = e.GuildID
		authorized, err := bot.isAuthorized(member)
		if err != nil {
			log.Println(err)
			return
		}
		if !authorized {
			err = bot.addUnauthorizedMember(member)
			switch err.(type) {
			case *ErrAlreadyRegistered:
				bot.privMsgDiscord(e.UserID, "Already registered for verification")
			case nil:
				// no-op
			default:
				log.Println(err)
				return
			}
		}
	}
}

//#region Cleaning handlers

func (bot *UsosBot) handlerChannelDelete(session *discordgo.Session, e *discordgo.ChannelDelete) {
	log.Println("Channel deleted")
	guildInfo := bot.getGuildUsosInfo(e.GuildID)
	delete(guildInfo.LogChannelIDs, e.Channel.ID)
}

func (bot *UsosBot) handlerGuildMemberRemove(session *discordgo.Session, e *discordgo.GuildMemberRemove) {
	log.Println("Guild member removed")
	delete(bot.tokenMap, e.User.ID)
}

func (bot *UsosBot) handlerGuildRoleDelete(session *discordgo.Session, e *discordgo.GuildRoleDelete) {
	log.Println("Guild role deleted")
	guildInfo := bot.getGuildUsosInfo(e.GuildID)
	if e.RoleID == guildInfo.AuthorizeRoleID {
		guildInfo.AuthorizeRoleID = ""
	}
}

func (bot *UsosBot) handlerMessageDelete(session *discordgo.Session, e *discordgo.MessageDelete) {
	log.Println("Mesage deleted")
	guildInfo := bot.getGuildUsosInfo(e.GuildID)
	delete(guildInfo.AuthorizeMessegeIDs[e.ChannelID], e.Message.ID)
}

func (bot *UsosBot) handlerGuildDelete(session *discordgo.Session, e *discordgo.GuildDelete) {
	log.Println("Guild deleted")
	delete(bot.guildUsosInfos, e.Guild.ID)
}

func (bot *UsosBot) handlerGuildCreate(session *discordgo.Session, e *discordgo.GuildCreate) {
	log.Println("Guild created")
	guildInfo := bot.getGuildUsosInfo(e.Guild.ID)

	// clean deleted authorize message ids
	for channelID, messageMap := range guildInfo.AuthorizeMessegeIDs {
		for messageID := range messageMap {
			_, err := bot.ChannelMessage(channelID, messageID)
			if err != nil {
				if IsNotFound(err) {
					delete(messageMap, messageID)
				} else {
					log.Println(err)
				}
			}
		}
		if len(messageMap) == 0 {
			delete(guildInfo.AuthorizeMessegeIDs, channelID)
		}
	}

	// clean authorize role
	if guildInfo.AuthorizeRoleID != "" {
		_, err := bot.guildRole(e.Guild.ID, guildInfo.AuthorizeRoleID)
		if err != nil {
			if IsNotFound(err) {
				guildInfo.AuthorizeRoleID = ""
			} else {
				log.Println(err)
			}
		}
	}

	// clean log channels
	for logChannelID := range guildInfo.LogChannelIDs {
		_, err := bot.Channel(logChannelID)
		if err != nil {
			if IsNotFound(err) {
				delete(guildInfo.LogChannelIDs, logChannelID)
			} else {
				log.Println(err)
			}
		}
	}

	log.Println("Guild cleaned")

}

//#endregion
