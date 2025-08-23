# golang - context  

## What is Context?

Context carries deadlines, cancellation signals, and request-scoped values across API boundaries. It's designed to be passed as the first parameter to functions and is immutable - you create new contexts derived from existing ones.

## Core Context Types

**Background and TODO contexts:**

-   `context.Background()` - root context, never cancelled
-   `context.TODO()` - placeholder when unsure which context to use

**Derived contexts:**

-   `WithCancel()` - creates a cancellable context
-   `WithTimeout()` - cancels after a duration
-   `WithDeadline()` - cancels at a specific time
-   `WithValue()` - carries request-scoped values  

## Example  

In this example I've provided derived context above with its own case, each derived context will have 2 cases: first is a case without context and the second when using context. It will show you how context interacts with function  

## When NOT to Use Context

### 1.  **Don't Use for Optional Parameters**

```go
// ❌ BAD - Context is not for optional parameters
func SaveUser(ctx context.Context, user User, isAdmin bool) error

// ✅ GOOD - Use regular parameters for optional settings
func SaveUser(user User, opts ...SaveOption) error
```

### 2.  **Don't Store Large Data in Context**
```go
// ❌ BAD - Context is not a data bag
ctx = context.WithValue(ctx, "largeData", hugeSlice)

// ✅ GOOD - Pass pointers or use request-scoped caching
func ProcessData(ctx context.Context, largeData []byte) error
```

### 3.  **Don't Use for Function Parameters That Aren't Related to Request Scope**
```go
// ❌ BAD - Math function doesn't need context
func Add(ctx context.Context, a, b int) int

// ✅ GOOD - Context-free function
func Add(a, b int) int
```

### 4.  **Don't Ignore Cancellation in Loops**
```go
func processBatch(ctx context.Context, items []Item) error {
    for _, item := range items {
        // ❌ BAD - Might continue processing after context cancellation
        processItem(item)

        // ✅ GOOD - Check context periodically
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            processItem(item)
        }
    }
    return nil
}
```


## Best Practices

### 1.  **Always Pass Context as First Parameter**
```go
// ✅ Good convention
func DoSomething(ctx context.Context, param1 string, param2 int) error
```

### 2.  **Always Call cancel() to Avoid Leaks**
```go
func example() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel() // ✅ This is crucial!

    // Use ctx...
}
```

### 3.  **Use context.TODO() When Unsure During Development**
```go
func workInProgress() {
    // Use TODO when you're not sure which context to use yet
    ctx := context.TODO()

    // Later replace with proper context
    _ = ctx
}
```

### 4.  **Check Context in Long-Running Operations**
```go
func longOperation(ctx context.Context) error {
    for i := 0; i < 1000; i++ {
        // Check context every iteration
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // Continue work
        }

        // Simulate work
        time.Sleep(10 * time.Millisecond)
    }
    return nil
}
```