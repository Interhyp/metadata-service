package cache

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDeepCopyStringSlice(t *testing.T) {
	input := []string{"first", "WORLD", "problems"}
	copyResult := deepCopyStringSlice(input)
	require.Equal(t, input, copyResult, "Deep copied slices do not contain same elements.")
	// Compare addresses of first slice elements to determine reference equality.
	// We may not use require.NotEqual as it compares pointer values and not addresses.
	// See https://stackoverflow.com/a/53010178
	require.False(t, &input[0] == &copyResult[0], "Deep copied slices are identical.")
}

func TestDeepCopyStruct(t *testing.T) {
	type deepCopyTestStruct struct {
		SomeString          string
		SomeInt             int
		SomeBool            bool
		SomeSlice           []string
		SomeMap             map[string]string
		SomePointer         *string
		SomeNilPointer      *string
		somePrivateField    string
		SomeInBetweenStruct struct {
			SomeInBetweenString string
			SomeInnerStruct     struct {
				SomeInnerString string
			}
		}
	}

	mapField := make(map[string]string)
	mapField["key"] = "value"
	pointerValue := "pointer"
	input := deepCopyTestStruct{
		SomeString:       "SomeString",
		SomeInt:          51324,
		SomeBool:         true,
		SomeSlice:        []string{"first", "second"},
		SomeMap:          mapField,
		SomePointer:      &pointerValue,
		SomeNilPointer:   nil,
		somePrivateField: "private",
		SomeInBetweenStruct: struct {
			SomeInBetweenString string
			SomeInnerStruct     struct {
				SomeInnerString string
			}
		}{
			SomeInBetweenString: "inBetweenString",
			SomeInnerStruct: struct {
				SomeInnerString string
			}{
				SomeInnerString: "innerString",
			},
		},
	}

	result := deepCopyTestStruct{}
	err := deepCopyStruct(input, &result)

	require.Nil(t, err)
	require.Empty(t, result.somePrivateField)
	require.EqualExportedValues(t, input, result)
	require.False(t, &input == &result, "Structs are equal by reference.")
}

func TestDeepCopyStruct_MarshalErrorIsReturned(t *testing.T) {
	type errorStruct struct{ SomeFunction func(int) int }
	input := errorStruct{
		SomeFunction: func(i int) int {
			return i
		},
	}
	result := errorStruct{}
	err := deepCopyStruct(input, &result)
	require.ErrorContains(t, err, "unsupported type")
}

func TestDeepCopyStruct_UnmarshalErrorIsReturned(t *testing.T) {
	type someStruct struct{ SomeString string }
	input := someStruct{
		SomeString: "someString",
	}
	err := deepCopyStruct(input, nil)
	require.ErrorContains(t, err, "json: Unmarshal(nil")
}
