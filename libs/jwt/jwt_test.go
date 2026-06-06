package jwt

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestIssueAndParseAccess(t *testing.T) {
	mgr := NewManager("test-secret", time.Minute)
	uid := uuid.New()

	token, err := mgr.IssueAccess(uid, "teacher", true)
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}

	claims, err := mgr.ParseAccess(token)
	if err != nil {
		t.Fatalf("ParseAccess: %v", err)
	}
	if claims.UserID != uid {
		t.Fatalf("uid: got %s want %s", claims.UserID, uid)
	}
	if claims.Role != "teacher" || !claims.IsAdmin {
		t.Fatalf("role/admin: %+v", claims)
	}
}

func TestParseAccessRejectsBadToken(t *testing.T) {
	mgr := NewManager("secret-a", time.Minute)
	other := NewManager("secret-b", time.Minute)
	uid := uuid.New()

	token, err := other.IssueAccess(uid, "student", false)
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}
	if _, err := mgr.ParseAccess(token); err == nil {
		t.Fatal("wrong secret should fail")
	}
	if _, err := mgr.ParseAccess("not-a-jwt"); err == nil {
		t.Fatal("garbage token should fail")
	}
}

func TestNewManagerDefaultTTL(t *testing.T) {
	mgr := NewManager("x", 0)
	if mgr.ttl != 15*time.Minute {
		t.Fatalf("default ttl: %v", mgr.ttl)
	}
}
