package main

import "testing"

func TestBot(t *testing.T) {
        bs := []byte("Test")
        t.Fatal(string(bs[:3]))
}
