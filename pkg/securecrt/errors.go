package securecrt

import "errors"

var (
	ErrFailedToExpandHomeDir   = errors.New("unable to expand user home dir")
	ErrFailedToLoadConfig      = errors.New("failed to get securecrt config")
	ErrFailedToLoadCredentials = errors.New("failed to load default credentials")
	ErrFailedToCreateSession   = errors.New("failed to create session")
)
