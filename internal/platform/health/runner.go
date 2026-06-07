package health

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

func runChecks(ctx context.Context, registrations []Config) []Result {
	if len(registrations) == 0 {
		return []Result{}
	}
	if ctx == nil {
		ctx = context.Background()
	}

	results := make([]Result, len(registrations))
	var wait sync.WaitGroup
	wait.Add(len(registrations))

	for i, registration := range registrations {
		i := i
		registration := registration

		go func() {
			defer wait.Done()
			results[i] = runCheck(ctx, registration)
		}()
	}

	wait.Wait()
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return results
}

func runCheck(parent context.Context, registration Config) Result {
	if parent == nil {
		parent = context.Background()
	}

	spec := registration.Spec
	startedAt := time.Now()
	ctx, cancel := context.WithTimeout(parent, spec.Timeout)
	defer cancel()

	done := make(chan Result, 1)
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				done <- failedResult(spec, "health check panicked", fmt.Sprint(recovered), startedAt)
			}
		}()

		if registration.Checker == nil {
			done <- failedResult(spec, "health check checker is not configured", ErrCheckCheckerRequired.Error(), startedAt)
			return
		}

		done <- normalizeResult(spec, registration.Checker.Check(ctx), startedAt)
	}()

	select {
	case result := <-done:
		if errors.Is(ctx.Err(), context.DeadlineExceeded) && time.Since(startedAt) >= spec.Timeout {
			return timeoutResult(spec, startedAt)
		}

		return result
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return timeoutResult(spec, startedAt)
		}

		return failedResult(spec, "health check canceled", ctx.Err().Error(), startedAt)
	}
}

func normalizeResult(spec Spec, result Result, startedAt time.Time) Result {
	result.Name = spec.Name
	result.Status = normalizeStatus(result.Status)
	result.Latency = latencyMillisecondsSince(startedAt)
	result.CheckedAt = time.Now().UTC()
	result.Target = spec.Target
	result.Criticality = normalizeCriticality(spec.Criticality)
	return result
}

func timeoutResult(spec Spec, startedAt time.Time) Result {
	return failedResult(
		spec,
		fmt.Sprintf("health check timed out after %s", spec.Timeout),
		context.DeadlineExceeded.Error(),
		startedAt,
	)
}

func failedResult(spec Spec, message string, err string, startedAt time.Time) Result {
	return Result{
		Name:        spec.Name,
		Status:      StatusUnhealthy,
		Message:     message,
		Error:       err,
		Latency:     latencyMillisecondsSince(startedAt),
		CheckedAt:   time.Now().UTC(),
		Target:      spec.Target,
		Criticality: normalizeCriticality(spec.Criticality),
	}
}

func latencyMillisecondsSince(startedAt time.Time) time.Duration {
	return time.Duration(time.Since(startedAt).Milliseconds())
}
