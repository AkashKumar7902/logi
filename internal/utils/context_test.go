package utils

import (
	"context"
	"testing"
	"time"
)

func TestSetDBOperationTimeoutAffectsDBContext(t *testing.T) {
	original := DBOperationTimeout()
	t.Cleanup(func() {
		SetDBOperationTimeout(original)
	})

	SetDBOperationTimeout(2 * time.Second)

	if got := DBOperationTimeout(); got != 2*time.Second {
		t.Fatalf("expected timeout to be updated to 2s, got %v", got)
	}

	ctx, cancel := DBContext(context.Background())
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected db context to have a deadline")
	}

	remaining := time.Until(deadline)
	if remaining > 2*time.Second || remaining < time.Second {
		t.Fatalf("expected deadline close to 2s from now, got %v", remaining)
	}
}
