package utils

import (
	"context"
	"sync/atomic"
	"time"
)

const defaultDBOperationTimeout = 5 * time.Second

var dbOperationTimeoutNanos atomic.Int64

func init() {
	dbOperationTimeoutNanos.Store(int64(defaultDBOperationTimeout))
}

func SetDBOperationTimeout(timeout time.Duration) {
	if timeout <= 0 {
		timeout = defaultDBOperationTimeout
	}
	dbOperationTimeoutNanos.Store(int64(timeout))
}

func DBOperationTimeout() time.Duration {
	return time.Duration(dbOperationTimeoutNanos.Load())
}

func DBContext(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithTimeout(parent, DBOperationTimeout())
}
