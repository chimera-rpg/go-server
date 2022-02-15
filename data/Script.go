package data

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos72/gomacro/fast"
)

// ScriptError is the error structure emitted via panic when a script fails to compile.
type ScriptError struct {
	lineIndex int
	charIndex int
	lines     []string
	s         string
}

// Error returns the main error string.
func (e *ScriptError) Error() string {
	return e.s
}

type ScriptEventResponse struct {
	Expr *fast.Expr
}

func (s *ScriptEventResponse) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var code string
	defer func() {
		if err := recover(); err != nil {
			parts := strings.Split(fmt.Sprintf("%v", err), " ")
			firstPart := parts[0]
			fileParts := strings.Split(firstPart, ":")
			lineIndex, _ := strconv.Atoi(fileParts[1])
			lineIndex--
			charIndex, _ := strconv.Atoi(fileParts[2])
			lines := strings.Split(code, "\n")

			scriptErr := ScriptError{
				lineIndex: lineIndex,
				charIndex: charIndex,
			}

			scriptErr.s = strings.Join(parts[1:], " ")
			for i := lineIndex - 1; i <= lineIndex+1; i++ {
				if i >= 0 && i < len(lines) {
					if i == lineIndex {
						scriptErr.lines = append(scriptErr.lines, fmt.Sprintf("--> %s <--", lines[i]))
					} else {
						scriptErr.lines = append(scriptErr.lines, fmt.Sprintf("    %s", lines[i]))
					}
				}
			}
			panic(scriptErr)
		}
	}()
	if err := unmarshal(&code); err != nil {
		return nil
	}
	s.Expr = Interpreter.Compile(code)
	return nil
}
