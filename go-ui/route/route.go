package route

import (
	"fmt"
	"strings"
)

type Route struct {
	path               string
	pathParts          []string
	indexToVariableMap map[int]string
}

func Parse(path string) (*Route, error) {
	r := &Route{
		path:               path,
		pathParts:          strings.Split(path, "/"),
		indexToVariableMap: map[int]string{},
	}
	for i, part := range r.pathParts {
		if strings.HasPrefix(part, ":") {
			variable := part[1:]
			if variable == "" {
				return nil, fmt.Errorf("invalid variable path: %s", part)
			}
			r.indexToVariableMap[i] = variable
		}
	}
	return r, nil
}

func (r Route) Match(path string) (matched bool, variableMap map[string]string) {
	variableMap = map[string]string{}

	pathParts := strings.Split(path, "/")
	if len(pathParts) != len(r.pathParts) {
		return false, nil
	}

	for i, part := range pathParts {
		if variableName := r.indexToVariableMap[i]; variableName != "" {
			variableMap[variableName] = part
		} else {
			if part != r.pathParts[i] {
				return false, nil
			}
		}
	}

	return true, variableMap
}

func (r Route) Regexp() string {
	output := "^"
	for i, part := range r.pathParts {
		if i > 0 {
			output += "/"
		}
		if variableName := r.indexToVariableMap[i]; variableName != "" {
			output += "([^/]+)"
		} else {
			output += part
		}
	}
	output += "$"
	return output
}
