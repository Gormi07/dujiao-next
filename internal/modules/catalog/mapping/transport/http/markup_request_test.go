package mappinghttp

import (
	"encoding/json"
	"testing"
)

func TestUpdateProductMappingMarkupRequestAcceptsNumberOrStringPercent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		raw  string
		want string
	}{
		{name: "json number", raw: `{"markup_mode":"markup_percent","markup_percent":40}`, want: "40"},
		{name: "json number with fraction", raw: `{"markup_mode":"markup_percent","markup_percent":40.5}`, want: "40.5"},
		{name: "json string", raw: `{"markup_mode":"markup_percent","markup_percent":"40"}`, want: "40"},
		{name: "omitted percent", raw: `{"markup_mode":"inherit"}`, want: ""},
		{name: "null percent", raw: `{"markup_mode":"inherit","markup_percent":null}`, want: ""},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var req UpdateProductMappingMarkupRequest
			if err := json.Unmarshal([]byte(tc.raw), &req); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if req.MarkupMode == "" {
				t.Fatal("markup_mode empty")
			}
			got := req.MarkupPercent.String()
			if got != tc.want {
				t.Fatalf("markup_percent = %q, want %q", got, tc.want)
			}
		})
	}
}
