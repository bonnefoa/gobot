package bsmeter

import (
    "testing"
    "testing/assert"
)

func TestExtractUrls(t *testing.T) {
        urls := ExtractUrls("test http://w1 test 2 https://w2 https://imgur.com/toto.jpg")
        assert.AssertStringSliceEquals(t, urls,
        []string{"http://w1", "https://w2", "https://imgur.com/toto"})
}

func checkTitle(t *testing.T, url, title string) {
        res, found := LookupTitle(url)
        assert.AssertEquals(t, true, found)
        assert.AssertEquals(t, res, title)
}

func TestLookupTitle(t *testing.T) {
        checkTitle(t, "http://i.imgur.com/EHILhaP",
        "Starcraft Units to scale - Imgur")
        checkTitle(t, "https://github.com/",
        "GitHub · Build software better, together.")
        checkTitle(t, "http://xkcd.com/1239/",
        "xkcd: Social Media")
}

func TestEvaluateBs(t *testing.T) {
        good := map[string]int{"jam": 10}
        bad := map[string]int{"ham": 9, "to":4}
        probs := map[string]float64{}
        emptyList := []string{}
        state := BsState{good, bad, probs, emptyList, emptyList}
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
        emptyList := []string{}
        state := BsState{good, bad, probs, emptyList, emptyList}


        saveBsState(&state)
        loaded := loadBsState()
        assert.AssertEquals(t, loaded.GoodWords["to"], 10)
}

func TestEnlargeCorpus(t *testing.T) {
        good := map[string]int{"to": 10}
        bad := map[string]int{"ta": 9}
        probs := map[string]float64{}
        emptyList := []string{}
        state := BsState{good, bad, probs, emptyList, emptyList}

        urls := []string{"https://ga"}
        bsQuery := BsQuery{urls, true, true, "test"}

        words := []string{"test"}
        enlargeCorpus(words, bsQuery, &state)
        assert.AssertEquals(t, state.BadWords["test"], 1)
        assert.AssertStringSliceEquals(t, state.BadUrls, urls)
}
