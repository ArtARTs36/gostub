package api

import (
	"context"
	"net/http"
)

type StubUserService struct {
}

func NewStubUserService() *StubUserService {
	return &StubUserService{}
}

func (s *StubUserService) List(ctx context.Context, r *http.Request) (*Response, error) {
	var anyArg any
	return anyArg.(*Response), errors.New("is not real method List")
}

func (s *StubUserService) Create(ctx context.Context, r *http.Request) (*Response, error) {
	var anyArg any
	return anyArg.(*Response), errors.New("is not real method Create")
}
