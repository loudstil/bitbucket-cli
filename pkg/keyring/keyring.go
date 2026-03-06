// Package keyring wraps 99designs/keyring with a simple interface for bb-cli.
// All secrets are stored under the service name "bb-cli".
package keyring

import (
	"fmt"

	"github.com/99designs/keyring"
)

const serviceName = "bb-cli"

// Set stores a token for the given context name in the system keyring.
func Set(contextName, token string) error {
	ring, err := open()
	if err != nil {
		return err
	}
	err = ring.Set(keyring.Item{
		Key:  contextName,
		Data: []byte(token),
		// Label is displayed in macOS Keychain Access and similar UIs.
		Label: fmt.Sprintf("bb-cli token for %q", contextName),
	})
	if err != nil {
		return fmt.Errorf("keyring: set %q: %w", contextName, err)
	}
	return nil
}

// Get retrieves the token for the given context name from the system keyring.
func Get(contextName string) (string, error) {
	ring, err := open()
	if err != nil {
		return "", err
	}
	item, err := ring.Get(contextName)
	if err != nil {
		return "", fmt.Errorf("keyring: get %q: %w", contextName, err)
	}
	return string(item.Data), nil
}

// Delete removes the token for the given context name from the system keyring.
func Delete(contextName string) error {
	ring, err := open()
	if err != nil {
		return err
	}
	if err := ring.Remove(contextName); err != nil {
		return fmt.Errorf("keyring: delete %q: %w", contextName, err)
	}
	return nil
}

// open returns a keyring handle using platform-appropriate backends.
// On Windows this uses the Windows Credential Manager; on macOS the
// system Keychain; on Linux it tries Secret Service (GNOME Keyring /
// KWallet) and falls back to a file-based keyring.
func open() (keyring.Keyring, error) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: serviceName,
		// Allow all backends so the library can pick the best one
		// available at runtime.
		AllowedBackends: keyring.AvailableBackends(),
	})
	if err != nil {
		return nil, fmt.Errorf("keyring: open: %w", err)
	}
	return ring, nil
}
