package types

type EntitlementSpec struct {
	ID        string `json:"id,omitempty"`
	Spec      string `json:"spec,omitempty"`
	Name      string `json:"name,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
}

type EntitlementValueResponse struct {
	ID         string `json:"id,omitempty"`
	SpecID     string `json:"specId,omitempty"`
	CustomerID string `json:"customerId,omitempty"`
	Key        string `json:"key,omitempty"`
	Value      string `json:"value,omitempty"`
}

type Entitlement struct {
	IsDefault bool   `json:"isDefault,omitempty"`
	Name      string `json:"name,omitempty"`
	Value     string `json:"value,omitempty"`
}
