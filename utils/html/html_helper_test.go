package html

import (
	"testing"
	"github.com/bonnefoa/gobot/testing/assert"
)

func TestExtractUrls(t *testing.T) {
	urls := ExtractUrls("test http://w1 test 2 https://w2 https://imgur.com/toto.jpg")
	assert.AssertStringSliceEquals(t, urls,
		[]string{"http://w1", "https://w2", "https://imgur.com/toto"})
}

