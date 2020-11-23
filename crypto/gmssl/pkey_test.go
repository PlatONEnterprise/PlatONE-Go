package gmssl

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGenerateECPrivateKey(t *testing.T) {
	type args struct {
		curve string
	}
	tests := []struct {
		name    string
		args    args
		want    *PrivateKey
		wantErr bool
	}{
		// TODO: Add test cases.
		{args:args{"SM2"}},
		{args:args{"secp256k1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateECPrivateKey(tt.args.curve)
			fmt.Println(got.GetText())

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateECPrivateKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateECPrivateKey() got = %v, want %v", got, tt.want)
			}
		})
	}
}