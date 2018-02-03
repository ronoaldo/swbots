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

	// Only works if we are mentinoed in the message
	for i := range m.Mentions {
		// If bot is mentioned in the message
		if m.Mentions[i].ID == clientID {
			logger.Printf("RECV %s", m.Content)

			// Always delete previous message
			logger.Printf("DELT %v", m.ID)
			if err := s.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
				logger.Printf("ERR unable to delete message %v", err)
			}

			// Lookup some names and detects target channel ID if one is in the message
			targetChannelName, targetChannelID := findChannelInMessage(s, m.Content)
			if targetChannelID == "" || targetChannelID == m.ChannelID {
				// Delete source message so we don't get double
				targetChannelID = m.ChannelID
			}
			authorNick := findAuthorNick(logger, s, m)

			// Prepare the message, removing bot mention and
			msg := m.ContentWithMentionsReplaced()
			msg = strings.Replace(msg, "@comlink", "", -1)
			msg = strings.Replace(msg, "@commlink", "", -1)
			msg = strings.TrimSpace(channelRe.ReplaceAllString(msg, ""))

			content := fmt.Sprintf("Incoming message from **@%s**", authorNick)
			if targetChannelName != "" {
				content += " to **#" + targetChannelName + "**"
			}

			_, err := s.ChannelMessageSendComplex(targetChannelID, &discordgo.MessageSend{
				Content: content,
				Embed: &discordgo.MessageEmbed{
					Color: 0x000011,
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL: m.Author.AvatarURL("128"),
					},
					Description: msg,
				},
			})
			if err != nil {
				logger.Printf("ERR unable to send message via comlink: %v", err)
			}
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

func findChannelInMessage(s *discordgo.Session, m string) (string, string) {
	sub := channelRe.FindAllStringSubmatch(m, -1)
	if len(sub) == 0 {
		return "", ""
	}
	if len(sub[0]) >= 2 {
		channelID := sub[0][1]
		ch, err := s.Channel(channelID)
		if err != nil {
			return "", channelID
		}

		return ch.Name, channelID
	}
	return "", ""
}

func findAuthorNick(logger *log.Logger, s *discordgo.Session, m *discordgo.MessageCreate) string {
	ch, err := s.Channel(m.ChannelID)
	if err != nil {
		logger.Printf("ERR looking up channel %v", err)
		return m.Author.Username
	}

	member, err := s.GuildMember(ch.GuildID, m.Author.ID)
	if err != nil {
		logger.Printf("ERR looking up member %v", err)
		return m.Author.Username
	}

	return member.Nick
}
