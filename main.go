package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
	log "github.com/charmbracelet/log"
	"github.com/pacsui/threadsinchannel/handlers"
)

// Flags
var (
	Debug       bool
	HandlerList []handlers.Handler
)

func init() {
	flag.BoolVar(&Debug, "d", false, "Set debug mode")
	flag.Parse()
	log.SetReportCaller(Debug)
	if Debug {
		log.SetLevel(log.DebugLevel)
		log.Debug("Running in Debug!")
		log.Debugf("main PPID: %d", os.Getpid())
	}
	dConVal, err := handlers.ReadConfigFile("config.yaml")
	if err != nil {
		log.Error(err)
		return
	}
	handlers.DiscordBotConfigValues = dConVal

	HandlerList = []handlers.Handler{
		{
			Name:     "starboard_handler",
			Function: handlers.HandleStarBoardAdd,
			File:     "starboard.go",
		},
		{
			Name:     "thread_creator",
			Function: handlers.HandleMessageInChannelPool,
			File:     "channelthread.go",
		},
		{
			Name:     "old commands handler",
			Function: handlers.OnMessageCommandHandler,
			File:     "miscported.go",
		},
	}

}

func main() {
	s, _ := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Infof("Bot running as %s", s.State.User.DisplayName())
	})

	// for _, handler := range HandlerList {
	// 	s.AddHandler(handler.Function)
	// 	log.Infof("[Old-will-deprecate-calling] Added Handler : %s", handler.Name)
	// }

	for _, handler := range handlers.ImportMemberHandlers() {
		s.AddHandler(handler.Function)
		log.Infof("[MemberHandler] Added : %s", handler.Name)
	}

	// for _, handler := range handlers.ImportAdminHandlers() {
	// 	s.AddHandler(handler.Function)
	// 	log.Infof("[AdminHandler] Added : %s", handler.Name)
	// }

	// for _, handler := range handlers.ImportCapboardHandlers() {
	// 	s.AddHandler(handler.Function)
	// 	log.Infof("[Capboard] Added : %s", handler.Name)
	// }

	// for _, handler := range handlers.ImportAnonHandlers() {
	// 	s.AddHandler(handler.Function)
	// 	log.Infof("[Anon] Added : %s", handler.Name)
	// }

	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if strings.HasPrefix(m.Content, handlers.C("mixins")) {
			mixinList := "Enabled Mixins :\n```"
			for _, handler := range HandlerList {
				mixinList += fmt.Sprintf("- %s (%s)\n", handler.Name, handler.File)
			}
			mixinList += "\n```"
			s.ChannelMessageSend(m.ChannelID, mixinList)
		}
	})
	s.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)
	s.State.MaxMessageCount = 500
	s.StateEnabled = true

	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Info("Graceful shutdown")

}
