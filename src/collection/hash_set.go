package collection

type HashSet struct {
	hash *Hash
}

type HashSetIterator struct {
	iter IIterator
}

func NewHashSet() *HashSet {
	result := &HashSet{hash: NewHash()}
	return result
}

func NewCustomHashSet(initialCapacity int, loadFactor, multiplier float32) *HashSet {
	result := &HashSet{hash: NewCustomHash(initialCapacity, loadFactor, multiplier)}
	return result
}

func (m *HashSet) Exist(k IHashable) bool {
	node := m.hash.Get(k)
	if !IsNil(node) {
		return true
	} else {
		return false
	}
}

func (m *HashSet) Put(k IHashable) IHashable {
	node := m.hash.Put(k, nil)
	if !IsNil(node) {
		return node.key
	} else {
		return nil
	}
}

func (m *HashSet) Remove(k IHashable) IHashable {
	node := m.hash.Remove(k)
	if !IsNil(node) {
		return node.key
	} else {
		return nil
	}
}

func (m *HashSet) Size() int {
	return m.hash.Size()
}

func (i *HashSetIterator) Next() IObject {
	node := i.iter.Next().(*HashNode)
	if !IsNil(node) {
		return node.key
	} else {
		return nil
	}
}

func (i *HashSetIterator) HasNext() bool {
	return i.iter.HasNext()
}

func (m *HashSet) Iterator() IIterator {
	return &HashSetIterator{iter: m.hash.Iterator()}
}
