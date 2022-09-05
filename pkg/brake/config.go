package brake

import (
	"encoding/json"
	"fmt"
	"github.com/cyrilix/robocar-throttle/pkg/types"
	"os"
)

var (
	defaultBrakeConfig = Config{
		DeltaSteps: []float32{0.05, 0.3, 0.5},
		Data:       []types.Throttle{-0.1, -0.5, -1.},
	}
)

func NewConfig() *Config {
	return &defaultBrakeConfig
}

func NewConfigFromJson(fileName string) (*Config, error) {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("unable to read content from %s file: %w", fileName, err)
	}
	var ft Config
	err = json.Unmarshal(content, &ft)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal json content from %s file: %w", fileName, err)
	}
	return &ft, nil
}

type Config struct {
	DeltaSteps []float32        `json:"delta_steps"`
	Data       []types.Throttle `json:"data"`
}

func (tc *Config) ValueOf(currentThrottle, targetThrottle types.Throttle) types.Throttle {
	delta := float32(currentThrottle - targetThrottle)

	if delta < tc.DeltaSteps[0] {
		return targetThrottle
	}
	if delta >= tc.DeltaSteps[len(tc.DeltaSteps)-1] {
		return tc.Data[len(tc.Data)-1]
	}
	for idx, step := range tc.DeltaSteps {
		if delta < step {
			return tc.Data[idx-1]
		}
	}
	return tc.Data[len(tc.Data)-1]
}
