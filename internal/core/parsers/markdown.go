package parsers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Fission-AI/openspec-go/internal/core/schemas"
)

type Section struct {
	Level    int
	Title    string
	Content  string
	Children []*Section
}

func NormalizeContent(content string) string {
	return strings.ReplaceAll(strings.ReplaceAll(content, "\r\n", "\n"), "\r", "\n")
}

func ParseSections(content string) []*Section {
	normalized := NormalizeContent(content)
	lines := strings.Split(normalized, "\n")
	var sections []*Section
	var stack []*Section

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		level, title := parseHeader(line)
		if level == 0 {
			continue
		}

		sectionContent := getContentUntilNextHeader(lines, i+1, level)
		section := &Section{
			Level:    level,
			Title:    title,
			Content:  sectionContent,
			Children: nil,
		}

		for len(stack) > 0 && stack[len(stack)-1].Level >= level {
			stack = stack[:len(stack)-1]
		}

		if len(stack) == 0 {
			sections = append(sections, section)
		} else {
			stack[len(stack)-1].Children = append(stack[len(stack)-1].Children, section)
		}
		stack = append(stack, section)
	}

	return sections
}

func parseHeader(line string) (int, string) {
	trimmed := line
	level := 0
	for i := 0; i < len(trimmed) && i < 6; i++ {
		if trimmed[i] == '#' {
			level++
		} else {
			break
		}
	}
	if level == 0 || level > 6 {
		return 0, ""
	}
	if len(trimmed) <= level || trimmed[level] != ' ' {
		return 0, ""
	}
	title := strings.TrimSpace(trimmed[level+1:])
	if title == "" {
		return 0, ""
	}
	return level, title
}

func getContentUntilNextHeader(lines []string, startLine, currentLevel int) string {
	var contentLines []string
	for i := startLine; i < len(lines); i++ {
		headerLevel, _ := parseHeader(lines[i])
		if headerLevel > 0 && headerLevel <= currentLevel {
			break
		}
		contentLines = append(contentLines, lines[i])
	}
	return strings.TrimSpace(strings.Join(contentLines, "\n"))
}

func FindSection(sections []*Section, title string) *Section {
	target := strings.ToLower(title)
	for _, s := range sections {
		if strings.ToLower(s.Title) == target {
			return s
		}
		if child := FindSection(s.Children, title); child != nil {
			return child
		}
	}
	return nil
}

func ParseSpec(name, content string) (*schemas.Spec, error) {
	sections := ParseSections(content)
	purposeSection := FindSection(sections, "Purpose")
	requirementsSection := FindSection(sections, "Requirements")

	if purposeSection == nil || purposeSection.Content == "" {
		return nil, fmt.Errorf("Spec must have a Purpose section")
	}
	if requirementsSection == nil {
		return nil, fmt.Errorf("Spec must have a Requirements section")
	}

	requirements := parseRequirements(requirementsSection)

	return &schemas.Spec{
		Name:         name,
		Overview:     strings.TrimSpace(purposeSection.Content),
		Requirements: requirements,
		Metadata: &schemas.Metadata{
			Version: "1.0.0",
			Format:  "openspec",
		},
	}, nil
}

func ParseChange(name, content string) (*schemas.Change, error) {
	sections := ParseSections(content)
	whySection := FindSection(sections, "Why")
	whatChangesSection := FindSection(sections, "What Changes")

	if whySection == nil || whySection.Content == "" {
		return nil, fmt.Errorf("Change must have a Why section")
	}
	if whatChangesSection == nil || whatChangesSection.Content == "" {
		return nil, fmt.Errorf("Change must have a What Changes section")
	}

	deltas := parseDeltas(whatChangesSection.Content)

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

func parseRequirements(section *Section) []schemas.Requirement {
	var requirements []schemas.Requirement

	for _, child := range section.Children {
		text := child.Title

		// Get content before any child sections (scenarios)
		if strings.TrimSpace(child.Content) != "" {
			lines := strings.Split(child.Content, "\n")
			var contentBeforeChildren []string
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "#") {
					break
				}
				contentBeforeChildren = append(contentBeforeChildren, line)
			}
			directContent := strings.TrimSpace(strings.Join(contentBeforeChildren, "\n"))
			if directContent != "" {
				for _, l := range strings.Split(directContent, "\n") {
					if strings.TrimSpace(l) != "" {
						text = strings.TrimSpace(l)
						break
					}
				}
			}
		}

		scenarioList := parseScenarios(child)

		requirements = append(requirements, schemas.Requirement{
			Text:      text,
			Scenarios: scenarioList,
		})
	}

	return requirements
}

func parseScenarios(requirementSection *Section) []schemas.Scenario {
	var scenarioList []schemas.Scenario
	for _, scenarioSection := range requirementSection.Children {
		if strings.TrimSpace(scenarioSection.Content) != "" {
			scenarioList = append(scenarioList, schemas.Scenario{
				RawText: scenarioSection.Content,
			})
		}
	}
	return scenarioList
}

var deltaRegex = regexp.MustCompile(`^\s*-\s*\*\*([^*:]+)(?::\*\*|\*\*:)\s*(.+)$`)

func parseDeltas(content string) []schemas.Delta {
	var deltas []schemas.Delta
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		matches := deltaRegex.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		specName := strings.TrimSpace(matches[1])
		description := strings.TrimSpace(matches[2])

		operation := schemas.DeltaModified
		lowerDesc := strings.ToLower(description)

		switch {
		case matchWord(lowerDesc, `\brename[sd]?\b`) || matchWord(lowerDesc, `\brenamed\s+(to|from)\b`) || matchWord(lowerDesc, `\brenaming\b`):
			operation = schemas.DeltaRenamed
		case matchWord(lowerDesc, `\badd[sed]?\b`) || matchWord(lowerDesc, `\bcreate[sd]?\b`) || matchWord(lowerDesc, `\bnew\b`) || matchWord(lowerDesc, `\badding\b`) || matchWord(lowerDesc, `\bcreating\b`):
			operation = schemas.DeltaAdded
		case matchWord(lowerDesc, `\bremove[sd]?\b`) || matchWord(lowerDesc, `\bdelete[sd]?\b`) || matchWord(lowerDesc, `\bremoving\b`) || matchWord(lowerDesc, `\bdeleting\b`):
			operation = schemas.DeltaRemoved
		}

		deltas = append(deltas, schemas.Delta{
			Spec:        specName,
			Operation:   operation,
			Description: description,
		})
	}

	return deltas
}

func matchWord(text, pattern string) bool {
	re := regexp.MustCompile(pattern)
	return re.MatchString(text)
}
