# Contributing to Vigilant

Thank you for your interest in contributing to Vigilant! This document provides guidelines and information for contributors.

## Code of Conduct

Be respectful, inclusive, and constructive. We're all here to make productivity fun!

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/vigilant.git`
3. Set up the development environment (see [Onboarding Guide](onboarding.md))
4. Create a feature branch: `git checkout -b feature/my-feature`

## Development Workflow

### 1. Pick an Issue

- Check existing issues for something to work on
- For new features, open an issue first to discuss
- Comment on the issue to claim it

### 2. Make Changes

```bash
# Start development server
make dev

# Make your changes...

# Run tests
make test

# Run tests with race detection
go test ./... -race
```

### 3. Commit Guidelines

We follow [Conventional Commits](https://conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Code style (formatting, semicolons, etc.)
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `perf`: Performance improvement
- `test`: Adding or correcting tests
- `build`: Changes to build system or dependencies
- `ci`: CI configuration changes
- `chore`: Other changes that don't modify src or test files

**Examples:**
```
feat(blocker): add support for browser extension detection
fix(monitor): resolve race condition in window polling
docs(readme): update installation instructions
refactor(player): simplify state machine transitions
test(stats): add edge case tests for focus rate calculation
```

### 4. Submit Pull Request

1. Push your branch: `git push origin feature/my-feature`
2. Open a Pull Request against `main`
3. Fill out the PR template
4. Wait for review

## Pull Request Guidelines

### PR Title
Use the same format as commit messages:
```
feat(component): brief description
```

### PR Description
Include:
- What changes were made
- Why the changes were needed
- How to test the changes
- Screenshots (for UI changes)

### PR Checklist
- [ ] Tests pass locally (`make test`)
- [ ] No race conditions (`go test ./... -race`)
- [ ] Code follows existing style
- [ ] Documentation updated (if needed)
- [ ] Commit messages follow convention

## Code Style

### Go Code

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Use meaningful variable names
- Add comments for exported functions
- Keep functions small and focused

```go
// Good: Clear function with documentation
// IsBlocked checks if the given window matches any blocklist rules.
// It returns true if the window should trigger the FBI meme.
func (b *BlocklistMatcher) IsBlocked(window *WindowInfo) bool {
    // Implementation...
}

// Bad: Unclear, no documentation
func (b *BlocklistMatcher) Check(w *WindowInfo) bool {
    // ...
}
```

### Svelte/TypeScript Code

- Use TypeScript for type safety
- Follow Svelte style guide
- Use meaningful component names
- Keep components focused on one purpose
- Import stores from `stores/app`
- Use Lucide icons for UI elements

```svelte
<!-- Good: Clear component structure -->
<script>
  import { onMount } from 'svelte';
  import { stats, focusState } from '../../stores/app';
  import { Focus, AlertCircle } from 'lucide-svelte';

  let focusedTime = 0;
  let distractedTime = 0;

  onMount(() => {
    // Subscribe to stats updates
    const unsubscribe = stats.subscribe((s) => {
      if (s) {
        focusedTime = s.focusedTime;
        distractedTime = s.distractedTime;
      }
    });

    return () => {
      unsubscribe();
    };
  });

  function formatDuration(ms) {
    // Implementation...
  }
</script>

<div class="stats-display">
  <!-- Template... -->
</div>

<style>
  /* Scoped styles... */
</style>
```

### Wails Event Listeners

When listening to backend events in Svelte components, initialize event listeners in `stores/app.ts`:

```typescript
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { writable } from 'svelte/store';

export const stats = writable<Stats | null>(null);

export function initializeEventListeners() {
  // Listen for player state changes (lofi/fbi)
  EventsOn('player:state-change', (data: { state: string; timestamp: number }) => {
    console.log('[stores] player:state-change received:', data);
    playerState.set(data.state as PlayerState);
  });

  // Listen for stats updates
  EventsOn('stats:update', (data: StatsData) => {
    stats.set(parseStatsData(data));
  });
}
```

Then call `initializeEventListeners()` once in `App.svelte`'s `onMount()`.

## Testing Guidelines

### Unit Tests

Every package should have corresponding tests. Use descriptive test names with the format `Test<Function>_<Case>`:

```go
// blocker_test.go
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
```

### Table-Driven Tests

Use table-driven tests for testing multiple cases:

```go
func TestCalculateFocusRateNormal(t *testing.T) {
    tests := []struct {
        name            string
        focusedTime     time.Duration
        distractedTime  time.Duration
        expectedRateMin float64
        expectedRateMax float64
    }{
        {
            name:            "50/50 split",
            focusedTime:     50 * time.Second,
            distractedTime:  50 * time.Second,
            expectedRateMin: 0.49,
            expectedRateMax: 0.51,
        },
        {
            name:            "100% focused",
            focusedTime:     100 * time.Second,
            distractedTime:  0,
            expectedRateMin: 0.99,
            expectedRateMax: 1.01,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation...
        })
    }
}
```

### Benchmark Tests

Add benchmarks for performance-critical code:

```go
func BenchmarkIsBlocked(b *testing.B) {
    blocklistCfg := config.BlocklistConfig{
        Patterns: []string{"discord", "slack", "steam", "reddit"},
    }
    bm, _ := NewBlocklistMatcher(blocklistCfg, []string{})

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
```

### Test Coverage

Aim for >80% coverage on new code:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Integration Tests

For complex flows, add integration tests in `internal/app/app_test.go`.

## Architecture Guidelines

### Adding New Features

1. **Interface First**: Define interfaces before implementations
2. **Platform Abstraction**: Use build tags for platform-specific code
3. **Event-Driven**: Use channels and events for communication
4. **Configuration**: Make behavior configurable via YAML

### Adding New Components

1. Create package in `internal/`
2. Define interface and types
3. Implement with thread safety in mind
4. Add unit tests
5. Integrate with App orchestrator
6. Update documentation

### Example: Adding a New Blocklist Pattern

The blocklist is regex-based and unified, so adding new patterns is straightforward:

```yaml
# 1. Update config/default.yaml
blocklist:
  patterns:
    - "discord"
    - "reddit"
    - "netflix"
    - "your-new-pattern"  # Add your regex pattern here
```

```go
// 2. Test the new pattern (internal/blocker/blocker_test.go)
func TestIsBlocked_NewPattern(t *testing.T) {
    blocklistCfg := config.BlocklistConfig{
        Patterns: []string{"your-new-pattern"},
    }

    bm, _ := NewBlocklistMatcher(blocklistCfg, []string{})

    window := &monitor.WindowInfo{
        PID:       1234,
        Title:     "Test Window with your-new-pattern",
        Process:   "app.exe",
        Timestamp: time.Now(),
    }

    if !bm.IsBlocked(window) {
        t.Error("Expected window to be blocked by new pattern")
    }
}
```

All patterns are regex with case-insensitive matching and apply to both window titles and process names.

## Documentation

### Code Documentation

- Add godoc comments to exported functions
- Include examples where helpful
- Keep comments up to date with code

### Project Documentation

- Update README.md for user-facing changes
- Update CLAUDE.md for development changes
- Add ADRs for architectural decisions

## Questions?

- Open an issue for bugs or feature requests
- Start a discussion for questions or ideas
- Check existing issues before creating new ones

## License

By contributing, you agree that your contributions will be licensed under the same license as the project (see LICENSE file).

---

**Thank you for contributing to Vigilant!**
