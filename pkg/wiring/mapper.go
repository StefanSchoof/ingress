package wiring

import (
	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/log"
)

// Mapper handles data transfer over wires
type Mapper struct {
	wires      *Wires
	connectors *Connectors
}

// NewMapper creates a data mapper that is able to Process() input messages
// by sending them to configured output wires
func NewMapper(wires *Wires, conn *Connectors) *Mapper {
	mapper := &Mapper{
		wires:      wires,
		connectors: conn,
	}
	return mapper
}

// Process data generated by data source by passing to all wired data targets
func (m *Mapper) Process(source string, d api.Data) {
	d.Normalize() // normalize data before sending to any target

	for _, wire := range m.wires.WiresForSource(source) {
		wire := wire // pin
		log.Context(
			log.EV, d.Name(),
			log.SRC, wire.Source,
			log.TGT, wire.Target,
		).Debug("routing")

		// map and publish async
		go m.processWire(&wire, d)
	}
}

// async function for publishing
func (m *Mapper) processWire(wire *Wire, d api.Data) {
	d = m.processActions(wire, d)
	if d == nil {
		return
	}

	target, err := m.connectors.TargetForName(wire.Target)
	if err != nil {
		log.Fatal("invalid target " + wire.Target)
		return
	}

	target.Publish(d)
}

func (m *Mapper) processActions(wire *Wire, d api.Data) api.Data {
	if len(wire.Actions) == 0 {
		return d
	}

	// initial := d
	for _, action := range wire.Actions {
		d = action.Process(d)
		if d == nil {
			// log.Context(
			// 	log.EV, initial.Name(),
			// 	// log.ACT, action.Name,
			// ).Debugf("dropped")
			return nil
		}
	}

	return d
}
