package loadtext

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Heliodex/tracker/types"
)

// formatHex formats a hex column, using ".." for zero
func formatHex(v uint8) string {
	if v == 0 {
		return ".."
	}

	return fmt.Sprintf("%02X", v)
}

func formatCell(c ChannelData) (string, error) {
	if c.Note > 97 {
		return "", fmt.Errorf("note outside range: %d", c.Note)
	}

	return strings.Join([]string{
		types.Note(c.Note).String(),
		formatHex(c.Instrument),
		formatHex(c.Volume),
		formatHex(c.Effect),
		formatHex(c.EffectParameter),
	}, " "), nil
}

// WriteText writes tf to w in the text format
func WriteText(w io.Writer, tf *TextFormat) error {
	if tf.NumChannels < 1 || tf.NumChannels > maxChannels {
		return fmt.Errorf("invalid NumChannels value: %d", tf.NumChannels)
	}

	if int(tf.NumPatterns) != len(tf.Patterns) {
		return fmt.Errorf("NumPatterns is %d but %d patterns are present", tf.NumPatterns, len(tf.Patterns))
	}

	name := strings.TrimRight(string(tf.Name[:]), "\x00")
	if strings.ContainsAny(name, "\r\n") {
		return errors.New("name contains a line break")
	}

	var b strings.Builder

	fmt.Fprintf(&b, "Name: %s\n", name)
	fmt.Fprintf(&b, "NumChannels: %d\n", tf.NumChannels)
	fmt.Fprintf(&b, "NumPatterns: %d\n", tf.NumPatterns)

	b.WriteString("Order:")

	for _, o := range tf.OrderTable {
		fmt.Fprintf(&b, " %d", o)
	}

	b.WriteString("\n")

	for i, pattern := range tf.Patterns {
		if len(pattern) < 1 || len(pattern) > maxRows {
			return fmt.Errorf("pattern %d: row count out of range: %d", i, len(pattern))
		}

		fmt.Fprintf(&b, "\nPattern %d\n", i)
		fmt.Fprintf(&b, "Rows: %d\n", len(pattern))

		for r, row := range pattern {
			if len(row) != int(tf.NumChannels) {
				return fmt.Errorf("pattern %d row %d: expected %d channels, got %d", i, r, tf.NumChannels, len(row))
			}

			cells := make([]string, len(row))

			for c, cd := range row {
				cell, err := formatCell(cd)
				if err != nil {
					return fmt.Errorf("pattern %d row %d channel %d: %w", i, r, c+1, err)
				}

				cells[c] = cell
			}

			fmt.Fprintf(&b, "Row %03d: %s\n", r, strings.Join(cells, " | "))
		}
	}

	_, err := io.WriteString(w, b.String())

	return err
}
