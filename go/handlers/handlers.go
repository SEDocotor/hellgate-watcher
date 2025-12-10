package handlers

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/SEDocotor/hellgate-watcher-go/models"
	"github.com/SEDocotor/hellgate-watcher-go/watcher"
)

// ChannelsHandler returns an http.HandlerFunc that handles GET/POST for channels
func ChannelsHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			var channels []models.Channel
			if err := db.Find(&channels).Error; err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(channels)
		case http.MethodPost:
			var c models.Channel
			if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}
			// Upsert by unique index
			if err := db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "server"}, {Name: "mode"}, {Name: "guild_id"}}, DoUpdates: clause.AssignmentColumns([]string{"channel_id"})}).Create(&c).Error; err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(c)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

// ReportedBattlesHandler handles GET/POST for reported battles
func ReportedBattlesHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			var reported []models.ReportedBattle
			if err := db.Find(&reported).Error; err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			out := map[string][]int64{}
			for _, rb := range reported {
				out[rb.Server] = append(out[rb.Server], rb.BattleID)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(out)
		case http.MethodDelete:
			// Optional ?server= query to clear a single server, otherwise clear all
			server := r.URL.Query().Get("server")
			var err error
			if server != "" {
				err = db.Where("server = ?", server).Delete(&models.ReportedBattle{}).Error
			} else {
				// allow global delete
				err = db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.ReportedBattle{}).Error
			}
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		case http.MethodPost:
			var body struct {
				Server   string `json:"server"`
				BattleID int64  `json:"battle_id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}
			rb := models.ReportedBattle{Server: body.Server, BattleID: body.BattleID}
			if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&rb).Error; err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(body)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

// RecentBattlesHandler returns parsed recent battles by polling the Albion APIs (uses watcher package)
func RecentBattlesHandler(gdb *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// Prefer cached recent battles from background poller if available
		if cached := watcher.LastRecentBattles(); cached != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(cached)
			return
		}
		res, err := watcher.GetRecentBattles(gdb)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	}
}

// HealthHandler returns a simple health check handler
func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}
}
