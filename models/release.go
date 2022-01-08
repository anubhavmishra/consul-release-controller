package models

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/looplab/fsm"
	"github.com/nicholasjackson/consul-canary-controller/plugins/interfaces"
)

type StateHistory struct {
	Time  time.Time
	State string
}

type Release struct {
	Name string `json:"name"`

	Created     time.Time `json:"created"`
	LastUpdated time.Time `json:"last_updated"`

	CurrentState string         `json:"current_state"`
	StateHistory []StateHistory `json:"state_history"`

	Releaser *PluginConfig `json:"releaser"`
	Runtime  *PluginConfig `json:"runtime"`
	Strategy *PluginConfig `json:"strategy"`
	Monitor  *PluginConfig `json:"monitor"`

	state *fsm.FSM

	releaserPlugin interfaces.Releaser
	runtimePlugin  interfaces.Runtime
	monitorPlugin  interfaces.Monitor
	strategyPlugin interfaces.Strategy
}

type PluginConfig struct {
	Name   string          `json:"plugin_name"`
	Config json.RawMessage `json:"config"`
}

// Build creates a new deployment setting the state to inactive
// unless current state is set, this indicates that the release
// has been de-serialzed
func (d *Release) Build(pluginProvider interfaces.Provider) error {
	d.StateHistory = []StateHistory{}

	// configure the setup plugin
	relP, err := pluginProvider.CreateReleaser(d.Releaser.Name)
	if err != nil {
		return err
	}

	// configure the releaser plugin
	relP.Configure(d.Releaser.Config)
	d.releaserPlugin = relP

	// configure the runtime plugin
	runP, err := pluginProvider.CreateRuntime(d.Runtime.Name)
	if err != nil {
		return err
	}

	// configure the runtime plugin
	runP.Configure(d.Runtime.Config)
	d.runtimePlugin = runP

	// configure the monitor plugin
	monP, err := pluginProvider.CreateMonitor(d.Monitor.Name)
	if err != nil {
		return err
	}

	// configure the monitor plugin
	monP.Configure(d.Monitor.Config)
	d.monitorPlugin = monP

	// configure the monitor plugin
	stratP, err := pluginProvider.CreateStrategy(d.Strategy.Name, monP)
	if err != nil {
		return err
	}

	// configure the strategy plugin
	stratP.Configure(d.Strategy.Config)
	d.strategyPlugin = stratP

	fsm := newFSM(d, relP, runP, stratP, pluginProvider.GetLogger())
	d.state = fsm

	// if we are rehydrating this we probably have an existing state
	if d.CurrentState != "" {
		d.state.SetState(d.CurrentState)
	}

	return err
}

// FromJsonBody decodes the json body into the Deployment type
func (d *Release) FromJsonBody(r io.ReadCloser) error {
	if r == nil {
		return fmt.Errorf("no json body provided")
	}

	return json.NewDecoder(r).Decode(d)
}

// ToJson serializes the deployment to json
func (d *Release) ToJson() []byte {
	// serialize the current state
	d.CurrentState = d.State()

	data, err := json.Marshal(d)
	if err != nil {
		panic(err)
	}

	return data
}

// RuntimePlugin returns the runtime plugin for this release
func (d *Release) RuntimePlugin() interfaces.Runtime {
	return d.runtimePlugin
}

func (d *Release) SetState(s string) {
	d.state.SetState(s)
}

// Save release to the datastore
func (d *Release) Save(state string) {
	d.StateHistory = append(d.StateHistory, StateHistory{Time: time.Now(), State: state})
	d.CurrentState = state
}

// StateIs returns true when the internal state matches the check state
func (d *Release) StateIs(s string) bool {
	if d.state == nil {
		return false
	}

	return d.state.Is(s)
}

// State returns the internal state of the deployment
func (d *Release) State() string {
	if d.state == nil {
		return ""
	}

	return d.state.Current()
}

// Configure the deployment and create any necessary configuration
func (d *Release) Configure() error {
	// trigger the configure event
	return d.state.Event(EventConfigure)
}

// Deploy the application
func (d *Release) Deploy() error {
	return d.state.Event(EventDeploy)
}
