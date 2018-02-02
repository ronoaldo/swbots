package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	token    string
	clientID string
)

func init() {
	flag.StringVar(&token, "token", os.Getenv("C2B5_TOKEN"), "The `token` to be used by the bot")
	flag.StringVar(&clientID, "clientid", os.Getenv("C2B5_CLIENTID"), "The `cilent id` to be used by the bot")
}

func main() {
	flag.Parse()

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Unable to stat discord: %v", err)
	}

	dg.AddHandler(handleMessage)

	err = dg.Open()
	if err != nil {
		log.Fatalf("Unable to open: %v", err)
	}

	fmt.Println("Bot is now running. Press CTRL+C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}

func handleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	logger := log.New(os.Stderr, fmt.Sprintf("[%v#%v] ", m.ID, m.ChannelID), log.LstdFlags)
	// Do not get stuck talking to himself
	if m.Author.ID == clientID {
		return
	}

	logger.Printf("RECV %s", m.Content)
	for i := range m.Mentions {
		// If bot is mentioned in the message
		if m.Mentions[i].ID == clientID {
			cmd := ""
			text := ""
			if index := strings.Index(m.Content, " decrypt "); index >= 0 {
				cmd = "rot13"
				text = m.Content[index+9:]
			} else if index := strings.Index(m.Content, " encrypt "); index >= 0 {
				cmd = "rot13"
				text = m.Content[index+9:]
			}

			switch cmd {
			case "rot13":
				logger.Printf("CMD: %v TEXT: %s", cmd, text)
				resp := fmt.Sprintf("*%s*", strings.Map(rot13, text))
				if _, err := s.ChannelMessageSend(m.ChannelID, resp); err != nil {
					logger.Printf("ERROR: %v", err)
				}
			}
		}
	}
}

// Source: https://www.dotnetperls.com/rot13-go
func rot13(r rune) rune {
	if r >= 'a' && r <= 'z' {
		// Rotate lowercase letters 13 places.
		if r > 'm' {
			return r - 13
		}
		return r + 13
	} else if r >= 'A' && r <= 'Z' {
		// Rotate uppercase letters 13 places.
		if r > 'M' {
			return r - 13
		}
		return r + 13
	}
	// Do nothing.
	return r
}
