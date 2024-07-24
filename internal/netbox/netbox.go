package netbox

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/netbox-community/go-netbox/v4"
)

type NetBox struct {
	url    string
	token  string
	client *netbox.APIClient
	limit  int32
	ctx    context.Context
}

func New(url string, token string, ctx context.Context) *NetBox {
	nb := netbox.NewAPIClientFor(url, token)
	var limit int32 = 1000

	return &NetBox{
		url:    url,
		token:  token,
		client: nb,
		limit:  limit,
		ctx:    ctx,
	}
}

func (nb *NetBox) TestConnection() error {
	_, _, err := nb.client.StatusAPI.StatusRetrieve(nb.ctx).Execute()
	if err != nil {
		slog.Error("Unable to connect to netbox", slog.String("url", nb.url), slog.String("error", err.Error()))
		return fmt.Errorf("unable to connect: %s", nb.url)
	}

	return nil
}

func (nb *NetBox) GetSites() ([]netbox.Site, error) {
	var results = make([]netbox.Site, 0)
	hasMorePages := true
	for hasMorePages {
		currentCount := int32(len(results))
		response, _, err := nb.client.DcimAPI.DcimSitesList(nb.ctx).Limit(nb.limit).Offset(currentCount).Execute()
		if err != nil {
			slog.Error("Failed to get sites from netbox", slog.String("error", err.Error()))
			return nil, ErrFailedToQuerySites
		}

		results = append(results, response.Results...)
		if len(response.Results) < int(nb.limit) {
			hasMorePages = false
		}
	}

	slog.Info("Retrived sites", slog.Int("count", len(results)))
	return results, nil
}

func (nb *NetBox) GetDevices() ([]netbox.DeviceWithConfigContext, error) {
	var results = make([]netbox.DeviceWithConfigContext, 0)
	hasMorePages := true
	for hasMorePages {
		currentCount := int32(len(results))

		response, _, err := nb.client.DcimAPI.DcimDevicesList(nb.ctx).HasPrimaryIp(true).Limit(nb.limit).Offset(currentCount).Execute()
		if err != nil {
			slog.Error("Failed to get devices from netbox", slog.String("error", err.Error()))
			return nil, ErrFailedToQueryDevices
		}
		results = append(results, response.Results...)
		if len(response.Results) < int(nb.limit) {
			hasMorePages = false
		}
	}

	slog.Info("Retrived devices", slog.Int("count", len(results)))
	return results, nil
}

func (nb *NetBox) GetVirtualMachines() ([]netbox.VirtualMachineWithConfigContext, error) {
	var results = make([]netbox.VirtualMachineWithConfigContext, 0)
	hasMorePages := true
	for hasMorePages {
		currentCount := int32(len(results))
		response, _, err := nb.client.VirtualizationAPI.VirtualizationVirtualMachinesList(nb.ctx).HasPrimaryIp(true).Limit(int32(nb.limit)).Offset(currentCount).Execute()
		if err != nil {
			slog.Error("Failed to get virtual machines from netbox", slog.String("error", err.Error()))
			return nil, ErrFailedToQueryDevices
		}
		results = append(results, response.Results...)
		if len(response.Results) < int(nb.limit) {
			hasMorePages = false
		}
	}

	slog.Info("Retrived virtual machines", slog.Int("count", len(results)))
	return results, nil
}
