package collection

type SortedSet struct {
	tree *AVLTree
}

type SortedSetIterator struct {
	iter IIterator
}

func NewSortedSet() *SortedSet {
	result := &SortedSet{tree: &AVLTree{}}
	return result
}

func (m *SortedSet) Exist(k IComparable) bool {
	node := m.tree.Get(k)
	if !IsNil(node) {
		return true
	} else {
		return false
	}
}

func (m *SortedSet) Put(k IComparable) IComparable {
	node := m.tree.Put(k, nil)
	if !IsNil(node) {
		return node.key
	} else {
		return nil
	}
}

func (m *SortedSet) Remove(k IComparable) IComparable {
	node := m.tree.Remove(k)
	if !IsNil(node) {
		return node.key
	} else {
		return nil
	}
}

func (m *SortedSet) Size() int {
	return m.tree.Size()
}

func (i *SortedSetIterator) Next() IObject {
	node := i.iter.Next().(*AVLNode)
	if !IsNil(node) {
		return node.key
	} else {
		return nil
	}
}

func (i *SortedSetIterator) HasNext() bool {
	return i.iter.HasNext()
}

func (i *SortedSetIterator) Peek() IObject {
	node := i.iter.Peek().(*AVLNode)
	if !IsNil(node) {
		return node.key
	} else {
		return nil
	}
}

func (m *SortedSet) Iterator() IIterator {
	return &SortedSetIterator{iter: m.tree.Iterator()}
}

func (m *SortedSet) RangeIterator(start, end IComparable) IIterator {
	return &SortedSetIterator{iter: m.tree.RangeIterator(start, end)}
}
