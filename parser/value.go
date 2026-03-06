package parser

type Dependency struct {
	Name string    `json:"name"`
	Type *TypeInfo `json:"type"`
}

type ReturnValue struct {
	Type *TypeInfo `json:"type"`
}
