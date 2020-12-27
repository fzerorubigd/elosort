package ranking

// Ranker is the ranking interface
type Ranker interface {
	Calculate(rankA, rankB int, scoreA, scoreB float64) (int, int, error)
}
