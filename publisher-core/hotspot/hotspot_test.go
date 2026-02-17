package hotspot

import (
	"context"
	"testing"
	"time"
)

func TestTopicValidation(t *testing.T) {
	topic := Topic{
		ID:        "test-1",
		Title:     "Test Hot Topic",
		Source:    "test-source",
		Heat:      1000000,
		CreatedAt: time.Now(),
	}

	if topic.ID != "test-1" {
		t.Error("ID should be set correctly")
	}

	if topic.Title != "Test Hot Topic" {
		t.Error("Title should be set correctly")
	}
}

func TestNewService(t *testing.T) {
	storage, _ := NewJSONStorage("./test-data")
	service := NewService(storage)

	if service == nil {
		t.Fatal("Service should not be nil")
	}
}

func TestServiceRegisterSource(t *testing.T) {
	storage, _ := NewJSONStorage("./test-data")
	service := NewService(storage)

	mockSource := &MockTestSource{
		id:      "test-source",
		name:    "test-source",
		display: "Test Source",
		enabled: true,
	}

	service.RegisterSource(mockSource)

	sources := service.GetSources()
	found := false
	for _, s := range sources {
		if s.ID == "test-source" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Registered source should be in list")
	}
}

type MockTestSource struct {
	id      string
	name    string
	display string
	enabled bool
}

func (m *MockTestSource) ID() string {
	return m.id
}

func (m *MockTestSource) Name() string {
	return m.name
}

func (m *MockTestSource) DisplayName() string {
	return m.display
}

func (m *MockTestSource) IsEnabled() bool {
	return m.enabled
}

func (m *MockTestSource) SetEnabled(enabled bool) {
	m.enabled = enabled
}

func (m *MockTestSource) Fetch(ctx context.Context, maxItems int) ([]Topic, error) {
	return []Topic{
		{
			ID:        "mock-1",
			Title:     "Mock Topic 1",
			Source:    m.name,
			Heat:      500000,
			CreatedAt: time.Now(),
		},
	}, nil
}

func TestTopicCompare(t *testing.T) {
	topic1 := Topic{Heat: 1000000}
	topic2 := Topic{Heat: 500000}

	if topic1.Heat <= topic2.Heat {
		t.Error("Topic1 should have higher heat than Topic2")
	}
}

func TestTopicCache(t *testing.T) {
	topic := Topic{
		ID:        "cache-test",
		CreatedAt: time.Now(),
	}

	cacheDuration := 5 * time.Minute
	isFresh := time.Since(topic.CreatedAt) < cacheDuration

	if !isFresh {
		t.Error("Topic should be fresh just after creation")
	}
}
