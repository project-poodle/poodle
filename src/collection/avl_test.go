package collection

import (
	"fmt"
	"testing"
	"time"
)

type intKey struct {
	int
}

func (k intKey) Compare(k2 IComparable) int {
	if k.int == k2.(intKey).int {
		return 0
	} else if k.int < k2.(intKey).int {
		return -1
	} else {
		return 1
	}
}

func (k intKey) Equal(k2 IObject) bool {
	return k.int == k2.(intKey).int
}

func TestAVL(t *testing.T) {
	tree := &AVLTree{}
	fmt.Println("Empty Tree:")
	//avl, _ := json.MarshalIndent(tree, "", "   ")
	//fmt.Println(string(avl))
	for iter := tree.Iterator(); iter.HasNext(); {
		time.Sleep(300 * time.Millisecond)
		fmt.Println(iter.Next())
	}

	fmt.Println("\nInsert Tree:")
	tree.Put(intKey{4})
	tree.Put(intKey{2})
	tree.Put(intKey{7})
	tree.Put(intKey{6})
	tree.Put(intKey{6})
	tree.Put(intKey{9})
	tree.Put(intKey{5})
	tree.Put(intKey{3})
	tree.Put(intKey{8})
	tree.Put(intKey{3})
	tree.Put(intKey{1})
	//tree.Put(intKey{3})
	//tree.Put(intKey{2})
	//avl, _ = json.MarshalIndent(tree, "", "   ")
	//fmt.Println(string(avl))
	for iter := tree.Iterator(); iter.HasNext(); {
		time.Sleep(300 * time.Millisecond)
		fmt.Println(iter.Next())
	}

	fmt.Println("\nRemove Tree:")
	tree.Remove(intKey{4})
	tree.Remove(intKey{6})
	tree.Remove(intKey{1})
	tree.Remove(intKey{4})
	tree.Remove(intKey{1})
	//avl, _ = json.MarshalIndent(tree, "", "   ")
	//fmt.Println(string(avl))
	for iter := tree.Iterator(); iter.HasNext(); {
		time.Sleep(300 * time.Millisecond)
		fmt.Println(iter.Next())
	}

}
