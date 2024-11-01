package ds

import "github.com/iancoleman/strcase"

type String struct {
	Value string
}

func NewString(val string) String {
	return String{Value: val}
}

func (s *String) Snake() String {
	return NewString(strcase.ToSnake(s.Value))
}

func (s *String) String() string {
	return s.Value
}

func (s *String) Pascal() String {
	return NewString(strcase.ToCamel(s.Value))
}
