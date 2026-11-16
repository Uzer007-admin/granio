package test

import (
	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/theme"
)

type mockCloud struct {
	configured bool
}

func (c *mockCloud) Cleanup(_ granio.App) {
	c.configured = false
}

func (*mockCloud) ProviderDescription() string {
	return "Mock cloud implementation"
}

func (*mockCloud) ProviderIcon() granio.Resource {
	return theme.ComputerIcon()
}

func (*mockCloud) ProviderName() string {
	return "mock"
}

func (c *mockCloud) Setup(_ granio.App) error {
	c.configured = true
	return nil
}
