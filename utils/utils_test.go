package utils_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/apourchet/commander/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringify(t *testing.T) {
	table := []struct {
		obj      interface{}
		expected string
	}{
		{1, `1`},
		{uint(20), `20`},
		{float64(2.1), `2.1`},
		{"asd", `asd`},
		{true, `true`},
		{[]int{1, 2, 3}, `[1,2,3]`},
		{map[string]int{"a": 1}, `{"a":1}`},
	}

	for _, test := range table {
		str, err := utils.Stringify(test.obj)
		require.NoError(t, err)
		require.Equal(t, test.expected, str)

		str, err = utils.StringifyValue(reflect.ValueOf(test.obj))
		require.NoError(t, err)
		require.Equal(t, test.expected, str)
	}
}

type MyStruct struct {
	B        bool
	S        string
	I        int
	I8       int8
	I16      int16
	I32      int32
	I64      int64
	UI       uint
	UI8      uint8
	UI16     uint16
	UI32     uint32
	UI64     uint64
	F32      float32
	F64      float64
	Slice    []string
	Map      map[string]string
	Duration time.Duration

	Ptr         *bool
	Unsupported map[int]string
}

func TestSetFieldFailures(t *testing.T) {
	obj := &MyStruct{}
	table := []struct {
		fieldname string
		value     string
	}{
		{"B", "true"},
		{"I", "1"},
		{"I8", "8"},
		{"I16", "16"},
		{"I32", "32"},
		{"I64", "64"},
		{"UI", "1"},
		{"UI8", "8"},
		{"UI16", "16"},
		{"UI32", "32"},
		{"UI64", "64"},
		{"F32", "2"},
		{"F64", "2.1"},
		{"Slice", `["a","b"]`},
		{"Map", `{"a":"b","c":"d"}`},
		{"Ptr", `true`},
		{"Unsupported", `{1: "asd"}`},
	}

	for _, test := range table {
		err := utils.SetField(obj, test.fieldname, test.value+"!")
		assert.Error(t, err)
	}
}

func TestSetGetField(t *testing.T) {
	obj := &MyStruct{}
	expected := &MyStruct{
		B:        true,
		S:        "something",
		I:        1,
		I8:       8,
		I16:      16,
		I32:      32,
		I64:      64,
		UI:       1,
		UI8:      8,
		UI16:     16,
		UI32:     32,
		UI64:     64,
		F32:      2,
		F64:      2.1,
		Slice:    []string{"a", "b"},
		Map:      map[string]string{"a": "b", "c": "d"},
		Duration: 1 * time.Hour,
	}
	b := true
	expected.Ptr = &b

	table := []struct {
		fieldname string
		value     string
	}{
		{"B", "true"},
		{"S", "something"},
		{"I", "1"},
		{"I8", "8"},
		{"I16", "16"},
		{"I32", "32"},
		{"I64", "64"},
		{"UI", "1"},
		{"UI8", "8"},
		{"UI16", "16"},
		{"UI32", "32"},
		{"UI64", "64"},
		{"F32", "2"},
		{"F64", "2.1"},
		{"Slice", `["a","b"]`},
		{"Map", `{"a":"b","c":"d"}`},
		{"Ptr", `true`},
	}

	for _, test := range table {
		err := utils.SetField(obj, test.fieldname, test.value)
		assert.NoError(t, err)
	}

	// Duration is special
	err := utils.SetField(obj, "Duration", "1h")
	assert.NoError(t, err)

	for _, test := range table {
		str, err := utils.GetFieldValue(obj, test.fieldname)
		assert.NoError(t, err)
		assert.Equal(t, test.value, str)
	}
	require.Equal(t, expected, obj)
}
