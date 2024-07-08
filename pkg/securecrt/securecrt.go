package securecrt

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type SecureCRTSession struct {
	DeviceName     string
	Path           string
	IP             string
	Protocol       string
	Description    string
	CredentialName *string
	Firewall       *string
}

type SecureCRT struct {
	configPath    string
	defaultConfig string
}

func New() (*SecureCRT, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	defaultConfig, err := loadDefaultSessionConfig(configPath)
	if err != nil {
		return nil, err
	}

	return &SecureCRT{
		configPath:    configPath,
		defaultConfig: defaultConfig,
	}, nil
}

func (scrt *SecureCRT) BuildSessionData(session *SecureCRTSession) (string, error) {
	firewallValue := getFirewall(session.Firewall)
	credentialValue, err := getCredentialHash(scrt.configPath, session.CredentialName)
	if err != nil {
		slog.Error("Failed to load securecrt credential", slog.String("error", err.Error()))
		return "", err
	}

	var data strings.Builder
	data.WriteString(scrt.defaultConfig)
	data.WriteString(firewallValue)
	data.WriteString(credentialValue)
	data.WriteString(fmt.Sprintf("\nS:\"Hostname\"=%s", session.IP))
	data.WriteString(fmt.Sprintf("\nS:\"Protocol Name\"=%s", session.Protocol))
	data.WriteString("\nZ:\"Description\"=00000003") // number of lines to display
	data.WriteString(session.Description)

	return data.String(), nil
}

func (scrt *SecureCRT) WriteSession(path string, data string) error {
	info, err := os.Stat(scrt.configPath)
	if err != nil {
		slog.Error("Failed to load securecrt session file", slog.String("error", err.Error()))
		return err
	}

	path = fmt.Sprintf("%s/Sessions/%s", scrt.configPath, path)
	err = os.MkdirAll(filepath.Dir(path), info.Mode())
	if err != nil {
		slog.Error("Failed to create securecrt session directory", slog.String("error", err.Error()))
		return errors.Join(ErrFailedToCreateSession, err)
	}

	err = os.WriteFile(path, []byte(data), info.Mode())
	if err != nil {
		slog.Error("Failed to write securecrt session", slog.String("error", err.Error()))
		return errors.Join(ErrFailedToCreateSession, err)
	}

	return nil
}

func (scrt *SecureCRT) RemoveSessions(path string) error {
	path = fmt.Sprintf("%s/Sessions/%s", scrt.configPath, path)
	return os.RemoveAll(path)
}
