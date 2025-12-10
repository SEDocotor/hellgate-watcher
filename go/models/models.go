package models

import "time"

// Channel represents a mapping of server/mode/guild -> discord channel
type Channel struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Server    string    `gorm:"uniqueIndex:idx_channel,priority:1" json:"server"`
	Mode      string    `gorm:"uniqueIndex:idx_channel,priority:2" json:"mode"`
	GuildID   string    `gorm:"uniqueIndex:idx_channel,priority:3" json:"guild_id"`
	ChannelID string    `json:"channel_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ReportedBattle stores which battles have already been reported per server
type ReportedBattle struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Server    string    `gorm:"index:idx_reported,priority:1" json:"server"`
	BattleID  int64     `gorm:"index:idx_reported,priority:2" json:"battle_id"`
	CreatedAt time.Time `json:"created_at"`
}
