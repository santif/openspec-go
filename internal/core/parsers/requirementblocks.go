package parsers

import (
	"regexp"
	"strings"

	"github.com/santif/openspec-go/internal/core/schemas"
)

type RequirementBlock struct {
	HeaderLine string
	Name       string
	Raw        string
}

type DeltaPlan struct {
	Added           []RequirementBlock
	Modified        []RequirementBlock
	Removed         []string // requirement names
	Renamed         []schemas.Rename
	SectionPresence struct {
		Added    bool
		Modified bool
		Removed  bool
		Renamed  bool
	}
}

func NormalizeRequirementName(name string) string {
	return strings.TrimSpace(name)
}

var requirementHeaderRegex = regexp.MustCompile(`^###\s*Requirement:\s*(.+)\s*$`)

func ParseDeltaSpec(content string) DeltaPlan {
	normalized := NormalizeContent(content)
	sections := splitTopLevelSections(normalized)

	addedLookup := getSectionCaseInsensitive(sections, "ADDED Requirements")
	modifiedLookup := getSectionCaseInsensitive(sections, "MODIFIED Requirements")
	removedLookup := getSectionCaseInsensitive(sections, "REMOVED Requirements")
	renamedLookup := getSectionCaseInsensitive(sections, "RENAMED Requirements")

	added := parseRequirementBlocksFromSection(addedLookup.body)
	modified := parseRequirementBlocksFromSection(modifiedLookup.body)
	removedNames := parseRemovedNames(removedLookup.body)
	renamedPairs := parseRenamedPairs(renamedLookup.body)

	plan := DeltaPlan{
		Added:    added,
		Modified: modified,
		Removed:  removedNames,
		Renamed:  renamedPairs,
	}
	plan.SectionPresence.Added = addedLookup.found
	plan.SectionPresence.Modified = modifiedLookup.found
	plan.SectionPresence.Removed = removedLookup.found
	plan.SectionPresence.Renamed = renamedLookup.found

	return plan
}

type sectionLookup struct {
	body  string
	found bool
}

func splitTopLevelSections(content string) map[string]string {
	lines := strings.Split(content, "\n")
	result := make(map[string]string)

	type sectionIndex struct {
		title string
		index int
	}
	var indices []sectionIndex

	headerRe := regexp.MustCompile(`^(##)\s+(.+)$`)
	for i, line := range lines {
		if m := headerRe.FindStringSubmatch(line); m != nil {
			indices = append(indices, sectionIndex{
				title: strings.TrimSpace(m[2]),
				index: i,
			})
		}
	}

	for i, si := range indices {
		var endIdx int
		if i+1 < len(indices) {
			endIdx = indices[i+1].index
		} else {
			endIdx = len(lines)
		}
		body := strings.Join(lines[si.index+1:endIdx], "\n")
		result[si.title] = body
	}

	return result
}

func getSectionCaseInsensitive(sections map[string]string, desired string) sectionLookup {
	target := strings.ToLower(desired)
	for title, body := range sections {
		if strings.ToLower(title) == target {
			return sectionLookup{body: body, found: true}
		}
	}
	return sectionLookup{body: "", found: false}
}

func parseRequirementBlocksFromSection(sectionBody string) []RequirementBlock {
	if sectionBody == "" {
		return nil
	}
	lines := strings.Split(NormalizeContent(sectionBody), "\n")
	var blocks []RequirementBlock
	i := 0

	for i < len(lines) {
		// Seek next requirement header
		for i < len(lines) && !requirementHeaderRegex.MatchString(lines[i]) {
			i++
		}
		if i >= len(lines) {
			break
		}

		headerLine := lines[i]
		m := requirementHeaderRegex.FindStringSubmatch(headerLine)
		if m == nil {
			i++
			continue
		}
		name := NormalizeRequirementName(m[1])
		buf := []string{headerLine}
		i++

		h2Re := regexp.MustCompile(`^##\s+`)
		for i < len(lines) && !requirementHeaderRegex.MatchString(lines[i]) {
			// Also stop at ## level headers (but not ### headers)
			if h2Re.MatchString(lines[i]) {
				break
			}
			buf = append(buf, lines[i])
			i++
		}

		raw := strings.TrimRight(strings.Join(buf, "\n"), " \t\n")
		blocks = append(blocks, RequirementBlock{
			HeaderLine: headerLine,
			Name:       name,
			Raw:        raw,
		})
	}

	return blocks
}

func parseRemovedNames(sectionBody string) []string {
	if sectionBody == "" {
		return nil
	}
	var names []string
	lines := strings.Split(NormalizeContent(sectionBody), "\n")
	bulletRe := regexp.MustCompile("^\\s*-\\s*`?###\\s*Requirement:\\s*(.+?)`?\\s*$")

	for _, line := range lines {
		if m := requirementHeaderRegex.FindStringSubmatch(line); m != nil {
			names = append(names, NormalizeRequirementName(m[1]))
			continue
		}
		if m := bulletRe.FindStringSubmatch(line); m != nil {
			names = append(names, NormalizeRequirementName(m[1]))
		}
	}

	return names
}

func parseRenamedPairs(sectionBody string) []schemas.Rename {
	if sectionBody == "" {
		return nil
	}
	var pairs []schemas.Rename
	lines := strings.Split(NormalizeContent(sectionBody), "\n")

	fromRe := regexp.MustCompile("^\\s*-?\\s*FROM:\\s*`?###\\s*Requirement:\\s*(.+?)`?\\s*$")
	toRe := regexp.MustCompile("^\\s*-?\\s*TO:\\s*`?###\\s*Requirement:\\s*(.+?)`?\\s*$")

	var current struct {
		from string
		to   string
	}

	for _, line := range lines {
		if m := fromRe.FindStringSubmatch(line); m != nil {
			current.from = NormalizeRequirementName(m[1])
		} else if m := toRe.FindStringSubmatch(line); m != nil {
			current.to = NormalizeRequirementName(m[1])
			if current.from != "" && current.to != "" {
				pairs = append(pairs, schemas.Rename{
					From: current.from,
					To:   current.to,
				})
				current.from = ""
				current.to = ""
			}
		}
	}

	return pairs
}
