package bsmeter

import (
	"math"
)

const scaleGood = 2
const minOcc = 5
const minProba = 0.1
const maxProba = 0.9
const defaultProba = 0.4

type prob struct {
	word  string
	proba float64
}

type probs []prob

func (p probs) Len() int {
	return len(p)
}

func (p probs) Less(i, j int) bool {
	dist1 := math.Abs(p[i].proba - 0.5)
	dist2 := math.Abs(p[j].proba - 0.5)
	return dist1 > dist2
}

func (p probs) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p probs) Combined() float64 {
	num := 1.0
	denum := 1.0
	for _, prob := range p {
		num *= prob.proba
		denum *= (1 - prob.proba)
	}
	return num / (num + denum)
}
