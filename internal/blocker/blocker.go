// Package blocker provides blocklist matching for processes, websites, and regex patterns.
// It determines whether a window should be blocked based on configuration rules and
// manages focus state transitions between working and distracted states.
//
// The BlocklistMatcher performs three types of matching:
// - Process matching: exact comparison of executable names (case-insensitive)
// - Website matching: substring search in window titles (case-insensitive)
// - Pattern matching: regex matching for complex rules
//
// The StateManager tracks focus state transitions and emits state change events
// to coordinate with other application components.
package blocker

import (
	"log"
	"regexp"
	"strings"
	"sync"

	"vigilant/internal/config"
	"vigilant/internal/monitor"
)

// Blocker interface defines the contract for matching logic against blocked items
type Blocker interface {
	// IsBlocked checks if a window is blocked by blocklist rules
	IsBlocked(window *monitor.WindowInfo) bool

	// GetBlockReason returns which rule matched (for debugging)
	GetBlockReason(window *monitor.WindowInfo) string
}

// BlocklistMatcher implements the Blocker interface with process, website, and pattern matching
type BlocklistMatcher struct {
	processes  []string         // Exact process name matches (lowercase)
	websites   []string         // Website substring matches (lowercase)
	patterns   []*regexp.Regexp // Compiled regex patterns
	exceptions []string         // Exceptions that bypass blocklist (lowercase)
	config     *config.Config
	mu         sync.RWMutex // Thread safety for patterns
}

// NewBlocklistMatcher creates a new BlocklistMatcher from configuration
// It pre-compiles regex patterns and stores process names in lowercase for case-insensitive matching
// Returns error only for critical issues; logs warnings for invalid patterns
func NewBlocklistMatcher(blocklistCfg config.BlocklistConfig, exceptions []string) (*BlocklistMatcher, error) {
	bm := &BlocklistMatcher{
		processes:  make([]string, 0, len(blocklistCfg.Processes)),
		websites:   make([]string, 0, len(blocklistCfg.Websites)),
		patterns:   make([]*regexp.Regexp, 0, len(blocklistCfg.Patterns)),
		exceptions: make([]string, 0, len(exceptions)),
	}

	// Store processes in lowercase for case-insensitive matching
	for _, proc := range blocklistCfg.Processes {
		bm.processes = append(bm.processes, strings.ToLower(proc))
	}

	// Store websites in lowercase for case-insensitive matching
	for _, website := range blocklistCfg.Websites {
		bm.websites = append(bm.websites, strings.ToLower(website))
	}

	// Store exceptions in lowercase for case-insensitive matching
	for _, exc := range exceptions {
		bm.exceptions = append(bm.exceptions, strings.ToLower(exc))
	}

	// Pre-compile all regex patterns with case-insensitive flag
	for _, pattern := range blocklistCfg.Patterns {
		// Prepend (?i) flag for case-insensitive matching if not already present
		patternStr := pattern
		if !strings.HasPrefix(pattern, "(?") {
			patternStr = "(?i)" + pattern
		}
		compiled, err := regexp.Compile(patternStr)
		if err != nil {
			log.Printf("Warning: Invalid regex pattern '%s': %v", pattern, err)
			continue
		}
		bm.patterns = append(bm.patterns, compiled)
	}

	return bm, nil
}

// IsBlocked checks if a window is blocked by blocklist rules
// It checks in order of priority:
// 1. Exceptions (fastest path - return false immediately)
// 2. Process name matching
// 3. Website matching
// 4. Regex pattern matching
func (bm *BlocklistMatcher) IsBlocked(window *monitor.WindowInfo) bool {
	if window == nil {
		return false
	}

	windowTitle := strings.ToLower(window.Title)

	// Exception checking FIRST (fastest path)
	for _, exc := range bm.exceptions {
		if strings.Contains(windowTitle, exc) {
			return false
		}
	}

	// Process name matching
	if bm.isProcessBlocked(window) {
		return true
	}

	// Website matching
	if bm.isWebsiteBlocked(windowTitle) {
		return true
	}

	// Regex pattern matching
	if bm.isPatternBlocked(windowTitle) {
		return true
	}

	return false
}

// isProcessBlocked checks if the process name matches any blocklist entry
func (bm *BlocklistMatcher) isProcessBlocked(window *monitor.WindowInfo) bool {
	// Extract process name and normalize it
	processName := strings.ToLower(window.Process)

	// Remove .exe suffix on Windows
	processName = strings.TrimSuffix(processName, ".exe")

	// Check against blocklist
	for _, blocked := range bm.processes {
		blockedName := strings.TrimSuffix(blocked, ".exe")
		if processName == blockedName {
			return true
		}
	}

	return false
}

// isWebsiteBlocked checks if any website domain appears in the window title
func (bm *BlocklistMatcher) isWebsiteBlocked(windowTitle string) bool {
	for _, website := range bm.websites {
		if strings.Contains(windowTitle, website) {
			return true
		}
	}
	return false
}

// isPatternBlocked checks if any regex pattern matches the window title
func (bm *BlocklistMatcher) isPatternBlocked(windowTitle string) bool {
	bm.mu.RLock()
	patterns := bm.patterns
	bm.mu.RUnlock()

	for _, pattern := range patterns {
		matched := false
		// Catch panic from regex matching if needed
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Warning: Regex pattern panic: %v", r)
				}
			}()

			matched = pattern.MatchString(windowTitle)
		}()

		if matched {
			return true
		}
	}

	return false
}

// GetBlockReason returns which rule caused blocking (for debugging/logging)
// Returns empty string if not blocked
func (bm *BlocklistMatcher) GetBlockReason(window *monitor.WindowInfo) string {
	if window == nil {
		return ""
	}

	windowTitle := strings.ToLower(window.Title)

	// Check exceptions first
	for _, exc := range bm.exceptions {
		if strings.Contains(windowTitle, exc) {
			return ""
		}
	}

	// Check process name
	processName := strings.ToLower(window.Process)
	processName = strings.TrimSuffix(processName, ".exe")

	for _, blocked := range bm.processes {
		blockedName := strings.TrimSuffix(blocked, ".exe")
		if processName == blockedName {
			return "Blocked by process rule (" + window.Process + ")"
		}
	}

	// Check websites
	for _, website := range bm.websites {
		if strings.Contains(windowTitle, website) {
			return "Blocked by website rule (" + website + ")"
		}
	}

	// Check patterns
	bm.mu.RLock()
	patterns := bm.patterns
	bm.mu.RUnlock()

	for _, pattern := range patterns {
		matched := false
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Warning: Regex pattern panic in GetBlockReason: %v", r)
				}
			}()

			matched = pattern.MatchString(windowTitle)
		}()

		if matched {
			return "Blocked by pattern rule"
		}
	}

	return ""
}
