package inventory

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jysk-network/netbox-securecrt-inventory/internal/config"
	"github.com/jysk-network/netbox-securecrt-inventory/pkg/evaluator"
	"github.com/jysk-network/netbox-securecrt-inventory/pkg/securecrt"
)

type SessionGenerator struct{}

func NewSessionGenerator() *SessionGenerator {
	return &SessionGenerator{}
}

func (sg *SessionGenerator) applyOverride(overrideTarget string, override interface{}, session *securecrt.SecureCRTSession) {
	val, ok := override.(string)
	if !ok {
		slog.Warn("Override is not string, skipping", slog.String("target", overrideTarget), slog.String("device", session.DeviceName), slog.Any("override", override))
		return
	}

	switch overrideTarget {
	case "path":
		session.Path = val
	case "device_name":
		session.DeviceName = val
	case "description":
		session.Description = val
	case "connection_protocol":
		session.Protocol = val
	case "credential":
		session.CredentialName = &val
	case "firewall":
		session.CredentialName = &val
	}
}

func (sg *SessionGenerator) getDefaultSession(env map[string]interface{}) (*securecrt.SecureCRTSession, error) {
	description := fmt.Sprintf("\n Site: %s", env["site_name"]) + fmt.Sprintf("\n Type: %s", env["device_type"]) + fmt.Sprintf("\n Adresse: %s", env["site_address"])
	session := securecrt.SecureCRTSession{
		IP:             env["device_ip"].(string),
		CredentialName: env["credential"].(*string),
		Description:    description,
		DeviceName:     env["device_name_template"].(string),
		Path:           env["path_template"].(string),
		Protocol:       env["connection_protocol_template"].(string),
		Firewall:       env["firewall_template"].(*string),
	}

	protocol, err := evaluator.EvaluateResult(session.Protocol, env)
	if err != nil {
		return nil, err
	}

	path, err := evaluator.EvaluateResult(session.Path, env)
	if err != nil {
		return nil, err
	}

	device_name, err := evaluator.EvaluateResult(session.DeviceName, env)
	if err != nil {
		return nil, err
	}

	if session.Firewall != nil {
		firewall, err := evaluator.EvaluateResult(*session.Firewall, env)
		if err != nil {
			return nil, err
		}

		val, ok := firewall.(string)
		if ok {
			session.Firewall = &val
		}
	}

	session.Protocol = protocol.(string)
	session.Path = path.(string)
	session.DeviceName = device_name.(string)
	return &session, nil
}

func (sg *SessionGenerator) GenerateSession(overrides []config.ConfigSessionOverride, env map[string]interface{}) (*securecrt.SecureCRTSession, error) {
	if slog.Default().Enabled(context.TODO(), slog.LevelDebug) {
		data, _ := json.Marshal(env)
		slog.Debug("Starting Override Evaluation", slog.String("device", env["device_name"].(string)), slog.String("env", string(data)))
	}

	session, err := sg.getDefaultSession(env)
	if err != nil {
		return nil, err
	}

	for _, override := range overrides {
		shouldOverride, err := evaluator.EvaluateCondition(override.Condition, env)
		if err != nil || !shouldOverride {
			continue
		}

		val, err := evaluator.EvaluateResult(override.Value, env)
		if err != nil {
			return nil, err
		}

		if val != nil {
			sg.applyOverride(override.Target, val, session)
		}
	}

	return session, nil
}
