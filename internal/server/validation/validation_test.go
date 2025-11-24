package validation

import (
	"net/url"
	"regexp"
	"testing"
)

func TestValidateStrings(t *testing.T) {
	rules := []Rule{
		{Name: "email", Required: true, Pattern: regexp.MustCompile(`^[^@]+@[^@]+$`)},
		{Name: "age", Type: FieldInt, MinLength: 1, Required: true},
		{Name: "role", Enum: []string{"user", "admin"}},
	}
	input := map[string]string{"email": "test@example.com", "age": "42", "role": "user"}
	res := ValidateStrings(input, rules)
	if !res.IsValid() {
		t.Fatalf("expected valid result, got errors: %v", res.Errors)
	}
	if res.Values["age"].(int) != 42 {
		t.Fatalf("expected age 42, got %v", res.Values["age"])
	}
}

func TestValidateStringsErrors(t *testing.T) {
	rules := []Rule{{Name: "email", Required: true, Pattern: regexp.MustCompile(`^[^@]+@[^@]+$`)}}
	input := map[string]string{"email": "invalid"}
	res := ValidateStrings(input, rules)
	if res.IsValid() {
		t.Fatalf("expected validation error")
	}
}

func TestValidateValues(t *testing.T) {
	rules := []Rule{{Name: "active", Type: FieldBool}}
	values := url.Values{"active": {"true"}}
	res := ValidateValues(values, rules)
	if !res.IsValid() {
		t.Fatalf("expected valid result")
	}
	if res.Values["active"].(bool) != true {
		t.Fatalf("expected true")
	}
}
