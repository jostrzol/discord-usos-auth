package bot

import (
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/Ogurczak/discord-usos-auth/usos"
)

func TestExportImportSettings(t *testing.T) {
	url, err := url.Parse("https://www.example.com/example.html")
	if err != nil {
		t.Error(err)
	}

	bot := &UsosBot{
		tokenMap: map[string]*requestTokenGuildPair{
			"userID": {
				GuildID: "guildID",
				RequestToken: &usos.RequestToken{
					Token:            "token",
					Secret:           "secret",
					AuthorizationURL: url,
				},
			},
		},
		guildUsosInfos: map[string]*guildUsosInfo{
			"guilID": {
				AuthorizeMessegeIDs: map[string]map[string]bool{
					"channelID": {
						"messageID1": true,
						"messageID2": true,
					},
				},
				AuthorizeRoleID: "authorizeRoleID",
				LogChannelIDs: map[string]bool{
					"logChannelID": true,
				},
				Filters: []*usos.User{
					{ID: "ID",
						FirstName: "Witold",
						LastName:  "Wysota",
						Programmes: []*usos.Programme{
							{ID: "ID",
								Name:        "name",
								Description: "Description"},
						},
					},
				},
			},
		},
	}

	dir := t.TempDir()
	file, err := os.Create(dir + "/settings.json")
	if err != nil {
		t.Error(err)
	}
	err = bot.ExportSettings(file)
	if err != nil {
		t.Error(err)
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		t.Error(err)
	}
	bot2 := &UsosBot{}
	err = bot2.ImportSettings(file)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(bot2.guildUsosInfos, bot.guildUsosInfos) {
		t.Errorf("GuildUsosInfos do not match")
	}

	if !reflect.DeepEqual(bot2.tokenMap, bot.tokenMap) {
		t.Errorf("TokenMap do not match")
	}
}
