package hotspot

import (
	"context"
	"time"
)

type Trend string

const (
	TrendUp     Trend = "up"
	TrendDown   Trend = "down"
	TrendStable Trend = "stable"
	TrendNew    Trend = "new"
	TrendHot    Trend = "hot"
)

type Category string

const (
	CategoryEntertainment Category = "娱乐"
	CategoryTech          Category = "科技"
	CategoryFinance       Category = "财经"
	CategorySports        Category = "体育"
	CategorySociety       Category = "社会"
	CategoryInternational Category = "国际"
	CategoryOther         Category = "其他"
)

type Topic struct {
	ID          string    `json:"_id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Category    Category  `json:"category"`
	Heat        int       `json:"heat"`
	Trend       Trend     `json:"trend"`
	Source      string    `json:"source"`
	SourceID    string    `json:"sourceId,omitempty"`
	SourceURL   string    `json:"sourceUrl,omitempty"`
	OriginalURL string    `json:"originalUrl,omitempty"`
	Keywords    []string  `json:"keywords,omitempty"`
	Suitability int       `json:"suitability,omitempty"`
	PublishedAt time.Time `json:"publishedAt,omitempty"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty"`
	Extra       *Extra    `json:"extra,omitempty"`
}

type Extra struct {
	HotValue    *int64  `json:"hotValue,omitempty"`
	OriginTitle *string `json:"originTitle,omitempty"`
}

type Source struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

type SourceInterface interface {
	ID() string
	Name() string
	Fetch(ctx context.Context, maxItems int) ([]Topic, error)
	IsEnabled() bool
	SetEnabled(enabled bool)
}

type Filter struct {
	Category Category
	Source   string
	MinHeat  int
	MaxHeat  int
	Limit    int
	Offset   int
	SortBy   string
	SortDesc bool
}

type Storage interface {
	Save(topics []Topic) error
	SaveOne(topic *Topic) error
	Get(id string) (*Topic, error)
	List(filter Filter) ([]Topic, int, error)
	Delete(id string) error
	DeleteBefore(t time.Time) error
	GetByTitle(title string) (*Topic, error)
	GetNewSince(t time.Time) ([]Topic, error)
}

type Service interface {
	FetchFromSource(ctx context.Context, sourceID string, maxItems int) ([]Topic, error)
	FetchFromAllSources(ctx context.Context, maxItemsPerSource int) (map[string][]Topic, error)
	List(filter Filter) ([]Topic, int, error)
	Get(id string) (*Topic, error)
	Refresh(ctx context.Context) (int, error)
	GetSources() []Source
	GetNewTopics(ctx context.Context, since time.Time) ([]Topic, error)
}

type Pagination struct {
	Page    int  `json:"page"`
	Limit   int  `json:"limit"`
	Total   int  `json:"total"`
	Pages   int  `json:"pages"`
	HasMore bool `json:"hasMore"`
}
