package prevotes

import "testing"

func TestValidatorColumns(t *testing.T) {
	tests := []struct {
		width, count, want int
	}{
		{0, 91, 1},
		{79, 91, 1},
		{80, 91, 2},
		{120, 91, 3},
		{200, 91, 5},
		{200, 2, 2},
		{120, 0, 3},
	}
	for _, tt := range tests {
		if got := validatorColumns(tt.width, tt.count); got != tt.want {
			t.Errorf("validatorColumns(%d, %d) = %d, want %d", tt.width, tt.count, got, tt.want)
		}
	}
}

func TestSplitVotes(t *testing.T) {
	votes := make([]VoteState, 91)
	for i := range votes {
		votes[i].Description = string(rune(i + 1))
	}
	groups := splitVotes(votes, 3)
	if len(groups) != 3 {
		t.Fatalf("got %d groups, want 3", len(groups))
	}
	if len(groups[0]) != 31 || len(groups[1]) != 30 || len(groups[2]) != 30 {
		t.Fatalf("unexpected group sizes: %d, %d, %d", len(groups), len(groups[0]), len(groups[1]), len(groups[2]))
	}
	index := 0
	for _, group := range groups {
		for _, vote := range group {
			if vote.Description != votes[index].Description {
				t.Fatalf("vote %d changed order", index)
			}
			index++
		}
	}
}
