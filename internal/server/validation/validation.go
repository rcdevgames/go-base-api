package validation

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// FieldType enumerates supported primitive conversions.
type FieldType string

const (
	FieldString FieldType = "string"
	FieldInt    FieldType = "int"
	FieldFloat  FieldType = "float"
	FieldBool   FieldType = "bool"
)

// Rule describes validation requirements for a single input.
type Rule struct {
	Name      string
	Required  bool
	Type      FieldType
	MinLength int
	MaxLength int
	Pattern   *regexp.Regexp
	Enum      []string
	Custom    func(string) (interface{}, error)
}

// Result represents validated values and per-field errors.
type Result struct {
	Values map[string]interface{}
	Errors ErrorMap
}

// ErrorMap represents a map of field names to validation error messages.
type ErrorMap map[string]string

// IsValid reports whether any validation error occurred.
func (r Result) IsValid() bool {
	return len(r.Errors) == 0
}

// ValidateStrings validates a simple map input (e.g. parsed JSON/body).
func ValidateStrings(input map[string]string, rules []Rule) Result {
	values := make(map[string]interface{}, len(rules))
	errs := make(ErrorMap)
	allowed := make(map[string]map[string]struct{})
	for _, rule := range rules {
		if rule.Enum != nil {
			set := make(map[string]struct{}, len(rule.Enum))
			for _, v := range rule.Enum {
				set[v] = struct{}{}
			}
			allowed[rule.Name] = set
		}
	}

	for _, rule := range rules {
		raw, ok := input[rule.Name]
		raw = strings.TrimSpace(raw)
		if !ok || raw == "" {
			if rule.Required {
				errs[rule.Name] = "required"
			}
			continue
		}

		if rule.Pattern != nil && !rule.Pattern.MatchString(raw) {
			errs[rule.Name] = "invalid format"
			continue
		}

		if rule.MinLength > 0 && len(raw) < rule.MinLength {
			errs[rule.Name] = fmt.Sprintf("minimum length %d", rule.MinLength)
			continue
		}

		if rule.MaxLength > 0 && len(raw) > rule.MaxLength {
			errs[rule.Name] = fmt.Sprintf("maximum length %d", rule.MaxLength)
			continue
		}

		if allowedSet, ok := allowed[rule.Name]; ok {
			if _, match := allowedSet[raw]; !match {
				errs[rule.Name] = "value not allowed"
				continue
			}
		}

		var parsed interface{}
		var err error
		if rule.Custom != nil {
			parsed, err = rule.Custom(raw)
		} else {
			switch rule.Type {
			case FieldInt:
				parsed, err = strconv.Atoi(raw)
			case FieldFloat:
				parsed, err = strconv.ParseFloat(raw, 64)
			case FieldBool:
				parsed, err = strconv.ParseBool(raw)
			case FieldString, "":
				parsed = raw
			default:
				err = fmt.Errorf("unsupported field type %s", rule.Type)
			}
		}
		if err != nil {
			errs[rule.Name] = err.Error()
			continue
		}
		values[rule.Name] = parsed
	}

	return Result{Values: values, Errors: errs}
}

// ValidateValues accepts url.Values (e.g. form submissions).
func ValidateValues(values url.Values, rules []Rule) Result {
	input := make(map[string]string, len(values))
	for key, vals := range values {
		if len(vals) > 0 {
			input[key] = vals[0]
		}
	}
	return ValidateStrings(input, rules)
}
