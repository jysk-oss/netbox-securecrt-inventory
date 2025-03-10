package netbox

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

type NetBoxRespone[T any] struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []T    `json:"results"`
}

type NetBox struct {
	url        string
	token      string
	limit      int32
	ctx        context.Context
	httpClient *http.Client
}

func New(url string, token string, ctx context.Context) *NetBox {
	schema := "https://"
	if strings.Contains(url, "http://") || strings.Contains(url, "https://") {
		schema = ""
	}

	url = fmt.Sprintf("%s%s", schema, url)
	var limit int32 = 1000

	return &NetBox{
		url:        url,
		token:      token,
		limit:      limit,
		ctx:        ctx,
		httpClient: &http.Client{},
	}
}

func (nb *NetBox) PrepareRequest(method string, url string) (*http.Request, error) {
	url = fmt.Sprintf("%s/api%s", nb.url, url)
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", nb.token))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	return req, nil
}

func (nb *NetBox) TestConnection() error {
	req, err := nb.PrepareRequest("GET", "/status")
	if err != nil {
		return err
	}

	_, err = nb.httpClient.Do(req)
	if err != nil {
		slog.Error("Unable to connect to netbox", slog.String("url", nb.url), slog.String("error", err.Error()))
		return fmt.Errorf("unable to connect: %s", nb.url)
	}

	return nil
}

func (nb *NetBox) GetSites() ([]Site, error) {
	var results = make([]Site, 0)
	hasMorePages := true
	for hasMorePages {
		currentCount := len(results)
		req, err := nb.PrepareRequest("GET", fmt.Sprintf("/dcim/sites/?limit=%d&offset=%d", nb.limit, currentCount))
		if err != nil {
			return results, err
		}

		response, err := nb.httpClient.Do(req)
		if err != nil {
			slog.Error("Failed to get sites from netbox", slog.String("error", err.Error()))
			return nil, ErrFailedToQuerySites
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			slog.Error("Failed to read body from sites request", slog.String("error", err.Error()))
			return nil, ErrFailedToQuerySites
		}

		var data NetBoxRespone[Site]
		err = json.Unmarshal(body, &data)
		if err != nil {
			slog.Error("Failed to parse sites from netbox", slog.String("error", err.Error()))
			return nil, ErrFailedToQuerySites
		}

		results = append(results, data.Results...)
		if len(data.Results) < int(nb.limit) {
			hasMorePages = false
		}
	}

	slog.Info("Retrieved sites", slog.Int("count", len(results)))
	return results, nil
}

func (nb *NetBox) GetDevices() ([]DeviceWithConfigContext, error) {
	var results = make([]DeviceWithConfigContext, 0)
	hasMorePages := true
	for hasMorePages {
		currentCount := len(results)
		req, err := nb.PrepareRequest("GET", fmt.Sprintf("/dcim/devices/?has_primary_ip=true&limit=%d&offset=%d", nb.limit, currentCount))
		if err != nil {
			return results, err
		}

		response, err := nb.httpClient.Do(req)
		if err != nil {
			slog.Error("Failed to get sites from netbox", slog.String("error", err.Error()))
			return nil, ErrFailedToQuerySites
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			slog.Error("Failed to read body from sites request", slog.String("error", err.Error()))
			return nil, ErrFailedToQuerySites
		}

		var data NetBoxRespone[DeviceWithConfigContext]
		err = json.Unmarshal(body, &data)
		if err != nil {
			slog.Error("Failed to parse sites from netbox", slog.String("error", err.Error()))
			return nil, ErrFailedToQuerySites
		}

		results = append(results, data.Results...)
		if len(data.Results) < int(nb.limit) {
			hasMorePages = false
		}
	}

	slog.Info("Retrieved devices", slog.Int("count", len(results)))
	return results, nil
}

func (nb *NetBox) GetVirtualMachines() ([]VirtualMachineWithConfigContext, error) {
	var results = make([]VirtualMachineWithConfigContext, 0)
	hasMorePages := true
	for hasMorePages {
		currentCount := len(results)
		req, err := nb.PrepareRequest("GET", fmt.Sprintf("/virtualization/virtual-machines/?has_primary_ip=true&limit=%d&offset=%d", nb.limit, currentCount))
		if err != nil {
			return results, err
		}

		response, err := nb.httpClient.Do(req)
		if err != nil {
			slog.Error("Failed to get sites from netbox", slog.String("error", err.Error()))
			return nil, ErrFailedToQuerySites
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			slog.Error("Failed to read body from sites request", slog.String("error", err.Error()))
			return nil, ErrFailedToQuerySites
		}

		var data NetBoxRespone[VirtualMachineWithConfigContext]
		err = json.Unmarshal(body, &data)
		if err != nil {
			slog.Error("Failed to parse sites from netbox", slog.String("error", err.Error()))
			return nil, ErrFailedToQuerySites
		}

		results = append(results, data.Results...)
		if len(data.Results) < int(nb.limit) {
			hasMorePages = false
		}
	}

	slog.Info("Retrieved virtual machines", slog.Int("count", len(results)))
	return results, nil
}

func (nb *NetBox) GetConsoleServerPorts() ([]ConsoleServerPort, error) {
	var results = make([]ConsoleServerPort, 0)
	hasMorePages := true
	for hasMorePages {
		currentCount := len(results)
		req, err := nb.PrepareRequest("GET", fmt.Sprintf("/dcim/console-server-ports?limit=%d&offset=%d", nb.limit, currentCount))
		if err != nil {
			return results, err
		}

		response, err := nb.httpClient.Do(req)
		if err != nil {
			slog.Error("Failed to get sites from netbox", slog.String("error", err.Error()))
			return nil, ErrFailedToQuerySites
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			slog.Error("Failed to read body from sites request", slog.String("error", err.Error()))
			return nil, ErrFailedToQuerySites
		}

		var data NetBoxRespone[ConsoleServerPort]
		err = json.Unmarshal(body, &data)
		if err != nil {
			slog.Error("Failed to parse sites from netbox", slog.String("error", err.Error()))
			return nil, ErrFailedToQuerySites
		}

		results = append(results, data.Results...)
		if len(data.Results) < int(nb.limit) {
			hasMorePages = false
		}
	}

	slog.Info("Retrieved console server ports", slog.Int("count", len(results)))
	return results, nil
}
