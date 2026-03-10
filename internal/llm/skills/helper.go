package skills

import (
	"fmt"
	"os"
	"strings"
)

func splitLines(s string) []string {
	// Split by \n and preserve structure
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}
	return result
}

func parseSkillMetadata(path string) (Metadata, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Metadata{}, fmt.Errorf("failed to read skill file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var name string
	var desc string
	var always bool
	var category string
	var requires Requirements
	frontmatter := map[string]string{}

	// Parse YAML frontmatter
	if len(lines) > 2 && strings.TrimSpace(lines[0]) == "---" {
		for i := 1; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			if line == "---" {
				lines = lines[i+1:]
				break
			}
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			value = strings.Trim(value, "\"'")
			if key != "" {
				frontmatter[key] = value
			}
		}
	}

	// Extract basic fields
	if frontmatter["name"] != "" {
		name = frontmatter["name"]
	}
	if frontmatter["description"] != "" {
		desc = frontmatter["description"]
	}
	if frontmatter["category"] != "" {
		category = frontmatter["category"]
	}

	// Parse always flag
	if val := frontmatter["always"]; val == "true" || val == "True" {
		always = true
	}

	// Parse requirements (simple format: requires_bins, requires_env)
	if bins := frontmatter["requires_bins"]; bins != "" {
		requires.Bins = parseCommaSeparated(bins)
	}
	if envs := frontmatter["requires_env"]; envs != "" {
		requires.Env = parseCommaSeparated(envs)
	}

	// Fallback: parse from markdown if name/desc not in frontmatter
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if name == "" && strings.HasPrefix(trimmed, "#") {
			name = strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
			continue
		}
		if name != "" && desc == "" {
			desc = trimmed
			break
		}
	}

	return Metadata{
		Name:        name,
		Description: desc,
		Path:        path,
		Always:      always,
		Category:    category,
		Requires:    requires,
		Frontmatter: frontmatter,
	}, nil
}

func parseCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
