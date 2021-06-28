package directives

type ResourceAttributes struct {
	Resource *string `json:"resource"`
	Role     *string `json:"role"`
	Mode     *string `json:"mode"`
}