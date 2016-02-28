package lcs

import (
	"fmt"
	"testing"
)

func Test1(t *testing.T) {
	a := "XMJYAUZZZ"
	b := "MZJAWXU"
	c := a

	offset := 0
	for i, op := range DiffText(a, b) {
		fmt.Printf("%d: %+v = ", i, op)
		if op.DeleteA > 0 {
			c = c[:op.StartA+offset] + c[op.StartA+offset+op.DeleteA:]
			offset -= op.DeleteA
			fmt.Printf("del at [%d] %s", op.StartA, c)
		}
		if op.InsertB > 0 {
			c = c[:op.StartA+offset] + b[op.StartB:op.StartB+op.InsertB] + c[op.StartA+offset:]
			offset += op.InsertB
			fmt.Printf("ins at [%d] %s", op.StartA, c)
		}
		fmt.Println()
	}

	if b != c {
		t.Fatal("patch failed")
	}
}
