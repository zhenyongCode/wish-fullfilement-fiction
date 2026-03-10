package skills

// Requirements describes skill dependencies.
type Requirements struct {
	Bins []string // Required CLI binaries (e.g., "gh", "tmux")
	Env  []string // Required environment variables (e.g., "GITHUB_TOKEN")
}

// Metadata describes a skill entry.
type Metadata struct {
	Name        string       // Skill name
	Description string       // Short description
	Path        string       // File path to SKILL.md
	Source      string       // "builtin" or "workspace"
	Always      bool         // Auto-load in system prompt
	Category    string       // Optional category
	Requires    Requirements // Dependencies
	Frontmatter map[string]string
}

// Skill represents a loaded skill document.
type Skill struct {
	Metadata Metadata
	Content  string // Full markdown content (without frontmatter)
}

// StripFrontmatter removes YAML frontmatter from content.
func StripFrontmatter(content string) string {
	lines := splitLines(content)
	if len(lines) > 2 && lines[0] == "---" {
		for i := 1; i < len(lines); i++ {
			if lines[i] == "---" {
				return joinLines(lines[i+1:])
			}
		}
	}
	return content
}
