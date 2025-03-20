package tfstate

import "testing"

func TestResourceID(t *testing.T) {
	tests := []struct {
		name     string
		resource Resource
		want     ResourceID
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
			if got := tt.resource.GetID(); got != tt.want {
				t.Errorf("ResourceID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSplitResourceID(t *testing.T) {
	tests := []struct {
		name string
		id   ResourceID
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
			got := tt.id.Split()
			if got.module != tt.want.module || got.resourceType != tt.want.resourceType || got.resourceName != tt.want.resourceName {
				t.Errorf("splitResourceID() = %v, want %v", got, tt.want)
			}
		})
	}
}
