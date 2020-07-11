package data

import (
	"reflect"
	"strconv"
)

type StringExpression struct {
	src    string
	result string
}

// NewStringExpression returns a new StringExpression from a provided string.
func NewStringExpression(src string) StringExpression {
	return StringExpression{
		src: src,
	}
}

// StrinngExpressionTransformer is the mergo transformer struct.
type StringExpressionTransformer struct{}

// Transformer checks if a StringExpression is empty, and if so, to replace it with the contents of another.
func (t StringExpressionTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	if typ == reflect.TypeOf(StringExpression{}) {
		return func(dst, src reflect.Value) error {
			if dst.CanSet() {
				isZero := dst.MethodByName("IsZero")
				result := isZero.Call([]reflect.Value{})
				if result[0].Bool() {
					dst.Set(src)
				}
			}
			return nil
		}
	}
	return nil
}

func (exp StringExpression) IsZero() bool {
	return exp.src == ""
}

func (exp *StringExpression) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string

	if err := unmarshal(&str); err != nil {
		return err
	}
	exp.src = str
	return nil
}

func (exp StringExpression) MarshalYAML() (interface{}, error) {
	return exp.src, nil
}

func (exp *StringExpression) Compile() error {
	exp.result = exp.src
	return nil
}

func (exp *StringExpression) Get() (string, error) {
	if exp.result == "" {
		if err := exp.Compile(); err != nil {
			return "", err
		}
	}
	return exp.result, nil
}

func (exp *StringExpression) GetInt() (int, error) {
	exp.Get()
	return strconv.Atoi(string(exp.result))
}
func (exp *StringExpression) GetString() (string, error) {
	exp.Get()
	return exp.result, nil
}

func BuildStringExpression(in string) (s StringExpression) {
	return s
}
