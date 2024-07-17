package inventory

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jysk-network/netbox-securecrt-inventory/internal/config"
	"github.com/jysk-network/netbox-securecrt-inventory/pkg/evaluator"
	"github.com/jysk-network/netbox-securecrt-inventory/pkg/securecrt"
)

func applyDefaultOverrides(env *evaluator.Environment) error {
	protocol, err := evaluator.EvaluateResult(env.ConnectionProtocolTemplate, env)
	if err != nil {
		return err
	}

	path, err := evaluator.EvaluateResult(env.PathTemplate, env)
	if err != nil {
		return err
	}

	device_name, err := evaluator.EvaluateResult(env.DeviceNameTemplate, env)
	if err != nil {
		return err
	}

	firewall, err := evaluator.EvaluateResult(env.FirewallTemplate, env)
	if err != nil {
		return err
	}

	//TODO: should we check for string type here? yeah..
	env.ConnectionProtocol = protocol.(string)
	env.Path = path.(string)
	env.DeviceName = device_name.(string)
	env.Firewall = firewall.(string)

	address := strings.ReplaceAll(env.SiteAddress, "\n", ", ")
	env.Description = fmt.Sprintf("Site: %s", env.SiteName) + fmt.Sprintf("\nType: %s", env.DeviceType) + fmt.Sprintf("\nAdresse: %s", address)
	return nil
}

func applyOverrides(overrides []config.ConfigSessionOverride, env *evaluator.Environment) error {
	if slog.Default().Enabled(context.TODO(), slog.LevelDebug) {
		data, _ := json.Marshal(env)
		slog.Debug("Starting Override Evaluation", slog.String("device", env.DeviceName), slog.String("env", string(data)))
	}

	err := applyDefaultOverrides(env)
	if err != nil {
		return err
	}

	for _, override := range overrides {
		shouldOverride, err := evaluator.EvaluateCondition(override.Condition, env)
		if err != nil || !shouldOverride {
			continue
		}

		val, err := evaluator.EvaluateResult(override.Value, env)
		if err != nil {
			return err
		}

		sVal, ok := val.(string)
		if val != nil && ok {
			switch override.Target {
			case "path":
				env.Path = sVal
			case "device_name":
				env.DeviceName = sVal
			case "description":
				env.Description = sVal
			case "connection_protocol":
				env.ConnectionProtocol = sVal
			case "credential":
				env.Credential = sVal
			case "firewall":
				if sVal == "" || sVal == "None" {
					env.Firewall = "None"
				} else {
					env.Firewall = "Session:" + sVal
				}
			}
		}
	}

	return err
}

func getSessionWithOverrides(fullPath string, env *evaluator.Environment) *securecrt.SecureCRTSession {
	session := securecrt.NewSession(fullPath)
	session.IP = env.DeviceIP
	session.Path = env.Path
	session.DeviceName = env.DeviceName
	session.CredentialName = env.Credential
	session.Description = env.Description
	session.Protocol = env.ConnectionProtocol
	session.Firewall = env.Firewall

	return session
}
