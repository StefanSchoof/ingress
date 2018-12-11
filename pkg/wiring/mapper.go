package wiring

import (
	"log"
	"sync/atomic"

	"github.com/andig/ingress/pkg/data"
)

type Mapper struct {
	wiring   *Wiring
	targets  targetMap
	inflight int64 // number of inflight requests
}

// NewMapper creates a data mapper that is able to Process() input messages
// by sending them to configured output wires
func NewMapper(wiring *Wiring, conn *Connectors) *Mapper {
	mapper := &Mapper{
		wiring:  wiring,
		targets: conn.Target,
	}
	return mapper
}

// Process data generated by source by passing to all affected targets
func (m *Mapper) Process(source string, d *data.Data) {
	d.Normalize() // normalize data before sending to any target

	for _, wire := range m.wiring.WiresForSource(source) {
		publisher, ok := m.targets[wire.Target]
		if !ok {
			log.Println("mapper: invalid target " + wire.Target)
			continue
		}

		log.Printf("mapper: routing %s -> %s ", wire.Source, wire.Target)

		// publish async
		go func(d *data.Data) {
			atomic.AddInt64(&m.inflight, 1)
			publisher.Publish(*d)
			atomic.AddInt64(&m.inflight, -1)
		}(d)
	}
}
