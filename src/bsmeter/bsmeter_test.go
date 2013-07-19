package bsmeter

import (
    "testing"
    "testing/assert"
    "math"
)

func TestExtractUrls(t *testing.T) {
        urls := ExtractUrls("test http://w1 test 2 https://w2 https://imgur.com/toto.jpg")
        assert.AssertStringSliceEquals(t, urls,
        []string{"http://w1", "https://w2", "https://imgur.com/toto"})
}

func TestEvaluateBs(t *testing.T) {
        good := map[string]int{"jam": 10}
        bad := map[string]int{"ham": 9, "to":4}
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
        state.trainWithFile("good/first", false)
        t.Logf("State after training is %v\n", state)
        assert.AssertEquals(t, state.GoodWords["the"], 8)
        assert.AssertFalse(t, math.IsNaN(state.BsProba["the"]))
        assert.AssertFloatInferior(t, state.BsProba["the"], 0.5)

        state.trainWithFile("bad/first", true)
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
        state.trainWithFile("bad/second", true)
        res := state.evaluateFile("bad/second")
        assert.AssertFloatSuperior(t, res.Score, 0.5)
}
