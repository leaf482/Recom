package experiments

type Weights struct {
	Genre      float64
	Artist     float64
	Popularity float64
	Freshness  float64
}

func WeightsForStrategy(strategy string) Weights {
	switch strategy {
	case StrategyArtistAffinity:
		return Weights{
			Genre:      0.25,
			Artist:     0.45,
			Popularity: 0.20,
			Freshness:  0.10,
		}
	case StrategyExploration:
		return Weights{
			Genre:      0.20,
			Artist:     0.10,
			Popularity: 0.35,
			Freshness:  0.35,
		}
	default:
		return Weights{
			Genre:      0.55,
			Artist:     0.20,
			Popularity: 0.15,
			Freshness:  0.10,
		}
	}
}
