package blocker

import (
	"strings"
	"testing"
	"time"

	"vigilant/internal/config"
	"vigilant/internal/monitor"
)

// TestNewBlocklistMatcher tests the constructor
func TestNewBlocklistMatcher(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Patterns: []string{"discord", "reddit", "netflix"},
	}
	exceptions := []string{"youtube music"}

	bm, err := NewBlocklistMatcher(blocklistCfg, exceptions)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if bm == nil {
		t.Fatal("Expected non-nil BlocklistMatcher")
	}

	// Verify patterns are compiled
	if len(bm.patterns) != 3 {
		t.Errorf("Expected 3 compiled patterns, got %d", len(bm.patterns))
	}

	// Verify exceptions are compiled
	if len(bm.exceptions) != 1 {
		t.Errorf("Expected 1 exception, got %d", len(bm.exceptions))
	}
}

// TestNewBlocklistMatcher_InvalidRegex tests handling of invalid regex patterns
func TestNewBlocklistMatcher_InvalidRegex(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Patterns: []string{"[invalid", "valid"},
	}

	bm, err := NewBlocklistMatcher(blocklistCfg, []string{})

	if err != nil {
		t.Fatalf("Expected no error for invalid pattern, got %v", err)
	}

	// Should have only 1 valid pattern compiled (invalid one skipped with warning)
	if len(bm.patterns) != 1 {
		t.Errorf("Expected 1 valid pattern, got %d", len(bm.patterns))
	}
}

// TestIsBlocked_ProcessMatch tests process name matching via regex
func TestIsBlocked_ProcessMatch(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Patterns: []string{"discord", "slack"},
	}

	bm, _ := NewBlocklistMatcher(blocklistCfg, []string{})

	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "Some Window",
		Process:   "Discord.exe",
		Timestamp: time.Now(),
	}

	if !bm.IsBlocked(window) {
		t.Error("Expected Discord.exe to be blocked by 'discord' pattern")
	}
}

// TestIsBlocked_ProcessMatchCase tests case-insensitive process matching
func TestIsBlocked_ProcessMatchCase(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Patterns: []string{"discord"},
	}

	bm, _ := NewBlocklistMatcher(blocklistCfg, []string{})

	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "Some Window",
		Process:   "DISCORD.EXE",
		Timestamp: time.Now(),
	}

	if !bm.IsBlocked(window) {
		t.Error("Expected DISCORD.EXE to match lowercase discord pattern")
	}
}

// TestIsBlocked_TitleMatch tests window title matching
func TestIsBlocked_TitleMatch(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Patterns: []string{"reddit"},
	}

	bm, _ := NewBlocklistMatcher(blocklistCfg, []string{})

	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "r/golang - reddit.com - Google Chrome",
		Process:   "chrome.exe",
		Timestamp: time.Now(),
	}

	if !bm.IsBlocked(window) {
		t.Error("Expected window with reddit in title to be blocked")
	}
}

// TestIsBlocked_TitleMatchCase tests case-insensitive title matching
func TestIsBlocked_TitleMatchCase(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Patterns: []string{"REDDIT"},
	}

	bm, _ := NewBlocklistMatcher(blocklistCfg, []string{})

	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "r/golang - Reddit.com - Google Chrome",
		Process:   "chrome.exe",
		Timestamp: time.Now(),
	}

	if !bm.IsBlocked(window) {
		t.Error("Expected case-insensitive title match")
	}
}

// TestIsBlocked_PatternMatch tests complex regex pattern matching
func TestIsBlocked_PatternMatch(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Patterns: []string{"netflix"},
	}

	bm, _ := NewBlocklistMatcher(blocklistCfg, []string{})

	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "Watching Netflix - Google Chrome",
		Process:   "chrome.exe",
		Timestamp: time.Now(),
	}

	if !bm.IsBlocked(window) {
		t.Error("Expected window matching Netflix pattern to be blocked")
	}
}

// TestIsBlocked_ExceptionBypassesTitle tests that exceptions bypass title matching
func TestIsBlocked_ExceptionBypassesTitle(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Patterns: []string{"youtube"},
	}
	exceptions := []string{"youtube music"}

	bm, _ := NewBlocklistMatcher(blocklistCfg, exceptions)

	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "YouTube Music - Google Chrome",
		Process:   "chrome.exe",
		Timestamp: time.Now(),
	}

	if bm.IsBlocked(window) {
		t.Error("Expected YouTube Music exception to bypass youtube block")
	}
}

// TestIsBlocked_ExceptionBypassesProcess tests that exceptions can match process names
func TestIsBlocked_ExceptionBypassesProcess(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Patterns: []string{"chrome"},
	}
	exceptions := []string{"work chrome"} // Won't match process

	bm, _ := NewBlocklistMatcher(blocklistCfg, exceptions)

	// This should still be blocked because exception matches title, not process
	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "Some Page - Google Chrome",
		Process:   "chrome.exe",
		Timestamp: time.Now(),
	}

	if !bm.IsBlocked(window) {
		t.Error("Expected chrome.exe to be blocked (exception doesn't match)")
	}

	// This should NOT be blocked because title matches exception
	window2 := &monitor.WindowInfo{
		PID:       1234,
		Title:     "Work Chrome Tasks",
		Process:   "chrome.exe",
		Timestamp: time.Now(),
	}

	if bm.IsBlocked(window2) {
		t.Error("Expected 'Work Chrome' title to bypass block via exception")
	}
}

// TestIsBlocked_NotBlocked tests that non-blocked windows are allowed
func TestIsBlocked_NotBlocked(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Patterns: []string{"discord", "reddit", "netflix"},
	}

	bm, _ := NewBlocklistMatcher(blocklistCfg, []string{})

	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "VSCode - my-project",
		Process:   "code.exe",
		Timestamp: time.Now(),
	}

	if bm.IsBlocked(window) {
		t.Error("Expected VSCode window to not be blocked")
	}
}

// TestIsBlocked_EmptyBlocklist tests that empty blocklist never blocks
func TestIsBlocked_EmptyBlocklist(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Patterns: []string{},
	}

	bm, _ := NewBlocklistMatcher(blocklistCfg, []string{})

	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "Discord",
		Process:   "Discord.exe",
		Timestamp: time.Now(),
	}

	if bm.IsBlocked(window) {
		t.Error("Expected empty blocklist to never block")
	}
}

// TestIsBlocked_NilWindow tests handling of nil window
func TestIsBlocked_NilWindow(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Patterns: []string{"discord"},
	}

	bm, _ := NewBlocklistMatcher(blocklistCfg, []string{})

	if bm.IsBlocked(nil) {
		t.Error("Expected nil window to not be blocked")
	}
}

// TestIsBlocked_SpecialRegexChars tests patterns with special regex characters
func TestIsBlocked_SpecialRegexChars(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Patterns: []string{"disney\\+", "battle\\.net"},
	}

	bm, _ := NewBlocklistMatcher(blocklistCfg, []string{})

	tests := []struct {
		title   string
		blocked bool
	}{
		{"Disney+ - Watching", true},
		{"Battle.net Launcher", true},
		{"Disneyplus - Browser", false}, // Should NOT match disney+
		{"Battlexnet", false},           // Should NOT match battle.net
	}

	for _, tc := range tests {
		window := &monitor.WindowInfo{
			PID:       1234,
			Title:     tc.title,
			Process:   "test.exe",
			Timestamp: time.Now(),
		}

		if bm.IsBlocked(window) != tc.blocked {
			t.Errorf("Title '%s': expected blocked=%v, got %v", tc.title, tc.blocked, !tc.blocked)
		}
	}
}

// TestIsBlocked_WordBoundaries tests patterns with word boundaries
func TestIsBlocked_WordBoundaries(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Patterns: []string{"\\bx\\.com\\b"},
	}

	bm, _ := NewBlocklistMatcher(blocklistCfg, []string{})

	tests := []struct {
		title   string
		blocked bool
	}{
		{"Home / X.com - Browser", true},
		{"x.com - Twitter", true},
		{"prefix.com", false},   // Should NOT match
		{"example.com", false},  // Should NOT match
		{"foxnews.com", false},  // Should NOT match (x.com inside word)
	}

	for _, tc := range tests {
		window := &monitor.WindowInfo{
			PID:       1234,
			Title:     tc.title,
			Process:   "browser.exe",
			Timestamp: time.Now(),
		}

		if bm.IsBlocked(window) != tc.blocked {
			t.Errorf("Title '%s': expected blocked=%v, got %v", tc.title, tc.blocked, !tc.blocked)
		}
	}
}

// TestIsBlocked_MultipleMatches tests that first matching rule triggers block
func TestIsBlocked_MultipleMatches(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Patterns: []string{"discord", "chat"},
	}

	bm, _ := NewBlocklistMatcher(blocklistCfg, []string{})

	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "Discord - chat channel",
		Process:   "Discord.exe",
		Timestamp: time.Now(),
	}

	if !bm.IsBlocked(window) {
		t.Error("Expected window to be blocked by at least one rule")
	}
}

// TestGetBlockReason_NoBlock tests that GetBlockReason returns empty for allowed windows
func TestGetBlockReason_NoBlock(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Patterns: []string{"discord"},
	}

	bm, _ := NewBlocklistMatcher(blocklistCfg, []string{})

	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "VSCode",
		Process:   "code.exe",
		Timestamp: time.Now(),
	}

	reason := bm.GetBlockReason(window)
	if reason != "" {
		t.Errorf("Expected empty reason for unblocked window, got: %s", reason)
	}
}

// TestGetBlockReason_TitleBlock tests GetBlockReason for title-based block
func TestGetBlockReason_TitleBlock(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Patterns: []string{"reddit"},
	}

	bm, _ := NewBlocklistMatcher(blocklistCfg, []string{})

	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "r/golang - reddit.com - Front Page",
		Process:   "chrome.exe",
		Timestamp: time.Now(),
	}

	reason := bm.GetBlockReason(window)
	if reason == "" {
		t.Error("Expected non-empty reason for title block")
	}
	if !contains(reason, "pattern") || !contains(reason, "title") {
		t.Errorf("Expected 'pattern' and 'title' in reason, got: %s", reason)
	}
}

// TestGetBlockReason_ProcessBlock tests GetBlockReason for process-based block
func TestGetBlockReason_ProcessBlock(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Patterns: []string{"discord"},
	}

	bm, _ := NewBlocklistMatcher(blocklistCfg, []string{})

	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "Random Window Title",
		Process:   "Discord.exe",
		Timestamp: time.Now(),
	}

	reason := bm.GetBlockReason(window)
	if reason == "" {
		t.Error("Expected non-empty reason for process block")
	}
	if !contains(reason, "pattern") || !contains(reason, "process") {
		t.Errorf("Expected 'pattern' and 'process' in reason, got: %s", reason)
	}
}

// TestGetBlockReason_NilWindow tests GetBlockReason with nil window
func TestGetBlockReason_NilWindow(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Patterns: []string{"discord"},
	}

	bm, _ := NewBlocklistMatcher(blocklistCfg, []string{})

	reason := bm.GetBlockReason(nil)
	if reason != "" {
		t.Errorf("Expected empty reason for nil window, got: %s", reason)
	}
}

// TestGetBlockReason_ExceptionOverride tests that GetBlockReason respects exceptions
func TestGetBlockReason_ExceptionOverride(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Patterns: []string{"discord"},
	}
	exceptions := []string{"work discord"}

	bm, _ := NewBlocklistMatcher(blocklistCfg, exceptions)

	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "Work Discord Channel",
		Process:   "Discord.exe",
		Timestamp: time.Now(),
	}

	reason := bm.GetBlockReason(window)
	if reason != "" {
		t.Errorf("Expected empty reason due to exception, got: %s", reason)
	}
}

// BenchmarkIsBlocked benchmarks the IsBlocked method
func BenchmarkIsBlocked(b *testing.B) {
	blocklistCfg := config.BlocklistConfig{
		Patterns: []string{"discord", "slack", "steam", "reddit", "youtube", "twitter", "netflix", "prime video"},
	}
	exceptions := []string{"youtube music"}

	bm, _ := NewBlocklistMatcher(blocklistCfg, exceptions)

	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "VSCode - my-project",
		Process:   "code.exe",
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bm.IsBlocked(window)
	}
}

// BenchmarkIsBlocked_Match benchmarks IsBlocked when there's a match
func BenchmarkIsBlocked_Match(b *testing.B) {
	blocklistCfg := config.BlocklistConfig{
		Patterns: []string{"discord", "slack", "steam", "reddit", "youtube", "twitter", "netflix", "prime video"},
	}
	exceptions := []string{"youtube music"}

	bm, _ := NewBlocklistMatcher(blocklistCfg, exceptions)

	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "Watching Netflix",
		Process:   "chrome.exe",
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bm.IsBlocked(window)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
