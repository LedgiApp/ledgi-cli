package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

// PrintJSON pretty-prints a JSON raw message to stdout.
func PrintJSON(data json.RawMessage) {
	var out bytes.Buffer
	if err := json.Indent(&out, data, "", "  "); err != nil {
		// Fallback: print raw.
		fmt.Println(string(data))
		return
	}
	fmt.Println(out.String())
}

// Table prints rows as an aligned table to stdout.
// headers defines column names; rows is a slice of string slices.
func Table(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	fmt.Fprintln(w, strings.Repeat("─\t", len(headers)))
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	w.Flush()
}
