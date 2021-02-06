package bot

import (
	"log"
	"strings"

	"github.com/Ogurczak/discord-usos-auth/usos"
	"github.com/bwmarrin/discordgo"
)

// messageCreateHandler handles messeges received by the bot
func messageCreateHandler(bot *UsosBot, e *discordgo.MessageCreate) {
	// ignore self
	if e.Author.ID == bot.State.User.ID {
		return
	}
	log.Println("Message created")

	trimmedMsg := strings.Trim(e.Content, "\n \t")

	switch trimmedMsg {
	case "!spawn-authorize-message":
		bot.spawnAuthorizeMessage(e.ChannelID)
	default:
		// ignore normal guild messages
		if e.GuildID != "" {
			return
		}

		err := bot.finalizeAuthorization(e.Author, trimmedMsg)
		if err != nil {
			switch err.(type) {
			case *ErrUnregisteredUnauthorizedUser:
				err = bot.privMsgDiscord(e.Author.ID, "You must first register for authorization by adding reaction to the bot's message on a server.")
				if err != nil {
					log.Println(err)
				}
			case *ErrFilteredOut:
				err = bot.privMsgDiscord(e.Author.ID, "You do not meet the requirements. Consult server administrators for details.")
				if err != nil {
					log.Println(err)
				}
			case *usos.ErrUnableToCall:
				err = bot.privMsgDiscord(e.Author.ID, "Cannot make call for required information to usos-api. Probably the verifier is wrong")
				if err != nil {
					log.Println(err)
				}

			default:
				log.Println(err)
			}
		}

	}
}

// readyHandler indicates that the bot is ready
func readyHandler(bot *UsosBot, e *discordgo.Ready) {
	log.Println("Ready")
	// for _, guild := range e.Guilds {
	// 	err := scanGuild(s, guild.ID)
	// 	if err != nil {
	// 		log.Println(err)
	// 	}
	// }
}

// reactionAddHandler handles reactions added to bot's messages
func reactionAddHandler(bot *UsosBot, e *discordgo.MessageReactionAdd) {
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
