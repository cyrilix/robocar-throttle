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

func TestConfig_ValueOf(t *testing.T) {
	type fields struct {
		SteeringValue []types.Steering
		Data          []types.Throttle
	}
	type args struct {
		s types.Steering
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   types.Throttle
	}{
		{
			name:   "Nil steering",
			fields: fields{[]types.Steering{0.0, 0.5, 1.0}, []types.Throttle{0.9, 0.6, 0.1}},
			args:   args{0.0},
			want:   0.9,
		},
		{
			name:   "Nil steering < min config",
			fields: fields{[]types.Steering{0.2, 0.5, 1.0}, []types.Throttle{0.9, 0.6, 0.3}},
			args:   args{0.1},
			want:   0.9,
		},
		{
			name:   "No nil steering",
			fields: fields{[]types.Steering{0.0, 0.5, 1.0}, []types.Throttle{0.9, 0.6, 0.1}},
			args:   args{0.2},
			want:   0.9,
		},
		{
			name:   "Intermediate steering",
			fields: fields{[]types.Steering{0.0, 0.5, 1.0}, []types.Throttle{0.9, 0.6, 0.1}},
			args:   args{0.5},
			want:   0.6,
		},
		{
			name:   "Max steering",
			fields: fields{[]types.Steering{0.0, 0.5, 1.0}, []types.Throttle{0.9, 0.6, 0.1}},
			args:   args{1.0},
			want:   0.1,
		},
		{
			name:   "Over steering",
			fields: fields{[]types.Steering{0.0, 0.5, 1.0}, []types.Throttle{0.9, 0.6, 0.1}},
			args:   args{1.1},
			want:   0.1,
		},
		{
			name:   "Negative steering < min config",
			fields: fields{[]types.Steering{0.2, 0.5, 1.0}, []types.Throttle{0.9, 0.6, 0.3}},
			args:   args{-0.1},
			want:   0.9,
		},
		{
			name:   "Negative steering",
			fields: fields{[]types.Steering{0.0, 0.5, 1.0}, []types.Throttle{0.9, 0.6, 0.1}},
			args:   args{-0.2},
			want:   0.9,
		},
		{
			name:   "Negative Intermediate steering",
			fields: fields{[]types.Steering{0.0, 0.5, 1.0}, []types.Throttle{0.9, 0.6, 0.1}},
			args:   args{-0.5},
			want:   0.6,
		},
		{
			name:   "Minimum steering",
			fields: fields{[]types.Steering{0.0, 0.5, 1.0}, []types.Throttle{0.9, 0.6, 0.1}},
			args:   args{-1.0},
			want:   0.1,
		},
		{
			name:   "Negative Over steering",
			fields: fields{[]types.Steering{0.0, 0.5, 1.0}, []types.Throttle{0.9, 0.6, 0.1}},
			args:   args{-1.1},
			want:   0.1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &Config{
				SteeringValues: tt.fields.SteeringValue,
				ThrottleSteps:  tt.fields.Data,
			}
			if got := tc.ValueOf(tt.args.s); got != tt.want {
				t.Errorf("ValueOf() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewConfigFromJson(t *testing.T) {
	type args struct {
		configContent string
	}

	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				configContent: `{
	"steering_values": [0.0, 0.5, 1.0],
	"throttle_steps": [0.9, 0.6, 0.1]
}
`,
			},
			want: &Config{
				SteeringValues: []types.Steering{0., 0.5, 1.},
				ThrottleSteps:  []types.Throttle{0.9, 0.6, 0.1},
			},
		},
		{
			name: "invalid config",
			args: args{
				configContent: `{ "steering_values" }`,
			},
			want:    &emptyConfig,
			wantErr: true,
		},
		{
			name: "empty config",
			args: args{
				configContent: `{
	"steering_values": [],
	"throttle_steps": []
}`,
			},
			want:    &emptyConfig,
			wantErr: true,
		},
		{
			name: "incoherent config",
			args: args{
				configContent: `{
	"steering_values": [0.0, 0.5, 1.0],
	"throttle_steps": [0.9, 0.1]
}`,
			},
			want:    &emptyConfig,
			wantErr: true,
		},
		{
			name: "steering in bad order",
			args: args{
				configContent: `{
	"steering_values": [0.0, 0.6, 0.5],
	"throttle_steps": [0.9, 0.5, 0.1]
}`,
			},
			want:    &emptyConfig,
			wantErr: true,
		},
		{
			name: "throttle in bad order",
			args: args{
				configContent: `{
	"steering_values": [0.0, 0.5, 0.9],
	"throttle_steps": [0.4, 0.5, 0.1]
}`,
			},
			want:    &emptyConfig,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configName := path.Join(t.TempDir(), "config.json")
			err := os.WriteFile(configName, []byte(tt.args.configContent), 0644)
			if err != nil {
				t.Errorf("unable to create test config: %v", err)
			}
			got, err := NewConfigFromJson(configName)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfigFromJson() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConfigFromJson() got = %v, want %v", got, tt.want)
			}
		})
	}
}
