package main

import "errors"

var (
	ErrNotFound                = errors.New("item not found")
	ErrContentTypeNotSupported = errors.New("content-type is not supported")
	ErrMarshalJson             = errors.New("failed to marshal response struct")
)
