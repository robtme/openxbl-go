package openxbl

import (
	"encoding/json"
	"testing"
)

func TestStatGetInt(t *testing.T) {
	// The API quotes most stat values but not all, and a title's stats aren't all
	// numeric, so decoding must survive every combination.
	payload := `[
		{"titleid": "1", "name": "MinutesPlayed", "type": "Integer", "value": 1250},
		{"titleid": "2", "name": "MinutesPlayed", "type": "Integer", "value": "480"},
		{"titleid": "3", "name": "Rank", "type": "String", "value": "Gold"},
		{"titleid": "4", "name": "Missing", "type": "Integer", "value": null}
	]`

	var stats []*Stat
	if err := json.Unmarshal([]byte(payload), &stats); err != nil {
		t.Fatalf("unmarshal stats: %v", err)
	}

	want := []struct {
		value int64
		ok    bool
	}{{1250, true}, {480, true}, {0, false}, {0, false}}

	for i, expected := range want {
		value, ok := stats[i].GetInt()
		if value != expected.value || ok != expected.ok {
			t.Errorf("stats[%d].GetInt() = %d, %t; want %d, %t", i, value, ok, expected.value, expected.ok)
		}
	}
}
