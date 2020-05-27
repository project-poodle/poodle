package pdb

import (
	"fmt"
	"testing"

	"../util"
)

func TestGet(t *testing.T) {

	d := util.RandInt32()

	fmt.Printf("RandInt32 : %x\n", d)
}
