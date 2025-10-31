package main

import (
	"context"
	"log"
	"os"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using env variables instead")
	}

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN not set in .env nor in env")
	}

	b, err := bot.New(token, bot.WithDefaultHandler(defaultHandler))
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, handleStart)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/voice", bot.MatchTypeExact, handleVoice)

	botname, berr := b.GetMe(context.Background())
	if berr != nil {
		log.Fatal("botname failed")

	}

	log.Printf("Bot started as @%s", botname.Username)
	b.Start(context.Background())
}

// Default handler (called for any unmatched message)
func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message != nil {
		log.Printf("[%s] %s", update.Message.From.Username, update.Message.Text)
	}
}

func handleStart(ctx context.Context, b *bot.Bot, update *models.Update) {
	msg := &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Hi! Send /voice to get your voice message 🎙️",
	}
	b.SendMessage(ctx, msg)
	log.Print("User /start")
}

func handleVoice(ctx context.Context, b *bot.Bot, update *models.Update) {
	// https://github.com/go-telegram/bot/blob/v1.17.0/models/input_file.go#L15
	f, err := os.Open("mda-ebat.ogg")
	if err != nil {
		log.Printf("failed to open voice file: %v", err)
		return
	}
	defer f.Close()

	voice := &bot.SendVoiceParams{
		ChatID:  update.Message.Chat.ID,
		Voice:   &models.InputFileUpload{Filename: "mda-ebat.ogg", Data: f},
		Caption: "Here’s your voice message!",
	}

	_, err = b.SendVoice(ctx, voice)
	if err != nil {
		log.Printf("failed to send voice: %v", err)
	}
}
