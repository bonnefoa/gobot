package assert

import "testing"

func AssertEquals(t *testing.T, a interface{}, b interface{}) {
        if a != b {
                t.Logf("Expected '%v', got '%v'\n", a, b)
                t.Fail()
        }
}
