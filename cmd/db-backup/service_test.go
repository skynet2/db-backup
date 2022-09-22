package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestServiceFileName(t *testing.T) {
	s := &Service{}

	result, _, _ := s.getFinalFilename("testdb")

	assert.Equal(t, "", result)
}
