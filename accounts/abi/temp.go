package abi

import "strings"

// todo: code reuse and relocate the temp.go file
func GetFuncParams(paramString string) []string {
	if paramString == "" {
		return nil
	}

	hasBracket := strings.Contains(paramString, "{") && strings.Contains(paramString, "}")
	if !hasBracket {
		return nil
	} else {
		paramString = paramString[strings.Index(paramString, "{")+1 : strings.Index(paramString, "}")]
	}

	splitPos := recordFuncParamSplitPos(paramString)
	return splitFuncParamByPos(paramString, splitPos)

}

// splitFuncParamByPos splits the function params which is in string format
// by the position index recorded by the recordFuncParamSplitPos method
func splitFuncParamByPos(paramString string, splitPos []int) []string {
	var params = make([]string, 0)
	var lastPos = 0

	for _, i := range splitPos {
		params = append(params, paramString[lastPos:i])
		lastPos = i + 1
	}
	params = append(params, paramString[lastPos:])

	//params := strings.Split(paramString, ",")
	for index, param := range params {
		if strings.HasPrefix(param, "\"") {
			params[index] = param[strings.Index(param, "\"")+1 : strings.LastIndex(param, "\"")]
		}
		if strings.HasPrefix(param, "'") {
			params[index] = param[strings.Index(param, "'")+1 : strings.LastIndex(param, "'")]
		}
	}

	return params
}

// recordFuncParamSplitPos record the index of the end of each parameter
func recordFuncParamSplitPos(paramString string) []int {
	var symStack []rune
	var splitPos []int

	for i, s := range paramString {
		switch s {
		case ',':
			if len(symStack) == 0 {
				splitPos = append(splitPos, i)
			}
		case '{':
			symStack = append(symStack, '{')
		case '}':
			if len(symStack) < 1 {
				panic("parameter's format is not write!!!")
			}
			if symStack[len(symStack)-1] == '{' {
				symStack = symStack[:len(symStack)-1]
			}
		case '[':
			symStack = append(symStack, '[')
		case ']':
			if len(symStack) < 1 {
				panic("parameter's format is not write!!!")
			}
			if symStack[len(symStack)-1] == '[' {
				symStack = symStack[:len(symStack)-1]
			}
		case '(':
			symStack = append(symStack, '(')
		case ')':
			if len(symStack) < 1 {
				panic("parameter's format is not write!!!")
			}
			if symStack[len(symStack)-1] == '(' {
				symStack = symStack[:len(symStack)-1]
			}
		case '"':
			if len(symStack) < 1 {
				symStack = append(symStack, '"')
			} else {
				if symStack[len(symStack)-1] == '"' {
					symStack = symStack[:len(symStack)-1]
				} else {
					symStack = append(symStack, '"')
				}
			}
		case '\'':
			if len(symStack) < 1 {
				symStack = append(symStack, '\'')
			} else {
				if symStack[len(symStack)-1] == '\'' {
					symStack = symStack[:len(symStack)-1]
				} else {
					symStack = append(symStack, '\'')
				}
			}
		}
	}

	return splitPos
}
