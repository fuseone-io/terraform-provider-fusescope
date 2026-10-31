package apiclients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type envelope[T any] struct {
	Data T `json:"data"`
}

// IsNotFound reports whether err is an API error with HTTP status 404.
func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

// doJSON sends in as the JSON body (when non-nil) and decodes the response's
// data envelope into out (when non-nil).
func (c *WrappedClient) doJSON(ctx context.Context, method, path string, in, out any) error {
	var body io.Reader
	if in != nil {
		raw, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(raw)
	}

	raw, err := c.Do(ctx, method, path, body)
	if err != nil {
		return err
	}

	if out == nil {
		return nil
	}

	if err := json.Unmarshal(raw, &envelope[any]{Data: out}); err != nil {
		return fmt.Errorf("decode %s %s response: %w", method, path, err)
	}

	return nil
}
