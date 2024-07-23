package securecrt

import (
	"bufio"
	"errors"
	"fmt"
	"io"
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

	// make sure to also remove empty folders so they dont clutter
	folder := filepath.Dir(s.fullPath)
	empty, err := isDirEmpty(folder)
	if err != nil {
		return err
	}

	if empty {
		return os.RemoveAll(folder)
	}

	return nil
}

func isDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	names, err := f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}

	// a folder is also empty if it only have one of these files
	if len(names) == 1 && (names[0] == ".DS_Store" || names[0] == "__FolderData__.ini") {
		return true, nil
	}

	return false, err // Either not empty or error, suits both cases
}
