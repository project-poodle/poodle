package util

////////////////////////////////////////////////////////////////////////////////
// Interfaces
////////////////////////////////////////////////////////////////////////////////

type IContext interface {
}

type IEncodable interface {

	////////////////////////////////////////
	// encode and decode

	// whether Data is encoded      - always return true for Mapped Data
	//                                  - return true for Constructed Data if encoded buf cache exists
	//                                  - return false for Constructed Data if no encoded buf cache
	IsEncoded() bool

	// encode Data                  - for Constructed Data only, return error for Mapped Data
	//                                  - if successful, encoded buf is kept as part of Data object
	//                                  - bool param indicate whether to encode with parent Record context
	//                                  - byte return value is parent bits, return 0xff if this is self-encoded as Data
	Encode(ctx IContext) error

	// whether Data is decoded      - always return true for Constructed Data
	//                                  - return true for Mapped Data if data is decoded
	//                                  - return false for Mapped Data if data is not decoded
	IsDecoded() bool

	// decode Data                  - for Mapped Data only, return error for Constructed Data
	//                                  - if successful, individual data array, record list, or primitive data are decoded and kept as part of Data object
	//                                  - parent param is data encode from parent: 0x00 is no length; 0x01 is 1 byte length; 0x02 is 2 byte length; 0x03 is custom encoding
	//                                  - use 0xff if no parent
	Decode(ctx IContext) (int, error)

	////////////////////////////////////////
	// copy & buf

	// copy
	// - for mapped object, copy the underlying mapped byte array (read only)
	// - for constructed object, make a copy of the constructed object (modifiable)
	Copy() IEncodable
	// make a constructed (modifiable) copy of the object
	CopyConstruct() (IEncodable, error)

	// return encoded buf (byte array)
	Buf() []byte
	// estimated buf size
	EstBufSize() int
}
