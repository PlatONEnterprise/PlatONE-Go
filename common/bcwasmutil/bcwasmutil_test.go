package bcwasmutil

import (
	"reflect"
	"testing"
)

func TestUnsignedInt_Bytes(t *testing.T) {
	tests := []struct {
		name string
		i    UnsignedInt
		want []byte
	}{
		// TODO: Add test cases.
		{
			i: 100,
			want: []byte{100},
		},
		{
			i: 0x80,
			want: []byte{0x80, 1},
		},
		{
			i: 0xff,
			want:[]byte{0xff, 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.Bytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnsignedInt.Bytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnsignedInt_Uint32(t *testing.T) {
	tests := []struct {
		name string
		i    UnsignedInt
		want uint32
	}{
		// TODO: Add test cases.
		{
			i: 30,
			want: 30,
		},
		{
			i: 200,
			want: 200,
		},
		{
			i: 300,
			want: 300,
		},
		{
			i: 1024,
			want: 1024,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.Uint32(); got != tt.want {
				t.Errorf("UnsignedInt.Uint32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnsignedInt_Int32(t *testing.T) {
	tests := []struct {
		name string
		i    UnsignedInt
		want int32
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.Int32(); got != tt.want {
				t.Errorf("UnsignedInt.Int32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnsignedInt_FromBytes(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		i       *UnsignedInt
		args    args
		wantPos int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotPos := tt.i.FromBytes(tt.args.data); gotPos != tt.wantPos {
				t.Errorf("UnsignedInt.FromBytes() = %v, want %v", gotPos, tt.wantPos)
			}
		})
	}
}
