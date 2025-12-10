package watcher

import "time"

// These types model the parsed battle objects that the Node bot will consume.

type Item struct {
	Type        string `json:"type"`
	Tier        int    `json:"tier"`
	Enchantment int    `json:"enchantment"`
	Quality     int    `json:"quality"`
}

type Equipment struct {
	MainHand *Item `json:"MainHand,omitempty"`
	OffHand  *Item `json:"OffHand,omitempty"`
	Armor    *Item `json:"Armor,omitempty"`
	Head     *Item `json:"Head,omitempty"`
	Shoes    *Item `json:"Shoes,omitempty"`
	Cape     *Item `json:"Cape,omitempty"`
	Bag      *Item `json:"Bag,omitempty"`
	Potion   *Item `json:"Potion,omitempty"`
	Food     *Item `json:"Food,omitempty"`
}

type Player struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Guild            string    `json:"guild"`
	Alliance         string    `json:"alliance"`
	Equipment        Equipment `json:"equipment"`
	AverageItemPower float64   `json:"average_item_power"`
}

type Event struct {
	EventID      int64    `json:"EventId"`
	Killer       Player   `json:"Killer"`
	Victim       Player   `json:"Victim"`
	KillFame     int64    `json:"TotalVictimKillFame"`
	Participants []Player `json:"Participants"`
	GroupMembers []Player `json:"GroupMembers"`
}

type Battle struct {
	ID        int64     `json:"id"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
	Events    []Event   `json:"events"`
	VictimIDs []string  `json:"victim_ids"`
	Players   []Player  `json:"players"`
	TeamAIDs  []string  `json:"team_a_ids"`
	TeamBIDs  []string  `json:"team_b_ids"`
	CreatedAt time.Time `json:"created_at"`
}
