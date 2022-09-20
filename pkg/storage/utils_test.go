package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSort(t *testing.T) {
	files := []File{
		{
			AbsolutePath: "new",
			CreatedAt:    time.Now().UTC(),
		},
		{
			AbsolutePath: "old",
			CreatedAt:    time.Now().UTC().Add(-1 * time.Hour),
		},
	}

	assert.Equal(t, "old", sortFiles(files)[0].AbsolutePath)
}
