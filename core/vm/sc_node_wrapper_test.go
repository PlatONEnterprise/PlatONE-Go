package vm

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
	"reflect"
	"testing"
)

func Test_eNodesToString(t *testing.T) {
	type args struct {
		enodes []*eNode
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := eNodesToString(tt.args.enodes); got != tt.want {
				t.Errorf("eNodesToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newSCNodeWrapper(t *testing.T) {
	tests := []struct {
		name string
		want *scNodeWrapper
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newSCNodeWrapper(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newSCNodeWrapper() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scNodeWrapper_RequiredGas(t *testing.T) {
	type fields struct {
		base *SCNode
	}
	type args struct {
		input []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   uint64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &scNodeWrapper{
				base: tt.fields.base,
			}
			if got := n.RequiredGas(tt.args.input); got != tt.want {
				t.Errorf("RequiredGas() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scNodeWrapper_Run(t *testing.T) {
	type fields struct {
		base *SCNode
	}
	type args struct {
		input []byte
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
			n := &scNodeWrapper{
				base: tt.fields.base,
			}
			got, err := n.Run(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Run() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scNodeWrapper_add(t *testing.T) {
	type fields struct {
		base *SCNode
	}
	type args struct {
		node *syscontracts.NodeInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &scNodeWrapper{
				base: tt.fields.base,
			}
			got, err := n.add(tt.args.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("add() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("add() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scNodeWrapper_allExportFns(t *testing.T) {
	type fields struct {
		base *SCNode
	}
	tests := []struct {
		name   string
		fields fields
		want   SCExportFns
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &scNodeWrapper{
				base: tt.fields.base,
			}
			if got := n.allExportFns(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("allExportFns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scNodeWrapper_getAllNodes(t *testing.T) {
	type fields struct {
		base *SCNode
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &scNodeWrapper{
				base: tt.fields.base,
			}
			got, err := n.getAllNodes()
			if (err != nil) != tt.wantErr {
				t.Errorf("getAllNodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getAllNodes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scNodeWrapper_getENodesOfAllDeletedNodes(t *testing.T) {
	type fields struct {
		base *SCNode
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &scNodeWrapper{
				base: tt.fields.base,
			}
			got, err := n.getENodesOfAllDeletedNodes()
			if (err != nil) != tt.wantErr {
				t.Errorf("getENodesOfAllDeletedNodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getENodesOfAllDeletedNodes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scNodeWrapper_getENodesOfAllNormalNodes(t *testing.T) {
	type fields struct {
		base *SCNode
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &scNodeWrapper{
				base: tt.fields.base,
			}
			got, err := n.getENodesOfAllNormalNodes()
			if (err != nil) != tt.wantErr {
				t.Errorf("getENodesOfAllNormalNodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getENodesOfAllNormalNodes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scNodeWrapper_getNodes(t *testing.T) {
	type fields struct {
		base *SCNode
	}
	type args struct {
		query *syscontracts.NodeInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &scNodeWrapper{
				base: tt.fields.base,
			}
			got, err := n.getNodes(tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("getNodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getNodes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scNodeWrapper_isPublicKeyExist(t *testing.T) {
	type fields struct {
		base *SCNode
	}
	type args struct {
		pub string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &scNodeWrapper{
				base: tt.fields.base,
			}
			got, err := n.isPublicKeyExist(tt.args.pub)
			if (err != nil) != tt.wantErr {
				t.Errorf("isPublicKeyExist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("isPublicKeyExist() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scNodeWrapper_nodesNum(t *testing.T) {
	type fields struct {
		base *SCNode
	}
	type args struct {
		query *syscontracts.NodeInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &scNodeWrapper{
				base: tt.fields.base,
			}
			got, err := n.nodesNum(tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("nodesNum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("nodesNum() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scNodeWrapper_update(t *testing.T) {
	type fields struct {
		base *SCNode
	}
	type args struct {
		name string
		node *syscontracts.UpdateNode
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &scNodeWrapper{
				base: tt.fields.base,
			}
			got, err := n.update(tt.args.name, tt.args.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("update() got = %v, want %v", got, tt.want)
			}
		})
	}
}