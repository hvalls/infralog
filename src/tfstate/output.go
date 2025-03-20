package tfstate

type Output struct {
	Value     any  `json:"value"`
	Type      any  `json:"type,omitempty"`
	Sensitive bool `json:"sensitive,omitempty"`
}
