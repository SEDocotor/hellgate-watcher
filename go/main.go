package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/SEDocotor/hellgate-watcher-go/db"
	"github.com/SEDocotor/hellgate-watcher-go/handlers"
	"github.com/SEDocotor/hellgate-watcher-go/watcher"
)

var dataDir = "data"

func main() {
	// Read DB settings from env; default to Postgres for pure-Go driver use.
	dialect := os.Getenv("DB_DIALECT")
	dsn := os.Getenv("DB_DSN")
	if dialect == "" {
		dialect = "postgres"
	}
	if dsn == "" {
		if dialect == "sqlite" {
			dsn = filepath.Join(dataDir, "hellgate.db")
		} else {
			//default 
			dsn = "host=localhost user=postgres password=postgres dbname=hellgate sslmode=disable"
		}
	}

	db.InitDB(dsn)

	go func() {
		watcher.StartBackgroundPoller(db.WithDB(), watcher.BattleCheckIntervalMins)
	}()

	mux := http.NewServeMux()

	mux.HandleFunc("/channels", handlers.ChannelsHandler(db.WithDB()))
	mux.HandleFunc("/reported_battles", handlers.ReportedBattlesHandler(db.WithDB()))
	mux.HandleFunc("/recent_battles", handlers.RecentBattlesHandler(db.WithDB()))
	mux.HandleFunc("/health", handlers.HealthHandler())

	addr := ":8080"
	log.Printf("Starting Go backend on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
