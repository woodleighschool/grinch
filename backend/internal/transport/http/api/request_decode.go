package apihttp

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const maxJSONBodyBytes = 1 << 20

type badRequestError string

func decodeJSONBody(r *http.Request, dst any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, maxJSONBodyBytes))

	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return badRequestError("request body is required")
		}

		return badRequestError("request body is invalid")
	}

	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return badRequestError("request body must contain a single JSON object")
	}

	return nil
}

func (e badRequestError) Error() string {
	return string(e)
}
