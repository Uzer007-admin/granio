package test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Uzer007-admin/granio/v2"
)

func TestTestApp_CloudProvider(t *testing.T) {
	a := NewApp()
	c := &mockCloud{}
	a.SetCloudProvider(c)

	assert.Equal(t, c, a.CloudProvider())
}

func TestFyneApp_transitionCloud(t *testing.T) {
	a := NewApp()
	p := &mockCloud{}
	preferenceChanged := false
	settingsChan := make(chan granio.Settings, 1)
	a.Preferences().AddChangeListener(func() {
		preferenceChanged = true
	})
	a.Settings().AddListener(func(s granio.Settings) {
		go func() { settingsChan <- s }()
	})
	a.SetCloudProvider(p)

	<-settingsChan // settings were updated
	assert.True(t, preferenceChanged)
}
