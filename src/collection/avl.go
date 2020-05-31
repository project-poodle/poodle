package collection

import (
	"fmt"
	"io"
	"strconv"
)

// This code is an adoption of the original implementation at:
// https://www.golangprograms.com/golang-program-for-implementation-of-avl-trees.html

type AVLTree struct {
	root *AVLNode
}

type AVLNode struct {
	key     IComparable
	value   IObject
	balance int
	link    [2]*AVLNode
}

type AVLIterator struct {
	paths    []*AVLNode
	currNode *AVLNode
	currPos  int
}

// create new AVL tree
func NewAVLTree() *AVLTree {
	return &AVLTree{}
}

// Get a node from the AVL tree.
func (t *AVLTree) Get(data IComparable) *AVLNode {
	if t.root == nil {
		return nil
	}

	return t.root.getR(data)
}

// Put a node into the AVL tree.
func (t *AVLTree) Put(key IComparable, value IObject) *AVLNode {
	if key == nil {
		panic("AVLTree::Put - key is nil")
	} else if t.root == nil {
		t.root = &AVLNode{key: key, value: value}
		return nil
	} else {
		var node *AVLNode
		t.root, node, _ = t.root.putR(key, value)
		return node
	}
}

// Remove a single item from an AVL tree.
func (t *AVLTree) Remove(key IComparable) *AVLNode {
	if key == nil {
		panic("AVLTree::Remove - data is nil")
	} else {
		var node *AVLNode
		t.root, node, _ = t.root.removeR(key)
		return node
	}
}

func (i *AVLIterator) Next() IObject {

	if len(i.paths) == 0 && i.currNode == nil {
		return nil
	}

	switch i.currPos {

	case 0:
		for i.currNode.link[0] != nil {
			i.paths = append(i.paths, i.currNode)
			i.currNode = i.currNode.link[0]
		}
		resultNode := i.currNode
		i.currPos = 1
		if i.currNode.link[1] != nil {
			i.currNode = i.currNode.link[1]
			i.currPos = 0
		}
		return resultNode

	case 1:
		if len(i.paths) != 0 {
			i.currNode = i.paths[len(i.paths)-1]
			i.currPos = 0
			i.paths = i.paths[:len(i.paths)-1]
		}
		if len(i.paths) == 0 && i.currPos == 1 {
			i.currNode = nil
			return nil
		} else {
			resultNode := i.currNode
			i.currPos = 1
			if i.currNode.link[1] != nil {
				i.currNode = i.currNode.link[1]
				i.currPos = 0
			}
			return resultNode
		}

	//case 2:
	default:
		panic(fmt.Sprintf("AVLIterator::Next - unknown currPos %d", i.currPos))
	}
}

func (i *AVLIterator) HasNext() bool {

	isEmpty := len(i.paths) == 0 && (i.currNode == nil || i.currPos == 1)

	return !isEmpty
}

// Return an iterator of the AVL tree.
func (t *AVLTree) Iterator() IIterator {

	iter := &AVLIterator{paths: []*AVLNode{}, currNode: t.root, currPos: 0}

	return iter
}

func (t *AVLTree) Print(w io.Writer, indent int) {
	fmt.Fprintf(w, "%"+strconv.Itoa(indent)+"vAVLTree:\n", "")
	if t.root != nil {
		t.root.Print(w, indent+4)
	}
}

func (t *AVLTree) ToString() string {
	return fmt.Sprintf("AVLTree: %s", Ternary(t.root == nil, nil, t.root.ToString()))
}

func opp(dir int) int {
	return 1 - dir
}

// single rotation
func (root *AVLNode) single(dir int) *AVLNode {
	save := root.link[opp(dir)]
	root.link[opp(dir)] = save.link[dir]
	save.link[dir] = root
	return save
}

// double rotation
func (root *AVLNode) double(dir int) *AVLNode {
	save := root.link[opp(dir)].link[dir]

	root.link[opp(dir)].link[dir] = save.link[opp(dir)]
	save.link[opp(dir)] = root.link[opp(dir)]
	root.link[opp(dir)] = save

	save = root.link[opp(dir)]
	root.link[opp(dir)] = save.link[dir]
	save.link[dir] = root
	return save
}

// adjust valance factors after double rotation
func (root *AVLNode) adjustBalance(dir, bal int) {
	n := root.link[dir]
	nn := n.link[opp(dir)]
	switch nn.balance {
	case 0:
		root.balance = 0
		n.balance = 0
	case bal:
		root.balance = -bal
		n.balance = 0
	default:
		root.balance = 0
		n.balance = bal
	}
	nn.balance = 0
}

func (root *AVLNode) putBalance(dir int) *AVLNode {
	n := root.link[dir]
	bal := 2*dir - 1
	if n.balance == bal {
		root.balance = 0
		n.balance = 0
		return root.single(opp(dir))
	}
	root.adjustBalance(dir, bal)
	return root.double(opp(dir))
}

// returns
// *AVLNode new node
// *AVLNode old node if matched
// bool whether balanced
func (root *AVLNode) putR(key IComparable, value IObject) (*AVLNode, *AVLNode, bool) {
	if root == nil {
		result := &AVLNode{key: key, value: value}
		return result, nil, false
	}
	if root.key.Equal(key) {
		found := &AVLNode{key: root.key, value: root.value}
		root.key = key
		root.value = value
		return root, found, true // return new node, old node, and balance flag
	}
	dir := 0
	if root.key.Compare(key) < 0 {
		dir = 1
	}
	var node *AVLNode
	var done bool
	root.link[dir], node, done = root.link[dir].putR(key, value)
	if done {
		return root, node, true
	}
	root.balance += 2*dir - 1
	switch root.balance {
	case 0:
		return root, node, true
	case 1, -1:
		return root, node, false
	}
	return root.putBalance(dir), node, true
}

func (root *AVLNode) removeBalance(dir int) (*AVLNode, bool) {
	n := root.link[opp(dir)]
	bal := 2*dir - 1
	switch n.balance {
	case -bal:
		root.balance = 0
		n.balance = 0
		return root.single(dir), false
	case bal:
		root.adjustBalance(opp(dir), -bal)
		return root.double(dir), false
	}
	root.balance = -bal
	n.balance = bal
	return root.single(dir), true
}

// returns
// *AVLNode new node
// *AVLNode old node if matched
// bool whether balanced
func (root *AVLNode) removeR(key IComparable) (*AVLNode, *AVLNode, bool) {
	if root == nil {
		//return nil, false
		return nil, nil, true
	}
	var found *AVLNode
	if root.key.Equal(key) {
		found = &AVLNode{key: root.key, value: root.value}
		switch {
		case root.link[0] == nil:
			return root.link[1], found, false
		case root.link[1] == nil:
			return root.link[0], found, false
		}
		heir := root.link[0]
		for heir.link[1] != nil {
			heir = heir.link[1]
		}
		root.key = heir.key
		root.value = heir.value
		key = heir.key
	}
	dir := 0
	if root.key.Compare(key) < 0 {
		dir = 1
	}
	var done bool
	var node *AVLNode
	root.link[dir], node, done = root.link[dir].removeR(key)
	if IsNil(found) {
		found = node
	}
	if done {
		return root, found, true
	}
	root.balance += 1 - 2*dir
	switch root.balance {
	case 1, -1:
		return root, found, true
	case 0:
		return root, found, false
	}
	resultRoot, resultDone := root.removeBalance(dir)
	return resultRoot, found, resultDone
}

func (root *AVLNode) getR(key IComparable) *AVLNode {
	if root.key.Equal(key) {
		return root
	}
	dir := 0
	if root.key.Compare(key) < 0 {
		dir = 1
	}
	if root.link[dir] != nil {
		return root.link[dir].getR(key)
	} else {
		return nil
	}
}

func (root *AVLNode) Print(w io.Writer, indent int) {
	var l1, l2 IComparable
	if root.link[0] != nil {
		l1 = root.link[0].key
	}
	if root.link[1] != nil {
		l2 = root.link[1].key
	}
	fmt.Fprintf(w, "%"+strconv.Itoa(indent)+"vk=%v, v=%v, b=%d, l=[%v, %v]\n",
		"",
		root.key,
		root.value,
		root.balance,
		l1,
		l2)
	if root.link[0] != nil {
		root.link[0].Print(w, indent+4)
	}
	if root.link[1] != nil {
		root.link[1].Print(w, indent+4)
	}
}

func (root *AVLNode) ToString() string {
	var l1, l2 IComparable
	if root.link[0] != nil {
		l1 = root.link[0].key
	}
	if root.link[1] != nil {
		l2 = root.link[1].key
	}
	return fmt.Sprintf("k=%v, v=%v, b=%d, l=[%v, %v]",
		root.key,
		root.value,
		root.balance,
		l1,
		l2)
}
