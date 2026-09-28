package project

import "errors"

var (
	ErrInvalidName       = errors.New("invalid project name")
	ErrInvalidTransition = errors.New("invalid project status transition")
	ErrProjectNotFound   = errors.New("project not found")
)

