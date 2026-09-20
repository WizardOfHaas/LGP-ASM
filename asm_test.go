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

func TestIsNumeric(t *testing.T){
	numbers := []string{"1", "0", "12345"}
	notNumbers := []string{"l", "A", "WORD"}

	for _, s := range numbers{
		if !isNumeric(s){
			t.Errorf("%s should be seen as numeric", s)
		}
	}

	for _, s := range notNumbers{
		if isNumeric(s){
			t.Errorf("%s should NOT be seen as numeric", s)
		}
	}
}

func TestIsBinary(t *testing.T){
	binary := []string{"0b1", "0B0", "0b01010"}
	notBinary := []string{"l", "01110", "0bBBBas", "0B1120", "0b", "0B"}

	for _, s := range binary{
		if !isBinary(s){
			t.Errorf("%s should be seen as binary", s)
		}
	}

	for _, s := range notBinary{
		if isBinary(s){
			t.Errorf("%s should NOT be seen as binary", s)
		}
	}
}

func TestIsStdHex(t *testing.T){
	hex := []string{"0x1", "0X0", "0x12FE"}
	notHex := []string{"l", "011FA0", "0xdfdsf3FSAD", "0xWJ01", "0x"}

	for _, s := range hex{
		if !isStdHex(s){
			t.Errorf("%s should be seen as hex", s)
		}
	}

	for _, s := range notHex{
		if isStdHex(s){
			t.Errorf("%s should NOT be seen as hex", s)
		}
	}
}

func TestIsLGPHex(t *testing.T){
	hex := []string{"0l1", "0L0", "0l1GFJ", "0lWJ01"}
	notHex := []string{"l", "011FA0", "0xdfdsf3FSAD", "0l12AA", "0l"}

	for _, s := range hex{
		if !isLgpHex(s){
			t.Errorf("%s should be seen as LGP hex", s)
		}
	}

	for _, s := range notHex{
		if isLgpHex(s){
			t.Errorf("%s should NOT be seen as LGP hex", s)
		}
	}
}


func TestIsSymbolic(t *testing.T){
	symbolic := []string{"hello", "AYO", "TEST_01"}
	notSymbolic := []string{"123", "0x123"}

	for _, s := range symbolic{
		if !isSymbolic(s){
			t.Errorf("%s should be seen as a symbol", s)
		}
	}

	for _, s := range notSymbolic{
		if isSymbolic(s){
			t.Errorf("%s should NOT be seen as a symbol", s)
		}
	}
}
