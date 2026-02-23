package output

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// ControlResult holds the aggregated savings for a single control.
type ControlResult struct {
	ControlName    string
	CloudProvider  string
	AssetCount     int
	MonthlySavings float64
}

// Formatter writes the savings report to the given writer.
type Formatter struct {
	w io.Writer
}

// NewFormatter creates a Formatter that writes to w.
func NewFormatter(w io.Writer) *Formatter {
	return &Formatter{w: w}
}

// Render outputs the savings report grouped by control.
func (f *Formatter) Render(results []ControlResult) {
	// Sort by savings descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].MonthlySavings > results[j].MonthlySavings
	})

	var totalAssets int
	var totalSavings float64

	fmt.Fprintln(f.w, "Cloud Savings Report")
	fmt.Fprintln(f.w, "====================")
	fmt.Fprintln(f.w)
	fmt.Fprintf(f.w, "  %-8s %-50s %6s   %15s\n", "Provider", "Control", "Assets", "Monthly Savings")
	fmt.Fprintf(f.w, "  %s\n", strings.Repeat("─", 83))

	for _, r := range results {
		if r.AssetCount == 0 {
			continue
		}
		fmt.Fprintf(f.w, "  %-8s %-50s %6d   %15s\n", strings.ToUpper(r.CloudProvider), truncate(r.ControlName, 50), r.AssetCount, formatUSD(r.MonthlySavings))
		totalAssets += r.AssetCount
		totalSavings += r.MonthlySavings
	}

	fmt.Fprintf(f.w, "  %s\n", strings.Repeat("═", 83))
	fmt.Fprintf(f.w, "  %-8s %-50s %6d   %15s\n", "", "TOTAL POTENTIAL MONTHLY SAVINGS", totalAssets, formatUSD(totalSavings))
	fmt.Fprintln(f.w)
}

func formatUSD(amount float64) string {
	// Format with 2 decimal places, then insert commas in the integer part.
	raw := fmt.Sprintf("%.2f", amount)
	parts := strings.SplitN(raw, ".", 2)
	intPart := parts[0]
	decPart := parts[1]

	// Insert commas every 3 digits from the right
	n := len(intPart)
	if n <= 3 {
		return "$" + intPart + "." + decPart
	}
	var b strings.Builder
	remainder := n % 3
	if remainder > 0 {
		b.WriteString(intPart[:remainder])
	}
	for i := remainder; i < n; i += 3 {
		if b.Len() > 0 {
			b.WriteByte(',')
		}
		b.WriteString(intPart[i : i+3])
	}
	return "$" + b.String() + "." + decPart
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
