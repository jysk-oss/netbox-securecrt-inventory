package netbox

import (
	"fmt"
	"log/slog"

	nbc "github.com/netbox-community/go-netbox/v3/netbox"
	"github.com/netbox-community/go-netbox/v3/netbox/client"
	"github.com/netbox-community/go-netbox/v3/netbox/client/dcim"
	"github.com/netbox-community/go-netbox/v3/netbox/client/status"
	"github.com/netbox-community/go-netbox/v3/netbox/client/virtualization"
	"github.com/netbox-community/go-netbox/v3/netbox/models"
)

type NetBox struct {
	url    string
	token  string
	client *client.NetBoxAPI
	limit  int64
}

func New(url string, token string) *NetBox {
	nb := nbc.NewNetboxWithAPIKey(url, token)
	var limit int64 = 1000

	return &NetBox{
		url:    url,
		token:  token,
		client: nb,
		limit:  limit,
	}
}

func (nb *NetBox) TestConnection() error {
	_, err := nb.client.Status.StatusList(status.NewStatusListParams(), nil)
	if err != nil {
		slog.Error("Unable to connect to netbox", slog.String("url", nb.url), slog.String("error", err.Error()))
		return fmt.Errorf("unable to connect: %s", nb.url)
	}

	return nil
}

func (nb *NetBox) GetSites() ([]*models.Site, error) {
	var results = make([]*models.Site, 0)
	hasMorePages := true
	for hasMorePages {
		currentCount := int64(len(results))
		query := dcim.NewDcimSitesListParams().WithLimit(&nb.limit).WithOffset(&currentCount)
		response, err := nb.client.Dcim.DcimSitesList(query, nil)
		if err != nil {
			slog.Error("Failed to get sites from netbox", slog.String("error", err.Error()))
			return nil, ErrFailedToQuerySites
		}

		results = append(results, response.Payload.Results...)
		if len(response.Payload.Results) < int(nb.limit) {
			hasMorePages = false
		}
	}

	slog.Info("Retrived sites", slog.Int("count", len(results)))
	return results, nil
}

func (nb *NetBox) GetDevices() ([]*models.DeviceWithConfigContext, error) {
	hasPrimaryIP := "true"

	var results = make([]*models.DeviceWithConfigContext, 0)
	hasMorePages := true
	for hasMorePages {
		currentCount := int64(len(results))
		query := dcim.NewDcimDevicesListParams().WithHasPrimaryIP(&hasPrimaryIP).WithLimit(&nb.limit).WithOffset(&currentCount)

		response, err := nb.client.Dcim.DcimDevicesList(query, nil)
		if err != nil {
			slog.Error("Failed to get devices from netbox", slog.String("error", err.Error()))
			return nil, ErrFailedToQueryDevices
		}
		results = append(results, response.Payload.Results...)
		if len(response.Payload.Results) < int(nb.limit) {
			hasMorePages = false
		}
	}

	slog.Info("Retrived devices", slog.Int("count", len(results)))
	return results, nil
}

func (nb *NetBox) GetVirtualMachines() ([]*models.VirtualMachineWithConfigContext, error) {
	hasPrimaryIP := "true"

	var results = make([]*models.VirtualMachineWithConfigContext, 0)
	hasMorePages := true
	for hasMorePages {
		currentCount := int64(len(results))
		query := virtualization.NewVirtualizationVirtualMachinesListParams().WithHasPrimaryIP(&hasPrimaryIP).WithLimit(&nb.limit).WithOffset(&currentCount)

		response, err := nb.client.Virtualization.VirtualizationVirtualMachinesList(query, nil)
		if err != nil {
			slog.Error("Failed to get virtual machines from netbox", slog.String("error", err.Error()))
			return nil, ErrFailedToQueryDevices
		}
		results = append(results, response.Payload.Results...)
		if len(response.Payload.Results) < int(nb.limit) {
			hasMorePages = false
		}
	}

	slog.Info("Retrived virtual machines", slog.Int("count", len(results)))
	return results, nil
}
