package openxbl

import (
	"encoding/json"
	"testing"
)

func TestAchievementRewards(t *testing.T) {
	payload := `{
		"progressState": "Achieved",
		"rewards": [
			{"value": "https://example.com/art.png", "type": "Art", "valueType": "String"},
			{"value": "10", "type": "Gamerscore", "valueType": "Int"}
		]
	}`

	var achievement Achievement
	if err := json.Unmarshal([]byte(payload), &achievement); err != nil {
		t.Fatalf("unmarshal achievement: %v", err)
	}

	if got := achievement.GetGamerscore(); got != 10 {
		t.Errorf("gamerscore = %d, want 10", got)
	}

	if !achievement.IsUnlocked() {
		t.Error("achievement should be unlocked")
	}

	var empty Achievement
	if got := empty.GetGamerscore(); got != 0 {
		t.Errorf("gamerscore without rewards = %d, want 0", got)
	}
}
