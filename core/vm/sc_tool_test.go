package vm

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/PlatONEnetwork/PlatONE-Go/common/byteutil"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"github.com/stretchr/testify/assert"
)

type fakeClass struct{}

func (fc *fakeClass) Fn(name string, age int64) (string, error) {
	return fmt.Sprintf("%s+%d", name, age), nil
}

func (fc *fakeClass) allExportFns() SCExportFns {
	return SCExportFns{
		"Fn": fc.Fn,
	}
}

func Test_retrieveFnNameAndParams(t *testing.T) {
	fnNameInput := "Fn"
	name := "wanxiang"
	var age int64 = 3
	var input = MakeInput(fnNameInput, name, age)
	txType, fnName, fn, params, err := retrieveFnAndParams(input, (&fakeClass{}).allExportFns())
	if nil != err {
		t.Error(err)
		return
	}

	assert.Equal(t, fnNameInput, fnName, "function name is invalid")
	assert.Equal(t, true, reflect.ValueOf(fn).IsValid(), "function is invalid")
	assert.Equal(t, 2, len(params), "params length is invalid")
	assert.Equal(t, name, params[0].String(), "params invalid")
	assert.Equal(t, age, params[1].Int(), "params invalid")
	assert.Equal(t, int(E_INVOKE_CONTRACT), txType)
}

func Test_execSC(t *testing.T) {
	fnNameInput := "Fn"
	name := "wanxiang"
	var age int64 = 3
	var input = MakeInput(fnNameInput, name, age)

	_, ret, err := execSC(input, (&fakeClass{}).allExportFns())
	if nil != err {
		t.Error(err)
		return
	}

	ret2, err := (&fakeClass{}).Fn(name, age)
	assert.NoError(t, err)
	assert.Equal(t, toContractReturnValueStringType(E_INVOKE_CONTRACT, []byte(ret2)), ret)

	input = MakeInput(fnNameInput, "bbb")
	_, _, err = execSC(input, (&fakeClass{}).allExportFns())
	assert.Error(t, err, "The params number invalid")
}

func MakeInput(fnName string, params ...interface{}) []byte {
	input := make([][]byte, 0)

	txTyp := byteutil.Int64ToBytes(int64(E_INVOKE_CONTRACT))

	input = append(input, txTyp)
	input = append(input, []byte(fnName))

	for _, v := range params {
		switch v.(type) {
		case int, int8, int16, int32, int64:
			param := byteutil.Int64ToBytes(reflect.ValueOf(v).Int())
			input = append(input, param)
		case uint, uint8, uint16, uint32, uint64:
			param := byteutil.Uint64ToBytes(reflect.ValueOf(v).Uint())
			input = append(input, param)
		case string:
			input = append(input, []byte(v.(string)))
		default:
			panic("unsupported type")
		}
	}

	encodedInput, err := rlp.EncodeToBytes(input)
	if nil != err {
		panic(err)
	}

	return encodedInput
}

const (
	E_INVOKE_CONTRACT = 1
)

func Test_checkNameFormat(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{
			args: args{"wanxiang"},
			want: true,
		},
		{
			args: args{"wan-xiang"},
			want: false,
		},
		{
			args: args{"wan xiang"},
			want: false,
		},
		{
			args: args{"wan_xiang"},
			want: true,
		},
		{
			args: args{"wan_xiang_123"},
			want: true,
		},
		{
			args: args{"_万_xiang_123"},
			want: true,
		},
		{
			args: args{"12_wan_xiang_123"},
			want: true,
		},
		{
			args: args{"1"},
			want: true,
		},
		{
			args: args{"13"},
			want: true,
		},
		{
			args: args{"12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"},
			want: true,
		},
		{
			args: args{"234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := checkNameFormat(tt.args.name); got != tt.want {
				t.Errorf("name=%v ,checkNameFormat() = %v, want %v, err %v", tt.args.name, got, tt.want, err)
			}
		})
	}
}

func Test_checkEmailFormat(t *testing.T) {
	type args struct {
		email string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{"gexin@wxblockchain.com"},
			want: true,
		},
		{
			args: args{"@wxblockchain.com"},
			want: false,
		},
		{
			args: args{"wxblockchain.com"},
			want: false,
		},
		{
			args: args{"gggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg@wxblockchainmmmmm.commmmmmm"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checkEmailFormat(tt.args.email)
			//if (err != nil) != tt.wantErr {
			//	t.Errorf("checkEmailFormat() error = %v, wantErr %v", err, tt.wantErr)
			//	return
			//}
			if got != tt.want {
				t.Errorf("checkEmailFormat() got = %v, want %v, err %v", got, tt.want, err)
			}
		})
	}
}

func Test_checkIpFormat(t *testing.T) {
	type args struct {
		ip string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checkIpFormat(tt.args.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkIpFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("checkIpFormat() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkIpFormat1(t *testing.T) {
	type args struct {
		ip string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{ip: "111"},
			want: false,
		},
		{
			args: args{ip: "111.11.11.1"},
			want: true,
		},
		{
			args: args{ip: "31.155.244.256"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checkIpFormat(tt.args.ip)
			//if (err != nil) != tt.wantErr {
			//	t.Errorf("checkIpFormat() error = %v, wantErr %v", err, tt.wantErr)
			//	return
			//}
			if got != tt.want {
				t.Errorf("checkIpFormat(%s) got = %v, want %v, err %v", tt.args.ip,got, tt.want, err)
			}
		})
	}
}
