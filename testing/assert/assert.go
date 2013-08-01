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

func AssertTrue(t *testing.T, cond bool) {
        if !cond {
                t.Logf("Excepted true, got %#v\n", cond)
                t.FailNow()
        }
}

func AssertFalse(t *testing.T, cond bool) {
        if cond {
                t.Logf("Excepted false, got %#v\n", cond)
                t.FailNow()
        }
}

func AssertEquals(t *testing.T, a interface{}, b interface{}) {
        if a != b {
                t.Logf("%#v != %#v\n", a, b)
                t.FailNow()
        }
}

func AssertNotEquals(t *testing.T, a interface{}, b interface{}) {
        if a == b {
                t.Logf("%#v == %#v\n", a, b)
                t.FailNow()
        }
}

func AssertFloatSuperior(t *testing.T, a float64, b float64) {
        if a < b {
                t.Logf("%#v < %#v\n", a, b)
                t.FailNow()
        }
}

func AssertFloatInferior(t *testing.T,a float64, b float64) {
        if a > b {
                t.Logf("%#v > %#v\n", a, b)
                t.FailNow()
        }
}

func AssertIntSliceEquals(t *testing.T, a []int, b []int) {
        if len(a) != len(b) {
                t.Logf("%#v != %#v\n", a, b)
                t.FailNow()
        }
        for i, _ := range a {
                if a[i] != b[i] {
                        t.Logf("%#v != %#v\n", a, b)
                        t.FailNow()
                }
        }
}

func AssertStringSliceEquals(t *testing.T, a []string, b []string) {
        if len(a) != len(b) {
                t.Logf("%#v != %#v\n", a, b)
                t.FailNow()
        }
        for i, _ := range a {
                if a[i] != b[i] {
                        t.Logf("%#v != %#v\n", a, b)
                        t.FailNow()
                }
        }
}

func AssertBytesEquals(t *testing.T, a, b []byte) {
        if bytes.Compare(a, b) != 0 {
                t.Logf("%#v != %#v\n", a, b)
                t.FailNow()
        }
}
