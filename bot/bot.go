package bot

import (
	"encoding/json"
	"fmt"
	"io"
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

type guildUsosInfo struct {
	AuthorizeRoleID string
	LogChannelIDs   map[string]bool
	Filters         []*usos.User
}

// UsosBot represents a session of usos authorization bot
type UsosBot struct {
	*discordgo.Session

	tokenMap            map[string]*requestTokenGuildPair
	authorizeMessegeIDs map[string]bool
	guildUsosInfos      map[string]*guildUsosInfo
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

		tokenMap:            make(map[string]*requestTokenGuildPair),
		authorizeMessegeIDs: make(map[string]bool),
		guildUsosInfos:      make(map[string]*guildUsosInfo),
	}

	bot.AddHandler(bot.messageCreateHandler)
	bot.AddHandler(bot.readyHandler)
	bot.AddHandler(bot.reactionAddHandler)

	return bot, err
}

// getGuildUsosInfo gives access to guild's usos info
func (bot *UsosBot) getGuildUsosInfo(guildID string) *guildUsosInfo {
	if bot.guildUsosInfos[guildID] == nil {
		bot.guildUsosInfos[guildID] = &guildUsosInfo{
			LogChannelIDs: make(map[string]bool),
			Filters:       make([]*usos.User, 0),
		}
		return bot.guildUsosInfos[guildID]
	}
	return bot.guildUsosInfos[guildID]
}

// addUnauthorizedMember creates a new oauth token bound to the given member
// and sends authorization instructions to that member
func (bot *UsosBot) addUnauthorizedMember(m *discordgo.Member) error {
	if bot.tokenMap[m.User.ID] != nil {
		return newErrAlreadyRegistered(m.User.ID, bot.tokenMap[m.User.ID])
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

// removeUnauthorizedUser removes an user from authorization list
func (bot *UsosBot) removeUnauthorizedUser(userID string) error {
	if _, exists := bot.tokenMap[userID]; !exists {
		return newErrAlreadyUnregisteredUser(userID)
	}
	delete(bot.tokenMap, userID)
	return nil
}

// authorizeMember authorizes the given member and gives him additional roles based on bot's roles function
func (bot *UsosBot) authorizeMember(member *discordgo.Member, usosUser *usos.User) error {
	var roles []*discordgo.Role

	var authorizeRole *discordgo.Role
	authorizeRoleID, err := bot.getAuthorizeRoleID(member.GuildID)
	if err != nil {
		return err
	}

	guildRoles, err := bot.GuildRoles(member.GuildID)
	if err != nil {
		return err
	}
	for _, role := range guildRoles {
		if role.ID == authorizeRoleID {
			authorizeRole = role
			break
		}
	}

	if authorizeRole == nil {
		// previous auth role was deleted
		authorizeRole, err = bot.createAuthorizeRole(member.GuildID)
		if err != nil {
			return err
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
		if err != nil {
			return err
		}
	}

	return nil
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
	for channelID := range bot.getGuildUsosInfo(guildID).LogChannelIDs {
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
				After that send me the authorization verifier using the %s command.
				You can also abort the authorization process using the %s command.`,
					tokenURL, utils.DiscordCodeSpan("!usos verify -c <verifier>"), utils.DiscordCodeSpan("!usos verify -a")),
				Inline: true,
			},
		},
	}
	_, err = bot.ChannelMessageSendEmbed(channel.ID, msg)
	return err
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

// getAuthorizeRoleID return authorization role id of the given guild
func (bot *UsosBot) getAuthorizeRoleID(GuildID string) (string, error) {
	guildInfo := bot.getGuildUsosInfo(GuildID)
	if guildInfo.AuthorizeRoleID != "" {
		return guildInfo.AuthorizeRoleID, nil
	}

	role, err := bot.createAuthorizeRole(GuildID)
	if err != nil {
		return "", err
	}

	guildInfo.AuthorizeRoleID = role.ID
	return role.ID, nil
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
	bot.authorizeMessegeIDs[msg.ID] = true
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
		return newErrWrongVerifier(err, user.ID, tokenGuilIDPair, verifier)
	}

	usosUser, err := usos.NewUsosUser(accessToken)
	switch err.(type) {
	case *usos.ErrUnableToCall:
		return newErrWrongVerifier(err, user.ID, tokenGuilIDPair, verifier)
	case nil:
		//no-op
	default:
		return err

	}

	message, err := json.MarshalIndent(usosUser, "", "    ")
	if err != nil {
		return err
	}
	err = bot.logDiscord(tokenGuilIDPair.GuildID, fmt.Sprintf("%s's authorization data:\n```json\n%s\n```", user.Username, message))
	if err != nil {
		return err
	}

	match, err := bot.filter(tokenGuilIDPair.GuildID, usosUser)
	if err != nil {
		return err
	}
	if !match {
		err := bot.removeUnauthorizedUser(user.ID)
		if err != nil {
			return err
		}
		return newErrFilteredOut(user.ID)
	}

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

	err = bot.privMsgDiscord(user.ID, "Authorization complete")
	if err != nil {
		return err
	}
	return nil
}

// addLogChannel adds a channel to log to authorization data from the guild
func (bot *UsosBot) addLogChannel(guildID string, channelID string) error {
	guildInfo := bot.getGuildUsosInfo(guildID)
	if guildInfo.LogChannelIDs[channelID] {
		return newErrLogChannelAlreadyAdded(channelID)
	}

	guildChannels, err := bot.GuildChannels(guildID)
	if err != nil {
		return err
	}

	// only add this guild's channels
	for _, guildChannel := range guildChannels {
		if guildChannel.ID == channelID {
			guildInfo.LogChannelIDs[channelID] = true
			return nil
		}
	}

	return newErrLogChannelNotInGuild(channelID, guildID)
}

// removeLogChannel removes a log channel
func (bot *UsosBot) removeLogChannel(guildID string, channelID string) error {
	guildInfo := bot.getGuildUsosInfo(guildID)
	if !guildInfo.LogChannelIDs[channelID] {
		return newErrLogChannelNotAdded(channelID)
	}
	delete(guildInfo.LogChannelIDs, channelID)
	return nil
}

// filter checks if an usos user passes at least one of the set filters
func (bot *UsosBot) filter(guildID string, user *usos.User) (bool, error) {
	guildInfo := bot.getGuildUsosInfo(guildID)
	if len(guildInfo.Filters) == 0 {
		return true, nil
	}
	for _, filter := range guildInfo.Filters {
		match, err := utils.FilterRec(filter, user)
		if err != nil {
			return false, err
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}

type settings struct {
	TokenMap            map[string]*requestTokenGuildPair `json:"tokenMap"`
	GuildUsosInfos      map[string]*guildUsosInfo         `json:"guildUsosInfos"`
	AuthorizeMessageIDs map[string]bool                   `json:"authorizeMessageIDs"`
}

// ExportSettings exports current bot settings on all servers to a json file
func (bot *UsosBot) ExportSettings(w io.Writer) error {
	stngs := settings{
		TokenMap:            bot.tokenMap,
		GuildUsosInfos:      bot.guildUsosInfos,
		AuthorizeMessageIDs: bot.authorizeMessegeIDs,
	}
	err := json.NewEncoder(w).Encode(&stngs)

	if err != nil {
		return err
	}

	return nil
}

// ImportSettings imports bot settings on all servers from a json file
// (overrides current settings)
func (bot *UsosBot) ImportSettings(r io.Reader) error {
	stngs := settings{}

	err := json.NewDecoder(r).Decode(&stngs)
	if err != nil {
		return err
	}

	bot.authorizeMessegeIDs = stngs.AuthorizeMessageIDs
	bot.guildUsosInfos = stngs.GuildUsosInfos
	bot.tokenMap = stngs.TokenMap

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
			err := bot.addUnauthorizedMember(member)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//#endregion
