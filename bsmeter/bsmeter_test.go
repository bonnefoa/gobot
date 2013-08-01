package bsmeter

import (
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"testing"
	"testing/assert"
)

func TestExtractUrls(t *testing.T) {
	urls := ExtractUrls("test http://w1 test 2 https://w2 https://imgur.com/toto.jpg")
	assert.AssertStringSliceEquals(t, urls,
		[]string{"http://w1", "https://w2", "https://imgur.com/toto"})
}

func TestEvaluateBs(t *testing.T) {
	good := map[string]int{"jam": 10}
	bad := map[string]int{"ham": 9, "to": 4}
	probs := map[string]float64{}
	state := BsState{good, bad, probs}
	state.buildProba()

	res := state.EvaluateBs([]string{"toto"})
	assert.AssertFloatInferior(t, res, 0.5)

	res = state.EvaluateBs([]string{"jam"})
	assert.AssertFloatInferior(t, res, 0.5)

	res = state.EvaluateBs([]string{"ham"})
	assert.AssertFloatSuperior(t, res, 0.5)
}

func TestSaveState(t *testing.T) {
	good := map[string]int{"to": 10}
	bad := map[string]int{"ta": 9}
	probs := map[string]float64{}
	state := BsState{good, bad, probs}
	state.save("to")
	loaded := loadBsState("to")
	assert.AssertEquals(t, loaded.GoodWords["to"], 10)
}

func TestEnlargeCorpus(t *testing.T) {
	good := map[string]int{"to": 10}
	bad := map[string]int{"ta": 9}
	probs := map[string]float64{}
	state := BsState{good, bad, probs}

	words := []string{"test"}
	state.enlargeCorpus(words, true)
	assert.AssertEquals(t, state.BadWords["test"], 1)
}

func TestParseFiles(t *testing.T) {
	state := defaultBsState()
	state.trainWithHtmlFile("good/first", false)
	t.Logf("State after training is %v\n", state)
	assert.AssertEquals(t, state.GoodWords["the"], 8)
	assert.AssertFalse(t, math.IsNaN(state.BsProba["the"]))
	assert.AssertFloatInferior(t, state.BsProba["the"], 0.5)

	state.trainWithHtmlFile("bad/first", true)
	t.Logf("State after training is %v\n", state)
	assert.AssertEquals(t, state.BadWords["learn"], 5)
	assert.AssertFalse(t, math.IsNaN(state.BsProba["learn"]))
	assert.AssertFalse(t, math.IsNaN(state.BsProba["learn"]))
	assert.AssertFloatSuperior(t, state.BsProba["learn"], 0.5)
}

func TestReload(t *testing.T) {
	state := defaultBsState()
	state.processReload()
	t.Logf("State after reload is %v\n", state)
	assert.AssertEquals(t, state.GoodWords["the"], 8)
}

func TestLoadAndEvaluate(t *testing.T) {
	state := defaultBsState()
	state.trainWithHtmlFile("bad/second", true)
	res := state.evaluateHtmlFile("bad/second")
	assert.AssertFloatSuperior(t, res.Score, 0.5)
}

func TestTrainPhrases(t *testing.T) {
	state := defaultBsState()
	state.trainWithPhrase("test bs bs bs bs bs", true)
	state.trainWithPhrase("test nobs nobs nobs nobs nobs", false)
	t.Logf("State %v", state)
	res := state.evaluatePhrase("bs")
	assert.AssertFloatSuperior(t, res.Score, 0.5)
	res = state.evaluatePhrase("nobs")
	assert.AssertFloatInferior(t, res.Score, 0.5)
}

func TestReloadPhrases(t *testing.T) {
	os.Remove(getPhraseStorage(true))
	os.Remove(getPhraseStorage(false))
	state := defaultBsState()
	state.processTrainQuery(BsQuery{Phrase: "test bs bs bs bs bs", Bs: true})
	state.processTrainQuery(BsQuery{Phrase: "test nobs nobs nobs nobs nobs", Bs: false})
	state.processReload()
	t.Logf("State %v", state)
	res := state.evaluatePhrase("bs")
	assert.AssertFloatSuperior(t, res.Score, 0.5)
	res = state.evaluatePhrase("nobs")
	assert.AssertFloatInferior(t, res.Score, 0.5)
}

func TestParsePdf(t *testing.T) {
	file, _ := os.Open("test.pdf")
	parsedUrl, _ := url.Parse("http://test.pdf")
	content, _ := ioutil.ReadAll(file)
	f := savePdf(parsedUrl, content, "/tmp/test_pdf")

	res, _ := os.Open(f)
	parsedText, _ := ioutil.ReadAll(res)
	t.Logf("Parsed text is %s\n", parsedText)
	assert.AssertNotEquals(t, parsedText, "")
	t.Fail()
}
