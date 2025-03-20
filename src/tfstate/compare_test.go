package tfstate

import (
	"reflect"
	"testing"
)

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
		want       []OutputDiff
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
			want: []OutputDiff{},
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
			want: []OutputDiff{
				{
					OutputName: "vpc_id",
					Status:     DiffStatusChanged,
					ValueDiff: ValueDiff{
						OldValue: "vpc-123456",
						NewValue: "vpc-654321",
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
			want: []OutputDiff{
				{
					OutputName: "vpc_id",
					Status:     DiffStatusAdded,
					ValueDiff:  ValueDiff{},
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
			want: []OutputDiff{
				{
					OutputName: "vpc_id",
					Status:     DiffStatusRemoved,
					ValueDiff:  ValueDiff{},
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
