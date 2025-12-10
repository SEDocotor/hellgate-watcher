package watcher

import (
	"fmt"
	"strings"
)

var QualityIP = map[int]float64{
	0: 0,
	1: 0,
	2: 20,
	3: 40,
	4: 60,
	5: 100,
}

var HealingWeapons = map[string]struct{}{
	"MAIN_HOLYSTAFF":           {},
	"2H_HOLYSTAFF":             {},
	"2H_DIVINESTAFF":           {},
	"MAIN_HOLYSTAFF_MORGANA":   {},
	"2H_HOLYSTAFF_HELL":        {},
	"2H_HOLYSTAFF_UNDEAD":      {},
	"MAIN_HOLYSTAFF_AVALON":    {},
	"2H_HOLYSTAFF_CRYSTAL":     {},
	"MAIN_NATURESTAFF":         {},
	"2H_NATURESTAFF":           {},
	"2H_WILDSTAFF":             {},
	"MAIN_NATURESTAFF_KEEPER":  {},
	"2H_NATURESTAFF_HELL":      {},
	"2H_NATURESTAFF_KEEPER":    {},
	"MAIN_NATURESTAFF_AVALON":  {},
	"MAIN_NATURESTAFF_CRYSTAL": {},
}

const (
	BaseIP                  = 300.0
	Lethal5v5IPCap          = 1100.0
	Lethal5v5SoftcapPercent = 35
	Lethal2v2IPCap          = 1100.0
	Lethal2v2SoftcapPercent = 35
)

// parseItemType parses strings like T5_MAIN_HOLYSTAFF@1 or MAIN_BLADE
func parseItemType(itemType string) (tier int, typ string, enchant int, err error) {
	if itemType == "" {
		return 0, "", 0, nil
	}
	tier = 0
	enchant = 0
	typ = itemType
	if len(itemType) > 0 && (itemType[0] == 'T' || itemType[0] == 't') {
		// expected format T{tier}_REST
		if len(itemType) < 3 {
			return 0, "", 0, fmt.Errorf("invalid item type: %s", itemType)
		}
		// parse tier digit(s)
		n := 1
		for n < len(itemType) && itemType[n] >= '0' && itemType[n] <= '9' {
			n++
		}
		// tier substring
		tierStr := itemType[1:n]
		var t int
		_, err := fmt.Sscanf(tierStr, "%d", &t)
		if err == nil {
			tier = t
		}
		// skip underscore
		if n+1 <= len(itemType) {
			typ = itemType[n+1:]
		} else {
			typ = itemType[n:]
		}
	}
	// enchant suffix like @1
	if len(typ) >= 2 {
		if idx := strings.LastIndex(typ, "@"); idx != -1 && idx == len(typ)-2 {
			var e int
			_, err := fmt.Sscanf(typ[idx+1:], "%d", &e)
			if err == nil {
				enchant = e
				typ = typ[:idx]
			}
		}
	}
	return tier, typ, enchant, nil
}

// ItemData represents parsed item
type ItemData = Item

func itemFromMap(m map[string]interface{}) Item {
	// m expected to have keys: "Type", "Quality"
	typ := ""
	if v, ok := m["Type"].(string); ok {
		typ = v
	}
	quality := 0
	if qf, ok := m["Quality"].(float64); ok {
		quality = int(qf)
	} else if qi, ok := m["Quality"].(int); ok {
		quality = qi
	}
	tier, t, enchant, _ := parseItemType(typ)
	return Item{Type: t, Tier: tier, Enchantment: enchant, Quality: quality}
}

func applyIPCap(ip float64, ipCap float64, softPercent int) float64 {
	if ip <= ipCap {
		return ip
	}
	return ipCap + (ip-ipCap)*float64(softPercent)/100.0
}

// EquipmentData packs items
type EquipmentData = Equipment

func equipmentFromMap(m map[string]interface{}) Equipment {
	var e Equipment
	if v, ok := m["MainHand"].(map[string]interface{}); ok {
		it := itemFromMap(v)
		e.MainHand = &it
	}
	if v, ok := m["OffHand"].(map[string]interface{}); ok {
		it := itemFromMap(v)
		e.OffHand = &it
	}
	if v, ok := m["Armor"].(map[string]interface{}); ok {
		it := itemFromMap(v)
		e.Armor = &it
	}
	if v, ok := m["Head"].(map[string]interface{}); ok {
		it := itemFromMap(v)
		e.Head = &it
	}
	if v, ok := m["Shoes"].(map[string]interface{}); ok {
		it := itemFromMap(v)
		e.Shoes = &it
	}
	if v, ok := m["Cape"].(map[string]interface{}); ok {
		it := itemFromMap(v)
		e.Cape = &it
	}
	if v, ok := m["Bag"].(map[string]interface{}); ok {
		it := itemFromMap(v)
		e.Bag = &it
	}
	if v, ok := m["Potion"].(map[string]interface{}); ok {
		it := itemFromMap(v)
		e.Potion = &it
	}
	if v, ok := m["Food"].(map[string]interface{}); ok {
		it := itemFromMap(v)
		e.Food = &it
	}
	return e
}

// Note: equipment merging and max-average calculations are provided
// by methods in `albion_objects.go` (Equipment.MaxAverageItemPower).
// The below helpers were removed because they were unused and duplicate
// functionality.

func isHealingWeapon(it *Item) bool {
	if it == nil {
		return false
	}
	if _, ok := HealingWeapons[it.Type]; ok {
		return true
	}
	return false
}

// updateFrom merges missing item slots from other into e (like Python's Equipment.update)
func (e *Equipment) updateFrom(other Equipment) {
	if e.MainHand == nil && other.MainHand != nil {
		e.MainHand = other.MainHand
	}
	if e.OffHand == nil && other.OffHand != nil {
		e.OffHand = other.OffHand
	}
	if e.Armor == nil && other.Armor != nil {
		e.Armor = other.Armor
	}
	if e.Head == nil && other.Head != nil {
		e.Head = other.Head
	}
	if e.Shoes == nil && other.Shoes != nil {
		e.Shoes = other.Shoes
	}
	if e.Cape == nil && other.Cape != nil {
		e.Cape = other.Cape
	}
	if e.Bag == nil && other.Bag != nil {
		e.Bag = other.Bag
	}
	if e.Potion == nil && other.Potion != nil {
		e.Potion = other.Potion
	}
	if e.Food == nil && other.Food != nil {
		e.Food = other.Food
	}
}
