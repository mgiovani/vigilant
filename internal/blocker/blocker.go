// Package blocker provides regex-based blocklist matching for window titles and processes.
// It determines whether a window should be blocked based on configuration rules and
// manages focus state transitions between working and distracted states.
//
// All matching is regex-based and case-insensitive:
// - Block patterns: regex patterns matching window title OR process name
// - Exception patterns: regex patterns that bypass blocking when matched
//
// The StateManager tracks focus state transitions and emits state change events
// to coordinate with other application components.
package blocker

import (
	"log"
	"regexp"
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

// BlocklistMatcher implements the Blocker interface with unified regex matching
type BlocklistMatcher struct {
	patterns   []*regexp.Regexp // Compiled block patterns (case-insensitive)
	exceptions []*regexp.Regexp // Compiled exception patterns (case-insensitive)
	mu         sync.RWMutex     // Thread safety for patterns
}

// NewBlocklistMatcher creates a new BlocklistMatcher from configuration
// It pre-compiles all regex patterns with case-insensitive flag
// Returns error only for critical issues; logs warnings for invalid patterns
func NewBlocklistMatcher(blocklistCfg config.BlocklistConfig, exceptions []string) (*BlocklistMatcher, error) {
	bm := &BlocklistMatcher{
		patterns:   make([]*regexp.Regexp, 0, len(blocklistCfg.Patterns)),
		exceptions: make([]*regexp.Regexp, 0, len(exceptions)),
	}

	// Compile block patterns with case-insensitive flag
	for _, pattern := range blocklistCfg.Patterns {
		compiled := compilePattern(pattern)
		if compiled != nil {
			bm.patterns = append(bm.patterns, compiled)
		}
	}

	// Compile exception patterns with case-insensitive flag
	for _, pattern := range exceptions {
		compiled := compilePattern(pattern)
		if compiled != nil {
			bm.exceptions = append(bm.exceptions, compiled)
		}
	}

	return bm, nil
}

// compilePattern compiles a regex pattern with case-insensitive flag
// Returns nil if pattern is invalid (logs warning)
func compilePattern(pattern string) *regexp.Regexp {
	// Prepend (?i) flag for case-insensitive matching if not already present
	patternStr := pattern
	if len(pattern) < 2 || pattern[0] != '(' || pattern[1] != '?' {
		patternStr = "(?i)" + pattern
	}
	compiled, err := regexp.Compile(patternStr)
	if err != nil {
		log.Printf("Warning: Invalid regex pattern '%s': %v", pattern, err)
		return nil
	}
	return compiled
}

// IsBlocked checks if a window is blocked by blocklist rules
// It checks in order of priority:
// 1. Exception patterns (return false if any matches)
// 2. Block patterns (return true if any matches title OR process)
func (bm *BlocklistMatcher) IsBlocked(window *monitor.WindowInfo) bool {
	if window == nil {
		return false
	}

	// Check exceptions first (on both title and process)
	if bm.matchesAnyException(window) {
		return false
	}

	// Check block patterns (on both title and process)
	return bm.matchesAnyPattern(window)
}

// matchesAnyException checks if window matches any exception pattern
func (bm *BlocklistMatcher) matchesAnyException(window *monitor.WindowInfo) bool {
	bm.mu.RLock()
	exceptions := bm.exceptions
	bm.mu.RUnlock()

	for _, pattern := range exceptions {
		if matchPattern(pattern, window.Title) || matchPattern(pattern, window.Process) {
			return true
		}
	}
	return false
}

// matchesAnyPattern checks if window matches any block pattern
func (bm *BlocklistMatcher) matchesAnyPattern(window *monitor.WindowInfo) bool {
	bm.mu.RLock()
	patterns := bm.patterns
	bm.mu.RUnlock()

	for _, pattern := range patterns {
		if matchPattern(pattern, window.Title) || matchPattern(pattern, window.Process) {
			return true
		}
	}
	return false
}

// matchPattern safely matches a regex pattern against text
func matchPattern(pattern *regexp.Regexp, text string) bool {
	if pattern == nil || text == "" {
		return false
	}

	matched := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Warning: Regex pattern panic: %v", r)
			}
		}()
		matched = pattern.MatchString(text)
	}()
	return matched
}

// GetBlockReason returns which rule caused blocking (for debugging/logging)
// Returns empty string if not blocked
func (bm *BlocklistMatcher) GetBlockReason(window *monitor.WindowInfo) string {
	if window == nil {
		return ""
	}

	// Check exceptions first
	if bm.matchesAnyException(window) {
		return ""
	}

	// Find matching pattern
	bm.mu.RLock()
	patterns := bm.patterns
	bm.mu.RUnlock()

	for _, pattern := range patterns {
		if matchPattern(pattern, window.Title) {
			return "Blocked by pattern (" + pattern.String() + ") matching title"
		}
		if matchPattern(pattern, window.Process) {
			return "Blocked by pattern (" + pattern.String() + ") matching process"
		}
	}

	return ""
}
