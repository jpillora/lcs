package lcs

import (
	"log"
	"testing"
)

func Test1(t *testing.T) {
	a := "XMJYAUZZZ"
	b := "MZJAWXU"

	for i, op := range DiffText(a, b) {
		log.Printf("%d: %+v", i, op)
	}
}
