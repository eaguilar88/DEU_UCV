package main

import (
	"context"
	"fmt"
	"time"
)

func saveToDB(ctx context.Context) error {
	// Simulate long-running DB operation with respect to context
	select {
	case <-time.After(time.Second * 9): // Simulate a 15-second DB operation
		return nil // Operation completed
	case <-ctx.Done():
		return ctx.Err() // Context canceled
	}
}

func main() {
	fmt.Println("working")
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		// Save things to a DB
		err := saveToDB(ctx)
		if err != nil {
			fmt.Println("Error saving to DB:", err)
		} else {
			fmt.Println("Saved to DB successfully")
		}
	}()

	fmt.Println("keep working")

	// Wait to see the output (not part of actual implementation)
	// time.Sleep(time.Second * 20)
}
