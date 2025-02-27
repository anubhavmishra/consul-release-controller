package v1

import (
	"encoding/json"
	"fmt"

	"github.com/nicholasjackson/consul-release-controller/models"
)

func (r *Release) ConvertToModel() *models.Release {
	mr := &models.Release{}
	mr.Name = r.Name
	mr.Namespace = r.Namespace
	// use
	mr.Version = fmt.Sprintf("%d", r.ObjectMeta.Generation)

	// Kubernetes uses CamelCase for it's CRDs while internally we use snake case
	// this means that the to JSON serialization does not work for not just
	// convert to the internal types. We can find a better way to do this later
	rpc := releaserConfigSnake(r.Spec.Releaser.Config)

	mr.Releaser = &models.PluginConfig{
		Name:   r.Spec.Releaser.PluginName,
		Config: getJSONRaw(rpc),
	}

	rupc := runtimeConfigSnake{
		Deployment: r.Spec.Runtime.Config.Deployment,
		Namespace:  r.Namespace,
	}

	mr.Runtime = &models.PluginConfig{
		Name:   r.Spec.Runtime.PluginName,
		Config: getJSONRaw(rupc),
	}

	spc := strategyConfigSnake(r.Spec.Strategy.Config)
	mr.Strategy = &models.PluginConfig{
		Name:   r.Spec.Strategy.PluginName,
		Config: getJSONRaw(spc),
	}

	mpq := []monitorQuerySnake{}
	for _, q := range r.Spec.Monitor.Config.Queries {
		mpq = append(mpq, monitorQuerySnake(q))
	}

	mpc := monitorConfigSnake{
		Address: r.Spec.Monitor.Config.Address,
		Queries: mpq,
	}

	mr.Monitor = &models.PluginConfig{
		Name:   r.Spec.Monitor.PluginName,
		Config: getJSONRaw(mpc),
	}

	return mr
}

func getJSONRaw(i interface{}) json.RawMessage {
	d, _ := json.Marshal(i)
	return d
}

type releaserConfigSnake struct {
	ConsulService string `json:"consul_service"`
}

type runtimeConfigSnake struct {
	Deployment string `json:"deployment,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
}

type strategyConfigSnake struct {
	InitialDelay   string `json:"initial_delay,omitempty"`
	Interval       string `json:"interval,omitempty"`
	InitialTraffic int    `json:"initial_traffic,omitempty"`
	TrafficStep    int    `json:"traffic_step,omitempty"`
	MaxTraffic     int    `json:"max_traffic,omitempty"`
	ErrorThreshold int    `json:"error_threshold,omitempty"`
}

type monitorConfigSnake struct {
	Address string              `json:"address,omitempty"`
	Queries []monitorQuerySnake `json:"queries,omitempty"`
}

type monitorQuerySnake struct {
	Name   string `json:"name,omitempty"`
	Preset string `json:"preset,omitempty"`
	Min    int    `json:"min,omitempty"`
	Max    int    `json:"max,omitempty"`
}
