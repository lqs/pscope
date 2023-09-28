package ui

import (
	"context"
	"time"
)

func NewReloader(ctx context.Context, cb func()) {
	go func() {
		for {
			cb()
			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Second):
			}
		}
	}()
}
