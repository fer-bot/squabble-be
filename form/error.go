package form

type SingleErrorResponse struct {
	Error string `json:"error"`
}

func SingleErrorResponseBuilder(err error) *SingleErrorResponse {
	e := SingleErrorResponse{}
	e.Error = err.Error()
	return &e
}
