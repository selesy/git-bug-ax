package issue

import "github.com/selesy/git-bug-ax/internal/types"

var _ types.TextCodec = (*Description)(nil)

type Description struct {
	// TODO: this is a placeholder for Markdown but we'll need to parse sections later
	description string
}

func (d *Description) MarshalText() ([]byte, error) {
	return []byte(d.description), nil
}

func (d *Description) UnmarshalText(s []byte) error {
	d.description = string(s)

	return nil
}
