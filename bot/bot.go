package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/Ogurczak/discord-usos-auth/usos"
	"github.com/Ogurczak/discord-usos-auth/utils"

	"github.com/bwmarrin/discordgo"
)

type requestTokenGuildPair struct {
	RequestToken *usos.RequestToken
	GuildID      string
}

// UsosBot represents a session of usos authorization bot
type UsosBot struct {
	*discordgo.Session

	tokenMap               map[string]*requestTokenGuildPair
	guildAuthorizeRolesMap map[string]string
	authorizeMessegeIDList []string
	logChannelIDMap        map[string]map[string]bool
}

// New creates a new session of usos authorization bot
func New(Token string) (*UsosBot, error) {
	session, err := discordgo.New(Token)
	if err != nil {
		return nil, err
	}

	session.Identify.Intents = discordgo.MakeIntent(
		discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages |
			discordgo.IntentsGuildMembers | discordgo.IntentsGuildPresences |
			discordgo.IntentsGuildMessages | discordgo.IntentsGuildMessageReactions)

	bot := &UsosBot{
		Session: session,

		tokenMap:               make(map[string]*requestTokenGuildPair),
		guildAuthorizeRolesMap: make(map[string]string),
		authorizeMessegeIDList: make([]string, 0),
		logChannelIDMap:        make(map[string]map[string]bool),
	}

	bot.AddHandler(bot.messageCreateHandler)
	bot.AddHandler(bot.readyHandler)
	bot.AddHandler(bot.reactionAddHandler)

	return bot, err
}

// registerUnauthorizedMember creates a new oauth token bound to the given member
// and sends authorization instructions to that member
func (bot *UsosBot) registerUnauthorizedMember(m *discordgo.Member) error {
	if bot.tokenMap[m.User.ID] != nil {
		return nil
	}
	token, err := usos.NewRequestToken()
	if err != nil {
		return err
	}
	bot.tokenMap[m.User.ID] = &requestTokenGuildPair{
		RequestToken: token,
		GuildID:      m.GuildID}

	err = bot.sendAuthorizationInstructions(m, token.AuthorizationURL)
	if err != nil {
		return err
	}
	return nil
}

// createAuthorizeRole creates an authorize role in the given guild
func (bot *UsosBot) createAuthorizeRole(GuildID string) (*discordgo.Role, error) {
	authorizeRole, err := bot.GuildRoleCreate(GuildID)
	if err != nil {
		return nil, err
	}
	authorizeRole, err = bot.GuildRoleEdit(GuildID, authorizeRole.ID, "authorized", 0, false, 0, true)
	if err != nil {
		return nil, err
	}
	return authorizeRole, nil
}

// authorizeMember authorizes the given member and gives him additional roles based on bot's roles function
func (bot *UsosBot) authorizeMember(member *discordgo.Member, usosUser *usos.User) error {
	var roles []*discordgo.Role

	var authorizeRole *discordgo.Role
	authorizeRoleID, err := bot.getAuthorizeRoleID(member.GuildID)
	if err != nil {
		return err
	}

	if authorizeRoleID == "" {
		authorizeRole, err = bot.createAuthorizeRole(member.GuildID)
		if err != nil {
			return err
		}
	} else {
		roles, err := bot.GuildRoles(member.GuildID)
		if err != nil {
			return err
		}
		for _, role := range roles {
			if role.ID == authorizeRoleID {
				authorizeRole = role
				break
			}
		}
	}

	// TODO: automatic roles assignment
	// if bot.UserRolesFunction != nil && usosUser != nil {
	// 	roles, err = bot.UserRolesFunction(usosUser)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	roles = append(roles, authorizeRole)

	for _, role := range roles {
		err = bot.GuildMemberRoleAdd(member.GuildID, member.User.ID, role.ID)
	}

	return err
}

// privMsgDiscord sends a private message to a user with the given text
func (bot *UsosBot) privMsgDiscord(userID string, text string) error {
	channel, err := bot.UserChannelCreate(userID)
	if err != nil {
		log.Println(err)
	}

	_, err = bot.ChannelMessageSend(channel.ID, text)
	if err != nil {
		log.Println(err)
	}
	return nil
}

// logDiscord logs a message to all log channels of a guild
func (bot *UsosBot) logDiscord(guildID string, text string) error {
	for channelID := range bot.logChannelIDMap[guildID] {
		_, err := bot.ChannelMessageSend(channelID, text)
		if err != nil {
			return err
		}
	}
	return nil
}

// sendAuthorizationInstructions sends instructions on authorization to the given member
func (bot *UsosBot) sendAuthorizationInstructions(member *discordgo.Member, tokenURL *url.URL) error {
	time.Sleep(time.Second)
	channel, err := bot.UserChannelCreate(member.User.ID)
	if err != nil {
		return err
	}
	msg := &discordgo.MessageEmbed{
		Type:  discordgo.EmbedTypeRich,
		URL:   tokenURL.String(),
		Title: "USOS Authorization required",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: "You must authorize yourself before proceeding on this server.",
				Value: fmt.Sprintf(`In order to do that visit [this page](%s) and authorize.
				After that send me the authorization verifier using the %s command.`, tokenURL, utils.DiscordCodeSpan("!usos verify -c <verifier>")),
				Inline: true,
			},
		},
	}
	_, err = bot.ChannelMessageSendEmbed(channel.ID, msg)
	return err
}

// getAuthorizeRoleID return authorization role id of the given guild
func (bot *UsosBot) getAuthorizeRoleID(GuildID string) (string, error) {
	if bot.guildAuthorizeRolesMap[GuildID] != "" {
		return bot.guildAuthorizeRolesMap[GuildID], nil
	}

	roles, err := bot.GuildRoles(GuildID)
	if err != nil {
		return "", err
	}
	for _, role := range roles {
		if role.Name == "authorized" {
			bot.guildAuthorizeRolesMap[GuildID] = role.ID
			return role.ID, nil
		}
	}
	return "", nil
}

// isAuthorized checks if a given member is authorized on his guild
func (bot *UsosBot) isAuthorized(member *discordgo.Member) (bool, error) {
	authorizeRoleID, err := bot.getAuthorizeRoleID(member.GuildID)
	if err != nil {
		return false, err
	}
	if authorizeRoleID == "" {
		return false, nil
	}

	for _, roleID := range member.Roles {
		if roleID == authorizeRoleID {
			return true, nil
		}
	}
	return false, nil
}

// spawnAuthorizeMessage spawns a messege in a given channel
// which adds unauthorized members if reacted to
func (bot *UsosBot) spawnAuthorizeMessage(ChannelID string, prompt string) error {
	msg, err := bot.ChannelMessageSend(ChannelID, prompt)
	if err != nil {
		return err
	}
	bot.authorizeMessegeIDList = append(bot.authorizeMessegeIDList, msg.ID)
	return nil
}

// finalizeAuthorization finalizes the user's authorization using the given verifier
func (bot *UsosBot) finalizeAuthorization(user *discordgo.User, verifier string) error {
	tokenGuilIDPair := bot.tokenMap[user.ID]
	if tokenGuilIDPair == nil {
		return newErrUnregisteredUnauthorizedUser(user.ID)
	}

	accessToken, err := tokenGuilIDPair.RequestToken.GetAccessToken(verifier)
	if err != nil {
		return err
	}

	usosUser, err := usos.NewUsosUser(accessToken)
	switch err.(type) {
	case *usos.ErrUnableToCall:
		return newErrWrongVerifier(err.(*usos.ErrUnableToCall), user.ID, tokenGuilIDPair, verifier)
	default:
		if err != nil {
			return err
		}
	}
	// TODO: usos user filter
	// if bot.UsosUserFilter != nil {
	// 	passed, err := bot.UsosUserFilter(usosUser)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if !passed {
	// 		return newErrFilteredOut(user.ID)
	// 	}
	// }

	member, err := bot.GuildMember(tokenGuilIDPair.GuildID, user.ID)
	if err != nil {
		return err
	}
	member.GuildID = tokenGuilIDPair.GuildID // because for some reason its empty (?)

	err = bot.authorizeMember(member, usosUser)
	if err != nil {
		return err
	}
	delete(bot.tokenMap, user.ID)

	message, err := json.MarshalIndent(usosUser, "", "    ")
	if err != nil {
		return err
	}
	err = bot.logDiscord(tokenGuilIDPair.GuildID, fmt.Sprintf("%s's authorization data:\n```json\n%s\n```", user.Username, message))
	if err != nil {
		return err
	}
	err = bot.privMsgDiscord(user.ID, "Authorization complete")
	if err != nil {
		return err
	}
	return nil
}

// addLogChannel adds a channel to log to authorization data from the guild
func (bot *UsosBot) addLogChannel(guildID string, channelID string) error {
	if bot.logChannelIDMap[guildID] == nil {
		bot.logChannelIDMap[guildID] = make(map[string]bool)
	}
	if bot.logChannelIDMap[guildID][channelID] {
		return newErrLogChannelAlreadyAdded(channelID)
	}
	bot.logChannelIDMap[guildID][channelID] = true
	return nil
}

// removeLogChannel removes a log channel
func (bot *UsosBot) removeLogChannel(guildID string, channelID string) error {
	if !bot.logChannelIDMap[guildID][channelID] {
		return newErrLogChannelNotAdded(channelID)
	}
	delete(bot.logChannelIDMap[guildID], channelID)
	return nil
}

//#region DEPRECATED

// scanGuild finds all unauthorized members of a guild and sends them authorization instructions
// MAY RESULT IN BOT BEEING BANNED IF TOO MANY USERS ARE MESSAGED AT ONCE!
func (bot *UsosBot) scanGuild(guildID string) error {
	members, err := bot.GuildMembers(guildID, "", 1000)
	if err != nil {
		return err
	}
	for _, member := range members {
		member.GuildID = guildID
		authorized, err := bot.isAuthorized(member)
		if err != nil {
			return err
		}

		if !authorized && !member.User.Bot {
			err := bot.registerUnauthorizedMember(member)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//#endregion
