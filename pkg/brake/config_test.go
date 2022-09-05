package brake

import (
	"github.com/cyrilix/robocar-throttle/pkg/types"
	"reflect"
	"testing"
)

func TestNewConfigFromJson(t *testing.T) {
	type args struct {
		fileName string
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		{
			name: "default config",
			args: args{
				fileName: "test_data/config.json",
			},
			want: &defaultBrakeConfig,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConfigFromJson(tt.args.fileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfigFromJson() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(*got, *tt.want) {
				t.Errorf("NewConfigFromJson() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got.DeltaSteps, tt.want.DeltaSteps) {
				t.Errorf("NewConfigFromJson(), bad DeltaSteps: got = %v, want %v", got.DeltaSteps, tt.want.DeltaSteps)
			}
		})
	}
}

func TestConfig_ValueOf(t *testing.T) {
	type fields struct {
		DeltaSteps []float32
		MinValue   int
		Data       []types.Throttle
	}
	type args struct {
		currentThrottle, targetThrottle types.Throttle
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   types.Throttle
	}{
		{
			name: "delta > 0",
			fields: fields{
				DeltaSteps: defaultBrakeConfig.DeltaSteps,
				Data:       defaultBrakeConfig.Data,
			},
			args: args{
				currentThrottle: 0.5,
				targetThrottle:  0.8,
			},
			want: 0.8,
		},
		{
			name: "no delta",
			fields: fields{
				DeltaSteps: defaultBrakeConfig.DeltaSteps,
				Data:       defaultBrakeConfig.Data,
			},
			args: args{
				currentThrottle: 0.5,
				targetThrottle:  0.5,
			},
			want: 0.5,
		},
		{
			name: "delta very low (< 1st step)",
			fields: fields{
				DeltaSteps: defaultBrakeConfig.DeltaSteps,
				Data:       defaultBrakeConfig.Data,
			},
			args: args{
				currentThrottle: 0.5,
				targetThrottle:  0.495,
			},
			want: 0.495,
		},
		{
			name: "low delta ( 1st step < delta < 2nd step )",
			fields: fields{
				DeltaSteps: defaultBrakeConfig.DeltaSteps,
				Data:       defaultBrakeConfig.Data,
			},
			args: args{
				currentThrottle: 0.5,
				targetThrottle:  0.38,
			},
			want: -0.1,
		},
		{
			name: "high delta",
			fields: fields{
				DeltaSteps: defaultBrakeConfig.DeltaSteps,
				Data:       defaultBrakeConfig.Data,
			},
			args: args{
				currentThrottle: 0.8,
				targetThrottle:  0.3,
			},
			want: -1.,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Config{
				DeltaSteps: tt.fields.DeltaSteps,
				Data:       tt.fields.Data,
			}
			got := f.ValueOf(tt.args.currentThrottle, tt.args.targetThrottle)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValueOf() = %v, want %v", got, tt.want)
			}
		})
	}
}
