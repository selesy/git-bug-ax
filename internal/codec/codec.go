package codec

import "encoding"

// TextCodec combines encoding.TextMarshaler and encoding.TextUnmarshaler
// for types that support bidirectional text serialization.
type TextCodec interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
}

var _ TextCodec = (*Label)(nil)

// Label is a string type that implements TextCodec for bidirectional text serialization.
type Label string

// MarshalText implements encoding.TextMarshaler, returning the label as a string.
func (l Label) MarshalText() ([]byte, error) {
	return []byte(l), nil
}

// UnmarshalText implements encoding.TextUnmarshaler, parsing a label string.
func (l *Label) UnmarshalText(text []byte) error {
	*l = Label(text)
	return nil
}
