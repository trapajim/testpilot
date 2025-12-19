package testpilot

type Expect struct {
	expectedResponseCode *int
	expectedResponseBody *AssertionFunc
	headerAssertions     map[string]func(string) error
}

// Status sets the expected response code
func (e *Expect) Status(code int) *Expect {
	e.expectedResponseCode = &code
	return e
}

// Body sets the expected response body
func (e *Expect) Body(assertionFunc AssertionFunc) *Expect {
	e.expectedResponseBody = &assertionFunc
	return e
}

// Header sets the expected response header
func (e *Expect) Header(key string, assertionFunc func(string) error) *Expect {
	if e.headerAssertions == nil {
		e.headerAssertions = make(map[string]func(string) error)
	}
	e.headerAssertions[key] = assertionFunc
	return e
}
