package handlers

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	MONGODB_URI string
	capboard    *mongo.Collection
)

type CapTransaction struct {
	MessageID string    `bson:"message_id"`
	ChannelID string    `bson:"channel_id"`
	GuildID   string    `bson:"guild_id"`
	GiverID   string    `bson:"giver_id"`
	TakerID   string    `bson:"taker_id"`
	CreatedAt time.Time `bson:"created_at"`
}

type LeaderboardEntry struct {
	UserID string `bson:"_id"`
	Count  int    `bson:"count"`
}

type PostLeaderboardEntry struct {
	MessageID string `bson:"_id"`
	ChannelID string `bson:"channel_id"`
	GuildID   string `bson:"guild_id"`
	Count     int    `bson:"count"`
}

func init() {
	log.Info("Connecting to MongoDB...")
	MONGODB_URI := os.Getenv("MONGODB_URI")
	client, err := mongo.Connect(options.Client().
		ApplyURI(MONGODB_URI))
	if err != nil {
		log.Debug(MONGODB_URI)
		log.Error(err.Error())
	}
	capboard = client.Database("Hibari").Collection("Capboard")
}

func ImportCapboardHandlers() []Handler {
	return []Handler{
		{
			Name:     "CapboardHandlers",
			Function: CapBoardHandler,
			File:     "capboard.go",
		},
		{
			Name:     "CapboardCommandHandlers",
			Function: CapBoardCommandHandler,
			File:     "capboard.go",
		},
	}
}

func CapBoardHandler(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	switch m.Emoji.Name {
	case "🧢":
		CapBoardProcessing(s, m)
	}
}

func CapBoardCommandHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if strings.HasPrefix(m.Content, C("caps")) {
		SendCapboardStats(s, m.ChannelID, 0x79AEA3)
	}
}

func CapBoardProcessing(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	TakerMessage, err := s.ChannelMessage(m.ChannelID, m.MessageID)
	if err != nil {
		log.Errorf("Failed to retrieve message: %v", err)
		return
	}

	GiverUser := m.MessageReaction.UserID
	Taker := TakerMessage.Author.ID

	if GiverUser == Taker {
		log.Debugf("User %s tried to cap themselves", GiverUser)
		return
	}

	capData := CapTransaction{
		MessageID: m.MessageID,
		ChannelID: m.ChannelID,
		GuildID:   m.GuildID,
		GiverID:   GiverUser,
		TakerID:   Taker,
		CreatedAt: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = capboard.InsertOne(ctx, capData)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Debugf("Duplicate cap blocked: %s on message %s", GiverUser, m.MessageID)
			return
		}
		log.Errorf("Failed to insert cap: %v", err)
		return
	}

	log.Debugf("Successfully recorded cap! Giver: %s -- Taker: %s", GiverUser, Taker)
}

func SendCapboardStats(s *discordgo.Session, channelID string, colHex int) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var topGivers []LeaderboardEntry
	giverCursor, _ := capboard.Aggregate(ctx, mongo.Pipeline{
		bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$giver_id"}, {Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}}}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
		bson.D{{Key: "$limit", Value: 5}},
	})
	if giverCursor != nil {
		giverCursor.All(ctx, &topGivers)
	}

	var topTakers []LeaderboardEntry
	takerCursor, _ := capboard.Aggregate(ctx, mongo.Pipeline{
		bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$taker_id"}, {Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}}}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
		bson.D{{Key: "$limit", Value: 5}},
	})
	if takerCursor != nil {
		takerCursor.All(ctx, &topTakers)
	}

	var topPosts []PostLeaderboardEntry
	postCursor, _ := capboard.Aggregate(ctx, mongo.Pipeline{
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$message_id"},
			{Key: "channel_id", Value: bson.D{{Key: "$first", Value: "$channel_id"}}},
			{Key: "guild_id", Value: bson.D{{Key: "$first", Value: "$guild_id"}}},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
		bson.D{{Key: "$limit", Value: 5}},
	})
	if postCursor != nil {
		postCursor.All(ctx, &topPosts)
	}

	giverStr, takerStr, postStr := "", "", ""

	for i, entry := range topGivers {
		giverStr += fmt.Sprintf("**%d ** - <@%s> - `%d`\n", i+1, entry.UserID, entry.Count)
	}
	for i, entry := range topTakers {
		takerStr += fmt.Sprintf("**%d** - <@%s> - `%d`\n", i+1, entry.UserID, entry.Count)
	}
	for i, entry := range topPosts {
		link := fmt.Sprintf("https://discord.com/channels/%s/%s/%s", entry.GuildID, entry.ChannelID, entry.MessageID)
		msg, err := s.ChannelMessage(entry.ChannelID, entry.MessageID)

		toSend := ""
		if err != nil {
			toSend = "Deleted Message"
		} else {
			if len(msg.Content) > 0 {
				toSend = msg.Content[:min(25, len(msg.Content))]
				if len(msg.Content) > 25 {
					toSend += "..."
				}
			} else {
				toSend = "Image/Attachment"
			}
		}
		toSend = strings.ReplaceAll(toSend, "\n", " ")

		postStr += fmt.Sprintf("**%d** - [%s](%s) - `%d caps`\n", i+1, toSend, link, entry.Count)
	}

	if giverStr == "" {
		giverStr = "No data yet!"
	}
	if takerStr == "" {
		takerStr = "No data yet!"
	}
	if postStr == "" {
		postStr = "No data yet!"
	}

	embedGen := discordgo.MessageEmbed{
		Title: "Capboard Statistics",
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://cdn.pacsui.me/imgs/hibari/hibari_look_cap.jpg",
		},
		Description: "Top capped data:",
		Color:       colHex,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Top Caps Given",
				Value:  giverStr,
				Inline: true,
			},
			{
				Name:   "Top Caps Taken",
				Value:  takerStr,
				Inline: true,
			},
			{
				Name:   "Top Capped Posts",
				Value:  postStr,
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Hibaricap!!",
		},
	}

	_, err := s.ChannelMessageSendEmbed(channelID, &embedGen)
	if err != nil {
		fmt.Printf("Failed to send leaderboard embed: %v\n", err)
	}
}

func CapBoardRemoveHandler(s *discordgo.Session, m *discordgo.MessageReactionRemove) {
	switch m.Emoji.Name {
	case "🧢":
		CapBoardRemoveProcessing(s, m)
	}
}

func CapBoardRemoveProcessing(s *discordgo.Session, m *discordgo.MessageReactionRemove) {
	GiverUser := m.UserID

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{
		{Key: "message_id", Value: m.MessageID},
		{Key: "giver_id", Value: GiverUser},
	}

	res, err := capboard.DeleteOne(ctx, filter)
	if err != nil {
		log.Errorf("Failed to remove cap from DB: %v", err)
		return
	}

	if res.DeletedCount > 0 {
		log.Debugf("Successfully removed cap! Giver: %s -- Message: %s", GiverUser, m.MessageID)
	} else {
		log.Debugf("Cap removal ignored: record not found for Giver %s on message %s", GiverUser, m.MessageID)
	}
}
