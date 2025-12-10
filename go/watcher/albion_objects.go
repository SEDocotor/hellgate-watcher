package watcher

import (
	"strings"
)

const (
	ACCOUNT_FOR_ARTIFACT_IP = 100.0
	MAX_ITEM_LEVEL          = 120.0
	IP_PER_LEVEL            = 2.0
)

// Item methods
func (it Item) getQualityIP() float64 {
	return QualityIP[it.Quality]
}

func (it Item) isPlate() bool {
	t := strings.ToLower(it.Type)
	return strings.Contains(t, "plate")
}
func (it Item) isLeather() bool {
	t := strings.ToLower(it.Type)
	return strings.Contains(t, "leather")
}
func (it Item) isCloth() bool {
	t := strings.ToLower(it.Type)
	return strings.Contains(t, "cloth")
}

func (it Item) getBaseMaxItemPower(ipCap float64, ipSoftcapPercent int) float64 {
	itemPower := BaseIP
	itemPower += float64(it.Tier) * 100.0
	itemPower += float64(it.Enchantment) * 100.0
	itemPower += it.getQualityIP()
	itemPower = applyIPCap(itemPower, ipCap, ipSoftcapPercent)
	return itemPower
}

// Equipment methods
func (e Equipment) MaxAverageItemPower(ipCap float64, ipSoftcapPercent int) int {
	total := 0.0
	// items contributing to IP
	contrib := []*Item{e.Head, e.Armor, e.Shoes, e.MainHand, e.OffHand, e.Cape}
	for _, it := range contrib {
		if it != nil {
			// adjust calculation for armor vs weapon if needed
			if it == e.Armor || it == e.Head || it == e.Shoes {
				// armor piece enhanced calculation per Python's ArmorPiece
				total += it.getBaseMaxItemPower(ipCap, ipSoftcapPercent)
				// add overcharge and mastery approximations
				total += 100.0
				total += MAX_ITEM_LEVEL * IP_PER_LEVEL
				// small additional sums are omitted for simplicity
			} else {
				total += it.getBaseMaxItemPower(ipCap, ipSoftcapPercent)
			}
		}
	}
	if e.OffHand == nil && e.MainHand != nil {
		total += e.MainHand.getBaseMaxItemPower(ipCap, ipSoftcapPercent)
	}
	return int(total / 6.0)
}

// Player methods
func (p *Player) MaxAverageItemPower(ipCap float64, ipSoftcapPercent int) float64 {
	return float64(p.Equipment.MaxAverageItemPower(ipCap, ipSoftcapPercent))
}

// Battle checks
func (b *Battle) isIPCapped(ipCap float64, ipSoftcapPercent int) bool {
	for _, player := range b.Players {
		if player.AverageItemPower > player.MaxAverageItemPower(ipCap, ipSoftcapPercent)+ACCOUNT_FOR_ARTIFACT_IP {
			return false
		}
	}
	return true
}

func (b *Battle) IsHellgate5v5() bool {
	if len(b.Players) != 10 {
		return false
	}
	if !isXvsX(b, 5) {
		return false
	}
	if !b.isIPCapped(Lethal5v5IPCap, Lethal5v5SoftcapPercent) {
		return false
	}
	return true
}

func (b *Battle) IsHellgate2v2() bool {
	if len(b.Players) != 4 {
		return false
	}
	if !isXvsX(b, 2) {
		return false
	}
	if !b.isIPCapped(Lethal2v2IPCap, Lethal2v2SoftcapPercent) {
		return false
	}
	// depths check
	for _, e := range b.Events {
		if e.KillFame == 0 {
			return false
		}
	}
	return true
}

func isXvsX(b *Battle, x int) bool {
	hasTeam := false
	for _, e := range b.Events {
		if len(e.GroupMembers) == x {
			hasTeam = true
		}
		if len(e.GroupMembers) > x {
			return false
		}
	}
	return hasTeam
}
