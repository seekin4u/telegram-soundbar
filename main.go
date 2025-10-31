package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"
)

const cacheFile = "cache.json"

var cacheMu sync.Mutex

// cache is a map of string - string key/values
// "mda-ebat.ogg" : "hashasd123"
type Cache map[string]string

func loadCache() Cache {
	cache := make(Cache)

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return cache // no cache, corrupt format?
		}
		log.Printf("failed to read cache from %s: %v", err, cacheFile)
		return cache
	}

	if err := json.Unmarshal(data, &cache); err != nil {
		log.Printf("failed to parse cache from %s: %v", err, cacheFile)
	}

	return cache
}

func saveCache(cache Cache) {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		log.Printf("failed to encode cache to %s: %v", cacheFile, err)
		return
	}

	if err := os.WriteFile(cacheFile, data, 0644); err != nil {
		log.Printf("failed to write cache file to %s: %v", cacheFile, err)
	}
}

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
	log.Printf("User /mda-ebat")

	filename := "mda-ebat.ogg"
	cache := loadCache()

	if fileID, ok := cache[filename]; ok {
		log.Printf("Using cached file_id for %s", filename)

		voice := &bot.SendVoiceParams{
			ChatID:  update.Message.Chat.ID,
			Voice:   &models.InputFileString{Data: fileID},
			Caption: "Kolomoiskiy(C)",
		}

		if _, err := b.SendVoice(ctx, voice); err != nil {
			log.Printf("failed to send cached voice: %v", err)
		}
		return
	}

	// ELSE, it is not cached - load for the first time amnd then save cache
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
		Caption: "Kolomoiskiy",
	}

	sent, err := b.SendVoice(ctx, voice)
	if err != nil {
		log.Printf("failed to send voice: %v", err)
		return
	}

	if sent == nil || sent.Voice == nil {
		log.Printf("no voice info returned from Telegram")
		return
	}

	//https://pkg.go.dev/github.com/go-telegram/bot@v1.17.0/models#Voice
	//https://core.telegram.org/bots/api#voice
	//FileID will be a cache string from Tg CDN.
	fileID := sent.Voice.FileID
	cache[filename] = fileID
	saveCache(cache)

	log.Printf("Cached new file_id for %s: %s", filename, fileID)
}
