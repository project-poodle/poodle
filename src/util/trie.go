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
	Size() int             // total size

	// Iterators
	Iterator(key IKey) ITrieIterator             // nil key param returns iterator for all keys, otherwise return iterator for specified key and children
	RangeIterator(start, end IKey) ITrieIterator // return iterator for keys within given range
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
	if k.IsNil() {
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

func (t *MappedTrie) Iterator(k IKey) ITrieIterator {

	if !t.decoded {
		panic(fmt.Sprintf("MappedTrie::Iterator - not decoded"))
	}

	if collection.IsNil(k) {
		return &TrieIterator{rootNode: t.root}
	} else {
		node := t.Get(k)
		if collection.IsNil(node) {
			return &TrieIterator{} // return an empty iterator
		} else {
			rootNode := node.(ITrieNode)
			return &TrieIterator{rootNode: rootNode} // return an iterator starting with given node
		}
	}
}

func (t *MappedTrie) RangeIterator(start, end IKey) func() IKey {

	if !t.decoded {
		panic(fmt.Sprintf("MappedTrie::RangeIterator - not decoded"))
	}

	// TODO:
	return nil
}

////////////////////////////////////////
// encode, decode, and buf

func (t *MappedTrie) Buf() []byte {
	return t.buf
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

	// parent
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

	// children size
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
// Constructed Trie & TrieNode
////////////////////////////////////////////////////////////////////////////////

type Trie struct {
	// elements
	root ITrieNode
	// buf
	encoded bool
	buf     []byte
}

type TrieNode struct {
	// elements
	parent   ITrieNode
	children []ITrieNode
	data     IData
	// buf
	encoded bool
	buf     []byte
	offset  uint32
}

////////////////////////////////////////////////////////////////////////////////
// TrieIterator
////////////////////////////////////////////////////////////////////////////////

type TrieIterator struct {
	paths    []collection.ISortedMapIterator
	rootNode ITrieNode
	start    IKey
	end      IKey
}

func (i *TrieIterator) Next() (IKey, IData) {

	i.advance()

	var returnNode ITrieNode

	// if rootNode has not been iterated
	if i.rootNode != nil {
		returnNode = i.rootNode
		iter := i.rootNode.Children().Iterator()
		if returnNode.ChildSize() != 0 {
			if i.paths == nil {
				i.paths = []collection.ISortedMapIterator{iter}
			} else {
				i.paths = append(i.paths, iter)
			}
		}
		i.rootNode = nil
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

func (i *TrieIterator) HasNext() bool {

	i.advance()

	return collection.IsNil(i.rootNode) && (collection.IsNil(i.paths) || len(i.paths) == 0)
}

func (i *TrieIterator) advance() {

	if i.rootNode != nil {
		if !collection.IsNil(i.start) {
			if i.rootNode.FullKey().Compare(i.start) < 0 { // TODO
				// if root node is not within range of start key
				//return nil
			}
		}
		if !collection.IsNil(i.end) {
			if i.rootNode.FullKey().Compare(i.end) > 0 { // TODO
				// if root node is not within range of end key
				//return nil
			}
		}
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
