package data

import "strconv"

type StringExpression struct {
	src    string
	result string
}

func (exp *StringExpression) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string

	if err := unmarshal(&str); err != nil {
		return err
	}
	exp.src = str
	return nil
}

func (exp *StringExpression) MarshalYAML() (interface{}, error) {
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
