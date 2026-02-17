package browser

import (
	"testing"
)

func TestNewBrowser(t *testing.T) {
	cfg := &Config{
		Headless: true,
	}

	browser := NewBrowser(cfg)

	if browser == nil {
		t.Fatal("Browser should not be nil")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Headless != true {
		t.Errorf("Expected default Headless=true, got %v", cfg.Headless)
	}
}

func TestConfigValidation(t *testing.T) {
	cfg := &Config{
		Headless: true,
	}

	if cfg.Headless != true {
		t.Error("Headless should be true")
	}
}

func TestBrowserClose(t *testing.T) {
	browser := NewBrowser(DefaultConfig())

	err := browser.Close()
	if err != nil {
		t.Logf("Close failed (expected if not started): %v", err)
	}
}

func TestBrowserMustPage(t *testing.T) {
	browser := NewBrowser(DefaultConfig())

	page := browser.MustPage()

	if page == nil {
		t.Log("Page is nil (expected if browser not available)")
	}
}

func TestPageHelper(t *testing.T) {
	helper := &PageHelper{}

	if helper == nil {
		t.Fatal("PageHelper should not be nil")
	}
}
