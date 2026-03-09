package parsers

import (
	"regexp"
	"strings"
)

type RequirementsSectionParts struct {
	Before     string             // Content before ## Requirements
	HeaderLine string             // The "## Requirements" line itself
	Preamble   string             // Content between header and first ### Requirement:
	BodyBlocks []RequirementBlock // Parsed requirement blocks in order
	After      string             // Content after the Requirements section
}

var reqSectionHeaderRegex = regexp.MustCompile(`(?i)^##\s+Requirements\s*$`)

func ExtractRequirementsSection(content string) RequirementsSectionParts {
	normalized := NormalizeContent(content)
	lines := strings.Split(normalized, "\n")

	// Find ## Requirements header
	reqHeaderIndex := -1
	for i, line := range lines {
		if reqSectionHeaderRegex.MatchString(line) {
			reqHeaderIndex = i
			break
		}
	}

	if reqHeaderIndex == -1 {
		before := strings.TrimRight(content, " \t\n\r")
		beforeStr := ""
		if before != "" {
			beforeStr = before + "\n\n"
		}
		return RequirementsSectionParts{
			Before:     beforeStr,
			HeaderLine: "## Requirements",
			Preamble:   "",
			BodyBlocks: nil,
			After:      "\n",
		}
	}

	// Find end of Requirements section: next ## header
	endIndex := len(lines)
	h2Re := regexp.MustCompile(`^##\s+`)
	for i := reqHeaderIndex + 1; i < len(lines); i++ {
		if h2Re.MatchString(lines[i]) {
			endIndex = i
			break
		}
	}

	before := strings.Join(lines[:reqHeaderIndex], "\n")
	headerLine := lines[reqHeaderIndex]
	sectionBodyLines := lines[reqHeaderIndex+1 : endIndex]

	// Parse preamble and requirement blocks
	cursor := 0
	var preambleLines []string

	reqBlockRe := regexp.MustCompile(`^###\s+Requirement:`)
	for cursor < len(sectionBodyLines) && !reqBlockRe.MatchString(sectionBodyLines[cursor]) {
		preambleLines = append(preambleLines, sectionBodyLines[cursor])
		cursor++
	}

	var blocks []RequirementBlock
	for cursor < len(sectionBodyLines) {
		headerLineCandidate := sectionBodyLines[cursor]
		m := RequirementHeaderRegex.FindStringSubmatch(headerLineCandidate)
		if m == nil {
			cursor++
			continue
		}
		name := NormalizeRequirementName(m[1])
		buf := []string{headerLineCandidate}
		cursor++
		for cursor < len(sectionBodyLines) && !reqBlockRe.MatchString(sectionBodyLines[cursor]) && !h2Re.MatchString(sectionBodyLines[cursor]) {
			buf = append(buf, sectionBodyLines[cursor])
			cursor++
		}
		raw := strings.TrimRight(strings.Join(buf, "\n"), " \t\n")
		blocks = append(blocks, RequirementBlock{
			HeaderLine: headerLineCandidate,
			Name:       name,
			Raw:        raw,
		})
	}

	after := strings.Join(lines[endIndex:], "\n")

	beforeTrimmed := strings.TrimRight(before, " \t\n")
	beforeStr := before
	if beforeTrimmed != "" {
		beforeStr = beforeTrimmed + "\n"
	}

	preamble := strings.TrimRight(strings.Join(preambleLines, "\n"), " \t\n")

	if !strings.HasPrefix(after, "\n") {
		after = "\n" + after
	}

	return RequirementsSectionParts{
		Before:     beforeStr,
		HeaderLine: headerLine,
		Preamble:   preamble,
		BodyBlocks: blocks,
		After:      after,
	}
}
