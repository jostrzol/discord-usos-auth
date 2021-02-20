package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ogurczak/discord-usos-auth/bot"
	"github.com/akamensky/argparse"
)

var programmeName *string
var botToken *string

func init() {
	parser := argparse.NewParser("discord-usos-auth", "Runs an Usos Authorization Bot instance using the given bot token")
	botToken = parser.String("t", "token", &argparse.Options{Required: true, Help: "bot token"})
	err := parser.Parse(os.Args)
	if err != nil {
		log.Fatal(err)
	}
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

	b, err := bot.New("Bot " + *botToken)
	// b.UsosUserFilter = filterFunc
	if err != nil {
		log.Fatal(err)
	}

	err = b.Open()
	if err != nil {
		log.Fatal(err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	b.Close()
}
