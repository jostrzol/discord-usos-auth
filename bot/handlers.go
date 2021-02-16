package bot

import (
	"log"
	"strings"

	"github.com/Ogurczak/discord-usos-auth/bot/commands"
	"github.com/bwmarrin/discordgo"
)

// messageCreateHandler handles messeges received by the bot
func (bot *UsosBot) messageCreateHandler(session *discordgo.Session, e *discordgo.MessageCreate) {
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
	_, isParseError := err.(*commands.ErrParse)
	if err != nil && !isParseError {
		log.Println(err)
		return
	}

	if !parser.ParsedHelp() && err == nil {
		err = parser.Handle(e)
		_, isHandleError := err.(*commands.ErrInCommandHandler)
		_, isScopeError := err.(*commands.ErrCommandInWrongScope)
		if err != nil && !isHandleError && !isScopeError {
			log.Println(err)
			return
		}
	}
}

// readyHandler indicates that the bot is ready
func (bot *UsosBot) readyHandler(session *discordgo.Session, e *discordgo.Ready) {
	log.Println("Ready")
}

// reactionAddHandler handles reactions added to bot's messages
func (bot *UsosBot) reactionAddHandler(session *discordgo.Session, e *discordgo.MessageReactionAdd) {
	for _, id := range bot.authorizeMessegeIDList {
		if e.MessageID == id {
			member, err := bot.GuildMember(e.GuildID, e.UserID)
			if err != nil {
				log.Println(err)
			}
			member.GuildID = e.GuildID
			authorized, err := bot.isAuthorized(member)
			if err != nil {
				log.Println(err)
			}
			if !authorized {
				err = bot.registerUnauthorizedMember(member)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
}

//#region DEPRECATED

// guildMemberUpdateHandler checks if the member is authorized or not after the update
// and sends the authorization instructions if not
// NO LONGER USED DUE TO REACTION-BASED HANDLING
func guildMemberUpdateHandler(bot *UsosBot, e *discordgo.GuildMemberUpdate) {
	if e.User.ID == bot.State.User.ID {
		return
	}
	log.Println("Member update")
	authorized, err := bot.isAuthorized(e.Member)
	if err != nil {
		log.Println(err)
	}

	if !authorized {
		err := bot.registerUnauthorizedMember(e.Member)
		if err != nil {
			log.Println(err)
		}
	}
}

// guildMemberAddHandler checks if the member is authorized or not
// and sends the authorization instructions if not
// NO LONGER USED DUE TO REACTION-BASED HANDLING
func guildMemberAddHandler(bot *UsosBot, e *discordgo.GuildMemberAdd) {
	log.Println("Member added")
	authorized, err := bot.isAuthorized(e.Member)
	if err != nil {
		log.Println(err)
	}

	if !authorized {
		err := bot.registerUnauthorizedMember(e.Member)
		if err != nil {
			log.Println(err)
		}
	}
}

// guildCreateHandler performs a guildScan on a newly created guild
// NO LONGER USED DUE TO REACTION-BASED HANDLING
func guildCreateHandler(bot *UsosBot, e *discordgo.GuildCreate) {
	log.Println("Guild created")
	err := bot.scanGuild(e.Guild.ID)
	if err != nil {
		log.Println(err)
	}
}

//#endregion
