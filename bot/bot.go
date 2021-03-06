package bot

import (
	"encoding/json"
	"io"
	"log"

	"github.com/Ogurczak/discord-usos-auth/usos"

	"github.com/bwmarrin/discordgo"
)

type requestTokenGuildPair struct {
	GuildID      string
	RequestToken *usos.RequestToken
}

type guildUsosInfo struct {
	AuthorizeRoleID     string
	Filters             []*usos.User
	LogChannelIDs       map[string]bool
	AuthorizeMessegeIDs map[string]map[string]bool // maps channelID to a set of message IDs
}

// UsosBot represents a session of usos authorization bot
type UsosBot struct {
	*discordgo.Session

	tokenMap       map[string]*requestTokenGuildPair // maps user id to their auth token
	guildUsosInfos map[string]*guildUsosInfo         // maps guild id to its info
}

// New creates a new session of usos authorization bot
func New(Token string) (*UsosBot, error) {
	session, err := discordgo.New(Token)
	if err != nil {
		return nil, err
	}

	session.Identify.Intents = discordgo.MakeIntent(
		discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages |
			discordgo.IntentsGuildMembers | discordgo.IntentsGuildMessageReactions |
			discordgo.IntentsGuilds)

	bot := &UsosBot{
		Session: session,

		tokenMap:       make(map[string]*requestTokenGuildPair),
		guildUsosInfos: make(map[string]*guildUsosInfo),
	}

	bot.AddHandler(bot.handlerMessageCreate)
	bot.AddHandler(bot.handlerReady)
	bot.AddHandler(bot.handlerReactionAdd)

	bot.AddHandler(bot.handlerGuildMemberRemove)
	bot.AddHandler(bot.handlerMessageDelete)
	bot.AddHandler(bot.handlerGuildRoleDelete)
	bot.AddHandler(bot.handlerChannelDelete)
	bot.AddHandler(bot.handlerGuildDelete)
	bot.AddHandler(bot.handlerGuildCreate)

	return bot, err
}

// getGuildUsosInfo gives access to guild's usos info
func (bot *UsosBot) getGuildUsosInfo(guildID string) *guildUsosInfo {
	if bot.guildUsosInfos[guildID] == nil {
		bot.guildUsosInfos[guildID] = &guildUsosInfo{
			LogChannelIDs:       make(map[string]bool),
			Filters:             make([]*usos.User, 0),
			AuthorizeMessegeIDs: make(map[string]map[string]bool),
		}
		return bot.guildUsosInfos[guildID]
	}
	return bot.guildUsosInfos[guildID]
}

// getLogChannel returns log channel if the given channel id is still valid or nil pointer otherwise
func (bot *UsosBot) getLogChannel(guildID string, channelID string) (*discordgo.Channel, error) {
	guildInfo := bot.getGuildUsosInfo(guildID)
	if !guildInfo.LogChannelIDs[channelID] {
		return nil, newErrLogChannelNotFound(channelID, guildID)
	}
	channel, err := bot.Channel(channelID)
	if err != nil {
		if IsNotFound(err) {
			// channel was deleted, remove it from log channels autmatically
			delete(guildInfo.LogChannelIDs, channelID)
			return nil, newErrChannelNotFound(err, channelID)
		}
		return nil, err
	}
	return channel, nil
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

type settings struct {
	TokenMap       map[string]*requestTokenGuildPair `json:"tokenMap"`
	GuildUsosInfos map[string]*guildUsosInfo         `json:"guildUsosInfos"`
}

// ExportSettings exports current bot settings on all servers to a json file
func (bot *UsosBot) ExportSettings(w io.Writer) error {
	stngs := settings{
		TokenMap:       bot.tokenMap,
		GuildUsosInfos: bot.guildUsosInfos,
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

	bot.guildUsosInfos = stngs.GuildUsosInfos
	bot.tokenMap = stngs.TokenMap

	return nil
}
