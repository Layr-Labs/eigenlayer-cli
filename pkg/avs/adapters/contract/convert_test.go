package contract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertEqual(t *testing.T, expected interface{}, value interface{}, kind string) {
	actual, err := convert(value, kind)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestConvertInt(t *testing.T) {
	assertEqual(t, int(12345678), int(12345678), "int")
	assertEqual(t, int(123), int8(123), "int")
	assertEqual(t, int(123), int16(123), "int")
	assertEqual(t, int(12345678), int32(12345678), "int")
	assertEqual(t, int(12345678), int64(12345678), "int")
	assertEqual(t, int(12345678), "12345678", "int")
	assertEqual(t, int(12345678), uint(12345678), "int")
	assertEqual(t, int(123), uint8(123), "int")
	assertEqual(t, int(123), uint16(123), "int")
	assertEqual(t, int(12345678), uint32(12345678), "int")
	assertEqual(t, int(12345678), uint64(12345678), "int")
}

func TestConvertInt8(t *testing.T) {
	assertEqual(t, int8(123), int(123), "int8")
	assertEqual(t, int8(123), int8(123), "int8")
	assertEqual(t, int8(123), int16(123), "int8")
	assertEqual(t, int8(123), int32(123), "int8")
	assertEqual(t, int8(123), int64(123), "int8")
	assertEqual(t, int8(123), "123", "int8")
	assertEqual(t, int8(123), uint(123), "int8")
	assertEqual(t, int8(123), uint8(123), "int8")
	assertEqual(t, int8(123), uint16(123), "int8")
	assertEqual(t, int8(123), uint32(123), "int8")
	assertEqual(t, int8(123), uint64(123), "int8")
}

func TestConvertInt16(t *testing.T) {
	assertEqual(t, int16(12345), int(12345), "int16")
	assertEqual(t, int16(123), int8(123), "int16")
	assertEqual(t, int16(12345), int16(12345), "int16")
	assertEqual(t, int16(12345), int32(12345), "int16")
	assertEqual(t, int16(12345), int64(12345), "int16")
	assertEqual(t, int16(12345), "12345", "int16")
	assertEqual(t, int16(12345), uint(12345), "int16")
	assertEqual(t, int16(123), uint8(123), "int16")
	assertEqual(t, int16(12345), uint16(12345), "int16")
	assertEqual(t, int16(12345), uint32(12345), "int16")
	assertEqual(t, int16(12345), uint64(12345), "int16")
}

func TestConvertInt32(t *testing.T) {
	assertEqual(t, int32(12345678), int(12345678), "int32")
	assertEqual(t, int32(123), int8(123), "int32")
	assertEqual(t, int32(123), int16(123), "int32")
	assertEqual(t, int32(12345678), int32(12345678), "int32")
	assertEqual(t, int32(12345678), int64(12345678), "int32")
	assertEqual(t, int32(12345678), "12345678", "int32")
	assertEqual(t, int32(12345678), uint(12345678), "int32")
	assertEqual(t, int32(123), uint8(123), "int32")
	assertEqual(t, int32(123), uint16(123), "int32")
	assertEqual(t, int32(12345678), uint32(12345678), "int32")
	assertEqual(t, int32(12345678), uint64(12345678), "int32")
}

func TestConvertInt64(t *testing.T) {
	assertEqual(t, int64(12345678), int(12345678), "int64")
	assertEqual(t, int64(123), int8(123), "int64")
	assertEqual(t, int64(123), int16(123), "int64")
	assertEqual(t, int64(12345678), int32(12345678), "int64")
	assertEqual(t, int64(12345678), int64(12345678), "int64")
	assertEqual(t, int64(12345678), "12345678", "int64")
	assertEqual(t, int64(12345678), uint(12345678), "int64")
	assertEqual(t, int64(123), uint8(123), "int64")
	assertEqual(t, int64(123), uint16(123), "int64")
	assertEqual(t, int64(12345678), uint32(12345678), "int64")
	assertEqual(t, int64(12345678), uint64(12345678), "int64")
}

func TestConvertUint(t *testing.T) {
	assertEqual(t, uint(12345678), int(12345678), "uint")
	assertEqual(t, uint(123), int8(123), "uint")
	assertEqual(t, uint(123), int16(123), "uint")
	assertEqual(t, uint(12345678), int32(12345678), "uint")
	assertEqual(t, uint(12345678), int64(12345678), "uint")
	assertEqual(t, uint(12345678), "12345678", "uint")
	assertEqual(t, uint(12345678), uint(12345678), "uint")
	assertEqual(t, uint(123), uint8(123), "uint")
	assertEqual(t, uint(123), uint16(123), "uint")
	assertEqual(t, uint(12345678), uint32(12345678), "uint")
	assertEqual(t, uint(12345678), uint64(12345678), "uint")
}

func TestConvertUint8(t *testing.T) {
	assertEqual(t, uint8(123), int(123), "uint8")
	assertEqual(t, uint8(123), int8(123), "uint8")
	assertEqual(t, uint8(123), int16(123), "uint8")
	assertEqual(t, uint8(123), int32(123), "uint8")
	assertEqual(t, uint8(123), int64(123), "uint8")
	assertEqual(t, uint8(123), "123", "uint8")
	assertEqual(t, uint8(123), uint(123), "uint8")
	assertEqual(t, uint8(123), uint8(123), "uint8")
	assertEqual(t, uint8(123), uint16(123), "uint8")
	assertEqual(t, uint8(123), uint32(123), "uint8")
	assertEqual(t, uint8(123), uint64(123), "uint8")
}

func TestConvertUint16(t *testing.T) {
	assertEqual(t, uint16(12345), int(12345), "uint16")
	assertEqual(t, uint16(123), int8(123), "uint16")
	assertEqual(t, uint16(12345), int16(12345), "uint16")
	assertEqual(t, uint16(12345), int32(12345), "uint16")
	assertEqual(t, uint16(12345), int64(12345), "uint16")
	assertEqual(t, uint16(12345), "12345", "uint16")
	assertEqual(t, uint16(12345), uint(12345), "uint16")
	assertEqual(t, uint16(123), uint8(123), "uint16")
	assertEqual(t, uint16(12345), uint16(12345), "uint16")
	assertEqual(t, uint16(12345), uint32(12345), "uint16")
	assertEqual(t, uint16(12345), uint64(12345), "uint16")
}

func TestConvertUint32(t *testing.T) {
	assertEqual(t, uint32(12345678), int(12345678), "uint32")
	assertEqual(t, uint32(123), int8(123), "uint32")
	assertEqual(t, uint32(123), int16(123), "uint32")
	assertEqual(t, uint32(12345678), int32(12345678), "uint32")
	assertEqual(t, uint32(12345678), int64(12345678), "uint32")
	assertEqual(t, uint32(12345678), "12345678", "uint32")
	assertEqual(t, uint32(12345678), uint(12345678), "uint32")
	assertEqual(t, uint32(123), uint8(123), "uint32")
	assertEqual(t, uint32(123), uint16(123), "uint32")
	assertEqual(t, uint32(12345678), uint32(12345678), "uint32")
	assertEqual(t, uint32(12345678), uint64(12345678), "uint32")
}

func TestConvertUint64(t *testing.T) {
	assertEqual(t, uint64(12345678), int(12345678), "uint64")
	assertEqual(t, uint64(123), int8(123), "uint64")
	assertEqual(t, uint64(123), int16(123), "uint64")
	assertEqual(t, uint64(12345678), int32(12345678), "uint64")
	assertEqual(t, uint64(12345678), int64(12345678), "uint64")
	assertEqual(t, uint64(12345678), "12345678", "uint64")
	assertEqual(t, uint64(12345678), uint(12345678), "uint64")
	assertEqual(t, uint64(123), uint8(123), "uint64")
	assertEqual(t, uint64(123), uint16(123), "uint64")
	assertEqual(t, uint64(12345678), uint32(12345678), "uint64")
	assertEqual(t, uint64(12345678), uint64(12345678), "uint64")
}
