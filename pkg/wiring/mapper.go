package wiring

import (
	"strings"

	"github.com/andig/ingress/pkg/data"
)

// Mapper handles data transfer over wires
type Mapper struct {
	wiring     *Wiring
	connectors *Connectors
}

// NewMapper creates a data mapper that is able to Process() input messages
// by sending them to configured output wires
func NewMapper(wiring *Wiring, conn *Connectors) *Mapper {
	mapper := &Mapper{
		wiring:     wiring,
		connectors: conn,
	}
	return mapper
}

// Process data generated by data source by passing to all wired data targets
func (m *Mapper) Process(source string, d data.Data) {
	d.Normalize() // normalize data before sending to any target

	for _, wire := range m.wiring.WiresForSource(source) {
		Log(
			"event", d.Name,
			"source", wire.Source,
			"target", wire.Target,
		).Debug("routing")

		// map and publish async
		go m.mapAndPublish(&wire, d)
	}
}

// async function for publishing
func (m *Mapper) mapAndPublish(wire *Wire, d data.Data) {
	if len(wire.Mappings) > 0 {
		dataName := strings.ToLower(d.Name)
		for mappingName, mapping := range wire.Mappings {
			for _, entry := range mapping {
				if dataName == strings.ToLower(entry.From) {
					Log(
						"event", d.Name,
						"mapping", mappingName,
					).Debugf("mapping %s -> %s ", d.Name, entry.To)
					d.Name = entry.To
					goto MAPPED
				}
			}
		}

		// not mapped
		Log("event", d.Name).Debugf("no mapping - dropped")
		return
	}
MAPPED:

	target, err := m.connectors.TargetForName(wire.Target)
	if err != nil {
		Log().Fatal("invalid target " + wire.Target)
		return
	}

	target.Publish(d)
}
