package config

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type ConfigFilter struct {
	Target    string `yaml:"target"`
	Condition string `yaml:"condition"`
}

type ConfigSessionOverride struct {
	Target    string `yaml:"target"`
	Condition string `yaml:"condition"`
	Value     string `yaml:"value"`
}

type ConfigNameOverwrite struct {
	Regex string `yaml:"regex"`
	Value string `yaml:"value"`
}

type ConfigSessionOptions struct {
	ConnectionProtocol string `yaml:"connection_protocol"`
	Credential         string `yaml:"credential"`
	Firewall           string `yaml:"firewall"`
}

type ConfigSession struct {
	Path           string                  `yaml:"path"`
	DeviceName     string                  `yaml:"device_name"`
	SessionOptions ConfigSessionOptions    `yaml:"session_options"`
	Overrides      []ConfigSessionOverride `yaml:"overrides"`
}

type Config struct {
	configPath              string
	LogLevel                string         `yaml:"log_level"`
	NetboxUrl               string         `yaml:"netbox_url"`
	NetboxToken             string         `yaml:"netbox_token"`
	RootPath                string         `yaml:"root_path"`
	Filters                 []ConfigFilter `yaml:"filters"`
	Session                 ConfigSession  `yaml:"session"`
	EnableConsoleServerSync bool           `yaml:"console_server_sync_enable"`
	EnablePeriodicSync      bool           `yaml:"periodic_sync_enable"`
	PeriodicSyncInterval    *int           `yaml:"periodic_sync_interval"`
}

func NewConfig(configPath string) (*Config, error) {
	config := &Config{
		configPath: configPath,
	}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	err = config.SetDefaultsAndValidate()
	if err != nil {
		return nil, err
	}

	config.Save()
	return config, nil
}

func (c *Config) SetDefaultsAndValidate() error {
	// setup defaults
	if c.LogLevel == "" {
		c.LogLevel = "ERROR"
	}

	if c.PeriodicSyncInterval == nil {
		defaultTime := 60
		c.PeriodicSyncInterval = &defaultTime
	}

	if c.Session.SessionOptions.ConnectionProtocol == "" {
		c.Session.SessionOptions.ConnectionProtocol = "SSH"
	}

	if c.Session.SessionOptions.Firewall == "" {
		c.Session.SessionOptions.Firewall = "None"
	}

	if c.Session.DeviceName == "" {
		c.Session.DeviceName = "{device_name}"
	}

	if c.Session.Path == "" {
		c.Session.Path = "{tenant_name}/{region_name}/{site_name}/{device_role}"
	}

	// validate overrides
	for _, override := range c.Session.Overrides {
		if override.Target == "" {
			return errors.New("override target can not be empty")
		}

		if override.Condition == "" {
			return errors.New("override condition can not be empty")
		}
	}

	// validate the netbox url, and allows us to strip http/https etc
	url, err := parseRawURL(c.NetboxUrl)
	if err != nil {
		return err
	}

	c.NetboxUrl = url.Host
	return nil
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

	// Validate the path first, and if empty create the config file
	s, err := os.Stat(configPath)
	if err != nil {
		_, err = os.Create(configPath)
		if err != nil {
			return "", err
		}
	}
	if s.IsDir() {
		return "", fmt.Errorf("'%s' is a directory, not a normal file", configPath)
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
