package throttle

import "testing"

func TestSteeringProcessor_Process(t *testing.T) {
	type fields struct {
		minThrottle float32
		maxThrottle float32
	}
	type args struct {
		steering float32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   float32
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
