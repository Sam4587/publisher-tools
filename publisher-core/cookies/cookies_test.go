package cookies

import (
	"context"
	"testing"

	"github.com/go-rod/rod/lib/proto"
)

func TestNewManager(t *testing.T) {
	manager := NewManager("./test-cookies")

	if manager == nil {
		t.Fatal("Manager should not be nil")
	}
}

func TestCookieJar(t *testing.T) {
	jar := NewCookieJar()

	if jar == nil {
		t.Fatal("CookieJar should not be nil")
	}
}

func TestManagerSaveAndGet(t *testing.T) {
	manager := NewManager("./test-cookies")

	platform := "douyin"

	cookies := []*proto.NetworkCookie{
		{
			Name:   "session_id",
			Value:  "abc123",
			Domain: "douyin.com",
		},
	}

	err := manager.Save(context.Background(), platform, cookies)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	retrieved, err := manager.Load(context.Background(), platform)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(retrieved) != len(cookies) {
		t.Errorf("Cookie count mismatch, got %d, want %d", len(retrieved), len(cookies))
	}
}

func TestManagerDelete(t *testing.T) {
	manager := NewManager("./test-cookies")

	platform := "douyin-test"

	_ = manager.Save(context.Background(), platform, []*proto.NetworkCookie{})

	err := manager.Delete(context.Background(), platform)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = manager.Load(context.Background(), platform)
	if err == nil {
		t.Error("Load should fail after Delete")
	}
}

func TestManagerExists(t *testing.T) {
	manager := NewManager("./test-cookies")

	platform := "test-exists"

	exists, err := manager.Exists(context.Background(), platform)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("Platform should not exist initially")
	}

	_ = manager.Save(context.Background(), platform, []*proto.NetworkCookie{
		{Name: "test", Value: "value"},
	})

	exists, err = manager.Exists(context.Background(), platform)
	if err != nil {
		t.Fatalf("Exists failed after save: %v", err)
	}
	if !exists {
		t.Error("Platform should exist after save")
	}
}

func TestExtractCookies(t *testing.T) {
	cookies := []*proto.NetworkCookie{
		{Name: "session_id", Value: "abc"},
		{Name: "csrf_token", Value: "def"},
		{Name: "other", Value: "ghi"},
	}

	keys := []string{"session_id", "csrf_token"}
	extracted := ExtractCookies(cookies, keys)

	if len(extracted) != 2 {
		t.Errorf("Expected 2 extracted cookies, got %d", len(extracted))
	}
}

func TestCookieJarSetAndGet(t *testing.T) {
	jar := NewCookieJar()

	cookies := []*proto.NetworkCookie{
		{
			Name:   "test",
			Value:  "123",
			Domain: "example.com",
			Path:   "/",
		},
	}

	jar.SetCookies("https://example.com", cookies)

	retrieved := jar.Cookies("https://example.com")
	if len(retrieved) != 1 {
		t.Errorf("Expected 1 cookie, got %d", len(retrieved))
	}
}
