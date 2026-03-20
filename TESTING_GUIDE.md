# Testing Guide

## Better Go Testing Practices

Instead of using `init()` functions to set up mocks, this codebase now uses dependency injection with interfaces for better testability.

### The Old Way (Problematic)

```go
// DON'T DO THIS - Global state mutation in init()
var mockExecCommand = func(name string, args ...string) *exec.Cmd {
    // mock implementation
}

func init() {
    internal.ExecCommand = mockExecCommand  // Mutates global state
}
```

**Problems with this approach:**
- Global state mutation affects all tests
- Tests can interfere with each other
- Difficult to reset between tests
- Not explicit about what's being mocked
- Makes parallel testing unsafe

### The New Way (Recommended)

```go
// Define interface for testability
type JobRunner interface {
    Execute(name string, args ...string) ([]byte, error)
}

// Real implementation
type RealSync struct{}
func (r *RealSync) Execute(name string, args ...string) ([]byte, error) {
    cmd := exec.Command(name, args...)
    return cmd.CombinedOutput()
}

// Test implementation
type MockCommandExecutor struct {
    CapturedCommands []MockCommand
}
func (m *MockCommandExecutor) Execute(name string, args ...string) ([]byte, error) {
    // Mock logic here
}
```

### Testing Examples

#### 1. Simple Test with Mock

```go
func TestExecuteJob(t *testing.T) {
    // Create mock executor for this test only
    mockExecutor := &MockCommandExecutor{}
    
    job := internal.Job{
        Name:    "test_job",
        Source:  "/home/test/",
        Target:  "/mnt/backup1/test/",
        Enabled: true,
    }
    
    // Use the executor-aware function
    status := internal.ExecuteJobWithExecutor(job, true, false, "", mockExecutor)
    
    if status != "SUCCESS" {
        t.Errorf("Expected SUCCESS, got %s", status)
    }
    
    // Verify the mock was called correctly
    if len(mockExecutor.CapturedCommands) == 0 {
        t.Error("Expected command to be executed")
    }
}
```

#### 2. Test Setup with t.Cleanup() (Alternative Pattern)

```go
func setupMockExecutor(t *testing.T) *MockCommandExecutor {
    t.Helper()
    
    // Store original to restore later
    originalExecutor := internal.DefaultExecutor
    
    // Create mock
    mockExecutor := &MockCommandExecutor{}
    
    // Set mock globally
    internal.DefaultExecutor = mockExecutor
    
    // Restore original after test
    t.Cleanup(func() {
        internal.DefaultExecutor = originalExecutor
    })
    
    return mockExecutor
}

func TestWithSetup(t *testing.T) {
    mock := setupMockExecutor(t)
    
    // Use regular ExecuteJob function - it uses DefaultExecutor
    status := internal.ExecuteJob(job, true, false, "")
    
    // Verify mock was used
    assert.Equal(t, 1, len(mock.CapturedCommands))
}
```

### Benefits of the New Approach

1. **Test Isolation**: Each test gets its own mock instance
2. **Explicit Dependencies**: Clear what's being mocked in each test
3. **Better Assertions**: Can inspect captured calls, arguments, etc.
4. **Parallel Safe**: Tests don't interfere with each other
5. **Easier Debugging**: Clearer test failures and state
6. **Production Safety**: Real code unchanged, only test behavior modified

### Key Principles

1. **Use interfaces for external dependencies** (file system, network, exec, etc.)
2. **Inject dependencies rather than using globals**
3. **Keep mocks scoped to individual tests**
4. **Use `t.Cleanup()` for setup/teardown patterns**
5. **Make test intentions explicit in the test code**

This approach follows Go testing best practices and makes the codebase more maintainable and reliable.