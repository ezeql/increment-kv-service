package testdata

import "testing"

func Test_fillMask(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// we want to generate a uuid with all items being the arg, being 0 in this case
		{name: "test1", args: args{0}, want: "00000000-0000-0000-0000-000000000000"},
		// save as above as using 2
		{name: "test2", args: args{2}, want: "22222222-2222-2222-2222-222222222222"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fillMask(tt.args.n); got != tt.want {
				t.Errorf("fillMask() = %v, want %v", got, tt.want)
			}
		})
	}
}
