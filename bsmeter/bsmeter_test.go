package bsmeter

import (
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"testing"
	"github.com/bonnefoa/gobot/testing/assert"
        "strings"
)

var testConf = BsConf{StateFile:"/tmp/to", StorageGood:"/tmp/good", StorageBad:"/tmp/bad"}

func TestEvaluateBs(t *testing.T) {
	good := map[string]int{"jam": 10}
	bad := map[string]int{"ham": 9, "to": 4}
	probs := map[string]float64{}
	state := BsState{good, bad, probs, BsConf{}}
	state.rebuildProbaMap()

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
        state := BsState{good, bad, probs, testConf}
	state.save()
	loaded := testConf.loadBsState()
	assert.AssertEquals(t, loaded.GoodWords["to"], 10)
}

func TestEnlargeCorpus(t *testing.T) {
	good := map[string]int{"to": 10}
	bad := map[string]int{"ta": 9}
	probs := map[string]float64{}
	state := BsState{good, bad, probs, testConf}

	words := []string{"test"}
	state.enlargeCorpus(words, true)
	assert.AssertEquals(t, state.BadWords["test"], 1)
}

func TestParseFiles(t *testing.T) {
	state := defaultBsState()
	state.trainWithHtmlFile("good/first", false)
	t.Logf("State after training is %v\n", state)
	assert.AssertEquals(t, state.GoodWords["test"], 5)
	assert.AssertFalse(t, math.IsNaN(state.BsProba["test"]))
	assert.AssertFloatInferior(t, state.BsProba["test"], 0.5)

	state.trainWithHtmlFile("bad/first", true)
	t.Logf("State after training is %v\n", state)
	assert.AssertEquals(t, state.BadWords["badtest"], 24)
	assert.AssertFalse(t, math.IsNaN(state.BsProba["badtest"]))
	assert.AssertFalse(t, math.IsNaN(state.BsProba["badtest"]))
	assert.AssertFloatSuperior(t, state.BsProba["badtest"], 0.5)
}

func TestReload(t *testing.T) {
        state := defaultBsState()
        state.BsConf = BsConf{StateFile:"to", StorageGood:"good", StorageBad:"bad"}

	t.Logf("State before reload is %v\n", state)
	state.processReload()
	t.Logf("State after reload is %v\n", state)
	assert.AssertEquals(t, state.GoodWords["test"], 5)
}

func TestLoadAndEvaluate(t *testing.T) {
        state := defaultBsState()
        state.trainWithHtmlFile("bad/second", true)
        res, _ := state.evaluateHtml("bad/second", "bad/second")
        assert.AssertFloatSuperior(t, res.Score, 0.5)
}

func TestTrainPhrases(t *testing.T) {
        state := defaultBsState()
        state.trainWithPhrase("test bs bs bs bs bs", true)
        state.trainWithPhrase("test nobs nobs nobs nobs nobs", false)
        t.Logf("State %v", state)
        res, _ := state.evaluatePhrase("bs")
        assert.AssertFloatSuperior(t, res.Score, 0.5)
        res, _ = state.evaluatePhrase("nobs")
        assert.AssertFloatInferior(t, res.Score, 0.5)
}

func TestReloadPhrases(t *testing.T) {
        os.Remove(testConf.getPhraseStorage(true))
        os.Remove(testConf.getPhraseStorage(false))
        state := defaultBsState()
        state.BsConf = testConf
        state.processTrainQuery(BsQuery{Phrase: "test bs bs bs bs bs", Bs: true})
        state.processTrainQuery(BsQuery{Phrase: "test nobs nobs nobs nobs nobs", Bs: false})
        state.processReload()
        t.Logf("State %v", state)
        res, _ := state.evaluatePhrase("bs")
        assert.AssertFloatSuperior(t, res.Score, 0.5)
        res, _ = state.evaluatePhrase("nobs")
        assert.AssertFloatInferior(t, res.Score, 0.5)
}

func TestParsePdf(t *testing.T) {
        file, _ := os.Open("test.pdf")
        parsedUrl, _ := url.Parse("test.pdf")
        content, _ := ioutil.ReadAll(file)
        f := savePdfToText(parsedUrl, content, "/tmp/test_pdf")
        res, _ := os.Open(f)
        parsedText, _ := ioutil.ReadAll(res)
        assert.AssertEquals(t, strings.TrimSpace(string(parsedText)), "Hello World\n\n1")
}
