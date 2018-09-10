package fnplot

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"

	"github.com/pkg/errors"
)

// A Values is a collection of any type of value. A Values can be converted to a
// scalar value (a floating point number).
type Values []interface{}

// smallestInt returns the smallest fixed-size signed or unsigned integer value
// necessary to store the given variable-size signed integer value.
func smallestInt(x int) interface{} {
	if x >= 0 {
		return smallestUint(uint(x))
	}
	switch {
	case x <= math.MaxInt8 && x >= math.MinInt8:
		return int8(x)
	case x <= math.MaxInt16 && x >= math.MinInt16:
		return int16(x)
	case x <= math.MaxInt32 && x >= math.MinInt32:
		return int32(x)
	}
	return int64(x)
}

// smallestUint returns the smallest fixed-size unsigned integer value necessary
// to store the given variable-size unsigned integer value.
func smallestUint(x uint) interface{} {
	switch {
	case x <= math.MaxUint8:
		return uint8(x)
	case x <= math.MaxUint16:
		return uint16(x)
	case x <= math.MaxUint32:
		return uint32(x)
	}
	return uint64(x)
}

func writeBinary(buf *bytes.Buffer, data interface{}) error {
	if data == nil {
		return nil
	}

	rv := reflect.ValueOf(data)
	// TODO: Iterative indirect? Kind of an edge case, but not a bad idea.
	rv = reflect.Indirect(rv)

	// Write types that a bytes.Buffer natively supports.
	switch v := rv.Interface().(type) {
	case byte:
		err := buf.WriteByte(v)
		return errors.WithMessage(err, "error writing byte to writer")
	case []byte:
		_, err := buf.Write(v)
		return errors.WithMessage(err, "error writing []byte to writer")
	case rune:
		_, err := buf.WriteRune(v)
		return errors.WithMessage(err, "error writing rune to writer")
	case string:
		_, err := buf.WriteString(v)
		return errors.WithMessage(err, "error writing string to writer")
	}

	// Unpack slice, array, and map types.
	switch rv.Type().Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			err := writeBinary(buf, rv.Index(i).Interface())
			if err != nil {
				return errors.WithMessage(
					err,
					"error writing binary for slice or array value at "+strconv.Itoa(i))
			}
		}
		return nil
	case reflect.Map:
		for _, mapKey := range rv.MapKeys() {
			mapVal := rv.MapIndex(mapKey)
			err := writeBinary(buf, Values{mapKey.Interface(), mapVal.Interface()})
			if err != nil {
				return errors.WithMessage(
					err,
					"error writing binary for map key/value pair "+mapKey.String())
			}
		}
		return nil
	}

	// Anything else, defer to binary.Write.
	// binary.Write can't handle "int" and "uint" types because they can be
	// variable bit widths, so convert them to the smallest possible bit width
	// integer that can hold the value before passing them to binary.Write.
	switch v := rv.Interface().(type) {
	case int:
		data = smallestInt(v)
	case uint:
		data = smallestUint(v)
	}
	err := binary.Write(buf, binary.BigEndian, data)
	return errors.WithMessage(
		err,
		fmt.Sprintf("error converting value to binary: %#v", data))
}

// Scalar converts a Values to an arbitrary precision floating point number. The
// scalar value conversion depends on the type of input value.
//
// Individual values that are already scalar values (floats and ints) are returned
// as their original value.
//
// Collections of values (slices, arrays, and maps) are unpacked into individual
// values. All individual values are converted to their binary representation and
// appended to a byte slice. When all values are appended to the byte buffer, the
// bytes are interpreted as a big-endian integer value.
func (vs Values) Scalar() (*big.Float, error) {
	// Return the zero value of a *big.Float if the input is empty.
	if len(vs) == 0 {
		return big.NewFloat(0), nil
	}

	// Convert some individual scalar values directly to a *big.Float
	if len(vs) == 1 && vs[0] != nil {
		switch vt := vs[0].(type) {
		case float32:
			return big.NewFloat(float64(vt)), nil
		case *float32:
			return big.NewFloat(float64(*vt)), nil
		case float64:
			return big.NewFloat(vt), nil
		case *float64:
			return big.NewFloat(*vt), nil
		}
	}

	// Convert everything else into bytes, interpret those bytes as a variable
	// precision integer, and return that integer represented as a *big.Float
	buf := bytes.NewBuffer(nil)
	if err := writeBinary(buf, vs); err != nil {
		return nil, errors.WithMessage(err, "error writing values as binary")
	}
	return big.NewFloat(0).SetInt(big.NewInt(0).SetBytes(buf.Bytes())), nil
}
