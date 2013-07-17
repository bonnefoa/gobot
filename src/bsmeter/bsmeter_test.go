package bsmeter

import (
    "testing"
    "testing/assert"
)

func TestExtractUrls(t *testing.T) {
        urls := ExtractUrls("test http://w1 test 2 https://w2 https://w2.jpg")
        assert.AssertStringSliceEquals(t, urls,
            []string{"http://w1", "https://w2", "https://w2"})
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
        checkTitle(t, "http://xkcd.com",
            "xkcd: Social Media")
}

