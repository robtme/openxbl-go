package openxbl

import (
	"encoding/json"
	"testing"
)

func TestTimeUnmarshal(t *testing.T) {
	// A friend carries both layouts at once: addedDateTimeUtc is zoned, while
	// lastSeenDateTimeUtc omits the zone despite also being UTC.
	payload := `{
		"zoned": "2026-01-24T22:18:24Z",
		"zonedFraction": "2024-04-13T21:29:27.4390000Z",
		"naive": "2026-06-14T20:55:06.6703089",
		"naiveNoFraction": "2026-06-14T20:55:06",
		"absent": null
	}`

	var stamps map[string]Time
	if err := json.Unmarshal([]byte(payload), &stamps); err != nil {
		t.Fatalf("unmarshal timestamps: %v", err)
	}

	for _, name := range []string{"zoned", "zonedFraction", "naive", "naiveNoFraction"} {
		stamp := stamps[name]
		if stamp.IsZero() {
			t.Errorf("%s did not parse", name)

			continue
		}

		if stamp.Location() != nil && stamp.UTC().Year() != 2026 && stamp.UTC().Year() != 2024 {
			t.Errorf("%s parsed to unexpected year %d", name, stamp.Year())
		}
	}

	if !stamps["absent"].IsZero() {
		t.Error("null should leave the timestamp zero")
	}

	var bad map[string]Time
	if err := json.Unmarshal([]byte(`{"x": "not a timestamp"}`), &bad); err == nil {
		t.Error("an unparseable timestamp should error")
	}
}
