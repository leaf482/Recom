package experiments

import "hash/fnv"

const (
	StrategyGenreAffinity  = "genre_affinity"
	StrategyArtistAffinity = "artist_affinity"
	StrategyExploration    = "exploration"
	DefaultExperimentID    = "default"
)

var Strategies = []string{
	StrategyGenreAffinity,
	StrategyArtistAffinity,
	StrategyExploration,
}

func AssignStrategy(userID string) string {
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(userID))
	index := hash.Sum32() % uint32(len(Strategies))
	return Strategies[index]
}
