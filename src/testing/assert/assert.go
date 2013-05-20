package assert

import "testing"
import "bytes"

func AssertMapEquals(t *testing.T, a map[string] int, b map[string] int) {
        for k, _ := range a {
                if a[k] != b[k] {
                        t.Logf("%#v != %#v\n", a, b)
                        t.FailNow()
                }
        }
}

func AssertNotNil(t *testing.T, a interface{}) {
        if a != nil {
                t.Logf("Expected not nil pointer, got %#v\n", a)
                t.FailNow()
        }
}

func AssertEquals(t *testing.T, a interface{}, b interface{}) {
        if a != b {
                t.Logf("%#v != %#v\n", a, b)
                t.FailNow()
        }
}

func AssertBytesEquals(t *testing.T, a, b []byte) {
        if bytes.Compare(a, b) != 0 {
                t.Logf("%#v != %#v\n", a, b)
                t.FailNow()
        }
}
