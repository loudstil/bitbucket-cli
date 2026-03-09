// Package factory constructs the correct BitbucketClient for the active context.
package factory

import (
	"fmt"

	"github.com/loudstil/bb/internal/config"
	"github.com/loudstil/bb/pkg/api"
	"github.com/loudstil/bb/pkg/api/cloud"
	"github.com/loudstil/bb/pkg/api/datacenter"
	bbkeyring "github.com/loudstil/bb/pkg/keyring"
)

// NewClient returns a BitbucketClient for the currently active context.
// It reads the active context name from config and retrieves the token from
// the system keyring.
func NewClient() (api.BitbucketClient, error) {
	name := config.ActiveContext()
	if name == "" {
		return nil, fmt.Errorf("no active context – run: bb auth login")
	}

	ctx, err := config.GetContext(name)
	if err != nil {
		return nil, fmt.Errorf("factory: %w", err)
	}

	token, err := bbkeyring.Get(name)
	if err != nil {
		return nil, fmt.Errorf("factory: retrieve token: %w", err)
	}

	switch ctx.Type {
	case config.ProviderCloud:
		return cloud.New(ctx.Username, token), nil
	case config.ProviderDC:
		return datacenter.New(ctx.BaseURL, ctx.Username, token), nil
	default:
		return nil, fmt.Errorf("factory: unknown provider type %q", ctx.Type)
	}
}
