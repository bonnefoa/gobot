package utilstring

import "testing"
import "testing/assert"

func TestReverse(t *testing.T) {
        expected := "ƃɐ è ƃɐ"
        r := RotateString("ga è ga")
        t.Logf("Got %q, expected %q\n", r, expected)
        assert.AssertEquals(t, expected, r)
}
