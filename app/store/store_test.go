package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataStore(t *testing.T) {
	d := NewDataStore()
	assert.NotNil(t, d)

	v := d.Get("myKey")
	assert.Nil(t, v)

	d.Set("myKey", "myVar")
	v = d.Get("myKey")
	assert.NotNil(t, v)
	assert.Equal(t, "myVar", *v)
}
