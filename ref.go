package observe

import "fmt"

// FormatRef formats a human-readable reference like OBS-0001 or FAQ-0042.
func FormatRef(prefix string, seq int64) string {
	if seq < 10000 {
		return fmt.Sprintf("%s-%04d", prefix, seq)
	}
	return fmt.Sprintf("%s-%d", prefix, seq)
}
