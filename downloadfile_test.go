package e621api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_E621PathExtractor_GetPath(t *testing.T) {
	p := &E621PathExtractor{}

	var (
		path string
		err  error
	)

	path, err = p.GetPath("https://static1.e621.net/data/cc/63/cc636a8276a532dc6909acdf7f19ea05.webm")
	assert.Equal(t, "cc636a8276a532dc6909acdf7f19ea05.webm", path, "")
	assert.Equal(t, nil, err, "")

	path, err = p.GetPath("https://static1.e621.net/data/preview/de/08/de08688f663a5cc8e44dfde508d84093.jpg")
	assert.Equal(t, "preview_de08688f663a5cc8e44dfde508d84093.jpg", path, "")
	assert.Equal(t, nil, err, "")

	path, err = p.GetPath("https://static1.e621.net/data/sample/de/08/de08688f663a5cc8e44dfde508d84093.jpg")
	assert.Equal(t, "sample_de08688f663a5cc8e44dfde508d84093.jpg", path, "")
	assert.Equal(t, nil, err, "")

	path, err = p.GetPath("https://static2.e621.net/data/sample/de/08/de08688f663a5cc8e44dfde508d84093.jpg")
	assert.Equal(t, "", path, "")
	assert.Equal(t, "Wrong host static2.e621.net", err.Error(), "")
}
