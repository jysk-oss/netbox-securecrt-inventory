package securecrt

import (
	"fmt"
	"os"
	"runtime"
)

func getConfigPath() (string, error) {
	if runtime.GOOS == "linux" {
		// On Linux, SecureCRT uses ~/.vandyke/SecureCRT/Config
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", ErrFailedToExpandHomeDir
		}
		return fmt.Sprintf("%s/.vandyke/SecureCRT/Config", homeDir), nil
	}
	
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

func loadDefaultSessionConfig(sessionPath string) (string, error) {
	data, err := os.ReadFile(fmt.Sprintf("%s/Default.ini", sessionPath))
	if err != nil {
		return "", ErrFailedToLoadConfig
	}

	return string(data), nil
}
