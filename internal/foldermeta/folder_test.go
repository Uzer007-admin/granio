package foldermeta

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Uzer007-admin/granio/v2/storage"
	"github.com/Uzer007-admin/granio/v2/test"
)

func TestDetailsForFolder(t *testing.T) {
	a := test.NewApp()
	t.Cleanup(a.Quit)
	testdata := storage.NewFileURI("testdata")
	ff, err := DetailsForFolder(testdata)

	require.NoError(t, err)
	require.NotNil(t, ff)
	path := ff.BackgroundURI.Path()
	assert.Equal(t, ".background.svg", path[len(path)-15:])
}

func TestDetailsForFolder_nil(t *testing.T) {
	ff, err := DetailsForFolder(nil)

	assert.Nil(t, ff)
	assert.Nil(t, err)
}
