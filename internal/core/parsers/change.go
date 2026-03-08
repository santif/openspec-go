package parsers

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/santif/openspec-go/internal/core/schemas"
)

func ParseChangeWithDeltas(name, content, changeDir string) (*schemas.Change, error) {
	sections := ParseSections(content)
	whySection := FindSection(sections, "Why")
	whatChangesSection := FindSection(sections, "What Changes")

	if whySection == nil || whySection.Content == "" {
		return nil, fmt.Errorf("Change must have a Why section")
	}
	if whatChangesSection == nil || whatChangesSection.Content == "" {
		return nil, fmt.Errorf("Change must have a What Changes section")
	}

	simpleDeltas := parseDeltas(whatChangesSection.Content)

	specsDir := filepath.Join(changeDir, "specs")
	deltaDeltas := parseDeltaSpecs(specsDir)

	deltas := simpleDeltas
	if len(deltaDeltas) > 0 {
		deltas = deltaDeltas
	}

	return &schemas.Change{
		Name:        name,
		Why:         strings.TrimSpace(whySection.Content),
		WhatChanges: strings.TrimSpace(whatChangesSection.Content),
		Deltas:      deltas,
		Metadata: &schemas.Metadata{
			Version: "1.0.0",
			Format:  "openspec-change",
		},
	}, nil
}

func parseDeltaSpecs(specsDir string) []schemas.Delta {
	var deltas []schemas.Delta

	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		specName := entry.Name()
		specFile := filepath.Join(specsDir, specName, "spec.md")

		content, err := os.ReadFile(specFile)
		if err != nil {
			continue
		}

		specDeltas := parseSpecDeltas(specName, string(content))
		deltas = append(deltas, specDeltas...)
	}

	return deltas
}

func parseSpecDeltas(specName, content string) []schemas.Delta {
	var deltas []schemas.Delta
	sections := ParseSections(content)

	// ADDED
	if addedSection := FindSection(sections, "ADDED Requirements"); addedSection != nil {
		for _, req := range parseRequirements(addedSection) {
			deltas = append(deltas, schemas.Delta{
				Spec:         specName,
				Operation:    schemas.DeltaAdded,
				Description:  "Add requirement: " + req.Text,
				Requirement:  &req,
				Requirements: []schemas.Requirement{req},
			})
		}
	}

	// MODIFIED
	if modifiedSection := FindSection(sections, "MODIFIED Requirements"); modifiedSection != nil {
		for _, req := range parseRequirements(modifiedSection) {
			deltas = append(deltas, schemas.Delta{
				Spec:         specName,
				Operation:    schemas.DeltaModified,
				Description:  "Modify requirement: " + req.Text,
				Requirement:  &req,
				Requirements: []schemas.Requirement{req},
			})
		}
	}

	// REMOVED
	if removedSection := FindSection(sections, "REMOVED Requirements"); removedSection != nil {
		for _, req := range parseRequirements(removedSection) {
			deltas = append(deltas, schemas.Delta{
				Spec:         specName,
				Operation:    schemas.DeltaRemoved,
				Description:  "Remove requirement: " + req.Text,
				Requirement:  &req,
				Requirements: []schemas.Requirement{req},
			})
		}
	}

	// RENAMED
	if renamedSection := FindSection(sections, "RENAMED Requirements"); renamedSection != nil {
		renames := parseRenames(renamedSection.Content)
		for _, rename := range renames {
			r := rename // capture
			deltas = append(deltas, schemas.Delta{
				Spec:        specName,
				Operation:   schemas.DeltaRenamed,
				Description: fmt.Sprintf("Rename requirement from %q to %q", rename.From, rename.To),
				Rename:      &r,
			})
		}
	}

	return deltas
}

func parseRenames(content string) []schemas.Rename {
	var renames []schemas.Rename
	lines := strings.Split(NormalizeContent(content), "\n")

	var current struct {
		from string
		to   string
	}

	fromRe := regexp.MustCompile(`^\s*-?\s*FROM:\s*` + "`?" + `###\s*Requirement:\s*(.+?)` + "`?" + `\s*$`)
	toRe := regexp.MustCompile(`^\s*-?\s*TO:\s*` + "`?" + `###\s*Requirement:\s*(.+?)` + "`?" + `\s*$`)

	for _, line := range lines {
		if m := fromRe.FindStringSubmatch(line); m != nil {
			current.from = strings.TrimSpace(m[1])
		} else if m := toRe.FindStringSubmatch(line); m != nil {
			current.to = strings.TrimSpace(m[1])
			if current.from != "" && current.to != "" {
				renames = append(renames, schemas.Rename{
					From: current.from,
					To:   current.to,
				})
				current.from = ""
				current.to = ""
			}
		}
	}

	return renames
}
