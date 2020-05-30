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
	if node != nil {
		return true
	} else {
		return false
	}
}

func (m *SortedSet) Put(k IComparable) IComparable {
	node := m.tree.Put(k, nil)
	if node != nil {
		return node.key
	} else {
		return nil
	}
}

func (m *SortedSet) Remove(k IComparable) IComparable {
	node := m.tree.Remove(k)
	if node != nil {
		return node.key
	} else {
		return nil
	}
}

func (i *SortedSetIterator) Next() IObject {
	node := i.iter.Next().(*AVLNode)
	if node != nil {
		return node.key
	} else {
		return nil
	}
}

func (i *SortedSetIterator) HasNext() bool {
	return i.iter.HasNext()
}

func (m *SortedSet) Iterator() IIterator {
	return &SortedSetIterator{iter: m.tree.Iterator()}
}
