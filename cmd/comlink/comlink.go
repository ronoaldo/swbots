package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	token    string
	clientID string
)

func init() {
	flag.StringVar(&token, "token", os.Getenv("CMLINK_TOKEN"), "The `token` to be used by the bot")
	flag.StringVar(&clientID, "clientid", os.Getenv("CMLINK_CLIENTID"), "The `cilent id` to be used by the bot")
}

func main() {
	flag.Parse()

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Unable to stat discord: %v", err)
	}

	dg.AddHandler(handleMessage)
	dg.AddHandler(ready)

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

	for i := range m.Mentions {
		// If bot is mentioned in the message
		if m.Mentions[i].ID == clientID {
			logger.Printf("RECV %s", m.Content)
			msg := m.ContentWithMentionsReplaced()
			msg = strings.Replace(msg, "@comlink", "", -1)
			msg = strings.Replace(msg, "@commlink", "", -1)

			targetChannel := findChannel(m.Content)
			if targetChannel == "" || targetChannel == m.ChannelID {
				// Delete source message so we don't get double
				targetChannel = m.ChannelID
				defer s.ChannelMessageDelete(m.ChannelID, m.ID)
			}
			s.ChannelMessageSendComplex(targetChannel, &discordgo.MessageSend{
				Embed: &discordgo.MessageEmbed{
					Color: 0x000011,
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL: m.Author.AvatarURL("128"),
					},
					Description: "Message from " + m.Author.Mention() + ":\n" + msg,
				},
			})
		}
	}
}

// ready is the registered callback for when the bot starts.
func ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateStatus(0, "@comlink #channel to send a message")
	if u, err := s.UserUpdate("", "", "comlink", "", ""); err != nil {
		log.Printf("Could not update profile: %v", err)
	} else {
		log.Printf("Profile updated: %v", u)
	}
}

var channelRe = regexp.MustCompile("<#(\\d+)>")

func findChannel(m string) string {
	sub := channelRe.FindAllStringSubmatch(m, -1)
	if len(sub) == 0 {
		return ""
	}
	if len(sub[0]) >= 2 {
		return sub[0][1]
	}
	return ""
}
