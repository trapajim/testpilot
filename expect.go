package testpilot

type Expect struct {
	expectedResponseCode *int
	expectedResponseBody *AssertionFunc
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
