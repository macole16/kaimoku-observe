package observe

import "testing"

func TestFormatRef(t *testing.T) {
	tests := []struct {
		prefix string
		seq    int64
		want   string
	}{
		{"OBS", 1, "OBS-0001"},
		{"OBS", 42, "OBS-0042"},
		{"OBS", 10000, "OBS-10000"},
		{"FAQ", 1, "FAQ-0001"},
	}
	for _, tt := range tests {
		got := FormatRef(tt.prefix, tt.seq)
		if got != tt.want {
			t.Errorf("FormatRef(%q, %d) = %q, want %q", tt.prefix, tt.seq, got, tt.want)
		}
	}
}
