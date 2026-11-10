package netbox

import "errors"

var (
	ErrFailedToQuerySites              = errors.New("unable to get sites")
	ErrFailedToQueryDevices            = errors.New("unable to get devices")
	ErrFailedToQueryVirtualMachines    = errors.New("unable to get virtual machines")
	ErrFailedToQueryConsoleServerPorts = errors.New("unable to get console server ports")
)
