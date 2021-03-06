package badger

import (
	"context"
	"testing"
	"time"
)

func TestServiceWithCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	service := NewService(
		WithContext(ctx),
		// WithSignal(false),
	)

	go func() {
		time.Sleep(5 * time.Second)
		t.Log("shutdown service with context")
		cancel()
	}()

	if err := service.Run(); err != nil {
		t.Fatal(err)
	}

	t.Log("service stopped")
}

// func TestService(t *testing.T) {
// 	service := NewService()

// 	if err := service.Run(); err != nil {
// 		t.Fatal(err)
// 	}

// 	t.Log("service stopped")
// }
