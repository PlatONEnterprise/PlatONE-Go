package vm

import (
	"reflect"
	"testing"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
)

func TestUserRoles_encode(t *testing.T) {
	tests := []struct {
		name    string
		roles   *UserRoles
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.roles.encode()
			if (err != nil) != tt.wantErr {
				t.Errorf("UserRoles.encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserRoles.encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserRoles_setRole(t *testing.T) {
	type args struct {
		role int32
	}
	tests := []struct {
		name    string
		roles   *UserRoles
		args    args
		targetRole uint32
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			roles: &UserRoles{roles: 0},
			args : args{role: -1},
			wantErr:true,
		},
		{
			roles: &UserRoles{roles: 0},
			args : args{role: ROLES_CNT},
			wantErr:true,
		},
		{
			roles: &UserRoles{roles: 0},
			args : args{role: SUPER_ADMIN},
			targetRole: 0b1,
			wantErr:false,
		},
		{
			roles: &UserRoles{roles: 1},
			args : args{role: CHAIN_ADMIN},
			targetRole: 0b11,
			wantErr:false,
		},
		{
			roles: &UserRoles{roles: 0b11},
			args : args{role: NODE_ADMIN},
			targetRole: 0b111,
			wantErr:false,
		},
		{
			roles: &UserRoles{roles: 0b111},
			args : args{role: CONTRACT_ADMIN},
			targetRole: 0b1111,
			wantErr:false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.roles.setRole(tt.args.role);
			if (err != nil) != tt.wantErr {
				t.Errorf("UserRoles.setRole() error = %v, wantErr %v", err, tt.wantErr)
			}

			if  !tt.wantErr && tt.roles.roles != tt.targetRole {
				t.Errorf("UserRoles.setRole() roles = %v, wantResult %v", tt.roles.roles, tt.targetRole )
			}
		})
	}
}

func TestUserRoles_unsetRole(t *testing.T) {
	type args struct {
		role int32
	}
	tests := []struct {
		name    string
		roles   *UserRoles
		args    args
		targetRoles uint32
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			roles: &UserRoles{roles: 0},
			args:args{role: -1},
			wantErr: true,
		},
		{
			roles: &UserRoles{roles: 0},
			args:args{role: ROLES_CNT},
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
		roles *UserRoles
		args  args
		want  bool
	}{
		// TODO: Add test cases.
		{
			roles: &UserRoles{roles: 0b1},
			args:args{role:SUPER_ADMIN},
			want: true,
		},
		{
			roles: &UserRoles{roles: 0},
			args:args{role:SUPER_ADMIN},
			want: false,
		},
		{
			roles: &UserRoles{roles: 0b10101},
			args:args{role:SUPER_ADMIN},
			want: true,
		},
		{
			roles: &UserRoles{roles: 0b10101},
			args:args{role:CHAIN_ADMIN},
			want: false,
		},
		{
			roles: &UserRoles{roles: 0b10101},
			args:args{role:NODE_ADMIN},
			want: true,
		},
		{
			roles: &UserRoles{roles: 0b10101},
			args:args{role:CONTRACT_ADMIN},
			want: false,
		},
		{
			roles: &UserRoles{roles: 0b10101},
			args:args{role:CONTRACT_DEPLOYER},
			want: true,
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

func TestRetrieveUserRoles(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *UserRoles
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RetrieveUserRoles(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveUserRoles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RetrieveUserRoles() = %v, want %v", got, tt.want)
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
			u := &UserManagement{
				Contract: tt.fields.Contract,
				Evm:      tt.fields.Evm,
			}
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
			u := &UserManagement{
				Contract: tt.fields.Contract,
				Evm:      tt.fields.Evm,
			}
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
			u := &UserManagement{
				Contract: tt.fields.Contract,
				Evm:      tt.fields.Evm,
			}
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
			u := &UserManagement{
				Contract: tt.fields.Contract,
				Evm:      tt.fields.Evm,
			}
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
			u := &UserManagement{
				Contract: tt.fields.Contract,
				Evm:      tt.fields.Evm,
			}
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
			u := &UserManagement{
				Contract: tt.fields.Contract,
				Evm:      tt.fields.Evm,
			}
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
				Contract: tt.fields.Contract,
				Evm:      tt.fields.Evm,
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
			u := &UserManagement{
				Contract: tt.fields.Contract,
				Evm:      tt.fields.Evm,
			}
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
			u := &UserManagement{
				Contract: tt.fields.Contract,
				Evm:      tt.fields.Evm,
			}
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
		want    *UserRoles
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserManagement{
				Contract: tt.fields.Contract,
				Evm:      tt.fields.Evm,
			}
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
		roles *UserRoles
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
			u := &UserManagement{
				Contract: tt.fields.Contract,
				Evm:      tt.fields.Evm,
			}
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
			u := &UserManagement{
				Contract: tt.fields.Contract,
				Evm:      tt.fields.Evm,
			}
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
			u := &UserManagement{
				Contract: tt.fields.Contract,
				Evm:      tt.fields.Evm,
			}
			if got := u.Caller(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserManagement.Caller() = %v, want %v", got, tt.want)
			}
		})
	}
}
