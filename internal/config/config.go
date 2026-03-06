package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	AppName    = "bb"
	ConfigFile = "config"
	ConfigType = "yaml"

	ProviderCloud = "cloud"
	ProviderDC    = "datacenter"
)

// Context represents a saved Bitbucket connection profile.
type Context struct {
	Name      string `mapstructure:"name"`
	Type      string `mapstructure:"type"`      // "cloud" or "datacenter"
	BaseURL   string `mapstructure:"base_url"`  // empty for Cloud
	Username  string `mapstructure:"username"`  // used for display / DC basic auth
	Workspace string `mapstructure:"workspace"` // Cloud workspace slug; empty for DC
}

// Init sets up Viper to load config from ~/.config/bb/config.yaml.
// It creates the config file and directory if they do not exist.
func Init() error {
	dir, err := configDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("config: create directory: %w", err)
	}

	viper.SetConfigName(ConfigFile)
	viper.SetConfigType(ConfigType)
	viper.AddConfigPath(dir)
	viper.SetEnvPrefix("BB")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// First run – create an empty config file.
			path := filepath.Join(dir, ConfigFile+"."+ConfigType)
			if err := os.WriteFile(path, []byte(""), 0o600); err != nil {
				return fmt.Errorf("config: create file: %w", err)
			}
			return nil
		}
		return fmt.Errorf("config: read: %w", err)
	}
	return nil
}

// ActiveContext returns the name of the currently active context.
func ActiveContext() string {
	return viper.GetString("active_context")
}

// SetActiveContext persists the active context to the config file.
func SetActiveContext(name string) error {
	viper.Set("active_context", name)
	return save()
}

// GetContext retrieves a named context from config.
func GetContext(name string) (*Context, error) {
	var contexts []Context
	if err := viper.UnmarshalKey("contexts", &contexts); err != nil {
		return nil, fmt.Errorf("config: unmarshal contexts: %w", err)
	}
	for i := range contexts {
		if contexts[i].Name == name {
			return &contexts[i], nil
		}
	}
	return nil, fmt.Errorf("config: context %q not found", name)
}

// ListContexts returns all saved contexts.
func ListContexts() ([]Context, error) {
	var contexts []Context
	if err := viper.UnmarshalKey("contexts", &contexts); err != nil {
		return nil, fmt.Errorf("config: unmarshal contexts: %w", err)
	}
	return contexts, nil
}

// AddContext appends (or replaces) a context in the config and saves.
func AddContext(ctx Context) error {
	var contexts []Context
	_ = viper.UnmarshalKey("contexts", &contexts)

	// Replace existing entry with the same name.
	replaced := false
	for i := range contexts {
		if contexts[i].Name == ctx.Name {
			contexts[i] = ctx
			replaced = true
			break
		}
	}
	if !replaced {
		contexts = append(contexts, ctx)
	}

	viper.Set("contexts", contexts)
	return save()
}

// save writes the current Viper state to disk.
func save() error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, ConfigFile+"."+ConfigType)
	if err := viper.WriteConfigAs(path); err != nil {
		return fmt.Errorf("config: write: %w", err)
	}
	return nil
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("config: home directory: %w", err)
	}
	return filepath.Join(home, ".config", AppName), nil
}
