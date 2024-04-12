package testpilot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var placeholderRegex = regexp.MustCompile(`{\.?([\w\.\d]+)}`)
var findPlaceholderRegex = regexp.MustCompile(`{[^}]+}`)

type TestPlan struct {
	t                *testing.T
	name             string
	client           *http.Client
	requests         []*Request
	responseStore    map[string][]byte
	lastResponseBody []byte
	runCalled        bool
}

// NewPlan creates a new TestPlan
// The TestPlan is a collection of requests that will be run in order
// failing to call Run will result in a test failure
func NewPlan(t *testing.T, name string) *TestPlan {
	plan := &TestPlan{
		t:             t,
		name:          name,
		client:        &http.Client{},
		requests:      make([]*Request, 0),
		responseStore: make(map[string][]byte),
	}
	t.Cleanup(func() {
		if !plan.runCalled {
			t.Errorf("TestPlan %s: Run was not called", plan.name)
		}
	})
	return plan
}

// Run runs the test plan
func (p *TestPlan) Run() {
	p.runCalled = true
	p.t.Run(p.name, func(t *testing.T) {
		for _, request := range p.requests {
			url, err := normalizeUrl(request.url, p.lastResponseBody, p.responseStore)
			if err != nil {
				t.Errorf(err.Error())
			}
			request.url = url
			t.Log(request.method, request.url)
			if err := request.send(context.TODO()); err != nil {
				t.Errorf(err.Error())
			}
		}
	})
}

// Request creates a new request in the test plan
func (p *TestPlan) Request(method, url string) *Request {
	request := &Request{
		method: method,
		url:    url,
		plan:   p,
	}
	p.requests = append(p.requests, request)
	return request
}

// Response returns the last response body
func (p *TestPlan) Response() []byte {
	return p.lastResponseBody
}

// ResponseForKey returns the response body for a given key
func (p *TestPlan) ResponseForKey(key string) []byte {
	return p.responseStore[key]
}

func navigateJSON(data any, path string) (interface{}, error) {
	keys := strings.Split(path, ".")
	var value = data
	for _, key := range keys {
		if idx, err := strconv.Atoi(key); err == nil {
			if arr, ok := value.([]interface{}); ok && idx < len(arr) {
				value = arr[idx]
			} else {
				return nil, errors.New("invalid index")
			}
		} else {
			if val, ok := value.(map[string]interface{})[key]; ok {
				value = val
			} else {
				return nil, errors.New("key not found")
			}
		}
	}
	return value, nil
}

func getResponseBody(placeholder string, lastResponseBody []byte, responseStore map[string][]byte) (any, error) {
	var data any
	if placeholder[1] == '.' {
		err := json.Unmarshal(lastResponseBody, &data)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	key := strings.Split(placeholder[1:len(placeholder)-1], ".")[0]
	body, ok := responseStore[key]
	if !ok {
		return nil, fmt.Errorf("key not found in responseStore: %s", key)
	}
	err := json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("something happend %w", err)
	}
	return data, nil
}

func normalizeUrl(url string, lastBody []byte, store map[string][]byte) (string, error) {
	matches := findPlaceholderRegex.FindAllString(url, -1)
	newUrl := url

	for _, match := range matches {
		data, err := getResponseBody(match, lastBody, store)
		if err != nil {
			return "", fmt.Errorf("error getting response body: %w", err)
		}

		path := extractPathFromPlaceholder(match)
		if path == "" {
			return "", fmt.Errorf("invalid placeholder: %s", match)
		}

		value, err := navigateJSON(data, path)
		if err != nil {
			log.Printf("Error navigating JSON: %v\n", err)
			return "", fmt.Errorf("error navigating JSON for path %s: %w", path, err)
		}

		newUrl = placeholderRegex.ReplaceAllStringFunc(newUrl, func(_ string) string {
			return fmt.Sprintf("%v", value)
		})
	}

	return newUrl, nil
}

// Helper function to extract path from placeholder
func extractPathFromPlaceholder(placeholder string) string {
	if placeholder[1] == '.' {
		return placeholder[2 : len(placeholder)-1]
	}
	path := strings.Trim(placeholder[1:len(placeholder)-1], ".")
	parts := strings.Split(path, ".")
	return strings.Join(parts[1:], ".")
}
