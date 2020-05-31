package collection

type HashMap struct {
	hash *Hash
}

type HashMapIterator struct {
	iter IIterator
}

func NewHashMap() *HashMap {
	result := &HashMap{hash: NewHash()}
	return result
}

func NewCustomHashMap(initialCapacity int, loadFactor, multiplier float32) *HashMap {
	result := &HashMap{hash: NewCustomHash(initialCapacity, loadFactor, multiplier)}
	return result
}

func (m *HashMap) Get(k IHashable) IObject {
	node := m.hash.Get(k)
	if !IsNil(node) {
		return node.value
	} else {
		return nil
	}
}

func (m *HashMap) Put(k IHashable, v IObject) IObject {
	node := m.hash.Put(k, v)
	if !IsNil(node) {
		return node.value
	} else {
		return nil
	}
}

func (m *HashMap) Remove(k IHashable) IObject {
	node := m.hash.Remove(k)
	if !IsNil(node) {
		return node.value
	} else {
		return nil
	}
}

func (m *HashMap) Size() int {
	return m.hash.Size()
}

func (i *HashMapIterator) Next() (IHashable, IObject) {
	node := i.iter.Next().(*HashNode)
	if !IsNil(node) {
		return node.key, node.value
	} else {
		return nil, nil
	}
}

func (i *HashMapIterator) HasNext() bool {
	return i.iter.HasNext()
}

func (i *HashMapIterator) Peek() (IHashable, IObject) {
	node := i.iter.Peek().(*HashNode)
	if !IsNil(node) {
		return node.key, node.value
	} else {
		return nil, nil
	}
}

func (m *HashMap) Iterator() IHashMapIterator {
	return &HashMapIterator{iter: m.hash.Iterator()}
}
