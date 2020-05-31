package collection

import (
	"fmt"
	"os"
	"testing"
	"time"
)

type intKey struct {
	int
}

func (k *intKey) HashUint32(f func([]byte) uint32) uint32 {
	return f([]byte{byte(k.int), byte(k.int >> 8), byte(k.int >> 16), byte(k.int >> 24)})
}

func (k *intKey) Compare(k2 IComparable) int {
	if k.int == k2.(*intKey).int {
		return 0
	} else if k.int < k2.(*intKey).int {
		return -1
	} else {
		return 1
	}
}

func (k *intKey) Equal(k2 IObject) bool {
	if IsNil(k2) {
		return false
	}
	return k.int == k2.(*intKey).int
}

var avlPutCases = []struct {
	k  *intKey
	v  *intKey
	rk *intKey
	rv *intKey
}{
	{&intKey{4}, &intKey{4 * 2}, nil, nil},
	{&intKey{2}, &intKey{2 * 2}, nil, nil},
	{&intKey{1}, &intKey{1 * 2}, nil, nil},
	{&intKey{7}, &intKey{7 * 2}, nil, nil},
	{&intKey{6}, &intKey{6 * 2}, nil, nil},
	{&intKey{6}, &intKey{6 * 3}, &intKey{6}, &intKey{6 * 2}},
	{&intKey{9}, &intKey{9 * 2}, nil, nil},
	{&intKey{3}, &intKey{3 * 2}, nil, nil},
	{&intKey{5}, &intKey{5 * 2}, nil, nil},
	{&intKey{8}, &intKey{8 * 2}, nil, nil},
	{&intKey{3}, &intKey{3 * 3}, &intKey{3}, &intKey{3 * 2}},
	{&intKey{1}, &intKey{1 * 3}, &intKey{1}, &intKey{1 * 2}},
}

var avlRemoveCases = []struct {
	k  *intKey
	rk *intKey
	rv *intKey
}{
	{&intKey{4}, &intKey{4}, &intKey{4 * 2}},
	{&intKey{6}, &intKey{6}, &intKey{6 * 3}},
	{&intKey{1}, &intKey{1}, &intKey{1 * 3}},
	{&intKey{4}, nil, nil},
	{&intKey{1}, nil, nil},
	{&intKey{9}, &intKey{9}, &intKey{9 * 2}},
	{&intKey{9}, nil, nil},
	{&intKey{1}, nil, nil},
}

func TestAVL(t *testing.T) {
	tree := NewAVLTree()
	fmt.Println("Empty Tree:")
	tree.Print(os.Stdout, 0)
	for iter := tree.Iterator(); iter.HasNext(); {
		time.Sleep(50 * time.Millisecond)
		fmt.Println(iter.Next().(*AVLNode).ToString())
	}

	// put predefined test cases
	fmt.Println("\nPut Cases:")
	for _, c := range avlPutCases {
		node := tree.Put(c.k, c.v)
		if IsNil(node) {
			if c.rk != nil || c.rv != nil {
				t.Errorf("Put Failed: param %v %v, expect %v, %v, got %v, %v", c.k, c.v, c.rk, c.rv, nil, nil)
			}
		} else {
			if (node.key == nil && c.rk != nil) || (node.key != nil && !node.key.Equal(c.rk)) {
				t.Errorf("Put Failed: param %v, %v, expect %v %v, got %v %v", c.k, c.v, c.rk, c.rv, node.key, node.value)
			}
			if (node.value == nil && c.rv != nil) || (node.value != nil && !node.value.(*intKey).Equal(c.rv)) {
				t.Errorf("Put Failed: param %v, %v, expect %v %v, got %v %v", c.k, c.v, c.rk, c.rv, node.key, node.value)
			}
		}
	}
	fmt.Println("Put Completed!")
	tree.Print(os.Stdout, 0)
	for iter := tree.Iterator(); iter.HasNext(); {
		time.Sleep(50 * time.Millisecond)
		fmt.Println(iter.Next().(*AVLNode).ToString())
	}

	// remove predefined test cases
	fmt.Println("\nRemove Cases:")
	for _, c := range avlRemoveCases {
		node := tree.Remove(c.k)
		if IsNil(node) {
			if c.rk != nil || c.rv != nil {
				t.Errorf("Remove Failed: param %v, expect %v %v, got %v %v", c.k, c.rk, c.rv, nil, nil)
			}
		} else {
			if (node.key == nil && c.rk != nil) || (node.key != nil && !node.key.Equal(c.rk)) {
				t.Errorf("Remove Failed: param %v, expect %v %v, got %v %v", c.k, c.rk, c.rv, node.key, node.value)
			}
			if (node.value == nil && c.rv != nil) || (node.value != nil && !node.value.(*intKey).Equal(c.rv)) {
				t.Errorf("Remove Failed: param %v, expect %v %v, got %v %v", c.k, c.rk, c.rv, node.key, node.value)
			}
		}
	}
	fmt.Println("Remove Completed!")
	tree.Print(os.Stdout, 0)
	for iter := tree.Iterator(); iter.HasNext(); {
		time.Sleep(50 * time.Millisecond)
		fmt.Println(iter.Next().(*AVLNode).ToString())
	}

	// put random
	fmt.Println("\nPut Tree Random:")
	putRandSize := int(randUint32() % 500)
	for i := 0; i < putRandSize; i++ {
		value := int(randUint32() % 100)
		tree.Put(&intKey{value}, &intKey{value * 5})
	}
	tree.Print(os.Stdout, 0)
	for iter := tree.Iterator(); iter.HasNext(); {
		fmt.Println(iter.Next().(*AVLNode).ToString())
	}

	// remove random
	fmt.Println("\nRemove Tree Random:")
	removeRandSize := int(randUint32() % 400)
	for i := 0; i < removeRandSize; i++ {
		value := int(randUint32() % 100)
		tree.Remove(&intKey{value})
	}
	tree.Print(os.Stdout, 0)
	for iter := tree.Iterator(); iter.HasNext(); {
		fmt.Println(iter.Next().(*AVLNode).ToString())
	}

}
