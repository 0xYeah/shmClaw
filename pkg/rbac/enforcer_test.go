package rbac

import "testing"

func TestEnforcer(t *testing.T) {
	e := NewEnforcer()

	// Grant user_1 read access to session_1
	e.AddPolicy("user_1", "session_1", "read")

	// Grant user_2 write access to any session starting with "public_"
	e.AddPolicy("user_2", "public_*", "write")

	// Grant admin all access to everything
	e.AddPolicy("admin", "*", "read")
	e.AddPolicy("admin", "*", "write")

	// Exact match tests
	if !e.Check("user_1", "session_1", "read") {
		t.Fatal("Expected user_1 to have read access to session_1")
	}
	if e.Check("user_1", "session_1", "write") {
		t.Fatal("Expected user_1 to NOT have write access to session_1")
	}
	if e.Check("user_1", "session_2", "read") {
		t.Fatal("Expected user_1 to NOT have read access to session_2")
	}

	// Wildcard tests
	if !e.Check("user_2", "public_123", "write") {
		t.Fatal("Expected user_2 to have write access to public_123")
	}
	if e.Check("user_2", "private_123", "write") {
		t.Fatal("Expected user_2 to NOT have write access to private_123")
	}

	// Admin wildcard tests
	if !e.Check("admin", "secret_session", "read") {
		t.Fatal("Expected admin to have read access to secret_session")
	}
	if !e.Check("admin", "secret_session", "write") {
		t.Fatal("Expected admin to have write access to secret_session")
	}

	// Unknown subject
	if e.Check("unknown_user", "session_1", "read") {
		t.Fatal("Expected unknown_user to NOT have access")
	}
}
