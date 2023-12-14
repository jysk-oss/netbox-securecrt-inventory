package netbox

import "errors"

var (
	ErrFailedToQuerySites   = errors.New("unable to get sites")
	ErrFailedToQueryDevices = errors.New("unable to get devices")
)
