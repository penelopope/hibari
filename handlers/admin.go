package handlers

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
)

func ImportAdminHandlers() []Handler {
	return []Handler{
		{
			Name:     "AdminHandler",
			Function: AdminHandler,
			File:     "admin.go",
		},
	}
}

func AdminHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	CommandSent := strings.TrimSpace(strings.TrimPrefix(m.Message.Content, C("admin")))
	if !strings.HasPrefix(m.Content, C("admin")) {
		return
	}
	s.MessageReactionAdd(m.ChannelID, m.ID, "pyes:1415778092500254741")
	switch {
	case strings.HasPrefix(CommandSent, "announce"):
		if m.ChannelID != DiscordBotConfigValues.ModChannel {
			return
		}
		handleAnnouncement(s, m)
	case strings.HasPrefix(CommandSent, "purge"):
		if m.ChannelID != DiscordBotConfigValues.ModChannel {
			return
		}
		handlePurge(s, m, "", 0)
	case strings.HasPrefix(CommandSent, "send"):
		if m.ChannelID != DiscordBotConfigValues.ModChannel {
			return
		}
		arg := strings.TrimSpace(strings.TrimPrefix(CommandSent, "send"))
		handleSend(s, m, arg)
	case strings.HasPrefix(CommandSent, "cdel"):
		if m.ChannelID != DiscordBotConfigValues.ModChannel {
			log.Debug(m.ChannelID, DiscordBotConfigValues.ModChannel)
			return
		}
		log.Debug("Running Anon Confession deletion")

		ConfessionForceCensor(s, strings.TrimSpace(strings.TrimPrefix(CommandSent, "cdel")))
	case strings.HasPrefix(CommandSent, "vdel"):
		if m.ChannelID != DiscordBotConfigValues.ModChannel {
			log.Debug(m.ChannelID, DiscordBotConfigValues.ModChannel)
			return
		}
		log.Debug("Running Anon Vent deletion")
		VentForceCensor(s, strings.TrimSpace(strings.TrimPrefix(CommandSent, "vdel")))
	default:
		return
	}
}

func handleAnnouncement(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.Warn("not impl")
}

func handleSend(s *discordgo.Session, m *discordgo.MessageCreate, a string) {
	// expecting args to be ChannelID::Message to send
	splitted := strings.Split(a, "::")
	if len(splitted) == 2 {
		s.ChannelMessageSend(splitted[0], splitted[1])
	} else {
		s.MessageReactionAdd(m.ChannelID, m.ID, "❌")
	}
}

func handlePurge(s *discordgo.Session, m *discordgo.MessageCreate, channelID string, count int) {
	log.Warn("not impl")
}
