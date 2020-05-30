package collection

type SortedMap struct {
	tree *AVLTree
}

type SortedMapIterator struct {
	iter IIterator
}

func NewSortedMap() *SortedMap {
	result := &SortedMap{tree: &AVLTree{}}
	return result
}

func (m *SortedMap) Get(k IComparable) IObject {
	node := m.tree.Get(k)
	if node != nil {
		return node.value
	} else {
		return nil
	}
}

func (m *SortedMap) Put(k IComparable, v IObject) IObject {
	node := m.tree.Put(k, v)
	if node != nil {
		return node.value
	} else {
		return nil
	}
}

func (m *SortedMap) Remove(k IComparable) IObject {
	node := m.tree.Remove(k)
	if node != nil {
		return node.value
	} else {
		return nil
	}
}

func (i *SortedMapIterator) Next() (IComparable, IObject) {
	node := i.iter.Next().(*AVLNode)
	if node != nil {
		return node.key, node.value
	} else {
		return nil, nil
	}
}

func (i *SortedMapIterator) HasNext() bool {
	return i.iter.HasNext()
}

func (m *SortedMap) Iterator() ISortedMapIterator {
	return &SortedMapIterator{iter: m.tree.Iterator()}
}
