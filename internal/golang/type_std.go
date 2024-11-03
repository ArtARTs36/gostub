package golang

import "slices"

const (
	TypeError  = "error"
	TypeAny    = "any"
	TypeString = "string"
	TypeBool   = "bool"
)

var (
	stdTypes = []string{
		"int", "int16", "int32",
		"uint8", "uint16", "uint32", "uint64",
		"float32", "float64",

		"string",

		"error",

		"any",

		"bool",

		"interface",
	}

	numericTypes = []string{
		"int", "int16", "int32",
		"uint8", "uint16", "uint32", "uint64",
		"float32", "float64",
	}
)

func IsNumericType(name string) bool {
	return slices.Contains(numericTypes, name)
}

func IsStdType(name string) bool {
	return slices.Contains(stdTypes, name)
}
