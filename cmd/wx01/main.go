package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jedediah/wx01/internal/api"
	"github.com/jedediah/wx01/internal/db"
	"github.com/jedediah/wx01/internal/ingest"
	"github.com/jedediah/wx01/internal/notify"
)

//go:embed dist
var embeddedDist embed.FS

func main() {
	cfg := loadConfig()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	store := db.NewStore(pool)

	udp, err := ingest.NewUDPListener(cfg.UDPPort, store)
	if err != nil {
		log.Fatalf("udp listener failed: %v", err)
	}

	hub := notify.NewHub()

	distFS, _ := fs.Sub(embeddedDist, "dist")
	server := api.NewServer(cfg.HTTPPort, pool, distFS, hub)

	// Start services
	go udp.Run(ctx)
	go hub.Listen(ctx, cfg.DatabaseURL)
	go server.Run(ctx)

	log.Printf("wx01 running — udp:%d http:%d", cfg.UDPPort, cfg.HTTPPort)
	<-ctx.Done()
	log.Println("shutting down")
}

type config struct {
	DatabaseURL string
	UDPPort     int
	HTTPPort    int
}

func loadConfig() config {
	dbURL := os.Getenv("WX01_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://wx01:wx01@localhost:5432/wx01?sslmode=disable"
	}

	udpPort := 50222
	httpPort := 3100

	if v := os.Getenv("WX01_UDP_PORT"); v != "" {
		var p int
		if _, err := parsePort(v, &p); err == nil {
			udpPort = p
		}
	}
	if v := os.Getenv("WX01_HTTP_PORT"); v != "" {
		var p int
		if _, err := parsePort(v, &p); err == nil {
			httpPort = p
		}
	}

	return config{
		DatabaseURL: dbURL,
		UDPPort:     udpPort,
		HTTPPort:    httpPort,
	}
}

func parsePort(s string, out *int) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, os.ErrInvalid
		}
		n = n*10 + int(c-'0')
	}
	*out = n
	return n, nil
}
