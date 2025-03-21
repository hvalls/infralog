package config

import "testing"

func TestFilter_MatchesResourceType(t *testing.T) {
	tests := []struct {
		name         string
		filter       Filter
		resourceType string
		want         bool
	}{
		{
			name:         "nil resource types should match any resource",
			filter:       Filter{ResourceTypes: nil},
			resourceType: "aws_instance",
			want:         true,
		},
		{
			name:         "empty resource types should match no resource",
			filter:       Filter{ResourceTypes: []string{}},
			resourceType: "aws_instance",
			want:         false,
		},
		{
			name:         "should match when resource type is in the list",
			filter:       Filter{ResourceTypes: []string{"aws_instance", "aws_vpc"}},
			resourceType: "aws_instance",
			want:         true,
		},
		{
			name:         "should not match when resource type is not in the list",
			filter:       Filter{ResourceTypes: []string{"aws_vpc", "aws_subnet"}},
			resourceType: "aws_instance",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filter.MatchesResourceType(tt.resourceType); got != tt.want {
				t.Errorf("Filter.MatchesResourceType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilter_MatchesOutput(t *testing.T) {
	tests := []struct {
		name   string
		filter Filter
		output string
		want   bool
	}{
		{
			name:   "nil outputs should match any output",
			filter: Filter{Outputs: nil},
			output: "instance_ip",
			want:   true,
		},
		{
			name:   "empty outputs should match no output",
			filter: Filter{Outputs: []string{}},
			output: "instance_ip",
			want:   false,
		},
		{
			name:   "should match when output is in the list",
			filter: Filter{Outputs: []string{"instance_ip", "vpc_id"}},
			output: "instance_ip",
			want:   true,
		},
		{
			name:   "should not match when output is not in the list",
			filter: Filter{Outputs: []string{"vpc_id", "subnet_id"}},
			output: "instance_ip",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filter.MatchesOutput(tt.output); got != tt.want {
				t.Errorf("Filter.MatchesOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}
