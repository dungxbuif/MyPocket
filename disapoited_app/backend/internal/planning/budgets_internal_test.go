package planning

import (
	"math"
	"testing"
)

func TestPercentSpentDoesNotOverflow(t *testing.T) {
	tests := []struct {
		name   string
		spent  int64
		amount int64
		want   int64
	}{
		{name: "max amount fully spent", spent: math.MaxInt64, amount: math.MaxInt64, want: 100},
		{name: "large amount half spent", spent: math.MaxInt64 / 2, amount: math.MaxInt64, want: 49},
		{name: "over budget", spent: math.MaxInt64, amount: math.MaxInt64 / 2, want: 200},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := percentSpent(tt.spent, tt.amount); got != tt.want {
				t.Fatalf("percentSpent(%d, %d) = %d, want %d", tt.spent, tt.amount, got, tt.want)
			}
		})
	}
}
