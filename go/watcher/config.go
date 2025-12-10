package watcher

// Config constants ported from Python config.py (subset used by watcher)
const (
	RateLimitDelaySeconds     = 0.5
	TimeoutSeconds            = 30
	BattlesLimit              = 20
	BattlesMaxAgeMinutes      = 2
	BattleCheckIntervalMins   = 1

	BaseURLEurope   = "https://gameinfo-ams.albiononline.com/api/gameinfo"
	BaseURLAmericas = "https://gameinfo.albiononline.com/api/gameinfo"
	BaseURLAsia     = "https://gameinfo-sgp.albiononline.com/api/gameinfo"
)

var ServerURLs = map[string]string{
	"europe":   BaseURLEurope,
	"americas": BaseURLAmericas,
	"asia":     BaseURLAsia,
}
