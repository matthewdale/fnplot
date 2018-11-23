package fnplot

import (
	"unicode"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
)

type Generator gopter.Gen

// Floating point generators.
// ==========================

func Float64Range(min, max float64) Generator {
	return Generator(gen.Float64Range(min, max))
}

func Float64() Generator {
	return Generator(gen.Float64())
}

func Float32Range(min, max float32) Generator {
	return Generator(gen.Float32Range(min, max))
}

func Float32() Generator {
	return Generator(gen.Float32())
}

// Rune generators.
// ================

func RuneRange(min, max rune) Generator {
	return Generator(gen.RuneRange(min, max))
}

func Rune() Generator {
	return Generator(gen.Rune())
}

func RuneNoControl() Generator {
	return Generator(RuneNoControl())
}

// Character generators.
// =====================

func NumChar() Generator {
	return Generator(gen.NumChar())
}

func AlphaUpperChar() Generator {
	return Generator(gen.AlphaUpperChar())
}

func AlphaLowerChar() Generator {
	return Generator(gen.AlphaLowerChar())
}

func AlphaChar() Generator {
	return Generator(gen.AlphaChar())
}

func AlphaNumChar() Generator {
	return Generator(gen.AlphaNumChar())
}

func UnicodeChar(table *unicode.RangeTable) Generator {
	return Generator(gen.UnicodeChar(table))
}

// String generators.
// ==================

func AnyString() Generator {
	return Generator(gen.AnyString())
}

func AlphaString() Generator {
	return Generator(gen.AlphaString())
}

func NumString() Generator {
	return Generator(gen.NumString())
}

func Identifier() Generator {
	return Generator(gen.Identifier())
}

func UnicodeString(table *unicode.RangeTable) Generator {
	return Generator(gen.UnicodeString(table))
}
