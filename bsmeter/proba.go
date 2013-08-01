package bsmeter

import (
	"encoding/json"
	"github.com/bonnefoa/gobot/utils/utilint"
	"log"
	"math"
	"os"
	"sort"
)

const scaleGood = 2
const minOcc = 5
const minProba = 0.1
const maxProba = 0.9
const defaultProba = 0.4

type BsState struct {
	GoodWords map[string]int
	BadWords  map[string]int
	BsProba   map[string]float64
}

func defaultBsState() *BsState {
	bsState := new(BsState)
	bsState.GoodWords = map[string]int{}
	bsState.BadWords = map[string]int{}
	bsState.BsProba = map[string]float64{}
	return bsState
}

func (state *BsState) enlargeCorpus(words []string, isBs bool) {
	log.Printf("Adding %d words to corpus", len(words))
	if isBs {
		for _, word := range words {
			state.BadWords[word] = state.BadWords[word] + 1
		}
	} else {
		for _, word := range words {
			state.GoodWords[word] = state.GoodWords[word] + 1
		}
	}
	state.buildProba()
}

func (state *BsState) computeProbaForWord(word string) {
	occGood := state.GoodWords[word]
	occGood = occGood * scaleGood
	occBad := state.BadWords[word]
	if occGood+occBad < minOcc {
		return
	}
	propGood := math.Min(1, float64(occGood)/math.Max(1, float64(len(state.GoodWords))))
	propBad := math.Min(1, float64(occBad)/math.Max(1, float64(len(state.BadWords))))
	proba := propBad / (propBad + propGood)
	proba = math.Min(maxProba, proba)
	proba = math.Max(minProba, proba)
	state.BsProba[word] = proba
}

func (state *BsState) buildProba() {
	state.BsProba = map[string]float64{}
	for word := range state.GoodWords {
		state.computeProbaForWord(word)
	}
	for word := range state.BadWords {
		state.computeProbaForWord(word)
	}
}

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

func (state *BsState) EvaluateBs(words []string) float64 {
	prbs := probs{}
	for _, word := range words {
		proba, found := state.BsProba[word]
		if !found {
			proba = 0.4
		}
		prbs = append(prbs, prob{word, proba})
	}
	sort.Sort(prbs)
	prbs = prbs[:utilint.MinInt(15, len(prbs))]
	log.Printf("Most significant probas are %v\n", prbs)
	return prbs.Combined()
}

func (state *BsState) save(filename string) {
	file, _ := os.OpenFile(filename,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 400)
	enc := json.NewEncoder(file)
	enc.Encode(state)
	file.Close()
}

func loadBsState(filename string) *BsState {
	file, err := os.Open(filename)
	bsState := defaultBsState()
	if err != nil {
		return bsState
	}
	defer file.Close()
	dec := json.NewDecoder(file)
	dec.Decode(bsState)
	return bsState
}
