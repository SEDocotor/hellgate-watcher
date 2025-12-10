package watcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/SEDocotor/hellgate-watcher-go/models"
	"gorm.io/gorm"
)

// helper for HTTP GET with timeout
func httpGetBytes(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// GetRecentBattles polls each server and returns parsed battles grouped by server and mode.
func GetRecentBattles(gdb *gorm.DB) (map[string]map[string][]Battle, error) {
	// result shape: server -> {"5v5": [], "2v2": []}
	res := map[string]map[string][]Battle{
		"europe":   {"5v5": {}, "2v2": {}},
		"americas": {"5v5": {}, "2v2": {}},
		"asia":     {"5v5": {}, "2v2": {}},
	}

	// load reported battles from DB
	reportedMap := map[string]map[int64]struct{}{}
	var reported []models.ReportedBattle
	if err := gdb.Find(&reported).Error; err == nil {
		for _, rb := range reported {
			if reportedMap[rb.Server] == nil {
				reportedMap[rb.Server] = map[int64]struct{}{}
			}
			reportedMap[rb.Server][rb.BattleID] = struct{}{}
		}
	}

	for server, base := range ServerURLs {
		_ = base
		// fetch pages until time window satisfied
		page := 0
		var battlesDicts []map[string]interface{}
		for {
			slice, err := get50Battles(server, page)
			if err != nil {
				break
			}
			battlesDicts = append(battlesDicts, slice...)
			page++
			if containsBattlesOutOfRange(battlesDicts) {
				break
			}
			// rate limit
			time.Sleep(time.Duration(RateLimitDelaySeconds * float64(time.Second)))
		}

		var toMark []int64
		for _, bdict := range battlesDicts {
			idFloat, ok := bdict["id"].(float64)
			if !ok {
				continue
			}
			id := int64(idFloat)
			if _, seen := reportedMap[server][id]; seen {
				continue
			}
			// fetch events for battle
			events, err := getBattleEvents(id, server)
			if err != nil {
				continue
			}
			battle := buildBattleFromDict(bdict, events)
			// determine mode using Battle methods (ported from Python)
			if battle.IsHellgate5v5() {
				res[server]["5v5"] = append(res[server]["5v5"], battle)
				toMark = append(toMark, id)
			} else if battle.IsHellgate2v2() {
				res[server]["2v2"] = append(res[server]["2v2"], battle)
				toMark = append(toMark, id)
			}
		}

		// persist reported battles (upsert/ignore duplicates)
		for _, bid := range toMark {
			rb := models.ReportedBattle{Server: server, BattleID: bid}
			// best-effort insert; ignore errors (unique constraint will error on duplicates)
			_ = gdb.Create(&rb).Error
		}
	}

	return res, nil
}

var (
	lastRecentBattles map[string]map[string][]Battle
	lastMu            sync.RWMutex
)

// StartBackgroundPoller launches a goroutine that polls Albion APIs every configured interval
// and stores the last found recent battles in memory. It also persists reported battles to DB.
func StartBackgroundPoller(gdb *gorm.DB, intervalMinutes int) {
	ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
	go func() {
		for {
			res, err := GetRecentBattles(gdb)
			if err != nil {
				fmt.Printf("watcher poll error: %v\n", err)
			} else {
				lastMu.Lock()
				lastRecentBattles = res
				lastMu.Unlock()
			}
			<-ticker.C
		}
	}()
}

// LastRecentBattles returns the last polled recent battles (may be nil)
func LastRecentBattles() map[string]map[string][]Battle {
	lastMu.RLock()
	defer lastMu.RUnlock()
	return lastRecentBattles
}

// get50Battles pulls a page of battles for a server
func get50Battles(server string, page int) ([]map[string]interface{}, error) {
	base, ok := ServerURLs[server]
	if !ok {
		return nil, fmt.Errorf("unknown server %s", server)
	}
	url := fmt.Sprintf("%s/battles?limit=%d&sort=recent&offset=%d", base, BattlesLimit, page*BattlesLimit)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(TimeoutSeconds)*time.Second)
	defer cancel()
	b, err := httpGetBytes(ctx, url)
	if err != nil {
		return nil, err
	}
	var arr []map[string]interface{}
	if err := json.Unmarshal(b, &arr); err != nil {
		return nil, err
	}
	return arr, nil
}

// getBattleEvents fetches events for a battle
func getBattleEvents(battleID int64, server string) ([]map[string]interface{}, error) {
	base, ok := ServerURLs[server]
	if !ok {
		return nil, fmt.Errorf("unknown server %s", server)
	}
	url := fmt.Sprintf("%s/events/battle/%d", base, battleID)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(TimeoutSeconds)*time.Second)
	defer cancel()
	b, err := httpGetBytes(ctx, url)
	if err != nil {
		return nil, err
	}
	var arr []map[string]interface{}
	if err := json.Unmarshal(b, &arr); err != nil {
		return nil, err
	}
	return arr, nil
}

// containsBattlesOutOfRange logic follows Python implementation: get times and compare window
func containsBattlesOutOfRange(battles []map[string]interface{}) bool {
	if len(battles) == 0 {
		return false
	}
	times := []time.Time{}
	for _, b := range battles {
		if st, ok := b["startTime"].(string); ok {
			if t, err := time.Parse(time.RFC3339, st); err == nil {
				times = append(times, t)
			}
		}
	}
	if len(times) < 2 {
		return false
	}
	sort.Slice(times, func(i, j int) bool { return times[i].Before(times[j]) })
	return times[len(times)-1].Sub(times[0]) > time.Duration(BattlesMaxAgeMinutes)*time.Minute
}

// buildBattleFromDict constructs a Battle from raw JSON maps
func buildBattleFromDict(bdict map[string]interface{}, events []map[string]interface{}) Battle {
	b := Battle{}
	if idf, ok := bdict["id"].(float64); ok {
		b.ID = int64(idf)
	}
	if st, ok := bdict["startTime"].(string); ok {
		b.StartTime = st
	}
	if et, ok := bdict["endTime"].(string); ok {
		b.EndTime = et
	}
	// map events
	for _, ed := range events {
		e := Event{}
		if eid, ok := ed["EventId"].(float64); ok {
			e.EventID = int64(eid)
		}
		// For brevity we only unmarshal Killer and Victim minimal fields here
		if k, ok := ed["Killer"].(map[string]interface{}); ok {
			e.Killer = mapToPlayer(k)
		}
		if v, ok := ed["Victim"].(map[string]interface{}); ok {
			e.Victim = mapToPlayer(v)
		}
		if kf, ok := ed["TotalVictimKillFame"].(float64); ok {
			e.KillFame = int64(kf)
		}
		// participants and group members are arrays; convert minimally
		if parts, ok := ed["Participants"].([]interface{}); ok {
			for _, p := range parts {
				if pm, ok := p.(map[string]interface{}); ok {
					e.Participants = append(e.Participants, mapToPlayer(pm))
				}
			}
		}
		if gms, ok := ed["GroupMembers"].([]interface{}); ok {
			for _, gm := range gms {
				if gmM, ok := gm.(map[string]interface{}); ok {
					e.GroupMembers = append(e.GroupMembers, mapToPlayer(gmM))
				}
			}
		}
		b.Events = append(b.Events, e)
		b.VictimIDs = append(b.VictimIDs, e.Victim.ID)
	}
	// find and update players list from events
	playerMap := map[string]Player{}
	for _, ev := range b.Events {
		addOrUpdatePlayer(&playerMap, ev.Killer)
		addOrUpdatePlayer(&playerMap, ev.Victim)
		for _, p := range ev.Participants {
			addOrUpdatePlayer(&playerMap, p)
		}
		for _, p := range ev.GroupMembers {
			addOrUpdatePlayer(&playerMap, p)
		}
	}
	for _, p := range playerMap {
		b.Players = append(b.Players, p)
	}
	// split teams and sort
	b.splitIDsByTeam()
	b.sortTeamsByClass()
	return b
}

func mapToPlayer(pm map[string]interface{}) Player {
	p := Player{}
	if id, ok := pm["Id"].(string); ok {
		p.ID = id
	}
	if name, ok := pm["Name"].(string); ok {
		p.Name = name
	}
	if guild, ok := pm["GuildName"].(string); ok {
		p.Guild = guild
	}
	if alliance, ok := pm["AllianceName"].(string); ok {
		p.Alliance = alliance
	}
	if aip, ok := pm["AverageItemPower"].(float64); ok {
		p.AverageItemPower = aip
	}
	if eq, ok := pm["Equipment"].(map[string]interface{}); ok {
		p.Equipment = equipmentFromMap(eq)
	}
	return p
}

func addOrUpdatePlayer(m *map[string]Player, p Player) {
	if p.ID == "" {
		return
	}
	if existing, ok := (*m)[p.ID]; ok {
		// Merge equipment slots from the new player into the existing record
		existing.Equipment.updateFrom(p.Equipment)
		// Prefer a non-zero AverageItemPower if available
		if existing.AverageItemPower == 0 && p.AverageItemPower > 0 {
			existing.AverageItemPower = p.AverageItemPower
		}
		(*m)[p.ID] = existing
	} else {
		(*m)[p.ID] = p
	}
}

// Classification and IP checks moved to albion_objects.go and are used via Battle methods.

// splitIDsByTeam assigns team ids using the same iterative approach as Python
func (b *Battle) splitIDsByTeam() {
	teamA := map[string]struct{}{}
	teamB := map[string]struct{}{}
	allIDs := map[string]struct{}{}
	for _, p := range b.Players {
		allIDs[p.ID] = struct{}{}
	}
	if len(b.Events) > 0 {
		first := b.Events[0].Killer.ID
		teamA[first] = struct{}{}
	}
	for i := 0; i < len(allIDs)+1; i++ {
		for _, e := range b.Events {
			killer := e.Killer.ID
			victim := e.Victim.ID
			group := map[string]struct{}{}
			for _, gm := range e.GroupMembers {
				group[gm.ID] = struct{}{}
			}
			if _, ok := teamA[killer]; ok {
				for id := range group {
					teamA[id] = struct{}{}
				}
				if _, inA := teamA[victim]; !inA {
					teamB[victim] = struct{}{}
				}
			} else if _, ok := teamB[killer]; ok {
				for id := range group {
					teamB[id] = struct{}{}
				}
				if _, inB := teamB[victim]; !inB {
					teamA[victim] = struct{}{}
				}
			}
			if _, inA := teamA[victim]; inA {
				if _, inA2 := teamA[killer]; !inA2 {
					teamB[killer] = struct{}{}
					for id := range group {
						teamB[id] = struct{}{}
					}
				}
			} else if _, inB := teamB[victim]; inB {
				if _, inB2 := teamB[killer]; !inB2 {
					teamA[killer] = struct{}{}
					for id := range group {
						teamA[id] = struct{}{}
					}
				}
			}
		}
	}
	// fill remaining
	if len(teamA) >= len(b.Players)/2 {
		for id := range allIDs {
			if _, inA := teamA[id]; !inA {
				teamB[id] = struct{}{}
			}
		}
	} else if len(teamB) >= len(b.Players)/2 {
		for id := range allIDs {
			if _, inB := teamB[id]; !inB {
				teamA[id] = struct{}{}
			}
		}
	}
	for id := range teamA {
		b.TeamAIDs = append(b.TeamAIDs, id)
	}
	for id := range teamB {
		b.TeamBIDs = append(b.TeamBIDs, id)
	}
}

// sortTeamsByClass placeholder: Python sorts by weapon/armor; we'll keep natural order for now
func (b *Battle) sortTeamsByClass() {
	// Reproduce Python's _sort_team behavior to order players by class and weapon
	b.TeamAIDs = b._sortTeam(b.TeamAIDs)
	b.TeamBIDs = b._sortTeam(b.TeamBIDs)
}

func (b *Battle) getPlayer(id string) *Player {
	for i := range b.Players {
		if b.Players[i].ID == id {
			return &b.Players[i]
		}
	}
	return nil
}

func (b *Battle) _sortTeam(team []string) []string {
	healers := []string{}
	melees := []string{}
	tanks := []string{}
	leathers := []string{}
	cloth := []string{}
	unknown := []string{}

	for _, playerID := range team {
		p := b.getPlayer(playerID)
		if p == nil {
			unknown = append(unknown, playerID)
			continue
		}

		if isHealingWeapon(p.Equipment.MainHand) {
			healers = append(healers, playerID)
			continue
		}

		if p.Equipment.Armor != nil {
			if p.Equipment.Armor.isPlate() {
				// classify melees vs tanks using type substrings PAL or SET1 heuristics from Python
				t := strings.ToUpper(p.Equipment.Armor.Type)
				if strings.Contains(t, "ROYAL") || strings.Contains(t, "SET1") {
					melees = append(melees, playerID)
					continue
				}
				tanks = append(tanks, playerID)
				continue
			}
			if p.Equipment.Armor.isLeather() {
				leathers = append(leathers, playerID)
				continue
			}
			if p.Equipment.Armor.isCloth() {
				cloth = append(cloth, playerID)
				continue
			}
		} else {
			unknown = append(unknown, playerID)
			continue
		}
	}

	// key function similar to Python: use mainhand type or Z if absent
	key := func(id string) string {
		p := b.getPlayer(id)
		if p == nil || p.Equipment.MainHand == nil {
			return "Z"
		}
		return p.Equipment.MainHand.Type
	}

	sort.Slice(cloth, func(i, j int) bool { return key(cloth[i]) < key(cloth[j]) })
	sort.Slice(unknown, func(i, j int) bool { return key(unknown[i]) < key(unknown[j]) })
	sort.Slice(tanks, func(i, j int) bool { return key(tanks[i]) < key(tanks[j]) })
	sort.Slice(melees, func(i, j int) bool { return key(melees[i]) < key(melees[j]) })
	sort.Slice(leathers, func(i, j int) bool { return key(leathers[i]) < key(leathers[j]) })
	sort.Slice(healers, func(i, j int) bool { return key(healers[i]) < key(healers[j]) })

	sorted := []string{}
	sorted = append(sorted, unknown...)
	sorted = append(sorted, tanks...)
	sorted = append(sorted, melees...)
	sorted = append(sorted, leathers...)
	sorted = append(sorted, cloth...)
	sorted = append(sorted, healers...)
	return sorted
}
