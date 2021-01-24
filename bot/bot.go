package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/Ogurczak/discord-usos-auth/usos"

	"github.com/bwmarrin/discordgo"
)

type requestTokenGuildPair struct {
	RequestToken *usos.RequestToken
	GuildID      string
}

var (
	tokenMap               = make(map[string]*requestTokenGuildPair)
	guildAuthorizeRolesMap = make(map[string]string)
)

var (
	// LogChannelID is the id of a channel for sending successful usos login harvested information
	LogChannelID string
	// LogUserID is used in the same way as LogChannelID, if it is not defined
	LogUserID string
	// UserRolesFunction is used to determine additional roles given upon authorization
	UserRolesFunction func(*usos.User) ([]*discordgo.Role, error)
	// UsosUserFilter is used to filter which users to authorize, every authorized user passes on default
	UsosUserFilter func(*usos.User) (bool, error) = func(*usos.User) (bool, error) { return true, nil }
)

// New creates a new discord session
func New(Token string) (*discordgo.Session, error) {
	dg, err := discordgo.New(Token)
	if err != nil {
		return nil, err
	}

	dg.Identify.Intents = discordgo.MakeIntent(
		discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages |
			discordgo.IntentsGuildMembers | discordgo.IntentsGuildPresences)

	dg.AddHandler(guildMemberAddHandler)
	dg.AddHandler(guildCreateHandler)
	dg.AddHandler(guildMemberUpdateHandler)
	dg.AddHandler(messageCreateHandler)
	dg.AddHandler(readyHandler)

	return dg, err
}

func addUnauthorizedMember(s *discordgo.Session, m *discordgo.Member) error {
	if tokenMap[m.User.ID] != nil {
		return nil
	}
	token, err := usos.NewRequestToken()
	if err != nil {
		return err
	}
	tokenMap[m.User.ID] = &requestTokenGuildPair{
		RequestToken: token,
		GuildID:      m.GuildID}

	err = sendAuthorizationInstructions(s, m, token.AuthorizationURL)
	if err != nil {
		return err
	}
	return nil
}

func authorizeMember(s *discordgo.Session, member *discordgo.Member, usosUser *usos.User) error {
	var roles []*discordgo.Role

	var authorizedRole *discordgo.Role
	authorizedRoleID, err := getAuthorizeRoleID(s, member.GuildID)
	if err != nil {
		return err
	}

	if authorizedRoleID == "" {
		authorizedRole, err = s.GuildRoleCreate(member.GuildID)
		if err != nil {
			return err
		}
		authorizedRole, err = s.GuildRoleEdit(member.GuildID, authorizedRole.ID, "authorized", 0, false, 0, true)
		if err != nil {
			return err
		}
	} else {
		roles, err := s.GuildRoles(member.GuildID)
		if err != nil {
			return err
		}
		for _, role := range roles {
			if role.ID == authorizedRoleID {
				authorizedRole = role
				break
			}
		}
	}

	if UserRolesFunction != nil {
		roles, err = UserRolesFunction(usosUser)
		if err != nil {
			return err
		}
	}
	roles = append(roles, authorizedRole)

	for _, role := range roles {
		err = s.GuildMemberRoleAdd(member.GuildID, member.User.ID, role.ID)
	}

	return err
}

func feedbackDiscord(s *discordgo.Session, user *discordgo.User, text string) error {
	channel, err := s.UserChannelCreate(user.ID)
	if err != nil {
		log.Println(err)
	}

	_, err = s.ChannelMessageSend(channel.ID, text)
	if err != nil {
		log.Println(err)
	}
	return nil
}

func logDiscord(s *discordgo.Session, text string) error {
	var channelID string
	if LogChannelID == "" {
		if LogUserID == "" {
			return nil
		}
		channel, err := s.UserChannelCreate(LogUserID)
		channelID = channel.ID
		if err != nil {
			log.Println(err)
		}
	} else {
		channelID = LogChannelID
	}
	_, err := s.ChannelMessageSend(channelID, text)
	if err != nil {
		log.Println(err)
	}
	return nil
}

func sendAuthorizationInstructions(s *discordgo.Session, member *discordgo.Member, tokenURL *url.URL) error {
	time.Sleep(time.Second)
	channel, err := s.UserChannelCreate(member.User.ID)
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
				After that send me the authorization verifier.`, tokenURL),
				Inline: true,
			},
		},
	}
	_, err = s.ChannelMessageSendEmbed(channel.ID, msg)
	return err
}

func getAuthorizeRoleID(s *discordgo.Session, GuildID string) (string, error) {
	if guildAuthorizeRolesMap[GuildID] != "" {
		return guildAuthorizeRolesMap[GuildID], nil
	}

	roles, err := s.GuildRoles(GuildID)
	if err != nil {
		return "", err
	}
	for _, role := range roles {
		if role.Name == "authorized" {
			guildAuthorizeRolesMap[GuildID] = role.ID
			return role.ID, nil
		}
	}
	return "", nil
}

func isAuthorized(s *discordgo.Session, member *discordgo.Member) (bool, error) {
	authorizeRoleID, err := getAuthorizeRoleID(s, member.GuildID)
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

func scanGuild(s *discordgo.Session, guildID string) error {
	members, err := s.GuildMembers(guildID, "", 1000)
	if err != nil {
		return err
	}
	for _, member := range members {
		member.GuildID = guildID
		authorized, err := isAuthorized(s, member)
		if err != nil {
			return err
		}

		if !authorized && !member.User.Bot {
			err := addUnauthorizedMember(s, member)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// ########## HANDLERS ##############

func guildMemberUpdateHandler(s *discordgo.Session, e *discordgo.GuildMemberUpdate) {
	if e.User.ID == s.State.User.ID {
		return
	}
	log.Println("Member update")
	authorized, err := isAuthorized(s, e.Member)
	if err != nil {
		log.Println(err)
	}

	if !authorized {
		err := addUnauthorizedMember(s, e.Member)
		if err != nil {
			log.Println(err)
		}
	}
}

func messageCreateHandler(s *discordgo.Session, e *discordgo.MessageCreate) {
	if e.Author.ID == s.State.User.ID {
		return
	}
	log.Println("Message created")
	tokenGuilIDPair := tokenMap[e.Author.ID]
	if tokenGuilIDPair == nil {
		return
	}

	verifier := strings.Trim(e.Content, "\t \n")
	accessToken, err := tokenGuilIDPair.RequestToken.GetAccessToken(verifier)
	if err != nil {
		log.Println(err)
		return
	}

	usosUser, err := usos.NewUsosUser(accessToken)
	if err != nil {
		log.Println(err)
		return
	}
	passed, err := UsosUserFilter(usosUser)
	if err != nil {
		log.Println(err)
		return
	}
	if !passed {
		err = feedbackDiscord(s, e.Author, "You do not meet the requirements. Consult server administrators for details")
		if err != nil {
			log.Println(err)
		}
		return
	}

	member, err := s.GuildMember(tokenGuilIDPair.GuildID, e.Author.ID)
	if err != nil {
		log.Println(err)
		return
	}
	member.GuildID = tokenGuilIDPair.GuildID // because for some reason its empty (?)

	err = authorizeMember(s, member, usosUser)
	if err != nil {
		log.Println(err)
		return
	}

	message, err := json.MarshalIndent(usosUser, "", "    ")
	if err != nil {
		log.Println(err)
		return
	}
	err = logDiscord(s, fmt.Sprintf("%s's authorization data:\n```json\n%s\n```", e.Author.Username, message))
	if err != nil {
		log.Println(err)
		return
	}
	err = feedbackDiscord(s, e.Author, "Authorization complete")
	if err != nil {
		log.Println(err)
		return
	}
}

func guildMemberAddHandler(s *discordgo.Session, e *discordgo.GuildMemberAdd) {
	log.Println("Member added")
	authorized, err := isAuthorized(s, e.Member)
	if err != nil {
		log.Println(err)
	}

	if !authorized {
		err := addUnauthorizedMember(s, e.Member)
		if err != nil {
			log.Println(err)
		}
	}
}

func guildCreateHandler(s *discordgo.Session, e *discordgo.GuildCreate) {
	log.Println("Guild created")
	for _, member := range e.Members {
		authorized, err := isAuthorized(s, member)
		if err != nil {
			log.Println(err)
		}

		if !authorized {
			err := addUnauthorizedMember(s, member)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func readyHandler(s *discordgo.Session, e *discordgo.Ready) {
	log.Println("Ready")
	for _, guild := range e.Guilds {
		scanGuild(s, guild.ID)
	}
}
