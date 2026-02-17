package hotspot

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type JSONStorage struct {
	mu       sync.RWMutex
	dataDir  string
	topics   map[string]*Topic
	filePath string
}

func NewJSONStorage(dataDir string) (*JSONStorage, error) {
	if dataDir == "" {
		dataDir = "./data"
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	s := &JSONStorage{
		dataDir:  dataDir,
		topics:   make(map[string]*Topic),
		filePath: filepath.Join(dataDir, "hotspot_topics.json"),
	}

	if err := s.load(); err != nil {
		logrus.Warnf("load existing topics failed: %v, starting fresh", err)
	}

	return s, nil
}

func (s *JSONStorage) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var topics []*Topic
	if err := json.Unmarshal(data, &topics); err != nil {
		return err
	}

	for _, t := range topics {
		s.topics[t.ID] = t
	}

	logrus.Infof("loaded %d topics from storage", len(s.topics))
	return nil
}

func (s *JSONStorage) save() error {
	topics := make([]*Topic, 0, len(s.topics))
	for _, t := range s.topics {
		topics = append(topics, t)
	}

	data, err := json.MarshalIndent(topics, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

func (s *JSONStorage) Save(topics []Topic) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for i := range topics {
		t := &topics[i]
		if existing, ok := s.topics[t.ID]; ok {
			t.CreatedAt = existing.CreatedAt
		} else {
			if t.CreatedAt.IsZero() {
				t.CreatedAt = now
			}
		}
		t.UpdatedAt = now
		s.topics[t.ID] = t
	}

	logrus.Infof("saved %d topics, total: %d", len(topics), len(s.topics))
	return s.save()
}

func (s *JSONStorage) SaveOne(topic *Topic) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if existing, ok := s.topics[topic.ID]; ok {
		topic.CreatedAt = existing.CreatedAt
	} else {
		if topic.CreatedAt.IsZero() {
			topic.CreatedAt = now
		}
	}
	topic.UpdatedAt = now
	s.topics[topic.ID] = topic

	return s.save()
}

func (s *JSONStorage) Get(id string) (*Topic, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	topic, ok := s.topics[id]
	if !ok {
		return nil, nil
	}
	return topic, nil
}

func (s *JSONStorage) List(filter Filter) ([]Topic, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []Topic
	for _, t := range s.topics {
		if filter.Category != "" && t.Category != filter.Category {
			continue
		}
		if filter.Source != "" && t.Source != filter.Source {
			continue
		}
		if filter.MinHeat > 0 && t.Heat < filter.MinHeat {
			continue
		}
		if filter.MaxHeat > 0 && t.Heat > filter.MaxHeat {
			continue
		}
		result = append(result, *t)
	}

	sortTopics(result, filter.SortBy, filter.SortDesc)

	total := len(result)

	if filter.Offset > 0 && filter.Offset < len(result) {
		result = result[filter.Offset:]
	}
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}

	return result, total, nil
}

func sortTopics(topics []Topic, sortBy string, desc bool) {
	sort.Slice(topics, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "heat":
			less = topics[i].Heat < topics[j].Heat
		case "createdAt":
			less = topics[i].CreatedAt.Before(topics[j].CreatedAt)
		case "publishedAt":
			less = topics[i].PublishedAt.Before(topics[j].PublishedAt)
		case "title":
			less = strings.Compare(topics[i].Title, topics[j].Title) < 0
		default:
			less = topics[i].Heat < topics[j].Heat
		}
		if desc {
			return !less
		}
		return less
	})
}

func (s *JSONStorage) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.topics, id)
	return s.save()
}

func (s *JSONStorage) DeleteBefore(t time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for id, topic := range s.topics {
		if topic.CreatedAt.Before(t) {
			delete(s.topics, id)
			count++
		}
	}

	logrus.Infof("deleted %d old topics before %s", count, t.Format(time.RFC3339))
	return s.save()
}

func (s *JSONStorage) GetByTitle(title string) (*Topic, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	title = strings.ToLower(strings.TrimSpace(title))
	for _, t := range s.topics {
		if strings.Contains(strings.ToLower(t.Title), title) {
			return t, nil
		}
	}
	return nil, nil
}

func (s *JSONStorage) GetNewSince(t time.Time) ([]Topic, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []Topic
	for _, topic := range s.topics {
		if topic.CreatedAt.After(t) {
			result = append(result, *topic)
		}
	}

	sortTopics(result, "createdAt", true)
	return result, nil
}

func (s *JSONStorage) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.topics)
}

func (s *JSONStorage) CountBySource() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	counts := make(map[string]int)
	for _, t := range s.topics {
		counts[t.Source]++
	}
	return counts
}
