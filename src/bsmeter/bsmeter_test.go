package bsmeter

import "testing"
import "testing/assert"

func TestExtractUrls(t *testing.T) {
        urls := ExtractUrls("test http://w1 test 2 https://w2 https://w2.jpg")
        assert.AssertStringSliceEquals(t, urls,
            []string{"http://w1", "https://w2", "https://w2})
}

func TestLookupTitle(t *testing.T) {
        title := LookupTitle("https://github.com")
        assert.AssertEquals(t, title, "GitHub · Build software better, together.")
}
