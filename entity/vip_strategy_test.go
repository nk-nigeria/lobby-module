package entity

import "testing"

func TestExchangeChipsToVipPoint(t *testing.T) {
	type args struct {
		chips int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		// TODO: Add test cases.
		{
			name: "not_sacrificed",
			args: args{
				chips: 7.5*100 - 1,
			},
			want: 0,
		},
		{
			name: "lv_1",
			args: args{
				chips: 15*1000 - 1,
			},
			want: 1,
		},
		{
			name: "lv_2",
			args: args{
				chips: 37.5*1000 - 1,
			},
			want: 2,
		},
		{
			name: "lv_3",
			args: args{
				chips: 90*1000 - 1,
			},
			want: 5,
		},
		{
			name: "lv_4",
			args: args{
				chips: 180*1000 - 1,
			},
			want: 10,
		},
		{
			name: "lv_5",
			args: args{
				chips: 450*1000 - 1,
			},
			want: 20,
		},
		{
			name: "lv_6",
			args: args{
				chips: 900*1000 - 1,
			},
			want: 50,
		},
		{
			name: "lv_6",
			args: args{
				chips: 900*1000 + 1,
			},
			want: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExchangeChipsToVipPoint(tt.args.chips); got != tt.want {
				t.Errorf("ExchangeChipsToVipPoint() = %v, want %v", got, tt.want)
			}
		})
	}
}
