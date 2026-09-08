package handlers

import (
	"encoding/json"
	"io"
	"net/http"
)

// decodeJSONObject decodes r.Body into a map, returning an error for the
// three MALFORMED_BODY cases the API contract defines uniformly across every
// endpoint: an empty body, invalid JSON, or JSON that isn't an object (e.g.
// an array or a bare scalar). A valid empty object ("{}") decodes to a
// non-nil, zero-length map and is not an error — required-field validation
// downstream is what turns that into 422, not this.
func decodeJSONObject(r *http.Request) (map[string]interface{}, error) {
	if r.Body == nil {
		return nil, io.EOF
	}
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	return body, nil
}
