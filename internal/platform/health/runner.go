package health

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

func runChecks(ctx context.Context, checks []Check) []Result {
	if len(checks) == 0 {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	results := make([]Result, len(checks))
	var wait sync.WaitGroup
	wait.Add(len(checks))

	for i, check := range checks {
		i := i
		check := check

		go func() {
			defer wait.Done()
			results[i] = runCheck(ctx, check)
		}()
	}

	wait.Wait()
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return results
}

func runCheck(parent context.Context, check Check) Result {
	if parent == nil {
		parent = context.Background()
	}

	startedAt := time.Now()
	ctx, cancel := context.WithTimeout(parent, check.Timeout)
	defer cancel()

	done := make(chan Result, 1)
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				done <- failedResult(check, "health check panicked", fmt.Sprint(recovered), startedAt)
			}
		}()

		if check.Checker == nil {
			done <- failedResult(check, "health check checker is not configured", ErrCheckCheckerRequired.Error(), startedAt)
			return
		}

		done <- normalizeResult(check, check.Checker.Check(ctx), startedAt)
	}()

	select {
	case result := <-done:
		if errors.Is(ctx.Err(), context.DeadlineExceeded) && time.Since(startedAt) >= check.Timeout {
			return timeoutResult(check, startedAt)
		}

		return result
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return timeoutResult(check, startedAt)
		}

		return failedResult(check, "health check canceled", ctx.Err().Error(), startedAt)
	}
}

func normalizeResult(check Check, result Result, startedAt time.Time) Result {
	result.Name = check.Name
	result.Status = normalizeStatus(result.Status)
	result.Latency = time.Since(startedAt)
	result.CheckedAt = time.Now().UTC()
	result.Target = check.Target
	result.Criticality = normalizeCriticality(check.Criticality)
	return result
}

func timeoutResult(check Check, startedAt time.Time) Result {
	return failedResult(
		check,
		fmt.Sprintf("health check timed out after %s", check.Timeout),
		context.DeadlineExceeded.Error(),
		startedAt,
	)
}

func failedResult(check Check, message string, err string, startedAt time.Time) Result {
	return Result{
		Name:        check.Name,
		Status:      StatusUnhealthy,
		Message:     message,
		Error:       err,
		Latency:     time.Since(startedAt),
		CheckedAt:   time.Now().UTC(),
		Target:      check.Target,
		Criticality: normalizeCriticality(check.Criticality),
	}
}
