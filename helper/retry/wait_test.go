// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package retry

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestRetry(t *testing.T) {
	t.Parallel()

	tries := 0
	f := func() *RetryError {
		tries++
		if tries == 3 {
			return nil
		}

		return RetryableError(fmt.Errorf("error"))
	}

	err := Retry(10*time.Second, f)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

// make sure a slow StateRefreshFunc is allowed to complete after timeout
func TestRetry_grace(t *testing.T) {
	t.Parallel()

	f := func() *RetryError {
		time.Sleep(1 * time.Second)
		return nil
	}

	err := Retry(10*time.Millisecond, f)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestRetry_timeout(t *testing.T) {
	t.Parallel()

	f := func() *RetryError {
		return RetryableError(fmt.Errorf("always"))
	}

	err := Retry(1*time.Second, f)
	if err == nil {
		t.Fatal("should error")
	}
}

func TestRetry_hang(t *testing.T) {
	old := refreshGracePeriod
	refreshGracePeriod = 50 * time.Millisecond
	defer func() {
		refreshGracePeriod = old
	}()

	f := func() *RetryError {
		time.Sleep(2 * time.Second)
		return nil
	}

	err := Retry(50*time.Millisecond, f)
	if err == nil {
		t.Fatal("should error")
	}
}

func TestRetry_error(t *testing.T) {
	t.Parallel()

	expected := fmt.Errorf("nope")
	f := func() *RetryError {
		return NonRetryableError(expected)
	}

	errCh := make(chan error)
	go func() {
		errCh <- Retry(1*time.Second, f)
	}()

	select {
	case err := <-errCh:
		if err != expected {
			t.Fatalf("bad: %#v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}
}

func TestRetry_nilNonRetryableError(t *testing.T) {
	t.Parallel()

	f := func() *RetryError {
		return NonRetryableError(nil)
	}

	err := Retry(1*time.Second, f)
	if err == nil {
		t.Fatal("should error")
	}
}

func TestRetry_nilRetryableError(t *testing.T) {
	t.Parallel()

	f := func() *RetryError {
		return RetryableError(nil)
	}

	err := Retry(1*time.Second, f)
	if err == nil {
		t.Fatal("should error")
	}
}

func TestRetryContext_cancel(t *testing.T) {
	t.Parallel()

	waitForever := make(chan struct{})
	defer close(waitForever)
	ctx, cancel := context.WithCancel(context.Background())

	f := func() *RetryError {
		<-waitForever
		return nil
	}

	cancel()

	if err := RetryContext(ctx, 10*time.Millisecond, f); err != context.Canceled {
		t.Fatalf("Expected context.Canceled error, got: %s", err)
	}
}

func TestRetryContext_deadline(t *testing.T) {
	t.Parallel()

	waitForever := make(chan struct{})
	defer close(waitForever)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	f := func() *RetryError {
		<-waitForever
		return nil
	}

	if err := RetryContext(ctx, 10*time.Second, f); err != context.DeadlineExceeded {
		t.Fatalf("Expected context.DeadlineExceeded error, got: %s", err)
	}
}
