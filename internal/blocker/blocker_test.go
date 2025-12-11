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
		Processes: []string{"Discord.exe", "Discord"},
		Websites:  []string{"reddit.com", "youtube.com"},
		Patterns:  []string{".*Netflix.*", ".*Prime Video.*"},
	}
	exceptions := []string{"YouTube Music"}

	bm, err := NewBlocklistMatcher(blocklistCfg, exceptions)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if bm == nil {
		t.Fatal("Expected non-nil BlocklistMatcher")
	}

	// Verify processes are stored (lowercase)
	if len(bm.processes) != 2 {
		t.Errorf("Expected 2 processes, got %d", len(bm.processes))
	}

	// Verify websites are stored (lowercase)
	if len(bm.websites) != 2 {
		t.Errorf("Expected 2 websites, got %d", len(bm.websites))
	}

	// Verify patterns are compiled
	if len(bm.patterns) != 2 {
		t.Errorf("Expected 2 compiled patterns, got %d", len(bm.patterns))
	}

	// Verify exceptions are stored
	if len(bm.exceptions) != 1 {
		t.Errorf("Expected 1 exception, got %d", len(bm.exceptions))
	}
}

// TestNewBlocklistMatcher_InvalidRegex tests handling of invalid regex patterns
func TestNewBlocklistMatcher_InvalidRegex(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Processes: []string{},
		Websites:  []string{},
		Patterns:  []string{"[invalid", ".*valid.*"},
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

// TestIsBlocked_ProcessMatchExact tests process name matching (exact)
func TestIsBlocked_ProcessMatchExact(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Processes: []string{"Discord", "Slack.exe"},
		Websites:  []string{},
		Patterns:  []string{},
	}

	bm, _ := NewBlocklistMatcher(blocklistCfg, []string{})

	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "Discord",
		Process:   "Discord.exe",
		Timestamp: time.Now(),
	}

	if !bm.IsBlocked(window) {
		t.Error("Expected Discord.exe to be blocked")
	}
}

// TestIsBlocked_ProcessMatchCase tests case-insensitive process matching
func TestIsBlocked_ProcessMatchCase(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Processes: []string{"discord"},
		Websites:  []string{},
		Patterns:  []string{},
	}

	bm, _ := NewBlocklistMatcher(blocklistCfg, []string{})

	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "Discord",
		Process:   "DISCORD.EXE",
		Timestamp: time.Now(),
	}

	if !bm.IsBlocked(window) {
		t.Error("Expected DISCORD.EXE to match lowercase discord")
	}
}

// TestIsBlocked_WebsiteMatch tests website substring matching
func TestIsBlocked_WebsiteMatch(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Processes: []string{},
		Websites:  []string{"reddit.com"},
		Patterns:  []string{},
	}

	bm, _ := NewBlocklistMatcher(blocklistCfg, []string{})

	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "r/golang - reddit.com - Google Chrome",
		Process:   "chrome.exe",
		Timestamp: time.Now(),
	}

	if !bm.IsBlocked(window) {
		t.Error("Expected window with reddit.com in title to be blocked")
	}
}

// TestIsBlocked_WebsiteMatchCase tests case-insensitive website matching
func TestIsBlocked_WebsiteMatchCase(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Processes: []string{},
		Websites:  []string{"REDDIT.COM"},
		Patterns:  []string{},
	}

	bm, _ := NewBlocklistMatcher(blocklistCfg, []string{})

	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "r/golang - Reddit.com - Google Chrome",
		Process:   "chrome.exe",
		Timestamp: time.Now(),
	}

	if !bm.IsBlocked(window) {
		t.Error("Expected case-insensitive website match")
	}
}

// TestIsBlocked_PatternMatch tests regex pattern matching
func TestIsBlocked_PatternMatch(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Processes: []string{},
		Websites:  []string{},
		Patterns:  []string{".*Netflix.*"},
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

// TestIsBlocked_ExceptionBypassesProcess tests that exceptions bypass process rules
func TestIsBlocked_ExceptionBypassesProcess(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Processes: []string{"Chrome"},
		Websites:  []string{},
		Patterns:  []string{},
	}
	exceptions := []string{"Work - YouTube"}

	bm, _ := NewBlocklistMatcher(blocklistCfg, exceptions)

	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "Work - YouTube in Chrome",
		Process:   "Chrome.exe",
		Timestamp: time.Now(),
	}

	if bm.IsBlocked(window) {
		t.Error("Expected exception to bypass process block")
	}
}

// TestIsBlocked_ExceptionBypassesWebsite tests that exceptions bypass website rules
func TestIsBlocked_ExceptionBypassesWebsite(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Processes: []string{},
		Websites:  []string{"youtube.com"},
		Patterns:  []string{},
	}
	exceptions := []string{"YouTube Music"}

	bm, _ := NewBlocklistMatcher(blocklistCfg, exceptions)

	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "YouTube Music - Google Chrome",
		Process:   "chrome.exe",
		Timestamp: time.Now(),
	}

	if bm.IsBlocked(window) {
		t.Error("Expected YouTube Music exception to bypass youtube.com website block")
	}
}

// TestIsBlocked_ExceptionBypassesPattern tests that exceptions bypass pattern rules
func TestIsBlocked_ExceptionBypassesPattern(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Processes: []string{},
		Websites:  []string{},
		Patterns:  []string{".*Netflix.*"},
	}
	exceptions := []string{"Work Netflix"}

	bm, _ := NewBlocklistMatcher(blocklistCfg, exceptions)

	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "Work Netflix - Learning Content",
		Process:   "chrome.exe",
		Timestamp: time.Now(),
	}

	if bm.IsBlocked(window) {
		t.Error("Expected exception to bypass pattern block")
	}
}

// TestIsBlocked_NotBlocked tests that non-blocked windows are allowed
func TestIsBlocked_NotBlocked(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Processes: []string{"Discord"},
		Websites:  []string{"reddit.com"},
		Patterns:  []string{".*Netflix.*"},
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
		Processes: []string{},
		Websites:  []string{},
		Patterns:  []string{},
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
		Processes: []string{"Discord"},
		Websites:  []string{},
		Patterns:  []string{},
	}

	bm, _ := NewBlocklistMatcher(blocklistCfg, []string{})

	if bm.IsBlocked(nil) {
		t.Error("Expected nil window to not be blocked")
	}
}

// TestIsBlocked_WindowsPathHandling tests Windows path handling
func TestIsBlocked_WindowsPathHandling(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Processes: []string{"Discord"},
		Websites:  []string{},
		Patterns:  []string{},
	}

	bm, _ := NewBlocklistMatcher(blocklistCfg, []string{})

	// Note: The monitor.WindowInfo.Process should be just the executable name,
	// but we test that our blocker handles it correctly
	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "Discord",
		Process:   "discord.exe",
		Timestamp: time.Now(),
	}

	if !bm.IsBlocked(window) {
		t.Error("Expected discord.exe from Windows path to be blocked")
	}
}

// TestIsBlocked_MultipleMatches tests that first matching rule triggers block
func TestIsBlocked_MultipleMatches(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Processes: []string{"Discord"},
		Websites:  []string{"discord.com"},
		Patterns:  []string{".*Discord.*"},
	}

	bm, _ := NewBlocklistMatcher(blocklistCfg, []string{})

	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "Discord - discord.com",
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
		Processes: []string{"Discord"},
		Websites:  []string{},
		Patterns:  []string{},
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

// TestGetBlockReason_ProcessBlock tests GetBlockReason for process block
func TestGetBlockReason_ProcessBlock(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Processes: []string{"Discord.exe"},
		Websites:  []string{},
		Patterns:  []string{},
	}

	bm, _ := NewBlocklistMatcher(blocklistCfg, []string{})

	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "Discord",
		Process:   "Discord.exe",
		Timestamp: time.Now(),
	}

	reason := bm.GetBlockReason(window)
	if reason == "" {
		t.Error("Expected non-empty reason for process block")
	}
	if !contains(reason, "process rule") {
		t.Errorf("Expected 'process rule' in reason, got: %s", reason)
	}
}

// TestGetBlockReason_WebsiteBlock tests GetBlockReason for website block
func TestGetBlockReason_WebsiteBlock(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Processes: []string{},
		Websites:  []string{"reddit.com"},
		Patterns:  []string{},
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
		t.Error("Expected non-empty reason for website block")
	}
	if !contains(reason, "website rule") {
		t.Errorf("Expected 'website rule' in reason, got: %s", reason)
	}
}

// TestGetBlockReason_PatternBlock tests GetBlockReason for pattern block
func TestGetBlockReason_PatternBlock(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Processes: []string{},
		Websites:  []string{},
		Patterns:  []string{".*Netflix.*"},
	}

	bm, _ := NewBlocklistMatcher(blocklistCfg, []string{})

	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "Netflix - Stranger Things",
		Process:   "chrome.exe",
		Timestamp: time.Now(),
	}

	reason := bm.GetBlockReason(window)
	if reason == "" {
		t.Error("Expected non-empty reason for pattern block")
	}
	if !contains(reason, "pattern rule") {
		t.Errorf("Expected 'pattern rule' in reason, got: %s", reason)
	}
}

// TestGetBlockReason_NilWindow tests GetBlockReason with nil window
func TestGetBlockReason_NilWindow(t *testing.T) {
	blocklistCfg := config.BlocklistConfig{
		Processes: []string{"Discord"},
		Websites:  []string{},
		Patterns:  []string{},
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
		Processes: []string{"Discord"},
		Websites:  []string{},
		Patterns:  []string{},
	}
	exceptions := []string{"Work Discord"}

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
		Processes: []string{"Discord", "Slack", "Steam"},
		Websites:  []string{"reddit.com", "youtube.com", "twitter.com"},
		Patterns:  []string{".*Netflix.*", ".*Prime Video.*"},
	}
	exceptions := []string{"YouTube Music"}

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

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
