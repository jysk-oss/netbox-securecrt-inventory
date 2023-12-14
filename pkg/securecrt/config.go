package securecrt

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

func getConfigPath() (string, error) {
	appDataDir, err := os.UserConfigDir()
	if err != nil {
		return "", ErrFailedToExpandHomeDir
	}

	path := "VanDyke/SecureCRT/Config"
	if runtime.GOOS == "windows" {
		path = "VanDyke/Config"
	}

	return fmt.Sprintf("%s/%s", appDataDir, path), nil
}

func loadDefaultSessionConfig(configPath string) (string, error) {
	data, err := os.ReadFile(fmt.Sprintf("%s/Sessions/Default.ini", configPath))
	if err != nil {
		return "", ErrFailedToLoadConfig
	}

	return string(data), nil
}

func getCredentialHash(configPath string, credentialName *string) (string, error) {
	if credentialName == nil {
		return "D:\"Session Password Saved\"=00000000", nil
	}

	paths, err := os.ReadDir(fmt.Sprintf("%s/Credentials", configPath))
	if err != nil {
		return "", ErrFailedToLoadCredentials
	}

	for _, path := range paths {
		if !strings.Contains(path.Name(), *credentialName) {
			continue
		}

		file := strings.ReplaceAll(path.Name(), ".ini", "")
		return fmt.Sprintf("S:\"Credential Title\"=%s", file), nil
	}

	return "", ErrFailedToLoadCredentials
}
