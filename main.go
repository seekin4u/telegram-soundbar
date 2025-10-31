package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	caption := "Kolomoiskiy"
	chatID := update.Message.Chat.ID

	if fileID, ok := getCachedFileID(filename); ok {
		log.Printf("Using cached file_id for %s", filename)
		if err := sendCachedVoice(ctx, b, chatID, fileID, caption+"(c)"); err != nil {
			log.Printf("Failed to send cached voice: %v", err)
		}
		return
	}

	// ELSE, it is not cached - load for the first time amnd then save cache

	if err := uploadAndCacheVoice(ctx, b, chatID, filename, caption); err != nil {
		log.Printf("Failed to upload and cache voice: %v", err)
	}
}

func getCachedFileID(filename string) (string, bool) {
	cache := loadCache()
	fileID, ok := cache[filename]
	return fileID, ok
}

func sendCachedVoice(ctx context.Context, b *bot.Bot, chatID int64, fileID, caption string) error {
	voice := &bot.SendVoiceParams{
		ChatID: chatID,
		// https://github.com/go-telegram/bot/blob/v1.17.0/models/input_file.go#L15
		Voice:   &models.InputFileString{Data: fileID},
		Caption: caption,
	}
	_, err := b.SendVoice(ctx, voice)
	return fmt.Errorf("failed to SendVoiceParams w/ InputFileString: %w", err)
}

func uploadAndCacheVoice(ctx context.Context, b *bot.Bot, chatID int64, filename, caption string) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to os.Open voice file: %w", err)
	}
	defer f.Close()

	voice := &bot.SendVoiceParams{
		ChatID: chatID,
		// https://github.com/go-telegram/bot/blob/v1.17.0/models/input_file.go#L15
		Voice:   &models.InputFileUpload{Filename: filename, Data: f},
		Caption: caption,
	}

	sent, err := b.SendVoice(ctx, voice)
	if err != nil {
		return fmt.Errorf("failed to SendVoiceParams w/ InputFileUpload: %w", err)
	}

	if sent == nil || sent.Voice == nil {
		return fmt.Errorf("no voice info returned from Telegram")
	}

	//https://pkg.go.dev/github.com/go-telegram/bot@v1.17.0/models#Voice
	//https://core.telegram.org/bots/api#voice
	//FileID will be a cache string from Tg CDN.
	cache := loadCache()
	cache[filename] = sent.Voice.FileID
	saveCache(cache)

	log.Printf("Cached new file_id for %s: %s", filename, sent.Voice.FileID)
	return nil
}
