package securecrt

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

type SecureCRT struct {
	configPath    string
	defaultConfig string
	rootPath      string
	sessionPath   string
}

func New(rootPath string) (*SecureCRT, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	defaultConfig, err := loadDefaultSessionConfig(fmt.Sprintf("%s/Sessions", configPath))
	if err != nil {
		return nil, err
	}

	return &SecureCRT{
		configPath:    configPath,
		defaultConfig: defaultConfig,
		rootPath:      rootPath,
		sessionPath:   fmt.Sprintf("%s/Sessions/%s", configPath, rootPath),
	}, nil
}

func (scrt *SecureCRT) GetSessions() ([]*SecureCRTSession, error) {
	var mu sync.Mutex
	var eg errgroup.Group
	var sessions []*SecureCRTSession
	err := filepath.WalkDir(scrt.sessionPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".ini") {
			return err
		}

		fileName := strings.ReplaceAll(d.Name(), ".ini", "")
		if strings.HasPrefix(fileName, "__") && strings.HasSuffix(fileName, "__") {
			return nil
		}

		eg.Go(func() error {
			session := NewSession(path)
			err = session.read()
			if err != nil {
				return err
			}

			mu.Lock()
			sessions = append(sessions, session)
			mu.Unlock()
			return nil
		})

		return nil
	})

	if err != nil {
		return nil, ErrFailedToReadSession
	}

	err = eg.Wait()
	return sessions, err
}

func (scrt *SecureCRT) RemoveSessions(sessions []*SecureCRTSession) error {
	currentSessions, err := scrt.GetSessions()
	if err != nil {
		return err
	}

	for i := 0; i < len(currentSessions); i++ {
		found := slices.ContainsFunc(sessions, func(e *SecureCRTSession) bool {
			return currentSessions[i].DeviceName == e.DeviceName
		})

		if !found {
			err = currentSessions[i].delete()
			if err != nil {
				return err
			}
		}
	}

	return err
}

func (scrt *SecureCRT) GetSessionPath() string {
	return scrt.sessionPath
}

func (scrt *SecureCRT) WriteSession(session *SecureCRTSession) error {
	info, err := os.Stat(scrt.configPath)
	if err != nil {
		slog.Error("Failed to load securecrt session file", slog.String("error", err.Error()))
		return err
	}

	err = session.write(scrt.defaultConfig, info.Mode())
	if err != nil {
		slog.Error("Failed to write securecrt session", slog.String("error", err.Error()))
		return errors.Join(ErrFailedToCreateSession, err)
	}

	return nil
}
