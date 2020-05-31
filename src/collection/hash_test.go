package collection

import (
	"fmt"
	"os"
	"testing"
	"time"
)

var hashPutCases = []struct {
	k  *intKey
	v  *intKey
	rk *intKey
	rv *intKey
	s  int
}{
	{&intKey{4}, &intKey{4 * 2}, nil, nil, 1},
	{&intKey{2}, &intKey{2 * 2}, nil, nil, 2},
	{&intKey{1}, &intKey{1 * 2}, nil, nil, 3},
	{&intKey{7}, &intKey{7 * 2}, nil, nil, 4},
	{&intKey{6}, &intKey{6 * 2}, nil, nil, 5},
	{&intKey{6}, &intKey{6 * 3}, &intKey{6}, &intKey{6 * 2}, 5},
	{&intKey{9}, &intKey{9 * 2}, nil, nil, 6},
	{&intKey{3}, &intKey{3 * 2}, nil, nil, 7},
	{&intKey{5}, &intKey{5 * 2}, nil, nil, 8},
	{&intKey{8}, &intKey{8 * 2}, nil, nil, 9},
	{&intKey{3}, &intKey{3 * 3}, &intKey{3}, &intKey{3 * 2}, 9},
	{&intKey{1}, &intKey{1 * 3}, &intKey{1}, &intKey{1 * 2}, 9},
}

var hashRemoveCases = []struct {
	k  *intKey
	rk *intKey
	rv *intKey
	s  int
}{
	{&intKey{4}, &intKey{4}, &intKey{4 * 2}, 8},
	{&intKey{6}, &intKey{6}, &intKey{6 * 3}, 7},
	{&intKey{1}, &intKey{1}, &intKey{1 * 3}, 6},
	{&intKey{4}, nil, nil, 6},
	{&intKey{1}, nil, nil, 6},
	{&intKey{9}, &intKey{9}, &intKey{9 * 2}, 5},
	{&intKey{9}, nil, nil, 5},
	{&intKey{1}, nil, nil, 5},
}

func TestHash(t *testing.T) {

	hash := NewHash()
	fmt.Println("Empty Hash:")
	hash.Print(os.Stdout, 0)
	for iter := hash.Iterator(); iter.HasNext(); {
		time.Sleep(50 * time.Millisecond)
		fmt.Println(iter.Next().(*HashNode).ToString())
	}

	// put predefined test cases
	fmt.Println("\nPut Cases:")
	for _, c := range hashPutCases {
		node := hash.Put(c.k, c.v)
		if c.s != hash.Size() {
			t.Errorf("Put Failed: param %v %v, expect size %d, got size %d", c.k, c.v, c.s, hash.Size())
			hash.Print(os.Stdout, 0)
		}
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
	hash.Print(os.Stdout, 0)
	for iter := hash.Iterator(); iter.HasNext(); {
		time.Sleep(50 * time.Millisecond)
		fmt.Println(iter.Next().(*HashNode).ToString())
	}

	// remove predefined test cases
	fmt.Println("\nRemove Cases:")
	for _, c := range hashRemoveCases {
		node := hash.Remove(c.k)
		if c.s != hash.Size() {
			t.Errorf("Put Failed: param %v, expect size %d, got size %d", c.k, c.s, hash.Size())
			hash.Print(os.Stdout, 0)
		}
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
	hash.Print(os.Stdout, 0)
	for iter := hash.Iterator(); iter.HasNext(); {
		time.Sleep(50 * time.Millisecond)
		fmt.Println(iter.Next().(*HashNode).ToString())
	}

	// put random
	fmt.Println("\nPut Hash Random:")
	putRandSize := int(randUint32() % 1000)
	for i := 0; i < putRandSize; i++ {
		value := int(randUint32() % 100)
		hash.Put(&intKey{value}, &intKey{value * 5})
	}
	hash.Print(os.Stdout, 0)
	for iter := hash.Iterator(); iter.HasNext(); {
		fmt.Println(iter.Next().(*HashNode).ToString())
	}

	// remove random
	fmt.Println("\nRemove Hash Random:")
	removeRandSize := int(randUint32() % 400)
	for i := 0; i < removeRandSize; i++ {
		value := int(randUint32() % 100)
		hash.Remove(&intKey{value})
	}
	hash.Print(os.Stdout, 0)
	for iter := hash.Iterator(); iter.HasNext(); {
		fmt.Println(iter.Next().(*HashNode).ToString())
	}

}
