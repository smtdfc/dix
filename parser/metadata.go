package parser

type Metadata struct {
	Compositions []*Composition `json:"compositions"`
	Root         *Composition   `json:"root"`
}
