package vm

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
)

func TestUserRoles_setRole(t *testing.T) {
	type args struct {
		role int32
	}
	tests := []struct {
		name       string
		roles      UserRoles
		args       args
		targetRole UserRoles
		wantErr    bool
	}{
		// TODO: Add test cases.
		{
			roles:   UserRoles(0),
			args:    args{role: -1},
			wantErr: true,
		},
		{
			roles:   UserRoles(0),
			args:    args{role: rolesCnt},
			wantErr: true,
		},
		{
			roles:      UserRoles(0),
			args:       args{role: superAdmin},
			targetRole: 0b1,
			wantErr:    false,
		},
		{
			roles:      UserRoles(1),
			args:       args{role: chainAdmin},
			targetRole: 0b11,
			wantErr:    false,
		},
		{
			roles:      UserRoles(0b11),
			args:       args{role: nodeAdmin},
			targetRole: 0b1011,
			wantErr:    false,
		},
		{
			roles:      UserRoles(0b111),
			args:       args{role: contractAdmin},
			targetRole: 0b10111,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.roles.setRole(tt.args.role)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserRoles.setRole() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.roles != tt.targetRole {
				t.Errorf("UserRoles.setRole() roles = %v, wantResult %v", tt.roles, tt.targetRole)
			}
		})
	}
}

func TestUserRoles_unsetRole(t *testing.T) {
	type args struct {
		role int32
	}
	tests := []struct {
		name        string
		roles       UserRoles
		args        args
		targetRoles UserRoles
		wantErr     bool
	}{
		// TODO: Add test cases.
		{
			roles:   UserRoles(0),
			args:    args{role: -1},
			wantErr: true,
		},
		{
			roles:   UserRoles(0),
			args:    args{role: rolesCnt},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.roles.unsetRole(tt.args.role); (err != nil) != tt.wantErr {
				t.Errorf("UserRoles.unsetRole() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserRoles_hasRole(t *testing.T) {
	type args struct {
		role int32
	}
	tests := []struct {
		name  string
		roles UserRoles
		args  args
		want  bool
	}{
		// TODO: Add test cases.
		{
			roles: UserRoles(0b1),
			args:  args{role: superAdmin},
			want:  true,
		},
		{
			roles: UserRoles(0),
			args:  args{role: superAdmin},
			want:  false,
		},
		{
			roles: UserRoles(0b10101),
			args:  args{role: superAdmin},
			want:  true,
		},
		{
			roles: UserRoles(0b10101),
			args:  args{role: chainAdmin},
			want:  false,
		},
		{
			roles: UserRoles(0b10101),
			args:  args{role: groupAdmin},
			want:  true,
		},
		{
			roles: UserRoles(0b10101),
			args:  args{role: contractAdmin},
			want:  true,
		},
		{
			roles: UserRoles(0b10101),
			args:  args{role: contractDeployer},
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.roles.hasRole(tt.args.role); got != tt.want {
				t.Errorf("UserRoles.hasRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserManagement_addChainAdminByAddress(t *testing.T) {
	type fields struct {
		Contract *Contract
		Evm      *EVM
	}
	type args struct {
		addr common.Address
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserManagement{}
			got, err := u.addChainAdminByAddress(tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserManagement.addChainAdminByAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserManagement.addChainAdminByAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserManagement_delChainAdminByAddress(t *testing.T) {
	type fields struct {
		Contract *Contract
		Evm      *EVM
	}
	type args struct {
		addr common.Address
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserManagement{}
			got, err := u.delChainAdminByAddress(tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserManagement.delChainAdminByAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserManagement.delChainAdminByAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserManagement_addNodeAdminByAddress(t *testing.T) {
	type fields struct {
		Contract *Contract
		Evm      *EVM
	}
	type args struct {
		addr common.Address
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserManagement{}
			got, err := u.addNodeAdminByAddress(tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserManagement.addNodeAdminByAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserManagement.addNodeAdminByAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserManagement_delNodeAdminByAddress(t *testing.T) {
	type fields struct {
		Contract *Contract
		Evm      *EVM
	}
	type args struct {
		addr common.Address
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserManagement{}
			got, err := u.delNodeAdminByAddress(tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserManagement.delNodeAdminByAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserManagement.delNodeAdminByAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserManagement_addContractAdminByAddress(t *testing.T) {
	type fields struct {
		Contract *Contract
		Evm      *EVM
	}
	type args struct {
		addr common.Address
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserManagement{}
			got, err := u.addContractAdminByAddress(tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserManagement.addContractAdminByAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserManagement.addContractAdminByAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserManagement_delContractAdminByAddress(t *testing.T) {
	type fields struct {
		Contract *Contract
		Evm      *EVM
	}
	type args struct {
		addr common.Address
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserManagement{}
			got, err := u.delContractAdminByAddress(tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserManagement.delContractAdminByAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserManagement.delContractAdminByAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserManagement_addContractDeployerByAddress(t *testing.T) {
	db := newMockStateDB()

	type fields struct {
		Contract *Contract
		Evm      *EVM
	}
	type args struct {
		addr common.Address
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserManagement{
				stateDB: db,
				caller:  ZeroAddress,
			}
			got, err := u.addContractDeployerByAddress(tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserManagement.addContractDeployerByAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserManagement.addContractDeployerByAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserManagement_delContractDeployerByAddress(t *testing.T) {
	type fields struct {
		Contract *Contract
		Evm      *EVM
	}
	type args struct {
		addr common.Address
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserManagement{}
			got, err := u.delContractDeployerByAddress(tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserManagement.delContractDeployerByAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserManagement.delContractDeployerByAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserManagement_getRolesByAddress(t *testing.T) {
	type fields struct {
		Contract *Contract
		Evm      *EVM
	}
	type args struct {
		addr common.Address
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserManagement{}
			got, err := u.getRolesByAddress(tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserManagement.getRolesByAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserManagement.getRolesByAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserManagement_getRole(t *testing.T) {
	type fields struct {
		Contract *Contract
		Evm      *EVM
	}
	type args struct {
		addr common.Address
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    UserRoles
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserManagement{}
			got, err := u.getRole(tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserManagement.getRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UserManagement.getRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserManagement_setRole(t *testing.T) {
	type fields struct {
		Contract *Contract
		Evm      *EVM
	}
	type args struct {
		addr  common.Address
		roles UserRoles
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserManagement{}
			if err := u.setRole(tt.args.addr, tt.args.roles); (err != nil) != tt.wantErr {
				t.Errorf("UserManagement.setRole() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserManagement_setRoleWithPermissionCheck(t *testing.T) {
	type fields struct {
		Contract *Contract
		Evm      *EVM
	}
	type args struct {
		addr       common.Address
		targetRole int32
		status     uint8
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserManagement{}
			if err := u.setRoleWithPermissionCheckByAddress(tt.args.addr, tt.args.targetRole, tt.args.status); (err != nil) != tt.wantErr {
				t.Errorf("UserManagement.setRoleWithPermissionCheckByAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserManagement_Caller(t *testing.T) {
	type fields struct {
		Contract *Contract
		Evm      *EVM
	}
	tests := []struct {
		name   string
		fields fields
		want   common.Address
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserManagement{}
			if got := u.Caller(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserManagement.Caller() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserManagement_getAddrList(t *testing.T) {
	addr := []common.Address{ZeroAddress}
	data, _ := json.Marshal(addr)
	fmt.Println(string(data))

	addr1 := []common.Address{}
	fmt.Println(json.Unmarshal(data, &addr1))
	fmt.Println(addr1)
}
func generateKey(targetRole int32) []byte {
	key := addressListKey
	switch targetRole {
	case superAdmin:
		key += "superAdmin"
	case chainAdmin:
		key += "chainAdmin"
	case nodeAdmin:
		key += "nodeAdmin"
	case contractAdmin:
		key += "contractAdmin"
	case contractDeployer:
		key += "contractAdmin"
	default:
		return nil
	}
	return []byte(key)
}

func TestUserManagement_addAddrList(t *testing.T) {
	type fields struct {
		state  StateDB
		caller common.Address
	}
	db := newMockStateDB()

	type args struct {
		key  []byte
		addr common.Address
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []common.Address
		wantErr error
	}{
		// TODO: Add test cases.
		{
			fields: fields{
				state:  db,
				caller: common.Address{},
			},
			args: args{
				key:  generateKey(superAdmin),
				addr: ZeroAddress,
			},
			want:    []common.Address{ZeroAddress},
			wantErr: nil,
		},
		{
			fields: fields{
				state:  db,
				caller: common.Address{},
			},
			args: args{
				key:  generateKey(superAdmin),
				addr: common.HexToAddress("0x0000000000000000000000000000000000000001"),
			},
			want:    []common.Address{ZeroAddress, common.HexToAddress("0x0000000000000000000000000000000000000001")},
			wantErr: nil,
		},
		{
			fields: fields{
				state:  db,
				caller: common.Address{},
			},
			args: args{
				key:  generateKey(superAdmin),
				addr: common.HexToAddress("0x0000000000000000000000000000000000000002"),
			},
			want:    []common.Address{ZeroAddress, common.HexToAddress("0x0000000000000000000000000000000000000001"), common.HexToAddress("0x0000000000000000000000000000000000000002")},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserManagement{
				stateDB: tt.fields.state,
				caller:  tt.fields.caller,
			}
			err := u.addAddrList(tt.args.key, tt.args.addr)
			if err != tt.wantErr {
				t.Errorf("UserManagement.addAddrList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			want, err := u.getAddrList(tt.args.key)
			if err != nil || !reflect.DeepEqual(want, tt.want) {
				t.Errorf("UserManagement.addAddrList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestUserManagement_delAddrList(t *testing.T) {
	type fields struct {
		state  StateDB
		caller common.Address
	}
	db := newMockStateDB()
	u := UserManagement{
		caller:  ZeroAddress,
		stateDB: db,
	}
	key := generateKey(superAdmin)
	u.addAddrList(key, ZeroAddress)
	u.addAddrList(key, common.HexToAddress("0x0000000000000000000000000000000000000001"))
	u.addAddrList(key, common.HexToAddress("0x0000000000000000000000000000000000000002"))

	type args struct {
		key  []byte
		addr common.Address
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []common.Address
		wantErr error
	}{
		// TODO: Add test cases.
		{
			fields: fields{
				state:  db,
				caller: common.Address{},
			},
			args: args{
				key:  generateKey(superAdmin),
				addr: ZeroAddress,
			},
			want:    []common.Address{common.HexToAddress("0x0000000000000000000000000000000000000001"), common.HexToAddress("0x0000000000000000000000000000000000000002")},
			wantErr: nil,
		},
		{
			fields: fields{
				state:  db,
				caller: common.Address{},
			},
			args: args{
				key:  generateKey(superAdmin),
				addr: common.HexToAddress("0x0000000000000000000000000000000000000001"),
			},
			want:    []common.Address{common.HexToAddress("0x0000000000000000000000000000000000000002")},
			wantErr: nil,
		},
		{
			fields: fields{
				state:  db,
				caller: common.Address{},
			},
			args: args{
				key:  generateKey(superAdmin),
				addr: common.HexToAddress("0x0000000000000000000000000000000000000002"),
			},
			want:    []common.Address{},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserManagement{
				stateDB: tt.fields.state,
				caller:  tt.fields.caller,
			}
			err := u.delAddrList(tt.args.key, tt.args.addr)
			if err != tt.wantErr {
				t.Errorf("UserManagement.delAddrList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			want, err := u.getAddrList(tt.args.key)
			if err != nil || !reflect.DeepEqual(want, tt.want) {
				t.Errorf("UserManagement.delAddrList() ret = %v, want %v", want, tt.want)
				return
			}
		})
	}
}
