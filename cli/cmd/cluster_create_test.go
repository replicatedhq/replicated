package cmd

import (
	"testing"

	"github.com/replicatedhq/replicated/pkg/kotsclient"
)

func Test_parseNodeGroups(t *testing.T) {
	type args struct {
		nodeGroups []string
	}
	tests := []struct {
		name    string
		args    args
		want    []kotsclient.NodeGroup
		wantErr bool
	}{
		{
			name: "valid node group with name, disk and instance type",
			args: args{
				nodeGroups: []string{
					"name=ng1,instance-type=t2.medium,nodes=3,disk=20",
				},
			},
			want: []kotsclient.NodeGroup{
				{
					Name:         "ng1",
					InstanceType: "t2.medium",
					Nodes:        3,
					Disk:         20,
				},
			},
			wantErr: false,
		},
		{
			name: "valid node group with name and instance type",
			args: args{
				nodeGroups: []string{
					"name=ng1,instance-type=t2.medium,nodes=3",
				},
			},
			want: []kotsclient.NodeGroup{
				{
					Name:         "ng1",
					InstanceType: "t2.medium",
					Nodes:        3,
				},
			},
			wantErr: false,
		},
		{
			name: "valid node group with name",
			args: args{
				nodeGroups: []string{
					"name=ng1,nodes=3",
				},
			},
			want: []kotsclient.NodeGroup{
				{
					Name:  "ng1",
					Nodes: 3,
				},
			},
			wantErr: false,
		},
		{
			name: "valid node group",
			args: args{
				nodeGroups: []string{
					"nodes=3",
				},
			},
			want: []kotsclient.NodeGroup{
				{
					Nodes: 3,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid node group field",
			args: args{
				nodeGroups: []string{
					"name=ng1,instance-type=t2.medium,nodes=3,disk=20,invalid=invalid",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid node group value (nodes)",
			args: args{
				nodeGroups: []string{
					"name=ng1,instance-type=t2.medium,nodes=invalid,disk=20",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid node group value (disk)",
			args: args{
				nodeGroups: []string{
					"name=ng1,instance-type=t2.medium,nodes=3,disk=invalid",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid node group format",
			args: args{
				nodeGroups: []string{
					"invalid",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		got, err := parseNodeGroups(tt.args.nodeGroups)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. parseNodeGroups() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if len(got) != len(tt.want) {
			t.Errorf("%q. parseNodeGroups() got = %v, want %v", tt.name, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("%q. parseNodeGroups() got = %v, want %v", tt.name, got, tt.want)
			}
		}
	}

}
