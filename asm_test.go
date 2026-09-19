package main

import (
	"testing"
)

func TestGetLabel(t *testing.T) {
	tokens := []string{"test:"}
	label := getLabel(tokens)

	if *label != "test" {
		t.Errorf("")
	}
}

func TestGetLabelWithNolabel(t *testing.T) {
	tokens := []string{"brg 1"}
	label := getLabel(tokens)

	if label != nil {
		t.Errorf("")
	}
}

