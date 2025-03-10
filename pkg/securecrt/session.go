package securecrt

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

type SecureCRTSession struct {
	DeviceName     string
	Path           string
	IP             string `session:"Hostname" type:"S"`
	Port           int    `session:"[SSH2] Port" type:"D"`
	Protocol       string `session:"Protocol Name" type:"S"`
	Description    string `session:"Description" type:"Z"`
	CredentialName string `session:"Credential Title" type:"S"`
	Firewall       string `session:"Firewall Name" type:"S"`
	fullPath       string
}

func NewSession(fullPath string) *SecureCRTSession {
	return &SecureCRTSession{
		Firewall:       "None",
		CredentialName: "",
		fullPath:       fullPath,
	}
}

func (s *SecureCRTSession) read() error {
	f, err := os.OpenFile(s.fullPath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	pattern, err := regexp.Compile("^([A-Z]):\\\"(.*)\"=(.*)$")
	if err != nil {
		return err
	}

	sc := bufio.NewScanner(f)

	for sc.Scan() {
		result := pattern.FindStringSubmatch(sc.Text())
		if result == nil {
			continue
		}

		// Z: indicates multiline content
		if result[1] == "Z" {
			multiLineContent := ""
			for sc.Scan() {
				result := pattern.FindAllString(sc.Text(), -1)
				if result == nil {
					multiLineContent += "\n" + sc.Text()
					continue
				}
				break
			}

			s.setInternalValue(result[2], multiLineContent)
			continue
		}

		// all other types are single line
		s.setInternalValue(result[2], result[3])
	}
	if err := sc.Err(); err != nil {
		return err
	}

	// set DeviceName and Path manually as they are file names
	configPath, _ := getConfigPath()
	configPath = configPath + "/Sessions"
	path, fileName := filepath.Split(s.fullPath)
	s.DeviceName = strings.ReplaceAll(fileName, ".ini", "")
	s.Path = strings.ReplaceAll(path, configPath, "")

	return nil
}

func (s *SecureCRTSession) setInternalValue(key string, value string) {
	val := reflect.ValueOf(s).Elem()
	for i := 0; i < val.NumField(); i++ {
		tag := val.Type().Field(i).Tag.Get("session")

		if tag == key && val.CanSet() {
			if val.Field(i).Kind() == reflect.String {
				val.Field(i).SetString(value)
			}

			if val.Field(i).Kind() == reflect.Pointer {
				val.Field(i).Set(reflect.ValueOf(&value))
			}
		}
	}
}

func (s *SecureCRTSession) write(defaultConfig string, mode fs.FileMode) error {
	var data strings.Builder
	data.WriteString(defaultConfig)

	// based on the tags we can generate the correct securecrt config format
	val := reflect.ValueOf(s).Elem()
	for i := 0; i < val.NumField(); i++ {
		itemType := val.Type().Field(i).Tag.Get("type")
		key := val.Type().Field(i).Tag.Get("session")
		if key == "" {
			continue
		}

		value := val.Field(i).String()
		if val.Field(i).Kind() == reflect.Pointer {
			value = val.Field(i).Elem().String()
		}

		if itemType == "Z" {
			items := strings.Split(value, "\n")
			itemsLengthPadded := fmt.Sprintf("%08d", len(items))
			data.WriteString(fmt.Sprintf("%s:\"%s\"=%s\n", itemType, key, itemsLengthPadded))
			for _, v := range items {
				if v != "" {
					data.WriteString(" " + v + "\n")
				}
			}
		} else if itemType == "D" {
			data.WriteString(fmt.Sprintf("%s:\"%s\"=%08X\n", itemType, key, val.Field(i).Int()))
		} else {
			data.WriteString(fmt.Sprintf("%s:\"%s\"=%s\n", itemType, key, value))
		}
	}

	err := os.MkdirAll(filepath.Dir(s.fullPath), mode)
	if err != nil {
		slog.Error("failed to create securecrt session directory", slog.String("error", err.Error()))
		return errors.Join(ErrFailedToCreateSession, err)
	}

	err = os.WriteFile(s.fullPath, []byte(data.String()), mode)
	if err != nil {
		slog.Error("failed to write securecrt session", slog.String("error", err.Error()))
		return errors.Join(ErrFailedToCreateSession, err)
	}

	return nil
}

func (s *SecureCRTSession) delete() error {
	err := os.Remove(s.fullPath)
	if err != nil {
		return err
	}

	// find the highest-level folder that is “empty” (ignoring this child folder and some ignored files)
	folder, err := getFolderToDelete(filepath.Dir(s.fullPath))
	if err != nil {
		return err
	}

	if folder != "" {
		return os.RemoveAll(folder)
	}

	return nil
}

// getFolderToDelete climbs the directory tree starting at 'folder'. It
// returns the highest ancestor directory that is “empty” except for a single
// subdirectory (the one we just deleted) and optionally a few ignored files.
func getFolderToDelete(folder string) (string, error) {
	// Start by ignoring the folder’s own basename.
	base := filepath.Base(folder)
	empty, err := isDirEmpty(folder, base)
	if err != nil {
		return "", err
	}

	if !empty {
		return "", nil
	}

	// Climb up the tree. In each parent directory, ignore the subdirectory (child)
	// that we came from.
	for {
		parent := filepath.Clean(filepath.Join(folder, ".."))
		empty, err := isDirEmpty(parent, filepath.Base(folder))
		if err != nil {
			return "", err
		}
		if !empty {
			break
		}
		folder = parent
	}
	return folder, nil
}

// isDirEmpty checks if the directory 'dir' is empty except for an entry
// named 'ignore' (and a couple of extra files that we want to ignore).
func isDirEmpty(dir, ignore string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}

	for _, entry := range entries {
		// Skip the ignored directory entry.
		if entry.IsDir() && entry.Name() == ignore {
			continue
		}

		// Skip known ignorable files.
		if entry.Name() == ".DS_Store" || entry.Name() == "__FolderData__.ini" {
			continue
		}

		// Any other file or directory means 'dir' is not empty.
		return false, nil
	}

	return true, nil
}
