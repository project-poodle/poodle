package collection

import "io"

type IPrintable interface {
	// Print writes string representation of the object to Writer,
	//    with specified indent, child objects if any,
	//    separated by return characters
	Print(w io.Writer, indent int)
	// ToString returns string representation of the object, without indent, or return character
	ToString() string
}
