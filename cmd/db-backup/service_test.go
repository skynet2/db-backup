package main

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestServiceFileName(t *testing.T) {
	viper.Set("DB_DUMP_DIR", "./")
	s := &Service{}

	result, _, _ := s.getFinalFilename("testdb")

	assert.Equal(t, "", result)
}
