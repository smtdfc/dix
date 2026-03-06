package parser

type Dependency struct {
	Name        string    `json:"name"`
	Type        *TypeInfo `json:"type"`
	IsSingleton bool      `json:"is_sng"`
}

type ReturnValue struct {
	Type *TypeInfo `json:"type"`
}
