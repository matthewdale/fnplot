package fnplot

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScalar(t *testing.T) {
	tests := []struct {
		description string
		values      Values
		expected    *big.Float
	}{
		{
			description: "Empty values",
			values:      Values{},
			expected:    big.NewFloat(0),
		},
		{
			description: "int value",
			values:      Values{123},
			expected:    big.NewFloat(123),
		},
		{
			description: "float value",
			values:      Values{123.456},
			expected:    big.NewFloat(123.456),
		},
		{
			description: "byte value",
			values:      Values{byte('d')},
			expected:    big.NewFloat(100),
		},
		{
			description: "byte slice value",
			values:      Values{[]byte("test")},
			expected:    big.NewFloat(1952805748),
		},
		{
			description: "rune value",
			values:      Values{'æ'},
			expected:    big.NewFloat(50086),
		},
		{
			description: "string value",
			values:      Values{"test"},
			expected:    big.NewFloat(1952805748),
		},
		{
			description: "Nil values should be ignored",
			values:      Values{nil, "test", nil},
			expected:    big.NewFloat(1952805748),
		},
		{
			description: "Large int value",
			values:      Values{math.MaxInt32 + 1},
			expected:    big.NewFloat(float64(math.MaxInt32 + 1)),
		},
		{
			description: "Large uint value",
			values:      Values{math.MaxUint32 + 1},
			expected:    big.NewFloat(float64(math.MaxUint32 + 1)),
		},
		{
			description: "map value",
			values:      Values{map[string]int{"a": 1}},
			expected:    big.NewFloat(24833),
		},
	}
	for _, test := range tests {
		test := test // Capture range variable.
		t.Run(test.description, func(t *testing.T) {
			s, err := test.values.Scalar()
			require.NoError(t, err, "Error calculating scalar value")
			assert.Equal(t, test.expected, s, "Expected and actual values are different")
		})
	}
}
