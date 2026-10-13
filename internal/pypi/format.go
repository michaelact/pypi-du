package pypi

import (
	"fmt"
	"strconv"
	"strings"
)

// FormatSize renders a byte count in human-readable units (binary, base-1024).
func FormatSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// ParseSize parses a human-friendly size such as "10GB", "500 MB", "1.5g",
// or a plain byte count like "1048576" into bytes.
func ParseSize(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" {
		return 0, fmt.Errorf("empty size")
	}
	// Split numeric prefix from unit suffix.
	i := 0
	for i < len(s) && (s[i] == '.' || (s[i] >= '0' && s[i] <= '9')) {
		i++
	}
	numPart := strings.TrimSpace(s[:i])
	unitPart := strings.TrimSpace(s[i:])
	if numPart == "" {
		return 0, fmt.Errorf("no number in size %q", s)
	}
	num, err := strconv.ParseFloat(numPart, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size %q: %w", s, err)
	}

	mult := int64(1)
	switch strings.TrimSuffix(unitPart, "B") {
	case "":
		mult = 1
	case "K":
		mult = 1 << 10
	case "M":
		mult = 1 << 20
	case "G":
		mult = 1 << 30
	case "T":
		mult = 1 << 40
	default:
		return 0, fmt.Errorf("unknown size unit %q", unitPart)
	}
	return int64(num * float64(mult)), nil
}
