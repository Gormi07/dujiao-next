package mappinghttp

import (
	"bytes"
	"encoding/json"
	"strings"
)

// flexibleDecimalText accepts a JSON string or number for decimal-like fields.
// Vue number inputs commonly serialize markup_percent as a JSON number.
type flexibleDecimalText string

func (v flexibleDecimalText) String() string { return string(v) }

func (v *flexibleDecimalText) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*v = ""
		return nil
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*v = flexibleDecimalText(strings.TrimSpace(s))
		return nil
	}
	*v = flexibleDecimalText(string(data))
	return nil
}
