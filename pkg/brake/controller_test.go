package brake

import (
	"github.com/cyrilix/robocar-throttle/pkg/types"
	"testing"
)

func TestController_AdjustThrottle(t *testing.T) {
	type fields struct {
		realThrottle      types.Throttle
		acceleratorFactor float64
	}
	type args struct {
		targetThrottle types.Throttle
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   types.Throttle
	}{
		{
			name:   "target same as current throttle with big acceleration",
			fields: fields{realThrottle: 0.2, acceleratorFactor: 2.},
			args:   args{targetThrottle: 0.2},
			want:   0.2,
		},
		{
			name:   "target > as current throttle with big acceleration",
			fields: fields{realThrottle: 0.2, acceleratorFactor: 2.},
			args:   args{targetThrottle: 0.3},
			want:   0.40000004,
		},
		{
			name:   "target >> as current throttle with big acceleration",
			fields: fields{realThrottle: 0.2, acceleratorFactor: 2.},
			args:   args{targetThrottle: 0.5},
			want:   0.8,
		},
		{
			name:   "target >> as current throttle with big acceleration, result > 1",
			fields: fields{realThrottle: 0.2, acceleratorFactor: 3.},
			args:   args{targetThrottle: 0.5},
			want:   1.0,
		},
		{
			name:   "target < as current throttle with big acceleration",
			fields: fields{realThrottle: 0.8, acceleratorFactor: 2.},
			args:   args{targetThrottle: 0.7},
			want:   -0.1,
		},
		{
			name:   "target << as current throttle with big acceleration",
			fields: fields{realThrottle: 0.8, acceleratorFactor: 2.},
			args:   args{targetThrottle: 0.5},
			want:   -0.5,
		},
		{
			name:   "target <<< as current throttle with big acceleration",
			fields: fields{realThrottle: 0.8, acceleratorFactor: 2.},
			args:   args{targetThrottle: 0.2},
			want:   -1.,
		},

		{
			name:   "target same as current throttle",
			fields: fields{realThrottle: 0.2, acceleratorFactor: 1.},
			args:   args{targetThrottle: 0.2},
			want:   0.2,
		},
		{
			name:   "target > as current throttle",
			fields: fields{realThrottle: 0.2, acceleratorFactor: 1.},
			args:   args{targetThrottle: 0.3},
			want:   0.3,
		},
		{
			name:   "target >> as current throttle",
			fields: fields{realThrottle: 0.2, acceleratorFactor: 1.},
			args:   args{targetThrottle: 0.5},
			want:   0.5,
		},
		{
			name:   "target < as current throttle",
			fields: fields{realThrottle: 0.8, acceleratorFactor: 1.},
			args:   args{targetThrottle: 0.7},
			want:   -0.1,
		},
		{
			name:   "target << as current throttle",
			fields: fields{realThrottle: 0.8, acceleratorFactor: 1.},
			args:   args{targetThrottle: 0.5},
			want:   -0.5,
		},
		{
			name:   "target <<< as current throttle",
			fields: fields{realThrottle: 0.8, acceleratorFactor: 1.},
			args:   args{targetThrottle: 0.2},
			want:   -1.,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &CustomController{cfg: NewConfig(), acceleratorFactor: tt.fields.acceleratorFactor}
			b.SetRealThrottle(tt.fields.realThrottle)
			if got := b.AdjustThrottle(tt.args.targetThrottle); got != tt.want {
				t.Errorf("AdjustThrottle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDisabledController_AdjustThrottle(t *testing.T) {
	type args struct {
		targetThrottle types.Throttle
	}
	tests := []struct {
		name string
		args args
		want types.Throttle
	}{
		{
			name: "doesn't modify value",
			args: args{targetThrottle: 0.5},
			want: 0.5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DisabledController{}
			if got := d.AdjustThrottle(tt.args.targetThrottle); got != tt.want {
				t.Errorf("AdjustThrottle() = %v, want %v", got, tt.want)
			}
		})
	}
}
