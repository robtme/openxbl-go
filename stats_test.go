package openxbl

import (
	"encoding/json"
	"testing"
)

func TestStatGetInt(t *testing.T) {
	// Values are always quoted, but a title's stats aren't all numeric.
	payload := `[
		{"titleid": "1", "name": "MinutesPlayed", "type": "Integer", "value": "1250"},
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

func TestGroupStats(t *testing.T) {
	// Trimmed from a real response: grouped stats carry no title ID of their own, a stat
	// not yet recorded has no value, and a title with no hero stats has an empty list.
	// Stats arrive out of Ordinal order, as some titles return them.
	payload := `{
		"groups": [
			{"name": "Hero", "titleid": "1278031123", "statlistscollection": [{"stats": [
				{"groupproperties": {"Ordinal": "7", "DisplayName": "Games Cleared with Jill", "DisplayFormat": "Integer"},
				 "name": "ClearPlayer.PlayerId.1", "type": "Integer"},
				{"groupproperties": {"Ordinal": "1", "DisplayName": "Zombie Kills", "DisplayFormat": "Integer", "DisplaySemantic": "Cumulative"},
				 "name": "DefeateCreatureOnly.EnemyId.0", "type": "Integer", "value": "25"}
			]}]},
			{"name": "Hero", "titleid": "2043073184", "statlistscollection": []}
		],
		"statlistscollection": []
	}`

	var response statsResponse
	if err := json.Unmarshal([]byte(payload), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	grouped := response.groupStats()

	stats, ok := grouped["1278031123"]
	if !ok || len(stats) != 2 {
		t.Fatalf("Resident Evil stats = %v; want 2", stats)
	}

	if stats[0].TitleID != "1278031123" || stats[0].GroupProperties.DisplayName != "Zombie Kills" || stats[0].Value != "25" {
		t.Errorf("first stat = %+v", stats[0])
	}

	if stats[1].Value != "" {
		t.Errorf("unrecorded stat value = %q; want empty", stats[1].Value)
	}

	if stats[1].GroupProperties.Ordinal != 7 {
		t.Errorf("second stat ordinal = %d; want 7", stats[1].GroupProperties.Ordinal)
	}

	if empty, present := grouped["2043073184"]; !present || empty == nil || len(empty) != 0 {
		t.Errorf("a title with no hero stats should be present and empty, got %v, %t", empty, present)
	}
}
