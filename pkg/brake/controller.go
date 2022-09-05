package brake

import (
	"github.com/cyrilix/robocar-throttle/pkg/types"
	"go.uber.org/zap"
	"sync"
)

type Controller interface {
	SetRealThrottle(t types.Throttle)
	AdjustThrottle(targetThrottle types.Throttle) types.Throttle
}

func NewCustomController() *CustomController {
	return &CustomController{cfg: NewConfig()}
}

func NewCustomControllerWithJsonConfig(filename string) *CustomController {
	config, err := NewConfigFromJson(filename)
	if err != nil {
		zap.S().Panicf("unable to init brake controller with json config '%s': %v", filename, err)
	}
	return &CustomController{cfg: config}
}

type CustomController struct {
	muRealThrottle sync.RWMutex
	realThrottle   types.Throttle
	cfg            *Config
}

func (b *CustomController) SetRealThrottle(t types.Throttle) {
	b.muRealThrottle.Lock()
	defer b.muRealThrottle.Unlock()
	b.realThrottle = t
}

func (b *CustomController) GetRealThrottle() types.Throttle {
	b.muRealThrottle.RLock()
	defer b.muRealThrottle.RUnlock()
	res := b.realThrottle
	return res
}

func (b *CustomController) AdjustThrottle(targetThrottle types.Throttle) types.Throttle {
	return b.cfg.ValueOf(b.GetRealThrottle(), targetThrottle)
}

type DisabledController struct{}

func (d *DisabledController) SetRealThrottle(_ types.Throttle) {}

func (d *DisabledController) AdjustThrottle(targetThrottle types.Throttle) types.Throttle {
	return targetThrottle
}
