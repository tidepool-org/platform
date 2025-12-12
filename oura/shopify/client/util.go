package client

func ptr[A any](a A) *A { return &a }
