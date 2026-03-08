package parser

type Metadata struct {
	Providers []*Provider `json:"compositions"`
	Root      *Provider   `json:"root"`
}
