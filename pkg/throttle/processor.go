package throttle

import (
	"encoding/json"
	"fmt"
	"github.com/cyrilix/robocar-protobuf/go/events"
	"github.com/cyrilix/robocar-throttle/pkg/types"
	"math"
	"os"
	"sync"
)

type Processor interface {
	// Process compute throttle from steering value
	Process(steering types.Steering) types.Throttle

	SetSpeedZone(sz events.SpeedZone)
}

func NewSteeringProcessor(minThrottle, maxThrottle types.Throttle) *SteeringProcessor {
	return &SteeringProcessor{
		minThrottle: minThrottle,
		maxThrottle: maxThrottle,
	}
}

type SteeringProcessor struct {
	minThrottle, maxThrottle types.Throttle
}

func (sp *SteeringProcessor) SetSpeedZone(_ events.SpeedZone) {
	return
}

// Process compute throttle from steering value
func (sp *SteeringProcessor) Process(steering types.Steering) types.Throttle {
	absSteering := math.Abs(float64(steering))
	return sp.minThrottle + types.Throttle(float64(sp.maxThrottle-sp.minThrottle)*(1-absSteering))
}

func NewSpeedZoneProcessor(slowThrottle, normalThrottle, fastThrottle types.Throttle,
	moderateSteering, fullSteering float64) *SpeedZoneProcessor {
	return &SpeedZoneProcessor{
		muSz:             sync.Mutex{},
		speedZone:        events.SpeedZone_UNKNOWN,
		slowThrottle:     slowThrottle,
		normalThrottle:   normalThrottle,
		fastThrottle:     fastThrottle,
		moderateSteering: moderateSteering,
		fullSteering:     fullSteering,
	}
}

type SpeedZoneProcessor struct {
	muSz                                       sync.Mutex
	speedZone                                  events.SpeedZone
	slowThrottle, normalThrottle, fastThrottle types.Throttle
	moderateSteering, fullSteering             float64
}

func (sp *SpeedZoneProcessor) SpeedZone() events.SpeedZone {
	sp.muSz.Lock()
	defer sp.muSz.Unlock()
	return sp.speedZone
}

func (sp *SpeedZoneProcessor) SetSpeedZone(sz events.SpeedZone) {
	sp.muSz.Lock()
	defer sp.muSz.Unlock()
	sp.speedZone = sz
}

// Process compute throttle from steering value
func (sp *SpeedZoneProcessor) Process(steering types.Steering) types.Throttle {
	st := math.Abs(float64(steering))

	switch sp.SpeedZone() {
	case events.SpeedZone_FAST:
		if st >= sp.fullSteering {
			return sp.slowThrottle
		} else if st >= sp.moderateSteering {
			return sp.normalThrottle
		}
		return sp.fastThrottle
	case events.SpeedZone_NORMAL:
		if st > sp.fullSteering {
			return sp.slowThrottle
		}
		return sp.normalThrottle
	case events.SpeedZone_SLOW:
		return sp.slowThrottle
	}
	return sp.slowThrottle
}

func NewCustomSteeringProcessor(cfg *Config) *CustomSteeringProcessor {
