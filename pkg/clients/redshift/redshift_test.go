/*
Copyright 2020 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package redshift

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/redshift/v1alpha1"
)

var (
	dbName     = "example-name"
	nodeType   = "dc1.large"
	singleNode = "single-node"
)

func fullClusterParam() *v1alpha1.ClusterParameters {
	o := &v1alpha1.ClusterParameters{
		AllowVersionUpgrade:              aws.Bool(false),
		AutomatedSnapshotRetentionPeriod: aws.Int64(0),
		AvailabilityZone:                 aws.String(""),
		ClusterVersion:                   aws.String("1.0"),
		ClusterSubnetGroupName:           aws.String("default"),
		ClusterType:                      aws.String("single-node"),
		DBName:                           aws.String("dev"),
		Encrypted:                        aws.Bool(false),
		EnhancedVPCRouting:               aws.Bool(false),
		KMSKeyID:                         aws.String(""),
		MaintenanceTrackName:             aws.String(""),
		ManualSnapshotRetentionPeriod:    aws.Int64(0),
		NodeType:                         "dc1.large",
		NumberOfNodes:                    aws.Int64(1),
		PreferredMaintenanceWindow:       aws.String(""),
		Port:                             aws.Int64(5439),
		PubliclyAccessible:               aws.Bool(false),
		SnapshotScheduleIdentifier:       aws.String(""),
		VPCSecurityGroupIDs:              []string{"sg-44444444"},
		//DeleteParams
		FinalClusterSnapshotIdentifier:      aws.String("doom"),
		FinalClusterSnapshotRetentionPeriod: aws.Int64(1),
		SkipFinalClusterSnapshot:            aws.Bool(true),
	}
	return o
}

func clusterParam(m ...func(*v1alpha1.ClusterParameters)) *v1alpha1.ClusterParameters {
	o := &v1alpha1.ClusterParameters{
		MasterUsername:           "admin",
		NodeType:                 "dc1.large",
		ClusterType:              aws.String("single-node"),
		SkipFinalClusterSnapshot: aws.Bool(true),
	}
	for _, f := range m {
		f(o)
	}
	return o
}

func cluster(m ...func(*redshift.Cluster)) *redshift.Cluster {
	o := &redshift.Cluster{
		AllowVersionUpgrade:              aws.Bool(true),
		AutomatedSnapshotRetentionPeriod: aws.Int64(1),
		AvailabilityZone:                 aws.String("us-east-1d"),
		ClusterParameterGroups: []redshift.ClusterParameterGroupStatus{
			{
				ParameterGroupName: aws.String("default"),
			},
		},
		ClusterSubnetGroupName: aws.String("default"),
		ClusterVersion:         aws.String("1.0"),
		DBName:                 aws.String("dev"),
		Encrypted:              aws.Bool(false),
		Endpoint: &redshift.Endpoint{
			Port: aws.Int64(5439),
		},
		EnhancedVpcRouting:            aws.Bool(false),
		MaintenanceTrackName:          aws.String("current"),
		ManualSnapshotRetentionPeriod: aws.Int64(-1),
		NodeType:                      aws.String("dc1.large"),
		NumberOfNodes:                 aws.Int64(1),
		PubliclyAccessible:            aws.Bool(true),
		VpcSecurityGroups: []redshift.VpcSecurityGroupMembership{
			{
				VpcSecurityGroupId: aws.String("sg-44444444"),
			},
		},
	}
	for _, f := range m {
		f(o)
	}

	return o
}

func TestCreatePatch(t *testing.T) {
	type args struct {
		cl *redshift.Cluster
		p  *v1alpha1.ClusterParameters
	}

	type want struct {
		patch *v1alpha1.ClusterParameters
	}

	cases := map[string]struct {
		args
		want
	}{
		"SameFields": {
			args: args{
				cl: &redshift.Cluster{
					NodeType:          &nodeType,
					ClusterIdentifier: aws.String(""),
					NumberOfNodes:     aws.Int64(1),
				},
				p: &v1alpha1.ClusterParameters{
					NodeType:    nodeType,
					ClusterType: &singleNode,
				},
			},
			want: want{
				patch: &v1alpha1.ClusterParameters{},
			},
		},
		"DifferentFields": {
			args: args{
				cl: &redshift.Cluster{
					NodeType:          &nodeType,
					ClusterIdentifier: aws.String(""),
					NumberOfNodes:     aws.Int64(2),
				},
				p: &v1alpha1.ClusterParameters{
					NodeType:      nodeType,
					NumberOfNodes: aws.Int64(1),
				},
			},
			want: want{
				patch: &v1alpha1.ClusterParameters{
					NumberOfNodes: aws.Int64(1),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			result, _ := CreatePatch(tc.args.p, tc.args.cl)
			if diff := cmp.Diff(tc.want.patch, result); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	ClusterSubnetGroupName := "example-subnet"

	type args struct {
		cl redshift.Cluster
		p  v1alpha1.ClusterParameters
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				cl: redshift.Cluster{
					NodeType:          &nodeType,
					ClusterIdentifier: aws.String(""),
					NumberOfNodes:     aws.Int64(1),
				},
				p: v1alpha1.ClusterParameters{
					NodeType:    nodeType,
					ClusterType: &singleNode,
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				cl: redshift.Cluster{
					NodeType:          &nodeType,
					ClusterIdentifier: aws.String(""),
					NumberOfNodes:     aws.Int64(2),
				},
				p: v1alpha1.ClusterParameters{
					NodeType:    nodeType,
					ClusterType: &singleNode,
				},
			},
			want: false,
		},
		"IgnoresRefs": {
			args: args{
				cl: redshift.Cluster{
					DBName:                 &dbName,
					ClusterSubnetGroupName: &ClusterSubnetGroupName,
				},
				p: v1alpha1.ClusterParameters{
					DBName:                 &dbName,
					ClusterSubnetGroupName: &ClusterSubnetGroupName,
				},
			},
			want: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, _ := IsUpToDate(tc.args.p, tc.args.cl)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitialize(t *testing.T) {
	type args struct {
		in *v1alpha1.ClusterParameters
		cl redshift.Cluster
	}
	cases := map[string]struct {
		args args
		want *v1alpha1.ClusterParameters
	}{
		"MinimalSpec": {
			args: args{
				in: clusterParam(),
				cl: *cluster(),
			},
			want: &v1alpha1.ClusterParameters{
				AllowVersionUpgrade:              aws.Bool(true),
				AutomatedSnapshotRetentionPeriod: aws.Int64(1),
				AvailabilityZone:                 aws.String("us-east-1d"),
				ClusterSubnetGroupName:           aws.String("default"),
				ClusterType:                      aws.String("single-node"),
				ClusterVersion:                   aws.String("1.0"),
				DBName:                           aws.String("dev"),
				Encrypted:                        aws.Bool(false),
				EnhancedVPCRouting:               aws.Bool(false),
				MaintenanceTrackName:             aws.String("current"),
				ManualSnapshotRetentionPeriod:    aws.Int64(-1),
				MasterUsername:                   "admin",
				NodeType:                         "dc1.large",
				NumberOfNodes:                    aws.Int64(1),
				Port:                             aws.Int64(5439),
				PubliclyAccessible:               aws.Bool(true),
				SkipFinalClusterSnapshot:         aws.Bool(true),
				VPCSecurityGroupIDs:              []string{"sg-44444444"},
			},
		},
		"EmptyExternalResponse": {
			args: args{
				in: clusterParam(),
				cl: redshift.Cluster{},
			},
			want: clusterParam(),
		},
		"FullSpec": {
			args: args{
				in: fullClusterParam(),
				cl: *cluster(),
			},
			want: fullClusterParam(),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			names := name
			fmt.Print(names)
			LateInitialize(tc.args.in, &tc.args.cl)
			if diff := cmp.Diff(tc.args.in, tc.want); diff != "" {
				t.Errorf("LateInitializeSpec(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	tests := map[string]struct {
		err  error
		want bool
	}{
		"validError": {
			err:  awserr.New(redshift.ErrCodeClusterNotFoundFault, "ClusterNotFound", errors.New(redshift.ErrCodeClusterNotFoundFault)),
			want: true,
		},
		"invalidAwsError": {
			err:  awserr.New(redshift.ErrCodeClusterAlreadyExistsFault, "ClusterAlreadyExists", errors.New(redshift.ErrCodeClusterAlreadyExistsFault)),
			want: false,
		},
		"randomError": {
			err:  errors.New("the specified hosted zone does not exist"),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateCreateClusterInput(t *testing.T) {
	cases := map[string]struct {
		in  *v1alpha1.ClusterParameters
		out *redshift.CreateClusterInput
	}{
		"MinimalSpec": {
			in: clusterParam(),
			out: &redshift.CreateClusterInput{
				ClusterIdentifier:  aws.String("unit-test"),
				ClusterType:        aws.String("single-node"),
				MasterUsername:     aws.String("admin"),
				MasterUserPassword: aws.String("very-strong-password"),
				NodeType:           aws.String("dc1.large"),
			},
		},
		"FullSpec": {
			in: fullClusterParam(),
			out: &redshift.CreateClusterInput{
				AllowVersionUpgrade:              aws.Bool(false),
				AutomatedSnapshotRetentionPeriod: aws.Int64(0),
				AvailabilityZone:                 aws.String(""),
				ClusterIdentifier:                aws.String("unit-test"),
				ClusterVersion:                   aws.String("1.0"),
				ClusterSubnetGroupName:           aws.String("default"),
				ClusterType:                      aws.String("single-node"),
				DBName:                           aws.String("dev"),
				Encrypted:                        aws.Bool(false),
				EnhancedVpcRouting:               aws.Bool(false),
				KmsKeyId:                         aws.String(""),
				MaintenanceTrackName:             aws.String(""),
				ManualSnapshotRetentionPeriod:    aws.Int64(0),
				MasterUsername:                   aws.String(""),
				MasterUserPassword:               aws.String("very-strong-password"),
				NodeType:                         aws.String("dc1.large"),
				NumberOfNodes:                    aws.Int64(1),
				PreferredMaintenanceWindow:       aws.String(""),
				Port:                             aws.Int64(5439),
				PubliclyAccessible:               aws.Bool(false),
				SnapshotScheduleIdentifier:       aws.String(""),
				VpcSecurityGroupIds:              []string{"sg-44444444"},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateCreateClusterInput(tc.in, aws.String("unit-test"), aws.String("very-strong-password"))
			if diff := cmp.Diff(r, tc.out); diff != "" {
				t.Errorf("GenerateCreateClusterInput(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateModifyClusterInput(t *testing.T) {
	type args struct {
		in *v1alpha1.ClusterParameters
		cl redshift.Cluster
	}
	cases := map[string]struct {
		args args
		want *redshift.ModifyClusterInput
	}{
		"Resize": {
			args: args{
				in: &v1alpha1.ClusterParameters{
					ClusterType:   aws.String("multi-node"),
					NodeType:      "dc1.large",
					NumberOfNodes: aws.Int64(2),
				},
				cl: *cluster(),
			},
			want: &redshift.ModifyClusterInput{
				ClusterType:   aws.String("multi-node"),
				NodeType:      aws.String("dc1.large"),
				NumberOfNodes: aws.Int64(2),
			},
		},
		"PublicAccessibility": {
			args: args{
				in: &v1alpha1.ClusterParameters{
					ElasticIP:          aws.String(""),
					PubliclyAccessible: aws.Bool(false),
				},
				cl: *cluster(),
			},
			want: &redshift.ModifyClusterInput{
				ElasticIp:          aws.String(""),
				PubliclyAccessible: aws.Bool(false),
			},
		},
		"EnhancedVPCRouting": {
			args: args{
				in: &v1alpha1.ClusterParameters{
					EnhancedVPCRouting: aws.Bool(true),
				},
				cl: *cluster(),
			},
			want: &redshift.ModifyClusterInput{
				EnhancedVpcRouting: aws.Bool(true),
			},
		},
		"Rename": {
			args: args{
				in: &v1alpha1.ClusterParameters{
					NewClusterIdentifier: aws.String("YouAreinMatrixNeo"),
				},
				cl: *cluster(),
			},
			want: &redshift.ModifyClusterInput{
				NewClusterIdentifier: aws.String("YouAreinMatrixNeo"),
			},
		},
		"EverythingElse": {
			args: args{
				in: &v1alpha1.ClusterParameters{
					AllowVersionUpgrade:              aws.Bool(false),
					AutomatedSnapshotRetentionPeriod: aws.Int64(2),
					Encrypted:                        aws.Bool(true),
					MaintenanceTrackName:             aws.String("r2d2"),
					ManualSnapshotRetentionPeriod:    aws.Int64(42),
					VPCSecurityGroupIDs:              []string{"sg-666"},
				},
				cl: redshift.Cluster{},
			},
			want: &redshift.ModifyClusterInput{
				AllowVersionUpgrade:              aws.Bool(false),
				AutomatedSnapshotRetentionPeriod: aws.Int64(2),
				Encrypted:                        aws.Bool(true),
				MaintenanceTrackName:             aws.String("r2d2"),
				ManualSnapshotRetentionPeriod:    aws.Int64(42),
				VpcSecurityGroupIds:              []string{"sg-666"},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateModifyClusterInput(tc.args.in, tc.args.cl)
			if diff := cmp.Diff(r, tc.want); diff != "" {
				t.Errorf("GenerateModifyClusterInput(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateDeleteClusterInput(t *testing.T) {
	cases := map[string]struct {
		in  *v1alpha1.ClusterParameters
		out *redshift.DeleteClusterInput
	}{
		"MinimalSpec": {
			in: clusterParam(),
			out: &redshift.DeleteClusterInput{
				ClusterIdentifier:        aws.String("unit-test"),
				SkipFinalClusterSnapshot: aws.Bool(true),
			},
		},
		"FullSpec": {
			in: fullClusterParam(),
			out: &redshift.DeleteClusterInput{
				ClusterIdentifier:                   aws.String("unit-test"),
				FinalClusterSnapshotIdentifier:      aws.String("doom"),
				FinalClusterSnapshotRetentionPeriod: aws.Int64(1),
				SkipFinalClusterSnapshot:            aws.Bool(true),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateDeleteClusterInput(tc.in, aws.String("unit-test"))
			if diff := cmp.Diff(r, tc.out); diff != "" {
				t.Errorf("GenerateDeleteClusterInput(...): -want, +got:\n%s", diff)
			}
		})
	}
}
