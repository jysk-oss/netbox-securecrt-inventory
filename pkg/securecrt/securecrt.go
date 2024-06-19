package securecrt

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type SecureCRT struct {
	credentialValue string
	configPath      string
	defaultConfig   string
}

func New(credentialName *string) (*SecureCRT, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	defaultConfig, err := loadDefaultSessionConfig(configPath)
	if err != nil {
		return nil, err
	}

	credentialValue, err := getCredentialHash(configPath, credentialName)
	if err != nil {
		return nil, err
	}

	return &SecureCRT{
		configPath:      configPath,
		defaultConfig:   defaultConfig,
		credentialValue: credentialValue,
	}, nil
}

func (scrt *SecureCRT) BuildSessionData(ip, protocol, site, address, machine_type string) string {
	var data strings.Builder
	data.WriteString(scrt.defaultConfig)
	data.WriteString(scrt.credentialValue)
	data.WriteString(fmt.Sprintf("\nS:\"Hostname\"=%s", ip))
	data.WriteString(fmt.Sprintf("\nS:\"Protocol Name\"=%s", protocol))
	data.WriteString("\nZ:\"Description\"=00000003") // number of lines to display
	data.WriteString(fmt.Sprintf("\n Site: %s", site))
	data.WriteString(fmt.Sprintf("\n Type: %s", machine_type))
	data.WriteString(fmt.Sprintf("\n Adresse: %s", address))

	return data.String()
}

func (scrt *SecureCRT) WriteSession(path string, data string) error {
	info, err := os.Stat(scrt.configPath)
	if err != nil {
		return err
	}

	path = fmt.Sprintf("%s/Sessions/%s", scrt.configPath, path)
	err = os.MkdirAll(filepath.Dir(path), info.Mode())
	if err != nil {
		return errors.Join(ErrFailedToCreateSession, err)
	}

	err = os.WriteFile(path, []byte(data), info.Mode())
	if err != nil {
		return errors.Join(ErrFailedToCreateSession, err)
	}

	return nil
}

func (scrt *SecureCRT) RemoveSessions(path string) error {
	path = fmt.Sprintf("%s/Sessions/%s", scrt.configPath, path)
	return os.RemoveAll(path)
}
