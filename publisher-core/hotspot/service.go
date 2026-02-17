package hotspot

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type HotspotService struct {
	mu      sync.RWMutex
	sources map[string]SourceInterface
	storage Storage
}

func NewService(storage Storage) *HotspotService {
	return &HotspotService{
		sources: make(map[string]SourceInterface),
		storage: storage,
	}
}

func (s *HotspotService) RegisterSource(source SourceInterface) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sources[source.ID()] = source
	logrus.Infof("registered hotspot source: %s", source.ID())
}

func (s *HotspotService) GetSources() []Source {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Source, 0, len(s.sources))
	for _, src := range s.sources {
		result = append(result, Source{
			ID:      src.ID(),
			Name:    src.Name(),
			Enabled: src.IsEnabled(),
		})
	}
	return result
}

func (s *HotspotService) FetchFromSource(ctx context.Context, sourceID string, maxItems int) ([]Topic, error) {
	s.mu.RLock()
	source, ok := s.sources[sourceID]
	s.mu.RUnlock()

	if !ok {
		return nil, ErrSourceNotFound
	}

	topics, err := source.Fetch(ctx, maxItems)
	if err != nil {
		return nil, err
	}

	if s.storage != nil && len(topics) > 0 {
		if err := s.storage.Save(topics); err != nil {
			logrus.Warnf("save topics failed: %v", err)
		}
	}

	return topics, nil
}

func (s *HotspotService) FetchFromAllSources(ctx context.Context, maxItemsPerSource int) (map[string][]Topic, error) {
	s.mu.RLock()
	sources := make([]SourceInterface, 0, len(s.sources))
	for _, src := range s.sources {
		if src.IsEnabled() {
			sources = append(sources, src)
		}
	}
	s.mu.RUnlock()

	results := make(map[string][]Topic)
	var allTopics []Topic

	for _, source := range sources {
		topics, err := source.Fetch(ctx, maxItemsPerSource)
		if err != nil {
			logrus.Warnf("fetch from %s failed: %v", source.ID(), err)
			continue
		}
		results[source.ID()] = topics
		allTopics = append(allTopics, topics...)
	}

	if s.storage != nil && len(allTopics) > 0 {
		if err := s.storage.Save(allTopics); err != nil {
			logrus.Warnf("save topics failed: %v", err)
		}
	}

	logrus.Infof("fetched %d topics from %d sources", len(allTopics), len(sources))
	return results, nil
}

func (s *HotspotService) List(filter Filter) ([]Topic, int, error) {
	if s.storage == nil {
		return nil, 0, ErrStorageNotInitialized
	}
	return s.storage.List(filter)
}

func (s *HotspotService) Get(id string) (*Topic, error) {
	if s.storage == nil {
		return nil, ErrStorageNotInitialized
	}
	return s.storage.Get(id)
}

func (s *HotspotService) Refresh(ctx context.Context) (int, error) {
	s.mu.RLock()
	sources := make([]SourceInterface, 0, len(s.sources))
	for _, src := range s.sources {
		if src.IsEnabled() {
			sources = append(sources, src)
		}
	}
	s.mu.RUnlock()

	var allTopics []Topic
	for _, source := range sources {
		topics, err := source.Fetch(ctx, 0)
		if err != nil {
			logrus.Warnf("fetch from %s failed: %v", source.ID(), err)
			continue
		}
		allTopics = append(allTopics, topics...)
	}

	if s.storage != nil && len(allTopics) > 0 {
		if err := s.storage.Save(allTopics); err != nil {
			return 0, err
		}
	}

	return len(allTopics), nil
}

func (s *HotspotService) GetNewTopics(ctx context.Context, since time.Time) ([]Topic, error) {
	if s.storage == nil {
		return nil, ErrStorageNotInitialized
	}
	return s.storage.GetNewSince(since)
}

func (s *HotspotService) Delete(id string) error {
	if s.storage == nil {
		return ErrStorageNotInitialized
	}
	return s.storage.Delete(id)
}

func (s *HotspotService) CleanupOldData(olderThan time.Duration) (int, error) {
	if s.storage == nil {
		return 0, ErrStorageNotInitialized
	}

	cutoff := time.Now().Add(-olderThan)
	if err := s.storage.DeleteBefore(cutoff); err != nil {
		return 0, err
	}

	return 0, nil
}

var (
	ErrSourceNotFound        = &Error{Code: "SOURCE_NOT_FOUND", Message: "source not found"}
	ErrStorageNotInitialized = &Error{Code: "STORAGE_NOT_INITIALIZED", Message: "storage not initialized"}
)

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string {
	return e.Message
}
