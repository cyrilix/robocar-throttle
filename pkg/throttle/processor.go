package throttle

import (
	"github.com/cyrilix/robocar-throttle/pkg/types"
	"math"
)

type SteeringProcessor struct {
	minThrottle, maxThrottle types.Throttle
}

// Process compute throttle from steering value
func (sp *SteeringProcessor) Process(steering float32) types.Throttle {
	absSteering := math.Abs(float64(steering))
	return sp.minThrottle + types.Throttle(float64(sp.maxThrottle-sp.minThrottle)*(1-absSteering))
}
