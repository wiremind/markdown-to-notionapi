package notion

import (
	"context"
	"testing"
)

func TestProcessBlocksInChunks(t *testing.T) {
	client := &Client{verbose: false}

	// Test with empty blocks
	err := client.processBlocksInChunks(context.Background(), []Block{}, func(ctx context.Context, chunk []Block) error {
		t.Error("Function should not be called with empty blocks")
		return nil
	})
	if err != nil {
		t.Errorf("Expected nil error for empty blocks, got %v", err)
	}

	// Test with blocks smaller than chunk size
	smallBlocks := make([]Block, 25)
	for i := range smallBlocks {
		smallBlocks[i] = Block{Type: "paragraph"}
	}

	callCount := 0
	err = client.processBlocksInChunks(context.Background(), smallBlocks, func(ctx context.Context, chunk []Block) error {
		callCount++
		if len(chunk) != 25 {
			t.Errorf("Expected chunk size 25, got %d", len(chunk))
		}
		return nil
	})
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}

	// Test with blocks larger than chunk size
	largeBlocks := make([]Block, 126) // This simulates the original problem
	for i := range largeBlocks {
		largeBlocks[i] = Block{Type: "paragraph"}
	}

	callCount = 0
	totalProcessed := 0
	err = client.processBlocksInChunks(context.Background(), largeBlocks, func(ctx context.Context, chunk []Block) error {
		callCount++
		totalProcessed += len(chunk)
		if len(chunk) > BlockChunkSize {
			t.Errorf("Chunk size %d exceeds maximum %d", len(chunk), BlockChunkSize)
		}
		return nil
	})
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	expectedCalls := (126 + BlockChunkSize - 1) / BlockChunkSize // Ceiling division
	if callCount != expectedCalls {
		t.Errorf("Expected %d calls, got %d", expectedCalls, callCount)
	}
	if totalProcessed != 126 {
		t.Errorf("Expected to process 126 blocks, processed %d", totalProcessed)
	}
}
