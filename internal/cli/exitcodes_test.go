package cli

import "testing"

func TestExitCodes_MatchSpec(t *testing.T) {
	tests := []struct {
		name     string
		got      int
		wantCode int
	}{
		{"ExitSuccess", ExitSuccess, 0},
		{"ExitGeneric", ExitGeneric, 1},
		{"ExitUsage", ExitUsage, 2},
		{"ExitNotFound", ExitNotFound, 3},
		{"ExitAlreadyExists", ExitAlreadyExists, 4},
		{"ExitPkgManager", ExitPkgManager, 5},
		{"ExitDBOperation", ExitDBOperation, 6},
		{"ExitAborted", ExitAborted, 7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.wantCode {
				t.Errorf("%s = %d, want %d", tt.name, tt.got, tt.wantCode)
			}
		})
	}
}

func TestExitCodes_AllUnique(t *testing.T) {
	codes := []int{
		ExitSuccess,
		ExitGeneric,
		ExitUsage,
		ExitNotFound,
		ExitAlreadyExists,
		ExitPkgManager,
		ExitDBOperation,
		ExitAborted,
	}
	seen := make(map[int]bool)
	for _, c := range codes {
		if seen[c] {
			t.Errorf("duplicate exit code: %d", c)
		}
		seen[c] = true
	}
}
