package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ogurczak/discord-usos-auth/bot"
	"github.com/Ogurczak/discord-usos-auth/usos"
)

var programmeID string
var botToken string
var logUserID string
var logChannelID string

func filterFunc(u *usos.User) (bool, error) {
	for _, prog := range u.Programmes {
		if prog.ID == programmeID {
			return true, nil
		}
	}
	return false, nil
}

func init() {

	flag.StringVar(&botToken, "t", "", "Bot Token")
	flag.StringVar(&programmeID, "p", "", "Desired Programme ID")
	flag.StringVar(&logUserID, "l", "", "ID of a user to send authorization data to")
	flag.StringVar(&logChannelID, "c", "", "ID of a channel to send authorization data to (has priority vs user log)")
	flag.Parse()
}

func main() {
	// if token == nil {
	// 	rt, err := usos.NewRequestToken()
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	fmt.Printf("Log in to usos from this url: %s\n", rt.AuthorizationURL)

	// 	var verifier string
	// 	fmt.Scan(&verifier)
	// 	token, err := rt.GetAccessToken(verifier)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	fmt.Printf("Token: %s\nSecret: %s\n", token.Token, token.TokenSecret)
	// }

	// user, err := usos.NewUsosUser(token)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// _ = user

	dg, err := bot.New("Bot " + botToken)
	bot.UsosUserFilter = filterFunc
	bot.LogUserID = logUserID
	bot.LogChannelID = logChannelID
	if err != nil {
		log.Fatal(err)
	}

	err = dg.Open()
	if err != nil {
		log.Fatal(err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}
