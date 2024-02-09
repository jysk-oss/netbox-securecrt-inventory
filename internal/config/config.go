package config

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type ConfigTemplateOverwrite struct {
	Key      string `yaml:"key"`
	Value    string `yaml:"value"`
	Template string `yaml:"template"`
}

type ConfigNameOverwrite struct {
	Regex string `yaml:"regex"`
	Value string `yaml:"value"`
}

type ConfigSessionPath struct {
	Template   string                    `yaml:"template"`
	Overwrites []ConfigTemplateOverwrite `yaml:"overwrites"`
}

type Config struct {
	configPath           string
	NetboxUrl            string                `yaml:"netbox_url"`
	NetboxToken          string                `yaml:"netbox_token"`
	RootPath             string                `yaml:"root_path"`
	NameOverwrites       []ConfigNameOverwrite `yaml:"name_overwrites"`
	SessionPath          ConfigSessionPath     `yaml:"session_path"`
	EnablePeriodicSync   bool                  `yaml:"periodic_sync_enable"`
	PeriodicSyncInterval *int                  `yaml:"periodic_sync_interval"`
	DefaultCredential    *string               `yaml:"default_credential"`
}

func NewConfig(configPath string) (*Config, error) {
	// Create config structure
	config := &Config{}

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	if config.PeriodicSyncInterval == nil {
		defaultTime := 60
		config.PeriodicSyncInterval = &defaultTime
	}

	// validate the netbox url, and allows us to strip http/https etc
	url, err := parseRawURL(config.NetboxUrl)
	if err != nil {
		return nil, err
	}

	config.NetboxUrl = url.Host
	config.configPath = configPath
	return config, nil
}

func (c *Config) Save() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	err = os.WriteFile(c.configPath, data, 0)
	if err != nil {
		return err
	}

	return nil
}

// ValidateConfigPath just makes sure, that the path provided is a file,
// that can be read
func ValidateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
}

func ParseFlags() (string, error) {
	// String that contains the configured configuration path
	var configPath string

	// Set up a CLI flag called "-config" to allow users
	// to supply the configuration file
	flag.StringVar(&configPath, "config", "~/.securecrt-inventory.yaml", "path to config file")

	// handle the users home dir
	usr, _ := user.Current()
	dir := usr.HomeDir
	if configPath == "~" {
		// In case of "~", which won't be caught by the "else if"
		configPath = dir
	} else if strings.HasPrefix(configPath, "~/") {
		// Use strings.HasPrefix so we don't match paths like
		// "/something/~/something/"
		configPath = filepath.Join(dir, configPath[2:])
	}

	// Actually parse the flags
	flag.Parse()

	// Validate the path first
	if err := ValidateConfigPath(configPath); err != nil {
		return "", err
	}

	// Return the configuration path
	return configPath, nil
}

func parseRawURL(rawurl string) (u *url.URL, err error) {
	u, err = url.ParseRequestURI(rawurl)
	if err != nil || u.Host == "" {
		u, repErr := url.ParseRequestURI("https://" + rawurl)
		if repErr != nil {
			return nil, err
		}
		return u, nil
	}

	return u, nil
}
