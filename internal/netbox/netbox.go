package netbox

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
)

type NetBoxRespone[T any] struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []T    `json:"results"`
}

type NetBox struct {
	url                  string
	token                string
	limit                int32
	excludeConfigContext bool
	ctx                  context.Context
	httpClient           *http.Client
	cursorPagination     atomic.Bool
}

func New(url string, token string, limit int32, excludeConfigContext bool, ctx context.Context) *NetBox {
	schema := "https://"
	if strings.Contains(url, "http://") || strings.Contains(url, "https://") {
		schema = ""
	}

	url = fmt.Sprintf("%s%s", schema, url)

	return &NetBox{
		url:                  url,
		token:                token,
		limit:                limit,
		excludeConfigContext: excludeConfigContext,
		ctx:                  ctx,
		httpClient:           &http.Client{},
	}
}

func (nb *NetBox) PrepareRequest(method string, url string) (*http.Request, error) {
	url = fmt.Sprintf("%s/api%s", nb.url, url)
	req, err := http.NewRequestWithContext(nb.ctx, method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", nb.token))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	return req, nil
}

func (nb *NetBox) get(path string, result any) error {
	req, err := nb.PrepareRequest(http.MethodGet, path)
	if err != nil {
		return err
	}
	response, err := nb.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: %s", path, response.Status)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, result)
}

func (nb *NetBox) TestConnection() error {
	var status struct {
		Version string `json:"netbox-version"`
	}
	if err := nb.get("/status/", &status); err != nil {
		slog.Error("Unable to connect to netbox", slog.String("url", nb.url), slog.String("error", err.Error()))
		return fmt.Errorf("unable to connect: %s: %w", nb.url, err)
	}

	var major, minor int
	_, err := fmt.Sscanf(status.Version, "%d.%d", &major, &minor)
	nb.cursorPagination.Store(err == nil && (major > 4 || major == 4 && minor >= 6))
	if err != nil {
		slog.Warn("Unknown NetBox version; using offset pagination", slog.String("version", status.Version))
	}
	return nil
}

func (nb *NetBox) GetSites() ([]Site, error) {
	results, err := getAll[Site](nb, "/dcim/sites/")
	if err != nil {
		slog.Error("Failed to get sites from netbox", slog.String("error", err.Error()))
		return nil, ErrFailedToQuerySites
	}
	slog.Info("Retrieved sites", slog.Int("count", len(results)))
	return results, nil
}

func (nb *NetBox) GetDevices() ([]DeviceWithConfigContext, error) {
	path := "/dcim/devices/?has_primary_ip=true"
	if nb.excludeConfigContext {
		path += "&exclude=config_context"
	}
	results, err := getAll[DeviceWithConfigContext](nb, path)
	if err != nil {
		slog.Error("Failed to get devices from netbox", slog.String("error", err.Error()))
		return nil, ErrFailedToQueryDevices
	}
	slog.Info("Retrieved devices", slog.Int("count", len(results)))
	return results, nil
}

func (nb *NetBox) GetVirtualMachines() ([]VirtualMachineWithConfigContext, error) {
	path := "/virtualization/virtual-machines/?has_primary_ip=true"
	if nb.excludeConfigContext {
		path += "&exclude=config_context"
	}
	results, err := getAll[VirtualMachineWithConfigContext](nb, path)
	if err != nil {
		slog.Error("Failed to get virtual machines from netbox", slog.String("error", err.Error()))
		return nil, ErrFailedToQueryVirtualMachines
	}
	slog.Info("Retrieved virtual machines", slog.Int("count", len(results)))
	return results, nil
}

func (nb *NetBox) GetConsoleServerPorts() ([]ConsoleServerPort, error) {
	results, err := getAll[ConsoleServerPort](nb, "/dcim/console-server-ports/")
	if err != nil {
		slog.Error("Failed to get console server ports from netbox", slog.String("error", err.Error()))
		return nil, ErrFailedToQueryConsoleServerPorts
	}
	slog.Info("Retrieved console server ports", slog.Int("count", len(results)))
	return results, nil
}

func getAll[T any](nb *NetBox, path string) ([]T, error) {
	endpoint, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	query := endpoint.Query()
	query.Set("limit", strconv.Itoa(int(nb.limit)))
	if nb.cursorPagination.Load() {
		query.Set("start", "0")
	} else {
		query.Set("offset", "0")
	}

	endpoint.RawQuery = query.Encode()
	results := make([]T, 0)
	for {
		var data NetBoxRespone[T]
		if err := nb.get(endpoint.String(), &data); err != nil {
			return nil, err
		}

		results = append(results, data.Results...)
		if data.Next == "" {
			return results, nil
		}

		next, err := url.Parse(data.Next)
		if err != nil {
			return nil, err
		}

		if next.RawQuery == "" || next.RawQuery == endpoint.RawQuery {
			return nil, fmt.Errorf("pagination did not advance for %s", path)
		}

		// Keep the configured server and endpoint when following pagination links.
		endpoint.RawQuery = next.RawQuery
	}
}
