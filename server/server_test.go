package main

import "testing"

func TestHarness(t *testing.T) {
	if 1 == 2 {
		t.Fatalf("Math stopped working")
	}
}
