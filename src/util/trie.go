package util

import "fmt"

////////////////////////////////////////////////////////////////////////////////
// Interfaces
////////////////////////////////////////////////////////////////////////////////

type ITrie interface {

	////////////////////////////////////////
	// accessor to elements
	Get(IKey) (IData, error)                      // get
	Set(IKey, IData) error                        // set
	Keys(IKey) []IKey                             // get a list of keys
	KeyIterator() func() IKey                     // return iterator for all keys
	KeyRangeIterator(start, end IKey) func() IKey // return iterator for key within given range

	////////////////////////////////////////
	// encode, decode, and buf
	Buf() []byte          // return buf
	IsEncoded() bool      // check if encoded
	Encode() error        // encode
	IsDecoded() bool      // check if decoded
	Decode() (int, error) // decode, returns bytes read, and error if any

	////////////////////////////////////////
	// copy
	Copy() ITrie                   // copy
	CopyConstruct() (ITrie, error) // copy construct

	////////////////////////////////////////
	// return in readable format
	ToString() string
}

type ITrieNode interface {

	////////////////////////////////////////
	// accessor to elements
	FullKey() IKey                              // return full key
	NodeKey() []byte                            // trie node key is a sub key of IKey
	NodeType() byte                             // node type : 0x01 is Key, 0x02 is Attribute Group
	Parent() ITrieNode                          // link to parent
	Children() map[string]ITrieNode             // a list of children
	ChildSize() int                             // child size
	ChildAt(nodeKey []byte) ITrieNode           // get i-th child
	PutChild(nodeKey []byte, n ITrieNode) error // add child (automatically sort)
	RemoveChild(nodeKey []byte) error           // remote child
	GetData() IData                             // get associated data
	SetData(IData) error                        // set associated data

	////////////////////////////////////////
	// encode, decode, and buf
	Buf() []byte            // return buf
	IsEncoded() bool        // check if encoded
	Encode() error          // encode
	IsDecoded() bool        // check if decoded
	Decode() (int, error)   // decode, returns bytes read, and error if any
	GetOffset() uint32      // get offset when this TrieNode is encoded
	SetOffset(uint32) error // set offset when this TrieNode is encoded to

	////////////////////////////////////////
	// copy
	Copy() ITrieNode                   // copy
	CopyConstruct() (ITrieNode, error) // copy construct

	////////////////////////////////////////
	// return in readable format
	ToString() string
}

func newEven() func() int {
	n := 0
	// closure captures variable n
	return func() int {
		n += 2
		return n
	}
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
	// hidden field
	known_nodes map[uint32]ITrieNode
}

////////////////////////////////////////
// constructor

func NewMappedTrie(buf []byte) (*MappedTrie, int, error) {

	result := &MappedTrie{buf: buf}

	length, err := result.Decode()
	if err != nil {
		return nil, length, fmt.Errorf("NewMappedTrie - %s", err)
	}

	return result, length, nil
}

////////////////////////////////////////
// accessor to elements

func (t *MappedTrie) Get(k IKey) (IData, error) {

	if !t.decoded {
		panic(fmt.Sprintf("MappedTrie::Get - not decoded"))
	}

	// return root data is key is nil
	if k.IsNil() {
		return t.root.GetData(), nil
	}

	currNode := t.root
	for _, subKey := range k.Key() {
		childNode := currNode.ChildAt(subKey)
		if childNode == nil {
			return nil, nil // if not found, return nil
		} else {
			currNode = childNode // traverse down
		}
	}

	if currNode != nil {
		return currNode.GetData(), nil
	} else {
		return nil, nil
	}
}

func (t *MappedTrie) Set(IKey, IData) error {
	return fmt.Errorf("MappedTrie::Set - set not supported")
}

func (t *MappedTrie) Keys(k IKey) []IKey {

	if !t.decoded {
		panic(fmt.Sprintf("MappedTrie::Get - not decoded"))
	}

	// return root data is key is nil
	currNode := t.root
	if !k.IsNil() {
		for _, subKey := range k.Key() {
			childNode := currNode.ChildAt(subKey)
			if childNode == nil {
				return nil // if not found, return nil
			} else {
				currNode = childNode // traverse down
			}
		}
	}

	if currNode == nil {
		return nil
	}

	result := []IKey{}

	// traverse children
	traverse_stack := []ITrieNode{currNode}
	for len(traverse_stack) != 0 {

		currNode := traverse_stack[0]
		if currNode.NodeType() == TRIE_NODE_TYPE_KEY {

			// add to result
			result = append(result, currNode.FullKey())
			for _, child := range currNode.Children() {
				if child.NodeType() == TRIE_NODE_TYPE_KEY {
					traverse_stack = append(traverse_stack, child)
				}
			}
		}

		traverse_stack = traverse_stack[1:]
	}

	return result
}

func (t *MappedTrie) KeyIterator() func() IKey {
	// TODO:
	return nil
}

func (t *MappedTrie) KeyRangeIterator(start, end IKey) func() IKey {
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

func (t *MappedTrie) Encode() error {
	return fmt.Errorf("MappedTrie::Encode - not supported")
}

func (t *MappedTrie) IsDecoded() bool {
	return t.decoded
}

func (t *MappedTrie) Decode() (int, error) {

	// initial setup
	t.known_nodes = map[uint32]ITrieNode{}                                                   // clear known knows
	t.root = &MappedTrieNode{parent: nil, buf: t.buf, offset: 0, known_nodes: t.known_nodes} // start with root node
	t.known_nodes[0] = t.root                                                                // add myself

	pos, err := t.root.Decode()
	if err != nil {
		return 0, fmt.Errorf("MappedTrie::Decode - decode %s", err)
	}

	currNode := (t.root).(*MappedTrieNode)
	for {
		node := &MappedTrieNode{parent: currNode, buf: t.buf[pos:], offset: uint32(pos), known_nodes: t.known_nodes}
		length, err := node.Decode()
		if err != nil {
			return 0, fmt.Errorf("MappedTrie::Decode - decode [%d] %s", pos, err)
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

func (t *MappedTrie) Copy() ITrie {

	buf := make([]byte, len(t.buf))
	copy(buf, t.buf)
	result, _, err := NewMappedTrie(buf)
	if err != nil {
		panic(fmt.Sprintf("MappedTrie::Copy - unexpected error %s", err))
	}

	return result
}

func (t *MappedTrie) CopyConstruct() (ITrie, error) {
	return nil, fmt.Errorf("MappedTrie::CopyConstruct - not supported")
}

////////////////////////////////////////
// return in readable format

func (t *MappedTrie) ToString() string {

	str := fmt.Sprintf("MappedTrie")

	str += fmt.Sprintf("\n    root = %s", t.root.ToString())
	str += fmt.Sprintf("\n    buf = %v", t.buf[:MinInt(len(t.buf), 32)])

	return str
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
	nodeKey   []byte
	nodeType  byte
	parent    ITrieNode
	children  map[string]ITrieNode
	childSize int // for verification only
	data      IData
	// hidden fields
	dummy       bool
	known_nodes map[uint32]ITrieNode
}

////////////////////////////////////////
// constructor

func NewMappedTrieNode(parent ITrieNode, buf []byte, offset uint32, known_nodes map[uint32]ITrieNode) (*MappedTrieNode, int, error) {

	result := &MappedTrieNode{parent: parent, buf: buf, offset: offset, children: map[string]ITrieNode{}, known_nodes: known_nodes}

	length, err := result.Decode()
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

func (tn *MappedTrieNode) NodeType() byte {

	if !tn.decoded {
		panic(fmt.Sprintf("MappedTrieNode::NodeType - not decoded"))
	}

	return tn.nodeType
}

func (tn *MappedTrieNode) Parent() ITrieNode {

	if !tn.decoded {
		panic(fmt.Sprintf("MappedTrieNode::Parent - not decoded"))
	}

	return tn.parent
}

func (tn *MappedTrieNode) Children() map[string]ITrieNode {

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

	return len(tn.children)
}

func (tn *MappedTrieNode) ChildAt(nodeKey []byte) ITrieNode {

	if !tn.decoded {
		panic(fmt.Sprintf("MappedTrieNode::ChildAt - not decoded"))
	}

	if tn.children == nil {
		return nil
	}

	return tn.children[string(nodeKey)]
}

func (tn *MappedTrieNode) PutChild(nodeKey []byte, n ITrieNode) error {
	return fmt.Errorf("MappedTrieNode::PutChild - not supported")
}

func (tn *MappedTrieNode) RemoveChild(nodeKey []byte) error {
	return fmt.Errorf("MappedTrieNode::RemoteChild - not supported")
}

func (tn *MappedTrieNode) GetData() IData {

	if !tn.decoded {
		panic(fmt.Sprintf("MappedTrieNode::GetData - not decoded"))
	}

	return tn.data
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

func (tn *MappedTrieNode) Encode() error {
	return fmt.Errorf("MappedTrieNode::Encode - not supported")
}

func (tn *MappedTrieNode) IsDecoded() bool {
	return tn.decoded
}

func (tn *MappedTrieNode) Decode() (int, error) {

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

	nodeType, nodeTypeN, err := DecodeUvarint(tn.buf[pos:])
	if err != nil {
		return 0, fmt.Errorf("MappedTrieNode::Decode - nodeType - %s", err)
	}
	switch byte(nodeType) {
	case TRIE_NODE_TYPE_KEY:
		tn.nodeType = TRIE_NODE_TYPE_KEY
	case TRIE_NODE_TYPE_ATTR_GROUP:
		tn.nodeType = TRIE_NODE_TYPE_ATTR_GROUP
	default:
		return 0, fmt.Errorf("MappedTrieNode::Decode - invalid nodeType %d", nodeType)
	}
	pos += nodeTypeN

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
	data, dataN, err := NewSimpleMappedData(0xff, tn.buf[pos:])
	if err != nil {
		return 0, fmt.Errorf("MappedTrieNode::Decode - data - %s", err)
	}
	tn.data = data
	pos += dataN

	// child size
	childSize, childSizeN, err := DecodeUvarint(tn.buf)
	if err != nil {
		return 0, fmt.Errorf("MappedTrieNode::Decode - child size - %s", err)
	}
	tn.childSize = int(childSize)
	pos += childSizeN

	// we are here if parsing has completed successfully
	parent.(*MappedTrieNode).children[string(nodeKey)] = tn // update pointer in parent
	tn.known_nodes[tn.offset] = tn                          // update known nodes

	return pos, nil
}

func (tn *MappedTrieNode) GetOffset() uint32 {
	return tn.offset
}

func (tn *MappedTrieNode) SetOffset(uint32) error {
	return fmt.Errorf("MappedTrieNode::SetOffset - not supported")
}

////////////////////////////////////////
// copy

func (tn *MappedTrieNode) Copy() ITrieNode {

	buf := make([]byte, len(tn.buf))
	copy(buf, tn.buf)

	result := &MappedTrieNode{buf: buf, offset: tn.offset, known_nodes: tn.known_nodes}
	_, err := result.Decode()
	if err != nil {
		panic(fmt.Sprintf("MappedTrieNode::Copy - unexpected error %s", err))
	}

	return result
}

func (tn *MappedTrieNode) CopyConstruct() (ITrieNode, error) {
	return nil, fmt.Errorf("MappedTrieNode::CopyConstruct - not supported")
}

////////////////////////////////////////
// return in readable format

func (tn *MappedTrieNode) ToString() string {

	str := fmt.Sprintf("MappedTrieNode")

	str += fmt.Sprintf("\n    parent = %x", tn.parent)
	for childKey, child := range tn.children {
		str += fmt.Sprintf("\n    child[%v] = %x", childKey, child)
	}
	str += fmt.Sprintf("\n    data = %v", tn.data)
	str += fmt.Sprintf("\n    offset = %d", tn.offset)
	str += fmt.Sprintf("\n    buf = %v", tn.buf[:MinInt(len(tn.buf), 32)])

	return str
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
