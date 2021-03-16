package bot

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var msg = "```go\n" +
	`package bot

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
	TokenMap       map[string]*requestTokenGuildPair
	GuildUsosInfos map[string]*guildUsosInfo
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
}` + "```"

var msgsWant = []string{"```go\npackage bot\n\nimport (\n    \"encoding/json\"\n    \"io\"\n    \"log\"\n\n    \"github.com/Ogurczak/discord-usos-auth/usos\"\n\n    \"github.com/bwmarrin/discordgo\"\n)\n\ntype requestTokenGuildPair struct {\n    GuildID      string\n    RequestToken *usos.RequestToken\n}\n\ntype guildUsosInfo struct {\n    AuthorizeRoleID     string\n    Filters             []*usos.User\n    LogChannelIDs       map[string]bool\n    AuthorizeMessegeIDs map[string]map[string]bool // maps channelID to a set of message IDs\n}\n\n// UsosBot represents a session of usos authorization bot\ntype UsosBot struct {\n    *discordgo.Session\n\n    tokenMap       map[string]*requestTokenGuildPair // maps user id to their auth token\n    guildUsosInfos map[string]*guildUsosInfo         // maps guild id to its info\n}\n\n// New creates a new session of usos authorization bot\nfunc New(Token string) (*UsosBot, error) {\n    session, err := discordgo.New(Token)\n    if err != nil {\n        return nil, err\n    }\n\n    session.Identify.Intents = discordgo.MakeIntent(\n        discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages |\n            discordgo.IntentsGuildMembers | discordgo.IntentsGuildMessageReactions |\n            discordgo.IntentsGuilds)\n\n    bot := &UsosBot{\n        Session: session,\n\n        tokenMap:       make(map[string]*requestTokenGuildPair),\n        guildUsosInfos: make(map[string]*guildUsosInfo),\n    }\n\n    bot.AddHandler(bot.handlerMessageCreate)\n    bot.AddHandler(bot.handlerReady)\n    bot.AddHandler(bot.handlerReactionAdd)\n\n    bot.AddHandler(bot.handlerGuildMemberRemove)\n    bot.AddHandler(bot.handlerMessageDelete)\n    bot.AddHandler(bot.handlerGuildRoleDelete)\n    bot.AddHandler(bot.handlerChannelDelete)\n    bot.AddHandler(bot.handlerGuildDelete)\n    bot.AddHandler(bot.handlerGuildCreate)\n\n    return bot, err\n}\n\n// getGuildUsosInfo gives access to guild's usos info\nfunc (bot *UsosBot) getGuildUsosInfo(guildID string) *guildUsosInfo {\n    if bot.guildUsosInfos[guildID] == nil {```",
	"```go\n        bot.guildUsosInfos[guildID] = &guildUsosInfo{\n            LogChannelIDs:       make(map[string]bool),\n            Filters:             make([]*usos.User, 0),\n            AuthorizeMessegeIDs: make(map[string]map[string]bool),\n        }\n        return bot.guildUsosInfos[guildID]\n    }\n    return bot.guildUsosInfos[guildID]\n}\n\n// getLogChannel returns log channel if the given channel id is still valid or nil pointer otherwise\nfunc (bot *UsosBot) getLogChannel(guildID string, channelID string) (*discordgo.Channel, error) {\n    guildInfo := bot.getGuildUsosInfo(guildID)\n    if !guildInfo.LogChannelIDs[channelID] {\n        return nil, newErrLogChannelNotFound(channelID, guildID)\n    }\n    channel, err := bot.Channel(channelID)\n    if err != nil {\n        if IsNotFound(err) {\n            // channel was deleted, remove it from log channels autmatically\n            delete(guildInfo.LogChannelIDs, channelID)\n            return nil, newErrChannelNotFound(err, channelID)\n        }\n        return nil, err\n    }\n    return channel, nil\n}\n\n// privMsgDiscord sends a private message to a user with the given text\nfunc (bot *UsosBot) privMsgDiscord(userID string, text string) error {\n    channel, err := bot.UserChannelCreate(userID)\n    if err != nil {\n        log.Println(err)\n    }\n\n    _, err = bot.ChannelMessageSend(channel.ID, text)\n    if err != nil {\n        log.Println(err)\n    }\n    return nil\n}\n\ntype settings struct {\n    TokenMap       map[string]*requestTokenGuildPair\n    GuildUsosInfos map[string]*guildUsosInfo\n}\n\n// ExportSettings exports current bot settings on all servers to a json file\nfunc (bot *UsosBot) ExportSettings(w io.Writer) error {\n    stngs := settings{\n        TokenMap:       bot.tokenMap,\n        GuildUsosInfos: bot.guildUsosInfos,\n    }\n    err := json.NewEncoder(w).Encode(&stngs)\n\n    if err != nil {\n        return err\n    }\n\n    return nil\n}\n\n// ImportSettings imports bot settings on all servers from a json file```",
	"```go\n// (overrides current settings)\nfunc (bot *UsosBot) ImportSettings(r io.Reader) error {\n    stngs := settings{}\n\n    err := json.NewDecoder(r).Decode(&stngs)\n    if err != nil {\n        return err\n    }\n\n    bot.guildUsosInfos = stngs.GuildUsosInfos\n    bot.tokenMap = stngs.TokenMap\n\n    return nil\n}```"}

func TestFragmentMsg(t *testing.T) {

	msgs := fragmentMsg(&msg)
	for i, msg := range msgs {
		fmt.Printf("Nr: %2d/%2d,\tLen: %4d\n##############################\n%s\n", i+1, len(msgs), len(msg), msg)
	}
	assert.ElementsMatch(t, msgs, msgsWant)
}
