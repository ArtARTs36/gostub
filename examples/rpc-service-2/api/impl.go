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
	panic("method StubUserService.List not implemented")
}

func (s *StubUserService) Create(ctx context.Context, r *http.Request) (*Response, error) {
	panic("method StubUserService.Create not implemented")
}
