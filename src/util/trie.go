package util

import (
	"fmt"
	"io"
	"strconv"

	"../collection"
)

////////////////////////////////////////////////////////////////////////////////
// Interfaces
////////////////////////////////////////////////////////////////////////////////

type ITrie interface {

	////////////////////////////////////////
	// embedded interfaces
	IEncodable
	collection.IPrintable

	////////////////////////////////////////
	// accessor to elements
	Get(IKey) IData        // get
	Put(IKey, IData) IData // put
	Remove(IKey) IData     // remove
	Entries() int          // total size

	// Iterators
	Iterator() ITrieIterator                     // this is same as nil key that iterates all keys
	KeyIterator(key IKey) ITrieIterator          // nil key param returns iterator for all keys, otherwise return iterator for specified key and children
	RangeIterator(start, end IKey) ITrieIterator // return iterator for keys within given range, start inclusive, end not inclusive
}

type ITrieNode interface {

	////////////////////////////////////////
	// embedded interfaces
	IEncodable
	collection.IPrintable

	////////////////////////////////////////
	// accessor to elements - parent, children, and keys
	FullKey() IKey                              // return full key
	NodeKey() []byte                            // trie node key is a sub key of IKey
	Parent() ITrieNode                          // link to parent
	Children() *collection.SortedMap            // a list of children
	ChildSize() int                             // child size
	GetChild(nodeKey []byte) ITrieNode          // get child by specified child key
	PutChild(nodeKey []byte, n ITrieNode) error // add child (automatically sort)
	RemoveChild(nodeKey []byte) error           // remote child

	////////////////////////////////////////
	// data
	Data() IData         // get associated data
	SetData(IData) error // set associated data

	////////////////////////////////////////
	// offset
	Offset() uint32         // get offset when this TrieNode is encoded
	SetOffset(uint32) error // set offset when this TrieNode is encoded to
}

type ITrieIterator interface {
	Next() (IKey, IData)
	HasNext() bool
	Peek() (IKey, IData)
}

////////////////////////////////////////////////////////////////////////////////
// MappedTrie
////////////////////////////////////////////////////////////////////////////////

type MappedTrie struct {
	// buf
	decoded bool
	buf     []byte
	// elements
	root ITrieNode
	size int
	// hidden field
	known_nodes map[uint32]ITrieNode
}

////////////////////////////////////////
// constructor

func NewMappedTrie(buf []byte) (*MappedTrie, int, error) {

	result := &MappedTrie{buf: buf}

	length, err := result.Decode(nil)
	if err != nil {
		return nil, length, fmt.Errorf("NewMappedTrie - %s", err)
	}

	return result, length, nil
}

////////////////////////////////////////
// accessor to elements

func (t *MappedTrie) Get(k IKey) IData {

	if !t.decoded {
		panic(fmt.Sprintf("MappedTrie::Get - not decoded"))
	}

	// return root data is key is nil
	if k.IsEmpty() {
		return t.root.Data()
	}

	currNode := t.root
	for _, subKey := range k.Key() {
		childNode := currNode.GetChild(subKey)
		if childNode == nil {
			return nil // if not found, return nil
		} else {
			currNode = childNode // traverse down
		}
	}

	if currNode != nil {
		return currNode.Data()
	} else {
		return nil
	}
}

func (t *MappedTrie) Set(IKey, IData) IData {
	panic(fmt.Sprintf("MappedTrie::Set - set not supported"))
}

func (t *MappedTrie) Iterator() ITrieIterator {
	return NewTrieKeyIterator(t.root, t.root)
}

func (t *MappedTrie) KeyIterator(k IKey) ITrieIterator {

	if !t.decoded {
		panic(fmt.Sprintf("MappedTrie::Iterator - not decoded"))
	}

	if collection.IsNil(k) || k.IsEmpty() {
		return NewTrieKeyIterator(t.root, t.root)
	} else {
		node := t.Get(k)
		if collection.IsNil(node) {
			return &TrieKeyIterator{} // return an empty iterator
		} else {
			currNode := node.(ITrieNode)
			return NewTrieKeyIterator(t.root, currNode) // return an iterator starting with given node
		}
	}
}

func (t *MappedTrie) RangeIterator(start, end IKey) ITrieIterator {

	if !t.decoded {
		panic(fmt.Sprintf("MappedTrie::RangeIterator - not decoded"))
	}

	return NewTrieRangeIterator(t.root, start, end)
}

////////////////////////////////////////
// encode, decode, and buf

func (t *MappedTrie) Buf() []byte {
	return t.buf
}

func (t *MappedTrie) EstBufSize() int {
	return len(t.buf)
}

func (t *MappedTrie) IsEncoded() bool {
	return true
}

func (t *MappedTrie) Encode(IContext) error {
	return fmt.Errorf("MappedTrie::Encode - not supported")
}

func (t *MappedTrie) IsDecoded() bool {
	return t.decoded
}

func (t *MappedTrie) Decode(IContext) (int, error) {

	// initial setup
	t.known_nodes = map[uint32]ITrieNode{}                                                   // clear known knows
	t.root = &MappedTrieNode{parent: nil, buf: t.buf, offset: 0, known_nodes: t.known_nodes} // start with root node
	t.known_nodes[0] = t.root                                                                // add myself

	pos, err := t.root.Decode(nil)
	if err != nil {
		return 0, fmt.Errorf("MappedTrie::Decode - %s", err)
	}

	currNode := (t.root).(*MappedTrieNode)
	for {
		node := &MappedTrieNode{parent: currNode, buf: t.buf[pos:], offset: uint32(pos), known_nodes: t.known_nodes}
		length, err := node.Decode(nil)
		if err != nil {
			return 0, fmt.Errorf("MappedTrie::Decode - pos [%d] %s", pos, err)
		}
		pos += length
		if node.dummy {
			if currNode == t.root {
				break // we have completed processing
			} else {
				currNode = currNode.parent.(*MappedTrieNode) // go back to the parent
			}
		} else {
			currNode = node // traverse down to child node
		}
	}

	// TODO - use size for validation

	t.decoded = true

	return pos, nil
}

////////////////////////////////////////
// copy

func (t *MappedTrie) Copy() IEncodable {

	buf := make([]byte, len(t.buf))
	copy(buf, t.buf)
	result, _, err := NewMappedTrie(buf)
	if err != nil {
		panic(fmt.Sprintf("MappedTrie::Copy - unexpected error %s", err))
	}

	return result
}

func (t *MappedTrie) CopyConstruct() (IEncodable, error) {
	return nil, fmt.Errorf("MappedTrie::CopyConstruct - not supported")
}

////////////////////////////////////////
// return in readable format

func (t *MappedTrie) Print(w io.Writer, indent int) {
	fmt.Fprintf(w, "%"+strconv.Itoa(indent)+"%s\n", "", t.ToString())
	if t.root != nil {
		t.root.Print(w, indent+4)
	}
}

func (t *MappedTrie) ToString() string {

	return fmt.Sprintf("MappedTrie: r=%s, buf=%v",
		t.root.ToString(),
		t.buf[:collection.MinInt(len(t.buf), 32)])
}

////////////////////////////////////////////////////////////////////////////////
// MappedTrieNode
////////////////////////////////////////////////////////////////////////////////

type MappedTrieNode struct {
	// buf
	decoded bool
	buf     []byte
	offset  uint32
	// elements
	nodeKey  []byte
	nodeData IData
	// parent and children
	parent    ITrieNode
	children  *collection.SortedMap
	childSize int // for verification only
	// hidden fields
	dummy       bool
	known_nodes map[uint32]ITrieNode
}

////////////////////////////////////////
// constructor

func NewMappedTrieNode(parent ITrieNode, buf []byte, offset uint32, known_nodes map[uint32]ITrieNode) (*MappedTrieNode, int, error) {

	result := &MappedTrieNode{parent: parent, buf: buf, offset: offset, children: collection.NewSortedMap(), known_nodes: known_nodes}

	length, err := result.Decode(nil)
	if err != nil {
		fmt.Errorf("NewMappedTrieNode - %s", err)
	}

	return result, length, nil
}

////////////////////////////////////////
// accessor to elements

func (tn *MappedTrieNode) FullKey() IKey {

	if !tn.decoded {
		panic(fmt.Sprintf("MappedTrieNode::NodeKey - not decoded"))
	}

	subKeys := []([]byte){}
	for currNode := tn; currNode != nil; currNode = currNode.parent.(*MappedTrieNode) {
		subKeys = append([]([]byte){currNode.nodeKey}, subKeys...)
	}

	fullKey := NewKey()
	for _, subKey := range subKeys {
		fullKey.Add(subKey)
	}

	return fullKey
}

func (tn *MappedTrieNode) NodeKey() []byte {

	if !tn.decoded {
		panic(fmt.Sprintf("MappedTrieNode::NodeKey - not decoded"))
	}

	return tn.nodeKey
}

func (tn *MappedTrieNode) Parent() ITrieNode {

	if !tn.decoded {
		panic(fmt.Sprintf("MappedTrieNode::Parent - not decoded"))
	}

	return tn.parent
}

func (tn *MappedTrieNode) Children() *collection.SortedMap {

	if !tn.decoded {
		panic(fmt.Sprintf("MappedTrieNode::Children - not decoded"))
	}

	return tn.children
}

func (tn *MappedTrieNode) ChildSize() int {

	if !tn.decoded {
		panic(fmt.Sprintf("MappedTrieNode::ChildSize - not decoded"))
	}

	if tn.children == nil {
		return 0
	}

	return tn.children.Size()
}

func (tn *MappedTrieNode) GetChild(nodeKey []byte) ITrieNode {

	if !tn.decoded {
		panic(fmt.Sprintf("MappedTrieNode::ChildAt - not decoded"))
	}

	if tn.children == nil {
		return nil
	}

	return tn.children.Get(collection.NewComparableByteSlice(nodeKey)).(ITrieNode)
}

func (tn *MappedTrieNode) PutChild(nodeKey []byte, n ITrieNode) error {
	return fmt.Errorf("MappedTrieNode::PutChild - not supported")
}

func (tn *MappedTrieNode) RemoveChild(nodeKey []byte) error {
	return fmt.Errorf("MappedTrieNode::RemoteChild - not supported")
}

func (tn *MappedTrieNode) Data() IData {

	if !tn.decoded {
		panic(fmt.Sprintf("MappedTrieNode::Data - not decoded"))
	}

	return tn.nodeData
}

func (tn *MappedTrieNode) SetData(IData) error {
	return fmt.Errorf("MappedTrieNode::SetData - not supported")
}

////////////////////////////////////////
// encode, decode, and buf

func (tn *MappedTrieNode) Buf() []byte {
	return tn.buf
}

func (tn *MappedTrieNode) EstBufSize() int {
	return len(tn.buf)
}

func (tn *MappedTrieNode) IsEncoded() bool {
	return true
}

func (tn *MappedTrieNode) Encode(IContext) error {
	return fmt.Errorf("MappedTrieNode::Encode - not supported")
}

func (tn *MappedTrieNode) IsDecoded() bool {
	return tn.decoded
}

func (tn *MappedTrieNode) Decode(IContext) (int, error) {

	pos := 0

	// check if this is a dummy Trie Node
	isDummy, isDummyN, err := DecodeUvarint(tn.buf[pos:])
	if err != nil {
		return 0, fmt.Errorf("MappedTrieNode::Decode - dummy - %s", err)
	}
	if isDummy == 0 {
		tn.dummy = true
		tn.decoded = true
		return isDummyN, nil
	}
	pos += isDummyN

	// node key
	nodeKey, nodeKeyN, err := DecodeVarchar(tn.buf[pos:])
	if err != nil {
		return 0, fmt.Errorf("MappedTrieNode::Decode - nodeKey - %s", err)
	} else if nodeKey == nil || len(nodeKey) == 0 {
		return 0, fmt.Errorf("MappedTrieNode::Decode - empty nodeKey")
	}
	tn.nodeKey = nodeKey
	pos += nodeKeyN

	// parent offset
	parentOffset, parentN, err := DecodeUvarint(tn.buf[pos:])
	if err != nil {
		return 0, fmt.Errorf("MappedTrieNode::Decode - parent offset - %s", err)
	}
	parent := tn.known_nodes[uint32(parentOffset)]
	if parent == nil {
		return 0, fmt.Errorf("MappedTrieNode::Decode - parent node not found [%d]", parentOffset)
	} else {
		tn.parent = parent
		pos += parentN
	}

	// data
	nodeData, nodeDataN, err := NewSimpleMappedData(0xff, tn.buf[pos:])
	if err != nil {
		return 0, fmt.Errorf("MappedTrieNode::Decode - data - %s", err)
	}
	tn.nodeData = nodeData
	pos += nodeDataN

	// child size
	childSize, childSizeN, err := DecodeUvarint(tn.buf)
	if err != nil {
		return 0, fmt.Errorf("MappedTrieNode::Decode - child size - %s", err)
	}
	tn.childSize = int(childSize)
	pos += childSizeN

	// we are here if parsing has completed successfully
	parent.(*MappedTrieNode).children.Put(collection.NewComparableByteSlice(nodeKey), tn) // update pointer in parent
	tn.known_nodes[tn.offset] = tn                                                        // update known nodes

	return pos, nil
}

func (tn *MappedTrieNode) Offset() uint32 {
	return tn.offset
}

func (tn *MappedTrieNode) SetOffset(uint32) error {
	return fmt.Errorf("MappedTrieNode::SetOffset - not supported")
}

////////////////////////////////////////
// copy

func (tn *MappedTrieNode) Copy() IEncodable {

	buf := make([]byte, len(tn.buf))
	copy(buf, tn.buf)

	result := &MappedTrieNode{buf: buf, offset: tn.offset, known_nodes: tn.known_nodes}
	_, err := result.Decode(nil)
	if err != nil {
		panic(fmt.Sprintf("MappedTrieNode::Copy - unexpected error %s", err))
	}

	return result
}

func (tn *MappedTrieNode) CopyConstruct() (IEncodable, error) {
	return nil, fmt.Errorf("MappedTrieNode::CopyConstruct - not supported")
}

////////////////////////////////////////
// return in readable format

func (tn *MappedTrieNode) Print(w io.Writer, indent int) {
	fmt.Fprintf(w, "%"+strconv.Itoa(indent)+"s%s", "", tn.ToString())
	for iter := tn.children.Iterator(); iter.HasNext(); {
		_, v := iter.Next()
		v.(ITrieNode).Print(w, indent+4)
	}
}

func (tn *MappedTrieNode) ToString() string {

	return fmt.Sprintf("MappedTrieNode: k=%x, child=[%d], data=%v, off=%d, buf=%v ",
		tn.nodeKey,
		tn.children.Size(),
		tn.nodeData,
		tn.offset,
		tn.buf[:collection.MinInt(len(tn.buf), 16)])
}

////////////////////////////////////////////////////////////////////////////////
// Constructed Trie
////////////////////////////////////////////////////////////////////////////////

type Trie struct {
	// elements
	root ITrieNode
	// buf
	encoded    bool
	buf        []byte
	estBufSize int
}

func NewTrie() *Trie {
	return &Trie{
		root:       NewTrieNode(nil, []byte{}, nil),
		estBufSize: 1,
	}
}

////////////////////////////////////////
// accessor to elements

func (t *Trie) Get(k IKey) IData {

	// return root data is key is nil
	if k.IsEmpty() {
		return t.root.Data()
	}

	currNode := t.root
	for _, subKey := range k.Key() {
		childNode := currNode.GetChild(subKey)
		if childNode == nil {
			return nil // if not found, return nil
		} else {
			currNode = childNode // traverse down
		}
	}

	if currNode != nil {
		return currNode.Data()
	} else {
		return nil
	}
}

func (t *Trie) Set(k IKey, d IData) IData {

	// set root data is key is nil
	if k.IsEmpty() {
		resultData := t.root.Data()
		t.root.SetData(d)
		return resultData
	}

	currNode := t.root
	for _, subKey := range k.Key() {
		childNode := currNode.GetChild(subKey)
		if childNode == nil {
			childNode = NewTrieNode(currNode, subKey, nil)
			currNode.PutChild(subKey, childNode)
			currNode = childNode
			t.estBufSize += childNode.EstBufSize()
		} else {
			currNode = childNode // traverse down
		}
	}

	if currNode == nil {
		panic("Trie::Set - unexpected empty currNode")
	}

	resultData := currNode.Data()
	currNode.SetData(d)
	t.estBufSize += estimateDataBufSize(d) - estimateDataBufSize(resultData)
	// if d is nil, clean up unused nodes
	if d == nil {
		currNode.Parent().RemoveChild(currNode.NodeKey())
		t.estBufSize -= currNode.EstBufSize()
		for currNode = currNode.Parent(); currNode.ChildSize() == 0 && currNode.Data() == nil; {
			currNode.Parent().RemoveChild(currNode.NodeKey())
			t.estBufSize -= currNode.EstBufSize()
			currNode = currNode.Parent()
		}
	}
	return resultData
}

func (t *Trie) Iterator() ITrieIterator {
	return NewTrieKeyIterator(t.root, t.root)
}

func (t *Trie) KeyIterator(k IKey) ITrieIterator {

	if collection.IsNil(k) || k.IsEmpty() {
		return NewTrieKeyIterator(t.root, t.root)
	} else {
		node := t.Get(k)
		if collection.IsNil(node) {
			return &TrieKeyIterator{} // return an empty iterator
		} else {
			currNode := node.(ITrieNode)
			return NewTrieKeyIterator(t.root, currNode) // return an iterator starting with given node
		}
	}
}

func (t *Trie) RangeIterator(start, end IKey) ITrieIterator {

	return NewTrieRangeIterator(t.root, start, end)
}

////////////////////////////////////////
// encode, decode, and buf

func (t *Trie) Buf() []byte {
	if !t.encoded {
		panic("Trie::Buf - not encoded")
	}

	return t.buf
}

func (t *Trie) EstBufSize() int {
	if t.estBufSize > 0 {
		return t.estBufSize
	} else {
		return 1
	}
}

func (t *Trie) IsEncoded() bool {
	return t.encoded
}

func (t *Trie) Encode(ctx IContext) error {

	buf := make([]byte, t.EstBufSize())
	pos := 0

	// encode root node
	t.root.SetOffset(uint32(pos))
	t.root.Encode(ctx)
	if len(buf) > pos+len(t.root.Buf()) {
		copy(buf[pos:], t.root.Buf())
	} else {
		buf = append(buf[:pos], t.root.Buf()...)
	}
	pos += len(t.root.Buf())

	// encode children
	iterList := []collection.ISortedMapIterator{t.root.Children().Iterator()}
	for lastIter := iterList[len(iterList)-1]; len(iterList) != 0; {
		if lastIter.HasNext() {

			// get next node, and encode
			_, data := lastIter.Next()
			node := data.(*TrieNode)
			node.SetOffset(uint32(pos))
			node.Encode(ctx)

			// encode node
			if len(buf) > pos+len(node.Buf()) {
				copy(buf[pos:], node.Buf())
			} else {
				buf = append(buf[:pos], node.Buf()...)
			}
			pos += len(node.Buf())

			// check if children exist
			if node.ChildSize() != 0 {
				iterList = append(iterList, node.Children().Iterator())
				lastIter = iterList[len(iterList)-1]
			}

		} else {

			// encode dummy node
			if len(buf) > pos+1 {
				buf[pos] = byte(0x00)
			} else {
				buf = append(buf[:pos], byte(0x00))
			}
			pos += 1

			// destack
			iterList = iterList[:len(iterList)-1]
		}
	}

	// we are here if encoding has completed successfully
	t.buf = buf
	return nil
}

func (t *Trie) IsDecoded() bool {
	return true
}

func (t *Trie) Decode(IContext) (int, error) {
	return 0, fmt.Errorf("Trie::Decode - not supported")
}

////////////////////////////////////////
// copy

func (t *Trie) Copy() IEncodable {

	buf := make([]byte, len(t.buf))
	copy(buf, t.buf)
	result, _, err := NewMappedTrie(buf)
	if err != nil {
		panic(fmt.Sprintf("MappedTrie::Copy - unexpected error %s", err))
	}

	return result
}

func (t *Trie) CopyConstruct() (IEncodable, error) {
	return t.Copy(), nil
}

////////////////////////////////////////
// return in readable format

func (t *Trie) Print(w io.Writer, indent int) {
	fmt.Fprintf(w, "%"+strconv.Itoa(indent)+"%s\n", "", t.ToString())
	if t.root != nil {
		t.root.Print(w, indent+4)
	}
}

func (t *Trie) ToString() string {

	return fmt.Sprintf("Trie: r=%s, buf=%v",
		t.root.ToString(),
		t.buf[:collection.MinInt(len(t.buf), 32)])
}

////////////////////////////////////////////////////////////////////////////////
// Constructed TrieNode
////////////////////////////////////////////////////////////////////////////////

type TrieNode struct {
	// elements
	parent   ITrieNode
	children *collection.SortedMap
	nodeKey  []byte
	data     IData
	// buf
	encoded bool
	buf     []byte
	offset  uint32
}

func NewTrieNode(parent ITrieNode, nodeKey []byte, data IData) *TrieNode {
	if collection.IsNil(parent) {
		return &TrieNode{
			parent:   nil,
			nodeKey:  nil,
			data:     data,
			children: collection.NewSortedMap(),
		}

	}
	return &TrieNode{
		parent:   parent,
		nodeKey:  nodeKey,
		data:     data,
		children: collection.NewSortedMap(),
	}
}

////////////////////////////////////////
// accessor to elements - parent, children, and keys
func (tn *TrieNode) FullKey() IKey {

	subKeys := []([]byte){}
	var currNode ITrieNode
	for currNode = tn; currNode != nil; currNode = currNode.Parent() {
		subKeys = append([]([]byte){currNode.NodeKey()}, subKeys...)
	}

	fullKey := NewKey()
	for _, subKey := range subKeys {
		fullKey.Add(subKey)
	}

	return fullKey
}

func (tn *TrieNode) NodeKey() []byte {
	return tn.nodeKey
}

func (tn *TrieNode) Parent() ITrieNode {
	return tn.parent
}

func (tn *TrieNode) Children() *collection.SortedMap {
	return tn.children
}

func (tn *TrieNode) ChildSize() int {
	return tn.children.Size()
}

func (tn *TrieNode) GetChild(nodeKey []byte) ITrieNode {
	return tn.children.Get(collection.NewComparableByteSlice(nodeKey)).(ITrieNode)
}

func (tn *TrieNode) PutChild(nodeKey []byte, n ITrieNode) error {
	tn.children.Put(collection.NewComparableByteSlice(nodeKey), n)
	return nil
}

func (tn *TrieNode) RemoveChild(nodeKey []byte) error {
	tn.children.Remove(collection.NewComparableByteSlice(nodeKey))
	return nil
}

func (tn *TrieNode) Data() IData {
	return tn.data
}

func (tn *TrieNode) SetData(data IData) error {
	tn.data = data
	return nil
}

////////////////////////////////////////
// encode, decode, and buf

func (tn *TrieNode) Buf() []byte {
	if !tn.encoded {
		panic("TrieNode::Buf - not encoded")
	}

	return tn.buf
}

func (tn *TrieNode) EstBufSize() int {
	return 4 + 1 + len(tn.nodeKey) + estimateDataBufSize(tn.data)
}

func (tn *TrieNode) IsEncoded() bool {
	return tn.encoded
}

func (tn *TrieNode) Encode(IContext) error {

	// node key
	nodeKeyBuf := EncodeVarchar(tn.nodeKey)

	// parent offset
	parentOffsetBuf := EncodeUvarint(uint64(tn.parent.Offset()))

	// data
	dataBuf, _, err := tn.data.Encode(false)
	if err != nil {
		return fmt.Errorf("TrieNode::Encode - data encode error %v", err)
	}

	// child size
	childSizeBuf := EncodeUvarint(uint64(tn.children.Size()))

	tn.buf = make([]byte, len(nodeKeyBuf)+len(parentOffsetBuf)+len(dataBuf)+len(childSizeBuf))

	pos := 0
	copy(tn.buf[pos:], nodeKeyBuf)
	pos += len(nodeKeyBuf)
	copy(tn.buf[pos:], parentOffsetBuf)
	pos += len(parentOffsetBuf)
	copy(tn.buf[pos:], dataBuf)
	pos += len(dataBuf)
	copy(tn.buf[pos:], childSizeBuf)
	pos += len(childSizeBuf)

	tn.encoded = true
	return nil
}

func (tn *TrieNode) IsDecoded() bool {
	return true
}

func (tn *TrieNode) Decode(IContext) (int, error) {
	return 0, fmt.Errorf("TrieNode::Decode - not supported")
}

func (tn *TrieNode) Offset() uint32 {
	return tn.offset
}

func (tn *TrieNode) SetOffset(offset uint32) error {
	tn.offset = offset
	return nil
}

////////////////////////////////////////
// copy

func (tn *TrieNode) Copy() IEncodable {

	buf := make([]byte, len(tn.buf))
	copy(buf, tn.buf)

	result := NewTrieNode(tn.parent, tn.nodeKey, tn.data)
	for iter := tn.Children().Iterator(); iter.HasNext(); {
		key, value := iter.Next()
		result.Children().Put(key, value)
	}

	return result
}

func (tn *TrieNode) CopyConstruct() (IEncodable, error) {
	return tn.Copy(), nil
}

////////////////////////////////////////
// return in readable format

func (tn *TrieNode) Print(w io.Writer, indent int) {
	fmt.Fprintf(w, "%"+strconv.Itoa(indent)+"s%s", "", tn.ToString())
	for iter := tn.children.Iterator(); iter.HasNext(); {
		_, v := iter.Next()
		v.(ITrieNode).Print(w, indent+4)
	}
}

func (tn *TrieNode) ToString() string {

	var buf []byte
	if tn.buf != nil {
		buf = tn.buf[:collection.MinInt(len(tn.buf), 16)]
	} else {
		buf = nil
	}

	return fmt.Sprintf("TrieNode: k=%x, child=[%d], data=%v, off=%d, buf=%v ",
		tn.nodeKey,
		tn.children.Size(),
		tn.data,
		tn.offset,
		buf)
}

////////////////////////////////////////////////////////////////////////////////
// TrieKeyIterator
////////////////////////////////////////////////////////////////////////////////

type TrieKeyIterator struct {
	paths    []collection.ISortedMapIterator
	rootNode ITrieNode
	currNode ITrieNode
}

func NewTrieKeyIterator(root, curr ITrieNode) *TrieKeyIterator {

	result := &TrieKeyIterator{paths: make([]collection.ISortedMapIterator, 0)}

	// check root validaty
	if collection.IsNil(root) {
		panic("NewTrieIterator - root cannot be nil")
	} else if !root.FullKey().IsEmpty() {
		panic("NewTrieIterator - root key is not empty")
	} else {
		result.rootNode = root
	}

	// check curr validaty
	if collection.IsNil(curr) {
		result.currNode = root
	} else {
		for _, subKey := range curr.FullKey().Key() {
			child := root.GetChild(subKey)
			if collection.IsNil(child) {
				panic(fmt.Sprintf("NewTrieIterator - curr node [%s] is not a child of root [%s]",
					root.FullKey().ToString(),
					curr.FullKey().ToString()))
			}
		}
		// we are here as curr is a descendent of root
		result.currNode = curr
	}

	return result
}

func (i *TrieKeyIterator) Next() (IKey, IData) {

	i.advance()

	var returnNode ITrieNode

	// if currNode has not been iterated
	if !collection.IsNil(i.currNode) {
		returnNode = i.currNode
		if returnNode.ChildSize() != 0 {
			iter := i.currNode.Children().Iterator()
			if i.paths == nil {
				i.paths = []collection.ISortedMapIterator{iter}
			} else {
				i.paths = append(i.paths, iter)
			}
		}
		i.currNode = nil
		return returnNode.FullKey(), returnNode.Data()
	}

	// we are here if root node has been iterated
	// check iterator paths
	for i.paths != nil && len(i.paths) != 0 {
		lastIter := i.paths[len(i.paths)-1]
		if lastIter.HasNext() {
			_, data := lastIter.Next()
			returnNode = data.(ITrieNode)
			if returnNode.ChildSize() != 0 {
				i.paths = append(i.paths, returnNode.Children().Iterator())
			}
			return returnNode.FullKey(), returnNode.Data()
		} else {
			i.paths = i.paths[:len(i.paths)-1]
		}
	}

	// we are here if no more paths left
	return nil, nil
}

func (i *TrieKeyIterator) HasNext() bool {

	i.advance()

	return collection.IsNil(i.currNode) && (collection.IsNil(i.paths) || len(i.paths) == 0)
}

func (i *TrieKeyIterator) Peek() (IKey, IData) {

	i.advance()

	var returnNode ITrieNode

	// if rootNode has not been iterated
	if !collection.IsNil(i.currNode) {
		returnNode = i.currNode
		return returnNode.FullKey(), returnNode.Data()
	}

	// we are here if root node has been iterated
	// check iterator paths
	for i.paths != nil && len(i.paths) != 0 {
		lastIter := i.paths[len(i.paths)-1]
		if lastIter.HasNext() {
			_, data := lastIter.Peek()
			returnNode = data.(ITrieNode)
			return returnNode.FullKey(), returnNode.Data()
		} else {
			i.paths = i.paths[:len(i.paths)-1]
		}
	}

	// we are here if no more paths left
	return nil, nil
}

func (i *TrieKeyIterator) advance() {

	if i.currNode != nil {
		return
	}

	if i.paths == nil {
		return
	}

	for len(i.paths) != 0 {
		lastIter := i.paths[len(i.paths)-1]
		if lastIter.HasNext() {
			return
		} else {
			i.paths = i.paths[:len(i.paths)-1]
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// TrieRangeIterator
////////////////////////////////////////////////////////////////////////////////

type TrieRangeIterator struct {
	rootNode         ITrieNode
	isRoot           bool
	startNodes       []ITrieNode
	startChild       []byte
	endNodes         []ITrieNode
	endChild         []byte
	paths            []collection.ISortedMapIterator
	startInitialized bool
	endMatches       []bool
}

func NewTrieRangeIterator(root ITrieNode, start, end IKey) *TrieRangeIterator {

	result := &TrieRangeIterator{
		startNodes:       make([]ITrieNode, 0),
		endNodes:         make([]ITrieNode, 0),
		paths:            make([]collection.ISortedMapIterator, 0),
		startInitialized: false,
		endMatches:       make([]bool, 0),
	}

	// check root validaty
	if collection.IsNil(root) {
		panic("NewTrieRangeIterator - root cannot be nil")
	} else if !root.FullKey().IsEmpty() {
		panic("NewTrieRangeIterator - root key is not empty")
	} else {
		result.rootNode = root
		result.isRoot = true
	}

	if !collection.IsNil(start) && !collection.IsNil(end) && start.Compare(end) > 0 {
		panic(fmt.Sprintf("NewTrieRangeIterator - start [%v] larget than end [%v]",
			start.ToString(),
			end.ToString()))
	}

	// process start key
	if collection.IsNil(start) {
		result.startChild = nil
	} else {
		currNode := result.rootNode
		for _, subKey := range start.Key() {
			child := currNode.GetChild(subKey)
			if !collection.IsNil(child) {
				// result.paths = append(result.paths,
				//    currNode.Children().RangeIterator(collection.NewComparableByteSlice(subKey), nil))
				result.startNodes = append(result.startNodes, child)
				currNode = child
			} else {
				result.startChild = subKey
				break
			}
		}
	}

	// process end key
	if collection.IsNil(end) {
		result.endChild = nil
	} else {
		currNode := result.rootNode
		for _, subKey := range end.Key() {
			child := currNode.GetChild(subKey)
			if !collection.IsNil(child) {
				result.endNodes = append(result.endNodes, child)
				//if len(result.paths) < len(result.endNodes) {
				result.endMatches = append(result.endMatches, false)
				//} else if len(result.endMatches) != 0 && !result.endMatches[len(result.endMatches)-1] {
				//	result.endMatches = append(result.endMatches, false)
				//} else {
				//	_, node := result.paths[len(result.endNodes)-1].Peek()
				//	if collection.IsNil(node) ||
				//		collection.CompareByteSlice(node.(ITrieNode).NodeKey(), child.NodeKey()) == 0 {
				//		result.endMatches = append(result.endMatches, true)
				//	} else {
				//		result.endMatches = append(result.endMatches, false)
				//	}
				//}
				currNode = child
			} else {
				result.endChild = subKey
				break
			}
		}
	}

	return result
}

func (i *TrieRangeIterator) Next() (IKey, IData) {

	i.checkStart() // this method returns with last iterator after proper start key
	i.checkEnd()   // this method by pass any unnessary last iterators if they passed end key

	// if currNode has not been iterated
	if i.isRoot {

		if len(i.paths) != 0 {
			panic(fmt.Sprintf("TrieRangeIterator::Next - paths length [%d] is not 0 when currNode is not nil", len(i.paths)))
		}
		if len(i.startNodes) != 0 {
			panic(fmt.Sprintf("TrieRangeIterator::Next - start nodes length [%d] is not 0 when currNode is not nil", len(i.startNodes)))
		}

		returnNode := i.rootNode
		if returnNode.ChildSize() != 0 {

			// create child iterator and add to paths
			iter := i.rootNode.Children().Iterator()
			i.paths = append(i.paths, iter)
			i.isRoot = false

		}

		i.checkEnd() // this method by pass any unnessary last iterators if they passed end key
		return returnNode.FullKey(), returnNode.Data()
	}

	// we are here if root node has been iterated
	// check iterator paths
	for i.paths != nil && len(i.paths) != 0 {

		lastIter := i.paths[len(i.paths)-1]
		if lastIter.HasNext() {

			_, data := lastIter.Next()

			returnNode := data.(ITrieNode)
			if returnNode.ChildSize() != 0 {

				iter := returnNode.Children().Iterator() // it is not possible to be here and have child iterator with start key
				i.paths = append(i.paths, iter)
			}

			i.checkEnd() // this method by pass any unnessary last iterators if they passed end key
			return returnNode.FullKey(), returnNode.Data()
		}
	}

	return nil, nil
}

func (i *TrieRangeIterator) HasNext() bool {

	i.checkStart()
	i.checkEnd()

	return i.isRoot || len(i.paths) != 0
}

func (i *TrieRangeIterator) Peek() (IKey, IData) {

	i.checkStart() // this method returns with last iterator after proper start key
	i.checkEnd()   // this method by pass any unnessary last iterators if they passed end key

	// if currNode has not been iterated
	if i.isRoot {

		if len(i.paths) != 0 {
			panic(fmt.Sprintf("TrieRangeIterator::Peek - paths length [%d] is not 0 when currNode is not nil", len(i.paths)))
		}
		if len(i.startNodes) != 0 {
			panic(fmt.Sprintf("TrieRangeIterator::Peek - start nodes length [%d] is not 0 when currNode is not nil", len(i.startNodes)))
		}

		returnNode := i.rootNode
		return returnNode.FullKey(), returnNode.Data()
	}

	// we are here if root node has been iterated
	// check iterator paths
	for i.paths != nil && len(i.paths) != 0 {

		lastIter := i.paths[len(i.paths)-1]
		if lastIter.HasNext() {

			_, data := lastIter.Peek()

			returnNode := data.(ITrieNode)
			return returnNode.FullKey(), returnNode.Data()
		}
	}

	return nil, nil
}

func (i *TrieRangeIterator) checkStart() {

	if i.startInitialized {
		return
	}

	if len(i.startNodes) == 0 {

		if collection.IsNil(i.startChild) {
			i.startInitialized = true
			return
		}

		if i.isRoot && i.rootNode.ChildSize() != 0 {
			// create child iterator and add to paths
			iter := i.rootNode.Children().RangeIterator(collection.NewComparableByteSlice(i.startChild), nil)
			i.paths = append(i.paths, iter)
			i.isRoot = false
		}

	} else {

		if i.isRoot && i.rootNode.ChildSize() != 0 {
			// create child iterator and add to paths
			iter := i.rootNode.Children().RangeIterator(collection.NewComparableByteSlice(i.startNodes[0].NodeKey()), nil)
			i.paths = append(i.paths, iter)
			i.isRoot = false
		}
	}

	i.startInitialized = true
	for len(i.paths) != 0 && len(i.startNodes) >= len(i.paths) {

		lastIter := i.paths[len(i.paths)-1]
		_, node := lastIter.Peek()

		comp := collection.CompareByteSlice(node.(ITrieNode).NodeKey(), i.startNodes[len(i.paths)-1].NodeKey())
		if comp < 0 {

			panic(fmt.Sprintf("TrieRangeIterator::checkStart - unexpected node [%v], less than but not matching start key [%v]",
				node.(ITrieNode).FullKey().ToString(),
				i.startNodes[len(i.paths)-1].FullKey().ToString()))

		} else if comp == 0 {

			if len(i.startNodes) < len(i.paths) {

				panic(fmt.Sprintf("TrieRangeIterator::checkStart - unexpected paths len [%d], greater than start nodes [%d]",
					len(i.paths),
					len(i.startNodes)))

			} else if len(i.startNodes) == len(i.paths) {

				if collection.IsNil(i.startChild) {
					return // iterate this node, as it matches start key, and there are no child key
				} else {
					_, node = lastIter.Next() // bypass this node, as it matches start key, and child key exist
					childIter := node.(ITrieNode).Children().RangeIterator(
						collection.NewComparableByteSlice(i.startChild),
						nil)
					i.paths = append(i.paths, childIter)
					return
				}

			} else {

				_, node = lastIter.Next() // bypass this node, as it matches start key, and child node(s) exist
				childIter := node.(ITrieNode).Children().RangeIterator(
					collection.NewComparableByteSlice(i.startNodes[len(i.paths)-1].NodeKey()),
					nil)
				i.paths = append(i.paths, childIter)
				// continue loop

			}

		} else { // comp > 0

			return

		}
	}
}

func (i *TrieRangeIterator) checkEnd() {

	if len(i.endNodes) != len(i.endMatches) {
		panic(fmt.Sprintf("TrieRangeIterator::checkEnd - endNodes len [%d] does not match endMatches [%d]",
			len(i.endNodes),
			len(i.endMatches)))
	}

	if len(i.endNodes) == 0 {
		return
	}

	if i.isRoot {
		//panic("TrieRangeIterator::checkEnd - isRoot set!")
		return
	}

	for len(i.paths) != 0 && len(i.endNodes) >= len(i.paths) {

		lastIter := i.paths[len(i.paths)-1]
		if !lastIter.HasNext() {
			i.paths = i.paths[:len(i.paths)-1] // destack if no more element left
			continue                           // continue to next loop
		}

		if len(i.paths) != 1 && !i.endMatches[len(i.paths)-2] {
			return // iterate this node, as some of previous end node(s) does not match
		}

		// we are here if all previous end node(s) match(es)
		_, node := i.paths[len(i.paths)-1].Peek()
		if collection.IsNil(node) {
			panic("TrieRangeIterator::checkEnd - unexpected nil node from iterator")
		}

		comp := collection.CompareByteSlice(node.(ITrieNode).NodeKey(), i.endNodes[len(i.paths)-1].NodeKey())
		if comp > 0 {

			i.endMatches[len(i.paths)-1] = true
			i.paths = i.paths[:len(i.paths)-1]
			continue // next loop

		} else if comp == 0 {

			i.endMatches[len(i.paths)-1] = true
			if len(i.paths) > len(i.endNodes) {

				panic(fmt.Sprintf("TrieRangeIterator::checkEnd - unexpected paths len [%d], greater than end nodes [%d]",
					len(i.paths),
					len(i.endNodes)))

			} else if len(i.paths) == len(i.endNodes) {

				if collection.IsNil(i.endChild) {

					i.paths = i.paths[:len(i.paths)-1] // bypass this node, this node matches full end key, (no child)
					continue                           // next loop

				} else {

					return // iterate this node, this node does not match full end key, child exist
				}

			} else { // len(i.paths) < len(i.endNodes)

				return // iterate this node, this node does not match ffull end key, child exist
			}

		} else { // comp < 0

			i.endMatches[len(i.paths)-1] = false // iterate this node, as more element exists
			return                               // just return
		}

	}
}
