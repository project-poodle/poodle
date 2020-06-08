package util

import "fmt"

////////////////////////////////////////////////////////////////////////////////
// Interfaces
////////////////////////////////////////////////////////////////////////////////

type IScheme interface {

	////////////////////////////////////////
	// embeded interfaces
	//collection.IPrintable
	IEncodable

	////////////////////////////////////////
	// accessor to elements
	Domain() []byte          // get Domain
	DomainName() string      // get Domain name
	Tablet() []byte          // get tablet
	TabletName() string      // get tablet name
	Buckets() [][]byte       // get buckets
	BucketAt(idx int) []byte // get bucket at idx

	////////////////////////////////////////
	// magic
	SchemeMagic() byte // scheme magic byte
}

////////////////////////////////////////////////////////////////////////////////
// MappedScheme
////////////////////////////////////////////////////////////////////////////////

type MappedScheme struct {
	// buf and encoding
	decoded bool
	buf     []byte
	// attributes
	domain  []byte
	tablet  []byte
	buckets [][]byte
}

func NewMappedSchemd(buf []byte) *MappedScheme {
	return &MappedScheme{buf: buf}
}

func (s *MappedScheme) Domain() []byte {

	if !s.decoded {
		panic(fmt.Sprintf("MappedScheme::Domain - not decoded"))
	}

	return s.domain
}

func (s *MappedScheme) DomainName() string {

	if !s.decoded {
		panic(fmt.Sprintf("MappedScheme::DomainName - not decoded"))
	}

	if s.domain == nil {
		return ""
	} else {
		return string(s.domain)
	}
}

func (s *MappedScheme) Tablet() []byte {

	if !s.decoded {
		panic(fmt.Sprintf("MappedScheme::Tablet - not decoded"))
	}

	return s.tablet
}

func (s *MappedScheme) TabletName() string {

	if !s.decoded {
		panic(fmt.Sprintf("MappedScheme::TabletName - not decoded"))
	}

	if s.tablet == nil {
		return ""
	} else {
		return string(s.tablet)
	}
}

func (s *MappedScheme) Buckets() [][]byte {

	if !s.decoded {
		panic(fmt.Sprintf("MappedScheme::Buckets - not decoded"))
	}

	return s.buckets
}

func (s *MappedScheme) BucketAt(idx int) []byte {

	if !s.decoded {
		panic(fmt.Sprintf("MappedScheme::TabBucketAtlet - not decoded"))
	}

	if s.buckets == nil {
		panic("MappedScheme::BucketAt - buckets is nil")
	}

	return s.buckets[idx]
}

////////////////////////////////////////
// encoding, decoding, and buf

func (s *MappedScheme) SchemeMagic() byte {
	return s.buf[0]
}

func (s *MappedScheme) Buf() []byte {
	return s.buf
}

func (s *MappedScheme) EstBufSize() int {
	length := 1
	if s.domain != nil {
		length += 1 + len(s.domain)
	}
	if s.tablet != nil {
		length += 1 + len(s.tablet)
	}
	if s.buckets != nil {
		length += 1
		for _, bucket := range s.buckets {
			length += 1 + len(bucket)
		}
	}
	return length
}

func (s *MappedScheme) IsEncoded() bool {
	return true
}

func (s *MappedScheme) Encode(IContext) error {
	return fmt.Errorf("MappedScheme::Encode - not supported")
}

func (s *MappedScheme) IsDecoded() bool {
	return s.decoded
}

func (s *MappedScheme) Decode(IContext) (int, error) {
	// TODO
	pos := 1
	var length int
	var err error

	// decode domain
	if s.buf[0]|(0x01<<7) != 0 {
		s.domain, length, err = DecodeVarchar(s.buf[pos:])
		if err != nil {
			return 0, fmt.Errorf("MappedScheme::Decode - domain error [%v]", err)
		}
		pos += length
	}

	// decode tablet
	if s.buf[0]|(0x01<<6) != 0 {
		s.tablet, length, err = DecodeVarchar(s.buf[pos:])
		if err != nil {
			return 0, fmt.Errorf("MappedScheme::Decode - tablet error [%v]", err)
		}
		pos += length
	}

	// decode buckets
	if s.buf[0]|(0x01<<5) != 0 {
		bucketSize, length, err := DecodeUvarint(s.buf[pos:])
		if err != nil {
			return 0, fmt.Errorf("MappedScheme::Decode - bucket size error [%v]", err)
		}
		pos += length
		s.buckets = make([][]byte, bucketSize)
		for idx, _ := range s.buckets {
			s.buckets[idx], length, err = DecodeVarchar(s.buf[pos:])
			if err != nil {
				return 0, fmt.Errorf("MappedScheme::Decode - bucket [%d] error [%v]", idx, err)
			}
			pos += length
		}
	}

	return pos, nil
}

////////////////////////////////////////
// deep copy

func (s *MappedScheme) Copy() IEncodable {

	c := NewScheme()

	if s.domain != nil {
		c.domain = s.domain
	}

	if s.tablet != nil {
		c.tablet = s.tablet
	}

	if s.buckets != nil {
		c.buckets = make([][]byte, len(s.buckets))
		copy(c.buckets, s.buckets)
	}

	return c
}

func (s *MappedScheme) CopyConstruct() (IEncodable, error) {

	return s.Copy(), nil
}

////////////////////////////////////////////////////////////////////////////////
// Scheme
////////////////////////////////////////////////////////////////////////////////

type Scheme struct {
	// buf and encoding
	encoded bool
	buf     []byte
	// attributes
	domain  []byte
	tablet  []byte
	buckets [][]byte
}

func NewScheme() *Scheme {
	return &Scheme{}
}

func (s *Scheme) Domain() []byte {
	return s.domain
}

func (s *Scheme) DomainName() string {
	if s.domain == nil {
		return ""
	} else {
		return string(s.domain)
	}
}

func (s *Scheme) Tablet() []byte {
	return s.tablet
}

func (s *Scheme) TabletName() string {
	if s.tablet == nil {
		return ""
	} else {
		return string(s.tablet)
	}
}

func (s *Scheme) Buckets() [][]byte {
	return s.buckets
}

func (s *Scheme) BucketAt(idx int) []byte {
	if s.buckets == nil {
		panic("Scheme::BucketAt - buckets is nil")
	}

	return s.buckets[idx]
}

////////////////////////////////////////
// encoding, decoding, and buf

func (s *Scheme) SchemeMagic() byte {

	if !s.encoded {
		panic(fmt.Sprintf("Scheme::SchemeMagic - not encoded"))
	}

	return s.buf[0]
}

func (s *Scheme) Buf() []byte {

	if !s.encoded {
		panic(fmt.Sprintf("Scheme::Buf - not encoded"))
	}

	return s.buf
}

func (s *Scheme) EstBufSize() int {
	length := 1
	if s.domain != nil {
		length += 1 + len(s.domain)
	}
	if s.tablet != nil {
		length += 1 + len(s.tablet)
	}
	if s.buckets != nil {
		length += 1
		for _, bucket := range s.buckets {
			length += 1 + len(bucket)
		}
	}
	return length
}

func (s *Scheme) IsEncoded() bool {
	return s.encoded
}

func (s *Scheme) Encode(IContext) error {

	buf := []byte{byte(0x00)}

	if s.domain != nil {
		buf[0] |= 0x00 << 7
		buf = append(buf, EncodeVarchar(s.domain)...)
	}

	if s.tablet != nil {
		buf[0] |= 0x00 << 6
		buf = append(buf, EncodeVarchar(s.tablet)...)
	}

	if s.buckets != nil {
		buf[0] |= 0x00 << 5
		buf = append(buf, EncodeUvarint(uint64(len(s.buckets)))...)
		for _, bucket := range s.buckets {
			buf = append(buf, EncodeVarchar(bucket)...)
		}
	}

	s.buf = buf
	s.encoded = true

	return nil
}

func (s *Scheme) IsDecoded() bool {
	return true
}

func (s *Scheme) Decode(IContext) (int, error) {
	return 0, fmt.Errorf("Scheme::Decode - decode not supported")
}

////////////////////////////////////////
// deep copy

func (s *Scheme) Copy() IEncodable {

	c := NewScheme()

	if s.domain != nil {
		c.domain = s.domain
	}

	if s.tablet != nil {
		c.tablet = s.tablet
	}

	if s.buckets != nil {
		c.buckets = make([][]byte, len(s.buckets))
		copy(c.buckets, s.buckets)
	}

	return c
}

func (s *Scheme) CopyConstruct() (IEncodable, error) {

	return s.Copy(), nil
}
