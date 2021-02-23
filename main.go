package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ogurczak/discord-usos-auth/bot"
	"github.com/akamensky/argparse"
	"github.com/dghubble/oauth1"
)

var programmeName *string
var botToken *string
var settingsFilename *string
var force *bool

func init() {
	parser := argparse.NewParser("discord-usos-auth", "Runs an Usos Authorization Bot instance using the given bot token")
	botToken = parser.String("t", "token", &argparse.Options{Required: true, Help: "bot token"})
	settingsFilename = parser.String("s", "settings", &argparse.Options{Required: false,
		Help: "settings filepath, if not specified no settings will be saved nor loaded"})
	force = parser.Flag("f", "force", &argparse.Options{Required: false,
		Help: "do not ask to overwrite the settings file on exit"})
	err := parser.Parse(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

var token *oauth1.Token

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

	if *settingsFilename == "" {
		log.Println("No settings file specified, no settings will be saved upon exit")
	} else {
		file, err := os.Open(*settingsFilename)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Fatal(err)
			}
		} else {
			err := b.ImportSettings(file)
			if err != nil {
				log.Fatal(err)
			}
			err = file.Close()
			if err != nil {
				log.Fatal(err)
			}
		}

		defer func() {
			exists := true
			_, err := os.Stat(*settingsFilename)
			if err != nil {
				if os.IsNotExist(err) {
					exists = false
				} else {
					log.Fatal(err)
				}
			}

			if !*force && exists {
				fmt.Printf("\nOverwrite file \"%s\"? [y/N]\n", *settingsFilename)

				correct := false
				for !correct {
					var answer string
					fmt.Scanln(&answer)
					switch answer {
					case "Y", "y":
						correct = true
					case "N", "n", "":
						return
					}
				}
			}
			file, err := os.OpenFile(*settingsFilename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
			if err != nil {
				log.Fatal(err)
			}
			err = b.ExportSettings(file)
			if err != nil {
				log.Fatal(err)
			}
			err = file.Close()
			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	err = b.Open()
	if err != nil {
		log.Fatal(err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	// time.Sleep(time.Second * 2)

	b.Close()
}
