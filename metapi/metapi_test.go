package metapi

import (
        "testing"
	"github.com/bonnefoa/gobot/testing/assert"
)

func TestConcatInt(t *testing.T) {
	assert.AssertIntSliceEquals(t, []int{1, 2},
		concatInt([]int{1}, []int{2}))
}

func checkReduce(t *testing.T, num, denum, expectedNum, expectedDenum []int) {
	resNum, resDenum := reduceSlices(num, denum)
	assert.AssertIntSliceEquals(t, resNum, expectedNum)
	assert.AssertIntSliceEquals(t, resDenum, expectedDenum)
}

func TestReduceSlices(t *testing.T) {
	checkReduce(t, []int{1}, []int{1}, []int{}, []int{})
	checkReduce(t, []int{1}, []int{2}, []int{1}, []int{2})
	checkReduce(t, []int{1, 3}, []int{3, 2}, []int{1}, []int{2})
	checkReduce(t, []int{1, 2, 3}, []int{4, 2, 0}, []int{1, 3}, []int{0, 4})
}

func TestPiDecimal(t *testing.T) {
	pi, _ := EstimatePi(10)
	assert.AssertEquals(t, "3.142", pi.FloatString(3))
}
