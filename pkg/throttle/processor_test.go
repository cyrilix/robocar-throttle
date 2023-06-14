package throttle

import (
	"github.com/cyrilix/robocar-protobuf/go/events"
	"github.com/cyrilix/robocar-throttle/pkg/types"
	"os"
	"path"
	"reflect"
	"testing"
)

func TestSteeringProcessor_Process(t *testing.T) {
	type fields struct {
		minThrottle types.Throttle
		maxThrottle types.Throttle
	}
	type args struct {
		steering types.Steering
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   types.Throttle
	}{
		{
			name: "steering straight",
			fields: fields{
				minThrottle: 0.2,
				maxThrottle: 0.5,
			},
			args: args{
				steering: 0.,
			},
			want: 0.5,
		},
		{
			name: "steering full left should return min throttle",
			fields: fields{
				minThrottle: 0.2,
				maxThrottle: 0.5,
			},
			args: args{
				steering: -1.,
			},
			want: 0.2,
		},
		{
			name: "steering full right should return min throttle",
			fields: fields{
				minThrottle: 0.2,
				maxThrottle: 0.5,
			},
			args: args{
				steering: 1.,
			},
			want: 0.2,
		},
		{
			name: "steering mid-left should return intermediate throttle",
			fields: fields{
				minThrottle: 0.3,
				maxThrottle: 0.5,
			},
			args: args{
				steering: -0.5,
			},
			want: 0.4,
		},
		{
			name: "steering mid-right should return intermediate throttle",
			fields: fields{
				minThrottle: 0.3,
				maxThrottle: 0.5,
			},
			args: args{
				steering: 0.5,
			},
			want: 0.4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &SteeringProcessor{
				minThrottle: tt.fields.minThrottle,
				maxThrottle: tt.fields.maxThrottle,
			}
			if got := sp.Process(tt.args.steering); got != tt.want {
				t.Errorf("Process() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpeedZoneProcessor_Process(t *testing.T) {
	type fields struct {
		slowThrottle   types.Throttle
		normalThrottle types.Throttle
		fastThrottle   types.Throttle
		speedZone      events.SpeedZone
	}
	type args struct {
		steering types.Steering
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   types.Throttle
	}{
		{
			name:   "steering straight, undefined zone",
			fields: fields{slowThrottle: 0.2, normalThrottle: 0.5, fastThrottle: 0.8, speedZone: events.SpeedZone_SLOW},
			args:   args{steering: 0.},
			want:   0.2,
		},
		{
			name:   "steering straight, slow zone",
			fields: fields{slowThrottle: 0.2, normalThrottle: 0.5, fastThrottle: 0.8, speedZone: events.SpeedZone_SLOW},
			args:   args{steering: 0.},
			want:   0.2,
		},
		{
			name:   "moderate left, slow speed",
			fields: fields{slowThrottle: 0.2, normalThrottle: 0.5, fastThrottle: 0.8, speedZone: events.SpeedZone_SLOW},
			args:   args{steering: -0.5},
			want:   0.2,
		},
		{
			name:   "moderate right, slow speed",
			fields: fields{slowThrottle: 0.2, normalThrottle: 0.5, fastThrottle: 0.8, speedZone: events.SpeedZone_SLOW},
			args:   args{steering: 0.5},
			want:   0.2,
		},
		{
			name:   "full left, slow speed",
			fields: fields{slowThrottle: 0.2, normalThrottle: 0.5, fastThrottle: 0.8, speedZone: events.SpeedZone_SLOW},
			args:   args{steering: -0.95},
			want:   0.2,
		},
		{
			name:   "full right, slow speed",
			fields: fields{slowThrottle: 0.2, normalThrottle: 0.5, fastThrottle: 0.8, speedZone: events.SpeedZone_SLOW},
			args:   args{steering: 0.95},
			want:   0.2,
		},
		{
			name:   "steering straight, normal zone",
			fields: fields{slowThrottle: 0.2, normalThrottle: 0.5, fastThrottle: 0.8, speedZone: events.SpeedZone_NORMAL},
			args:   args{steering: 0.},
			want:   0.5,
		},
		{
			name:   "moderate left, normal speed",
			fields: fields{slowThrottle: 0.2, normalThrottle: 0.5, fastThrottle: 0.8, speedZone: events.SpeedZone_NORMAL},
			args:   args{steering: -0.5},
			want:   0.5,
		},
		{
			name:   "moderate right, normal speed",
			fields: fields{slowThrottle: 0.2, normalThrottle: 0.5, fastThrottle: 0.8, speedZone: events.SpeedZone_NORMAL},
			args:   args{steering: 0.5},
			want:   0.5,
		},
		{
			name:   "full left, normal speed",
			fields: fields{slowThrottle: 0.2, normalThrottle: 0.5, fastThrottle: 0.8, speedZone: events.SpeedZone_NORMAL},
			args:   args{steering: -0.95},
			want:   0.2,
		},
		{
			name:   "full right, normal speed",
			fields: fields{slowThrottle: 0.2, normalThrottle: 0.5, fastThrottle: 0.8, speedZone: events.SpeedZone_NORMAL},
			args:   args{steering: 0.95},
			want:   0.2,
		},
		{
			name:   "steering straight, fast zone",
			fields: fields{slowThrottle: 0.2, normalThrottle: 0.5, fastThrottle: 0.8, speedZone: events.SpeedZone_FAST},
			args:   args{steering: 0.},
			want:   0.8,
		},
		{
			name:   "moderate left, fast speed",
			fields: fields{slowThrottle: 0.2, normalThrottle: 0.5, fastThrottle: 0.8, speedZone: events.SpeedZone_FAST},
			args:   args{steering: -0.5},
			want:   0.5,
		},
		{
			name:   "moderate right, fast speed",
			fields: fields{slowThrottle: 0.2, normalThrottle: 0.5, fastThrottle: 0.8, speedZone: events.SpeedZone_FAST},
			args:   args{steering: 0.5},
			want:   0.5,
		},
		{
			name:   "full left, fast speed",
			fields: fields{slowThrottle: 0.2, normalThrottle: 0.5, fastThrottle: 0.8, speedZone: events.SpeedZone_FAST},
			args:   args{steering: -0.95},
			want:   0.2,
		},
		{
			name:   "full right, fast speed",
			fields: fields{slowThrottle: 0.2, normalThrottle: 0.5, fastThrottle: 0.8, speedZone: events.SpeedZone_FAST},
			args:   args{steering: 0.95},
			want:   0.2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &SpeedZoneProcessor{
				slowThrottle:     tt.fields.slowThrottle,
				normalThrottle:   tt.fields.normalThrottle,
				fastThrottle:     tt.fields.fastThrottle,
				moderateSteering: 0.4,
				fullSteering:     0.8,
			}
			sp.SetSpeedZone(tt.fields.speedZone)
			if got := sp.Process(tt.args.steering); got != tt.want {
				t.Errorf("Process() = %v, want %v", got, tt.want)
			}
		})
	}
}

