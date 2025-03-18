package tfstate

import (
	"reflect"
	"testing"
)

func TestResourceID(t *testing.T) {
	tests := []struct {
		name     string
		resource Resource
		want     string
	}{
		{
			name: "simple resource",
			resource: Resource{
				Type: "aws_instance",
				Name: "web",
			},
			want: "aws_instance.web",
		},
		{
			name: "resource with module",
			resource: Resource{
				Module: "module.network",
				Type:   "aws_vpc",
				Name:   "main",
			},
			want: "module.network.aws_vpc.main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResourceID(tt.resource); got != tt.want {
				t.Errorf("ResourceID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSplitResourceID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want ResourceIDParts
	}{
		{
			name: "simple resource",
			id:   "aws_instance.web",
			want: ResourceIDParts{
				resourceType: "aws_instance",
				resourceName: "web",
			},
		},
		{
			name: "resource with module",
			id:   "module.network.aws_vpc.main",
			want: ResourceIDParts{
				module:       "module.network",
				resourceType: "aws_vpc",
				resourceName: "main",
			},
		},
		{
			name: "resource with nested module",
			id:   "module.network.module.subnets.aws_subnet.public",
			want: ResourceIDParts{
				module:       "module.network.module.subnets",
				resourceType: "aws_subnet",
				resourceName: "public",
			},
		},
		{
			name: "invalid format",
			id:   "invalid",
			want: ResourceIDParts{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitResourceID(tt.id)
			if got.module != tt.want.module || got.resourceType != tt.want.resourceType || got.resourceName != tt.want.resourceName {
				t.Errorf("splitResourceID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareInstances(t *testing.T) {
	tests := []struct {
		name         string
		oldInstances []ResourceInstance
		newInstances []ResourceInstance
		want         map[string]ValueDiff
	}{
		{
			name: "identical instances",
			oldInstances: []ResourceInstance{
				{
					Attributes: map[string]interface{}{
						"id":   "i-123456",
						"type": "t2.micro",
					},
				},
			},
			newInstances: []ResourceInstance{
				{
					Attributes: map[string]interface{}{
						"id":   "i-123456",
						"type": "t2.micro",
					},
				},
			},
			want: map[string]ValueDiff{},
		},
		{
			name: "changed attribute",
			oldInstances: []ResourceInstance{
				{
					Attributes: map[string]interface{}{
						"id":   "i-123456",
						"type": "t2.micro",
					},
				},
			},
			newInstances: []ResourceInstance{
				{
					Attributes: map[string]interface{}{
						"id":   "i-123456",
						"type": "t2.small",
					},
				},
			},
			want: map[string]ValueDiff{
				"type": {
					OldValue: "t2.micro",
					NewValue: "t2.small",
				},
			},
		},
		{
			name: "added attribute",
			oldInstances: []ResourceInstance{
				{
					Attributes: map[string]interface{}{
						"id": "i-123456",
					},
				},
			},
			newInstances: []ResourceInstance{
				{
					Attributes: map[string]interface{}{
						"id":   "i-123456",
						"type": "t2.micro",
					},
				},
			},
			want: map[string]ValueDiff{
				"type": {
					NewValue: "t2.micro",
				},
			},
		},
		{
			name: "removed attribute",
			oldInstances: []ResourceInstance{
				{
					Attributes: map[string]interface{}{
						"id":   "i-123456",
						"type": "t2.micro",
					},
				},
			},
			newInstances: []ResourceInstance{
				{
					Attributes: map[string]interface{}{
						"id": "i-123456",
					},
				},
			},
			want: map[string]ValueDiff{
				"type": {
					OldValue: "t2.micro",
				},
			},
		},
		{
			name:         "no instances",
			oldInstances: []ResourceInstance{},
			newInstances: []ResourceInstance{},
			want:         map[string]ValueDiff{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareInstances(tt.oldInstances, tt.newInstances)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("compareInstances() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareOutputs(t *testing.T) {
	tests := []struct {
		name       string
		oldOutputs map[string]Output
		newOutputs map[string]Output
		want       []ResourceDiff
	}{
		{
			name: "identical outputs",
			oldOutputs: map[string]Output{
				"vpc_id": {
					Value: "vpc-123456",
				},
			},
			newOutputs: map[string]Output{
				"vpc_id": {
					Value: "vpc-123456",
				},
			},
			want: []ResourceDiff{},
		},
		{
			name: "changed output",
			oldOutputs: map[string]Output{
				"vpc_id": {
					Value: "vpc-123456",
				},
			},
			newOutputs: map[string]Output{
				"vpc_id": {
					Value: "vpc-654321",
				},
			},
			want: []ResourceDiff{
				{
					ResourceName: "vpc_id",
					ResourceType: "output",
					Status:       DiffStatusChanged,
					AttributeDiffs: map[string]ValueDiff{
						"value": {
							OldValue: "vpc-123456",
							NewValue: "vpc-654321",
						},
					},
				},
			},
		},
		{
			name:       "added output",
			oldOutputs: map[string]Output{},
			newOutputs: map[string]Output{
				"vpc_id": {
					Value: "vpc-123456",
				},
			},
			want: []ResourceDiff{
				{
					ResourceName:   "vpc_id",
					ResourceType:   "output",
					Status:         DiffStatusAdded,
					AttributeDiffs: map[string]ValueDiff{},
				},
			},
		},
		{
			name: "removed output",
			oldOutputs: map[string]Output{
				"vpc_id": {
					Value: "vpc-123456",
				},
			},
			newOutputs: map[string]Output{},
			want: []ResourceDiff{
				{
					ResourceName:   "vpc_id",
					ResourceType:   "output",
					Status:         DiffStatusRemoved,
					AttributeDiffs: map[string]ValueDiff{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareOutputs(tt.oldOutputs, tt.newOutputs)
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("compareOutputs() = %v, want %v", got, tt.want)
			}
		})
	}
}
