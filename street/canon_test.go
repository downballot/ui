package street

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanon(t *testing.T) {
	rows := []struct {
		input        string
		splitEvenOdd bool
		output       string
	}{
		{
			input:        "122 Main St, Anytown, USA",
			splitEvenOdd: true,
			output:       "Main St e000000122#, Anytown, USA",
		},
		{
			input:        "123 Main St, Anytown, USA",
			splitEvenOdd: true,
			output:       "Main St o000000123#, Anytown, USA",
		},
		{
			input:        "123 Main St, Anytown, USA",
			splitEvenOdd: false,
			output:       "Main St 000000123#, Anytown, USA",
		},
		{
			input:        "123A Main St, Anytown, USA",
			splitEvenOdd: true,
			output:       "Main St o000000123A#, Anytown, USA",
		},
		{
			input:        "123 Main St #1, Anytown, USA",
			splitEvenOdd: true,
			output:       "Main St o000000123#000000001, Anytown, USA",
		},
		{
			input:        "123 Main St #1, Anytown, USA",
			splitEvenOdd: false,
			output:       "Main St 000000123#000000001, Anytown, USA",
		},
		{
			input:        "123 Main St #1A, Anytown, USA",
			splitEvenOdd: true,
			output:       "Main St o000000123#000000001A, Anytown, USA",
		},
		{
			input:        "123 Main St #10, Anytown, USA",
			splitEvenOdd: true,
			output:       "Main St o000000123#000000010, Anytown, USA",
		},
		{
			input:        "11 Thorn Ln #10, Anytown, DE 12345",
			splitEvenOdd: true,
			output:       "Thorn Ln o000000011#000000010, Anytown, DE 12345",
		},
		{
			input:        "11 Thorn Ln #2, Anytown, DE 12345",
			splitEvenOdd: true,
			output:       "Thorn Ln o000000011#000000002, Anytown, DE 12345",
		},
		{
			input:        "OFFICE\\CLUBHOUSE, 312 Christina Mill Dr, Newark, DE 19711",
			splitEvenOdd: true,
			output:       "OFFICE\\CLUBHOUSE, 312 Christina Mill Dr, Newark, DE 19711",
		},
	}
	for rowIndex, row := range rows {
		t.Run(fmt.Sprintf("%d", rowIndex), func(t *testing.T) {
			output := Canon(row.input, row.splitEvenOdd)
			assert.Equal(t, row.output, output)
		})
	}
}
