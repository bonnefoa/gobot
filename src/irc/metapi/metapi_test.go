package metapi

import "testing"
import "testing/assert"

func TestPiDecimal(t *testing.T) {
        t.Logf("pi is %q", estimatePi(40).FloatString(10000))
        t.Logf("pi is %q", estimatePi(80).FloatString(10000))
        t.Logf("pi is %q", estimatePi(400).FloatString(10000))

        assert.AssertEquals(t, 1, 2)
}
