package skills

import (
	"context"
	"errors"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

type Hub struct {
	skills map[string]*Metadata
	mu     sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		skills: make(map[string]*Metadata),
	}
}

func (h *Hub) Discover(workspaceRoot string) error {
	skills := make(map[string]*Metadata)

	//Discover workspace skills (overrides builtin)
	if workspaceRoot != "" {
		wsSkills, err := h.discoverPath(workspaceRoot+"/skills", "workspace")
		if err != nil && !os.IsNotExist(err) && !errors.Is(err, syscall.ENOENT) {
			return fmt.Errorf("failed to discover workspace skills: %w", err)
		}
		for name, meta := range wsSkills {
			if _, exists := skills[name]; exists {
				g.Log().Debugf(context.Background(), "Workspace skill '%s' overrides builtin", name)
			}
			skills[name] = &meta
		}
		g.Log().Debugf(context.Background(), "Discovered %d workspace skills", len(wsSkills))
	}

	h.mu.Lock()
	h.skills = skills
	h.mu.Unlock()

	g.Log().Infof(context.Background(), "Skills registry initialized: %d skills total", len(skills))
	return nil
}
func (h *Hub) discoverPath(root, source string) (map[string]Metadata, error) {
	skills := make(map[string]Metadata)

	if root == "" {
		return skills, nil
	}

	// Expand ~ to user home directory
	if strings.HasPrefix(root, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return skills, fmt.Errorf("failed to get user home directory: %w", err)
		}
		root = filepath.Join(home, root[1:])
	}

	// Check if root directory exists before walking
	if _, err := os.Stat(root); os.IsNotExist(err) || errors.Is(err, syscall.ENOENT) {
		g.Log().Debugf(context.Background(), "Skills directory does not exist: %s", root)
		return skills, nil
	}

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Skip entries that are not accessible (e.g., deleted during walk)
			if os.IsNotExist(err) || errors.Is(err, syscall.ENOENT) {
				g.Log().Debugf(context.Background(), "Path no longer exists during walk: %s", path)
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(d.Name(), "SKILL.md") {
			meta, parseErr := parseSkillMetadata(path)
			if parseErr != nil {
				g.Log().Debugf(context.Background(), "Failed to parse skill at %s: %v", path, parseErr)
				return nil // Continue discovering other skills
			}
			if meta.Name == "" {
				meta.Name = filepath.Base(filepath.Dir(path))
			}
			meta.Source = source
			skills[meta.Name] = meta
		}
		return nil
	})

	return skills, err
}

func (h *Hub) BuildSummary(workspaceRoot string) string {
	err := h.Discover(workspaceRoot)
	if err != nil {
		return ""
	} // Get all skills including unavailable

	var lines []string
	lines = append(lines, "<skills>")

	for _, skill := range h.skills {
		available, missing := CheckRequirements(skill.Requires)
		availStr := "true"
		if !available {
			availStr = "false"
		}

		lines = append(lines, fmt.Sprintf("  <skill available=\"%s\">", availStr))
		lines = append(lines, fmt.Sprintf("    <name>%s</name>", xmlEscape(skill.Name)))
		lines = append(lines, fmt.Sprintf("    <description>%s</description>", xmlEscape(skill.Description)))
		lines = append(lines, fmt.Sprintf("    <location>%s</location>", xmlEscape(skill.Path)))

		if !available && len(missing) > 0 {
			lines = append(lines, fmt.Sprintf("    <requires>%s</requires>", xmlEscape(FormatMissingRequirements(missing))))
		}

		if skill.Category != "" {
			lines = append(lines, fmt.Sprintf("    <category>%s</category>", xmlEscape(skill.Category)))
		}

		lines = append(lines, "  </skill>")
	}

	lines = append(lines, "</skills>")
	return strings.Join(lines, "\n")
}
func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// FormatMissingRequirements formats missing requirements as a string.
func FormatMissingRequirements(missing []string) string {
	if len(missing) == 0 {
		return ""
	}
	return strings.Join(missing, ", ")
}

// List returns all discovered skill metadata.
// If filterUnavailable is true, only returns skills with met requirements.
func (h *Hub) List(filterUnavailable bool) []*Metadata {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]*Metadata, 0, len(h.skills))
	for _, meta := range h.skills {
		if filterUnavailable {
			available, _ := CheckRequirements(meta.Requires)
			if !available {
				continue
			}
		}
		result = append(result, meta)
	}
	return result
}

func (h *Hub) GetAlwaysSkills() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var result []string
	for name, meta := range h.skills {
		if meta.Always {
			available, _ := CheckRequirements(meta.Requires)
			if available {
				result = append(result, name)
			}
		}
	}
	return result
}

// CheckRequirements checks if skill requirements are met.
// Returns (available, missing_requirements).
func CheckRequirements(req Requirements) (bool, []string) {
	var missing []string

	// Check binary dependencies
	for _, bin := range req.Bins {
		if _, err := exec.LookPath(bin); err != nil {
			missing = append(missing, fmt.Sprintf("CLI: %s", bin))
		}
	}

	// Check environment variables
	for _, env := range req.Env {
		if os.Getenv(env) == "" {
			missing = append(missing, fmt.Sprintf("ENV: %s", env))
		}
	}

	return len(missing) == 0, missing
}
