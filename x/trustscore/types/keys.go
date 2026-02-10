package types

const (
	// ModuleName defines the module name
	ModuleName = "trustscore"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// Store key prefixes for the trustscore module
var (
	// ScoreKeyPrefix is the prefix for trust score storage
	// Format: ScoreKeyPrefix | Address -> HumanityResult
	ScoreKeyPrefix = []byte{0x01}

	// ConfigKeyPrefix is the prefix for trust score config
	ConfigKeyPrefix = []byte{0x02}
)

// GetScoreKey returns the store key for a trust score
func GetScoreKey(address string) []byte {
	return append(ScoreKeyPrefix, []byte(address)...)
}
