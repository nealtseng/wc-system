package teamdata

// ModelWeights holds the current prediction model weight configuration.
type ModelWeights struct {
	W1, W2, W3 float64
	ClipDelta  float64
	DeltaMax   float64
	KellyScale float64
}

func defaultWeights() ModelWeights {
	return ModelWeights{W1: 0.30, W2: 0.30, W3: 0.40, ClipDelta: 0.05, DeltaMax: 0.15, KellyScale: 0.25}
}
