package data

import (
	"reflect"
	"strconv"
)

// StringExpression represents a semi-complex string-based expression used for calculations.
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

// StringExpressionTransformer is the mergo transformer struct.
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

// IsZero returns if the expression is equal to a zero value.
func (exp StringExpression) IsZero() bool {
	return exp.src == ""
}

// UnmarshalYAML unmarshals a string into a StringExpression's src.
func (exp *StringExpression) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string

	if err := unmarshal(&str); err != nil {
		return err
	}
	exp.src = str
	return nil
}

// MarshalYAML saves a StringExpression's src property.
func (exp StringExpression) MarshalYAML() (interface{}, error) {
	return exp.src, nil
}

// Compile compiles a StringExpression.
func (exp *StringExpression) Compile() error {
	exp.result = exp.src
	return nil
}

// Get returns the underlying compiled result string of an expression.
func (exp *StringExpression) Get() (string, error) {
	if exp.result == "" {
		if err := exp.Compile(); err != nil {
			return "", err
		}
	}
	return exp.result, nil
}

// GetInt returns the compiled integer result of an expression.
func (exp *StringExpression) GetInt() (int, error) {
	exp.Get()
	return strconv.Atoi(string(exp.result))
}

// GetString returns the compiled string result of an expression.
func (exp *StringExpression) GetString() (string, error) {
	exp.Get()
	return exp.result, nil
}

// BuildStringExpression builds a StringExpression from a given string.
func BuildStringExpression(in string) (s StringExpression) {
	return s
}

// Add adds another StringExpression's src to this one's.
func (exp *StringExpression) Add(other StringExpression) {
	exp.src = exp.src + other.src
}
