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
type Values []reflect.Value

func NewValues(args ...interface{}) Values {
	values := make(Values, len(args))
	for i := range args {
		values[i] = reflect.ValueOf(args[i])
	}
	return values
}

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

func writeBinary(buf *bytes.Buffer, value reflect.Value) error {
	if !value.IsValid() {
		return nil
	}

	value = indirect(value)

	// Unpack slice, array, and map types.
	switch value.Type().Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			err := writeBinary(buf, value.Index(i))
			if err != nil {
				return errors.WithMessage(
					err,
					"error writing binary for slice or array value at "+strconv.Itoa(i))
			}
		}
		return nil
	case reflect.Map:
		for _, mapKey := range value.MapKeys() {
			err := writeBinary(buf, mapKey)
			if err != nil {
				return errors.WithMessage(
					err,
					"error writing binary for map key "+mapKey.String())
			}
			err = writeBinary(buf, value.MapIndex(mapKey))
			if err != nil {
				return errors.WithMessage(
					err,
					"error writing binary for map value at key "+mapKey.String())
			}
		}
		return nil
	}

	// Handle the rest of the types as interface{} and defer to binary.Write. If
	// the value cannot be converted to interface{} here, we don't know how to
	// handle it.
	if !value.CanInterface() {
		return errors.New("Unsupported type: " + value.Type().String())
	}
	iValue := value.Interface()
	if iValue == nil {
		return nil
	}

	// Write any types that a bytes.Buffer natively supports to the buffer. If
	// the type is "int" or "uint", convert the value to the smallest possible
	// bit width integer that can hold the value because binary.Write can't
	// handle "int" and "uint" types because they can be variable bit widths.
	switch v := iValue.(type) {
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
	case int:
		iValue = smallestInt(v)
	case uint:
		iValue = smallestUint(v)
	}

	err := binary.Write(buf, binary.BigEndian, iValue)
	return errors.WithMessage(
		err,
		fmt.Sprintf("error converting value to binary: %#v", value))
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
	if len(vs) == 1 {
		if !vs[0].IsValid() {
			return big.NewFloat(0), nil
		}
		value := indirect(vs[0])
		if value.Kind() == reflect.Float32 || value.Kind() == reflect.Float64 {
			return big.NewFloat(value.Float()), nil
		}
	}

	// Convert everything else into bytes, interpret those bytes as a variable
	// precision integer, and return that integer represented as a *big.Float
	buf := bytes.NewBuffer(nil)
	for _, value := range vs {
		if err := writeBinary(buf, value); err != nil {
			return nil, errors.WithMessage(err, "error writing values as binary")
		}
	}
	return big.NewFloat(0).SetInt(big.NewInt(0).SetBytes(buf.Bytes())), nil
}

func indirect(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}
	return v
}
