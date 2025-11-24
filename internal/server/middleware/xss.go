package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

// XSS sanitizes query params, form data, and JSON bodies before they reach handlers.
func XSS() func(http.Handler) http.Handler {
	policy := bluemonday.UGCPolicy()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sanitizeQuery(r, policy)
			sanitizeForm(r, policy)
			r = sanitizeJSONBody(r, policy)
			next.ServeHTTP(w, r)
		})
	}
}

func sanitizeQuery(r *http.Request, policy *bluemonday.Policy) {
	q := r.URL.Query()
	changed := false
	for key, values := range q {
		for i, v := range values {
			sanitized := policy.Sanitize(v)
			if sanitized != v {
				q[key][i] = sanitized
				changed = true
			}
		}
	}
	if changed {
		r.URL.RawQuery = q.Encode()
	}
}

func sanitizeForm(r *http.Request, policy *bluemonday.Policy) {
	if !shouldParseForm(r.Header.Get("Content-Type")) {
		return
	}
	if err := r.ParseForm(); err != nil {
		return
	}
	for key, values := range r.Form {
		for i, v := range values {
			r.Form[key][i] = policy.Sanitize(v)
		}
	}
	r.PostForm = r.Form
}

func shouldParseForm(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return strings.Contains(contentType, "application/x-www-form-urlencoded") || strings.Contains(contentType, "multipart/form-data")
}

func sanitizeJSONBody(r *http.Request, policy *bluemonday.Policy) *http.Request {
	if r.Body == nil {
		return r
	}
	contentType := strings.ToLower(r.Header.Get("Content-Type"))
	if !strings.Contains(contentType, "application/json") {
		return r
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		r.Body = io.NopCloser(bytes.NewReader(nil))
		return r
	}
	if len(data) == 0 {
		r.Body = io.NopCloser(bytes.NewReader(data))
		return r
	}
	var payload interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		r.Body = io.NopCloser(bytes.NewReader(data))
		return r
	}
	payload = sanitizeInterface(policy, payload)
	sanitized, err := json.Marshal(payload)
	if err != nil {
		r.Body = io.NopCloser(bytes.NewReader(data))
		return r
	}
	r.Body = io.NopCloser(bytes.NewReader(sanitized))
	r.ContentLength = int64(len(sanitized))
	r.Header.Set("Content-Length", strconv.FormatInt(int64(len(sanitized)), 10))
	return r
}

func sanitizeInterface(policy *bluemonday.Policy, value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		for key, nested := range v {
			v[key] = sanitizeInterface(policy, nested)
		}
		return v
	case []interface{}:
		for i, nested := range v {
			v[i] = sanitizeInterface(policy, nested)
		}
		return v
	case string:
		return policy.Sanitize(v)
	default:
		return value
	}
}
