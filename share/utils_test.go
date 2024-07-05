package share

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUint32ToByte4(t *testing.T) {
	t.Run("should convert a uint to [4]byte", func(t *testing.T) {
		type test struct {
			input  uint32
			result [4]byte
		}

		tests := []test{
			{4, [4]byte{0b00000000, 0b00000000, 0b00000000, 0b00000100}},
			{515, [4]byte{0b00000000, 0b00000000, 0b00000010, 0b00000011}},
			{245080, [4]byte{0b00000000, 0b00000011, 0b10111101, 0b01011000}},
		}

		for _, test := range tests {
			actual := Uint32ToByte4(test.input)
			assert.True(
				t,
				bytes.Equal(test.result[:], actual[:]),
				fmt.Sprintf("converting %d to [4]byte failed", test.input),
			)
		}
	})
}
