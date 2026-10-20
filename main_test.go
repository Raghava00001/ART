package main

import (
	"testing"
)

func TestFlagDefaults(t *testing.T) {
	if Alpha != 128 {
		t.Errorf("Expected default Alpha=128, got %d", Alpha)
	}
	if Mode != 1 {
		t.Errorf("Expected default Mode=1, got %d", Mode)
	}
	if Port != 8080 {
		t.Errorf("Expected default Port=8080, got %d", Port)
	}
}
