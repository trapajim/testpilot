package testpilot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Request struct {
	method  string
	url     string
	body    func() io.Reader
	headers map[string]string
	plan    *TestPlan
	expect  *Expect
	key     string
}

// Body sets the request body
func (r *Request) Body(body func() io.Reader) *Request {
	r.body = body
	return r
}

// Headers sets the request headers
func (r *Request) Headers(headers map[string]string) *Request {
	r.headers = headers
	return r
}

// Store stores the response body in the response store
func (r *Request) Store(key string) *Request {
	r.key = key
	return r
}

// Expect sets the expectations for the response
func (r *Request) Expect() *Expect {
	r.expect = &Expect{}
	return r.expect
}

func (r *Request) send(ctx context.Context) error {
	var reqBody io.Reader
	if r.body != nil {
		reqBody = r.body()
	}
	req, err := http.NewRequestWithContext(ctx, r.method, r.url, reqBody)
	if err != nil {
		return err
	}
	for key, value := range r.headers {
		req.Header.Set(key, value)
	}
	resp, err := r.plan.client.Do(req)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	defer resp.Body.Close()
	r.plan.lastResponseBody = body
	if r.key != "" {
		r.plan.responseStore[r.key] = body
	}
	if r.expect != nil {
		err := r.expectations(resp.StatusCode, body)
		if err != nil {
			return err
		}
	}
	return nil
}

// JSON helper function to create a JSON request body
// panics if marshalling fails
func JSON(v any) func() io.Reader {
	j, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return func() io.Reader { return bytes.NewReader(j) }
}

// URLValues helper function to create a URL encoded request body
func URLValues(val url.Values) func() io.Reader {
	return func() io.Reader { return strings.NewReader(val.Encode()) }
}

func (r *Request) expectations(statusCode int, body []byte) error {
	if r.expect.expectedResponseCode != nil && statusCode != *r.expect.expectedResponseCode {
		return fmt.Errorf("expected response code %d got %d", *r.expect.expectedResponseCode, statusCode)
	}
	if r.expect.expectedResponseBody != nil {
		fn := *r.expect.expectedResponseBody
		err := fn(body)
		if err != nil {
			return err
		}
	}
	return nil
}
