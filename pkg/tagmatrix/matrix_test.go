package tagmatrix

import (
	"sort"
	"testing"
)

func TestTagMatrix(t *testing.T) {
	matrix := NewMatrix()

	// Add tags for block 100
	matrix.AddTags(100, []Tag{
		{Key: "type", Value: "context"},
		{Key: "session", Value: "sess_1"},
		{Key: "role", Value: "user"},
	})

	// Add tags for block 200
	matrix.AddTags(200, []Tag{
		{Key: "type", Value: "context"},
		{Key: "session", Value: "sess_1"},
		{Key: "role", Value: "assistant"},
	})

	// Add tags for block 300
	matrix.AddTags(300, []Tag{
		{Key: "type", Value: "summary"},
		{Key: "session", Value: "sess_2"},
	})

	// Query 1: session=sess_1
	res1 := matrix.Query([]Tag{{Key: "session", Value: "sess_1"}})
	sort.Slice(res1, func(i, j int) bool { return res1[i] < res1[j] })
	if len(res1) != 2 || res1[0] != 100 || res1[1] != 200 {
		t.Fatalf("Query 1 failed: expected [100 200], got %v", res1)
	}

	// Query 2: session=sess_1 AND role=user
	res2 := matrix.Query([]Tag{
		{Key: "session", Value: "sess_1"},
		{Key: "role", Value: "user"},
	})
	if len(res2) != 1 || res2[0] != 100 {
		t.Fatalf("Query 2 failed: expected [100], got %v", res2)
	}

	// Query 3: Non-existent tag
	res3 := matrix.Query([]Tag{{Key: "type", Value: "unknown"}})
	if len(res3) != 0 {
		t.Fatalf("Query 3 failed: expected empty, got %v", res3)
	}

	// Query 4: Empty query
	res4 := matrix.Query([]Tag{})
	if len(res4) != 0 {
		t.Fatalf("Query 4 failed: expected empty, got %v", res4)
	}

	// Remove block 100
	matrix.RemoveBlock(100)

	// Query 5: session=sess_1 after removal
	res5 := matrix.Query([]Tag{{Key: "session", Value: "sess_1"}})
	if len(res5) != 1 || res5[0] != 200 {
		t.Fatalf("Query 5 failed: expected [200], got %v", res5)
	}

	// Query 6: session=sess_1 AND role=user after removal
	res6 := matrix.Query([]Tag{
		{Key: "session", Value: "sess_1"},
		{Key: "role", Value: "user"},
	})
	if len(res6) != 0 {
		t.Fatalf("Query 6 failed: expected empty, got %v", res6)
	}
}
