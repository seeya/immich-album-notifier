package main

import (
	"fmt"
	immich "immich_album_notifier/internal/api"
	"immich_album_notifier/internal/telegram"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

func main() {
	slog.Info("immich album notifier running")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	telegram.Init()
	api := immich.InitImmichClient(os.Getenv("IMMICH_API_KEY"), os.Getenv("IMMICH_ENDPOINT"))

	interval := os.Getenv("CRON_SCHEDULE")

	if interval == "" {
		interval = "@hourly"
	}

	slog.Info("update frequency", slog.String("interval", interval))

	c := cron.New()
	c.AddFunc(os.Getenv("CRON_SCHEDULE"), func() {
		job(api)
	})

	c.Start()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("server running", slog.String("port", port))
	err = http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", port), nil)
	if err != nil {
		panic(err)
	}
}

func job(api *immich.ImmichClient) {
	slog.Info("running job")

	cachedAlbums := api.ReadCached()

	searchAlbum := "Pearly"

	newAlbums := api.GetAllAlbums()
	if newAlbums == nil {
		slog.Info("No albums found")
		return
	}

	for _, album := range *newAlbums {
		if album.AlbumName == searchAlbum {
			foundAlbum := findAlbum(cachedAlbums, searchAlbum)

			count := 0
			if foundAlbum != nil {
				count = foundAlbum.AssetCount
			}

			difference := album.AssetCount - count

			slog.Info("changes", slog.String("album", searchAlbum), slog.Int("difference", difference))

			if difference != 0 {
				msg := fmt.Sprintf("[%s] album has %d new media uploaded!", searchAlbum, difference)
				telegram.SendMessage(os.Getenv("TELEGRAM_CHAT_ID"), msg)
			}
		}
	}
}

func findAlbum(albums []immich.Album, name string) *immich.Album {
	for _, album := range albums {
		if album.AlbumName == name {
			return &album
		}
	}

	return nil
}
