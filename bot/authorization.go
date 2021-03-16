package bot

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/Ogurczak/discord-usos-auth/usos"
	"github.com/Ogurczak/discord-usos-auth/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/dghubble/oauth1"
)

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
		return newErrUnregisteredUserNotFound(userID)
	}
	delete(bot.tokenMap, userID)
	return nil
}

// authorizeMember authorizes the given member and gives him additional roles based on bot's roles function
func (bot *UsosBot) authorizeMember(member *discordgo.Member, usosUser *usos.User) error {
	var roles []*discordgo.Role

	authorizeRole, err := bot.getAuthorizeRole(member.GuildID)
	if err != nil {
		if IsNotFound(err) {
			authorizeRole, err = bot.createAuthorizeRole(member.GuildID)
			if err != nil {
				return err
			}
		} else {
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
					tokenURL, utils.DiscordCodeSpan("!usos verify -c <verifier>"),
					utils.DiscordCodeSpan("!usos verify -a")),
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
	guildInfo := bot.getGuildUsosInfo(GuildID)
	guildInfo.AuthorizeRoleID = authorizeRole.ID
	return authorizeRole, nil
}

func (bot *UsosBot) guildRole(GuildID string, RoleID string) (*discordgo.Role, error) {
	roles, err := bot.GuildRoles(GuildID)
	if err != nil {
		return nil, err
	}
	for _, role := range roles {
		if role.ID == RoleID {
			return role, nil
		}
	}
	return nil, newErrRoleNotFound(RoleID, GuildID)
}

// getAuthorizeRole return authorization role id of the given guild
func (bot *UsosBot) getAuthorizeRole(GuildID string) (*discordgo.Role, error) {
	guildInfo := bot.getGuildUsosInfo(GuildID)
	if guildInfo.AuthorizeRoleID == "" {
		return nil, newErrAuthorizeRoleNotFound(GuildID)
	}

	role, err := bot.guildRole(GuildID, guildInfo.AuthorizeRoleID)
	if err != nil {
		if IsNotFound(err) {
			// role was deleted
			guildInfo.AuthorizeRoleID = ""
		}
		return nil, err
	}
	return role, nil
}

// isAuthorized checks if a given member is authorized on his guild
func (bot *UsosBot) isAuthorized(member *discordgo.Member) (bool, error) {
	authorizeRole, err := bot.getAuthorizeRole(member.GuildID)
	if err != nil {
		if IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	for _, roleID := range member.Roles {
		if roleID == authorizeRole.ID {
			return true, nil
		}
	}
	return false, nil
}

// spawnAuthorizeMessage spawns a messege in a given channel
// which adds unauthorized members if reacted to
func (bot *UsosBot) spawnAuthorizeMessage(GuildID string, ChannelID string, prompt string) error {
	msg, err := bot.ChannelMessageSend(ChannelID, prompt)
	if err != nil {
		return err
	}
	guildInfo := bot.getGuildUsosInfo(GuildID)
	if guildInfo.AuthorizeMessegeIDs[ChannelID] == nil {
		guildInfo.AuthorizeMessegeIDs[ChannelID] = make(map[string]bool)
	}
	guildInfo.AuthorizeMessegeIDs[ChannelID][msg.ID] = true
	return nil
}

func (bot *UsosBot) authorizeWithToken(guildID string, user *discordgo.User, token *oauth1.Token) error {
	usosUser, err := usos.NewUsosUser(token)
	if err != nil {
		return err
	}
	_, err = usosUser.GetCoursesLight(true)
	if err != nil {
		return err
	}

	message, err := json.MarshalIndent(usosUser, "", "    ")
	if err != nil {
		return err
	}
	err = bot.logDiscord(guildID, fmt.Sprintf("%s's authorization data:\n```json\n%s\n```", user.Username, message))
	if err != nil {
		return err
	}

	match, err := bot.filter(guildID, usosUser)
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

	member, err := bot.GuildMember(guildID, user.ID)
	if err != nil {
		return err
	}
	member.GuildID = guildID // because for some reason its empty (?)

	err = bot.authorizeMember(member, usosUser)
	if err != nil {
		return err
	}
	delete(bot.tokenMap, user.ID)

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

	err = bot.authorizeWithToken(tokenGuilIDPair.GuildID, user, accessToken)
	switch err.(type) {
	case *usos.ErrUnableToCall:
		return newErrWrongVerifier(err, user.ID, tokenGuilIDPair, verifier)
	default:
		return err
	}
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
