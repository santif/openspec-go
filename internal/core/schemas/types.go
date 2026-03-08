package schemas

type DeltaOperation string

const (
	DeltaAdded    DeltaOperation = "ADDED"
	DeltaModified DeltaOperation = "MODIFIED"
	DeltaRemoved  DeltaOperation = "REMOVED"
	DeltaRenamed  DeltaOperation = "RENAMED"
)

type Scenario struct {
	RawText string `json:"rawText"`
}

type Requirement struct {
	Text      string     `json:"text"`
	Scenarios []Scenario `json:"scenarios"`
}

type Metadata struct {
	Version    string `json:"version"`
	Format     string `json:"format"`
	SourcePath string `json:"sourcePath,omitempty"`
}

type Spec struct {
	Name         string        `json:"name"`
	Overview     string        `json:"overview"`
	Requirements []Requirement `json:"requirements"`
	Metadata     *Metadata     `json:"metadata,omitempty"`
}

type Rename struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Delta struct {
	Spec         string         `json:"spec"`
	Operation    DeltaOperation `json:"operation"`
	Description  string         `json:"description"`
	Requirement  *Requirement   `json:"requirement,omitempty"`
	Requirements []Requirement  `json:"requirements,omitempty"`
	Rename       *Rename        `json:"rename,omitempty"`
}

type Change struct {
	Name        string    `json:"name"`
	Why         string    `json:"why"`
	WhatChanges string    `json:"whatChanges"`
	Deltas      []Delta   `json:"deltas"`
	Metadata    *Metadata `json:"metadata,omitempty"`
}
