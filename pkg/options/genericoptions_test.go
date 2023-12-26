package options

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlag(t *testing.T) {
	assert := assert.New(t)
	err := flag.Set("kubeconfig", "aa")
	assert.Nil(err)
	assert.Equal("aa", Kubeconfig)
	assert.Equal("", MasterURL)
}
