package throttle

import "math"

type SteeringProcessor struct {
	minThrottle, maxThrottle float32
}

// Process compute throttle from steering value
func (sp *SteeringProcessor) Process(steering float32) float32 {
	absSteering := math.Abs(float64(steering))
	return sp.minThrottle + float32(float64(sp.maxThrottle-sp.minThrottle)*(1-absSteering))
}
