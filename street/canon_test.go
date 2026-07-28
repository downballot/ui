package street

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanon(t *testing.T) {
	rows := []struct {
		input  string
		output string
	}{
		{
			input:  "123 Main St, Anytown, USA",
			output: "Main St o000000123#, Anytown, USA",
		},
		{
			input:  "123 Main St #1, Anytown, USA",
			output: "Main St o000000123#000000001, Anytown, USA",
		},
		{
			input:  "123 Main St #1A, Anytown, USA",
			output: "Main St o000000123#000000001A, Anytown, USA",
		},
		{
			input:  "123 Main St #10, Anytown, USA",
			output: "Main St o000000123#000000010, Anytown, USA",
		},
		{
			input:  "11 Thorn Ln #10, Anytown, DE 12345",
			output: "Thorn Ln o000000011#000000010, Anytown, DE 12345",
		},
		{
			input:  "11 Thorn Ln #2, Anytown, DE 12345",
			output: "Thorn Ln o000000011#000000002, Anytown, DE 12345",
		},
	}
	for rowIndex, row := range rows {
		t.Run(fmt.Sprintf("%d", rowIndex), func(t *testing.T) {
			output := Canon(row.input)
			assert.Equal(t, row.output, output)
		})
	}
}
