package share

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntToBytes(t *testing.T) {
	t.Run("should convert an int to []byte", func(t *testing.T) {
		type test struct {
			input  int
			result []byte
		}

		tests := []test{
			{4, []byte{0b00000000, 0b00000000, 0b00000000, 0b00000100}},
			{515, []byte{0b00000000, 0b00000000, 0b00000010, 0b00000011}},
			{245080, []byte{0b00000000, 0b00000011, 0b10111101, 0b01011000}},
		}

		for _, test := range tests {
			actual := IntToBytes(test.input)
			assert.True(
				t,
				bytes.Equal(test.result[:], actual[:]),
				fmt.Sprintf("converting %d to []byte failed", test.input),
			)
		}
	})
}
