package assert

import "testing"

func AssertMapEquals(t *testing.T, a map[string] int, b map[string] int) {
        for k, _ := range a {
                if a[k] != b[k] {
                        t.Fatal("'%v' != '%v'\n", a, b)
                }
        }
}

func AssertNotNil(t *testing.T, a interface{}) {
        if a != nil {
                t.Fatal("Expected not nil pointer, got '%v'\n", a)
        }
}

func AssertEquals(t *testing.T, a interface{}, b interface{}) {
        if a != b {
                t.Fatal("Expected '%v', got '%v'\n", a, b)
        }
}
