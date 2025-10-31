package main

import (
	"context"
	"log"
	"os"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func main() {
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("Please set BOT_TOKEN environment variable")
	}

	b, err := bot.New(token, bot.WithDefaultHandler(defaultHandler))
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	// b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, handleStart)
	// b.RegisterHandler(bot.HandlerTypeMessageText, "/voice", bot.MatchTypeExact, handleVoice)

	botname, berr := b.GetMe(context.Background())
	if berr != nil {
		log.Fatal("botname failed")

	}

	log.Printf("Bot started as @%s", botname.Username)
	b.Start(context.Background())
}

func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message != nil {
		log.Printf("[%s] %s", update.Message.From.Username, update.Message.Text)
	}
}
