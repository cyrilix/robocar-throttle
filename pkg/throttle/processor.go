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
	return &CustomSteeringProcessor{
		cfg: cfg,
	}
}

type CustomSteeringProcessor struct {
	cfg *Config
}

func (cp *CustomSteeringProcessor) Process(steering types.Steering) types.Throttle {
	return cp.cfg.ValueOf(steering)
}

func (cp *CustomSteeringProcessor) SetSpeedZone(_ events.SpeedZone) {
	return
}

var emptyConfig = Config{
	SteeringValues: []types.Steering{},
	ThrottleSteps:  []types.Throttle{},
}

func NewConfigFromJson(fileName string) (*Config, error) {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("unable to read content from %s file: %w", fileName, err)
	}
	var ft Config
	err = json.Unmarshal(content, &ft)
	if err != nil {
		return &emptyConfig, fmt.Errorf("unable to unmarshal json content from %s file: %w", fileName, err)
	}
	if len(ft.SteeringValues) == 0 {
		return &emptyConfig, fmt.Errorf("invalid configuration, none steering value'")
	}
	if len(ft.SteeringValues) != len(ft.ThrottleSteps) {
		return &emptyConfig, fmt.Errorf("invalid config, steering value number must be equals "+
			"to throttle value number: %v/%v", len(ft.SteeringValues), len(ft.ThrottleSteps))
	}
	lastT := types.Throttle(1.)
	for _, t := range ft.ThrottleSteps {
		if t < 0. || t > 1. {
			return &emptyConfig, fmt.Errorf("invalid throttle value: 0.0 < %v <= 1.0", t)
		}
		if t >= lastT {
			return &emptyConfig, fmt.Errorf("invalid throttle value, all values must be decreasing: %v <= %v", lastT, t)
		}
		lastT = t
	}
	lastS := types.Steering(-0.001)
	for _, s := range ft.SteeringValues {
		if s < 0. || s > 1. {
			return &emptyConfig, fmt.Errorf("invalid steering value: 0.0 < %v <= 1.0", s)
		}
		if s <= lastS {
			return &emptyConfig, fmt.Errorf("invalid steering value, all values must be increasing: %v <= %v", lastS, s)
		}
		lastS = s
	}
	return &ft, nil
}

type Config struct {
	SteeringValues []types.Steering `json:"steering_values"`
	ThrottleSteps  []types.Throttle `json:"throttle_steps"`
}

func (tc *Config) ValueOf(s types.Steering) types.Throttle {
	st := s
	if s < 0. {
		st = s * -1
	}
	if st < tc.SteeringValues[0] {
		return tc.ThrottleSteps[0]
	}

	for i, steeringStep := range tc.SteeringValues {
		if st < steeringStep {
			return tc.ThrottleSteps[i-1]
		}
	}
	return tc.ThrottleSteps[len(tc.ThrottleSteps)-1]
}
