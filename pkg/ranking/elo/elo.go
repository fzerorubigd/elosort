package elo

import (
	"errors"
	"math"

	"elbix.dev/elosort/pkg/ranking"
)

const (
	defaultK         = 32
	defaultDeviation = 400
)

type eloRanker struct {
	// k factor , it is normally 32 or 16
	k float64
	// deviation , the default is 400
	deviation float64
}

func (e *eloRanker) Calculate(rankA, rankB int, scoreA, scoreB float64) (int, int, error) {
	if scoreA+scoreB != 1 {
		return 0, 0, errors.New("sum of scores should be one")
	}

	qA := math.Pow(10, float64(rankA)/e.deviation)
	qB := math.Pow(10, float64(rankB)/e.deviation)

	eA := qA / (qA + qB)
	eB := qB / (qB + qA)

	newRankA := rankA + int(e.k*(scoreA-eA))
	newRankB := rankB + int(e.k*(scoreB-eB))

	return newRankA, newRankB, nil
}

// NewEloRankerDefault return a ranker with default options
func NewEloRankerDefault() ranking.Ranker {
	return &eloRanker{
		k:         defaultK,
		deviation: defaultDeviation,
	}
}
