package sources

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"

	"publisher-core/hotspot"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// RSSSource RSS 数据源
type RSSSource struct {
	id      string
	name    string
	enabled bool
	feedURL string
}

// RSSConfig RSS 配置
type RSSConfig struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	FeedURL string `json:"feed_url"`
	Enabled bool   `json:"enabled"`
}

// NewRSSSource 创建 RSS 数据源
func NewRSSSource(config RSSConfig) *RSSSource {
	return &RSSSource{
		id:      config.ID,
		name:    config.Name,
		enabled: config.Enabled,
		feedURL: config.FeedURL,
	}
}

func (s *RSSSource) ID() string {
	return s.id
}

func (s *RSSSource) Name() string {
	return s.name
}

func (s *RSSSource) IsEnabled() bool {
	return s.enabled
}

func (s *RSSSource) SetEnabled(enabled bool) {
	s.enabled = enabled
}

// RSSFeed RSS Feed 结构
type RSSFeed struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Items       []RSSItem `xml:"item"`
	} `xml:"channel"`
}

// RSSItem RSS 条目
type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Author      string `xml:"author"`
	Category    string `xml:"category"`
	GUID        string `xml:"guid"`
}

// AtomFeed Atom Feed 结构
type AtomFeed struct {
	XMLName xml.Name `xml:"feed"`
	Title   string   `xml:"title"`
	Link    string   `xml:"link,attr"`
	Entries []AtomEntry `xml:"entry"`
}

// AtomEntry Atom 条目
type AtomEntry struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	Summary string `xml:"summary"`
	Published string `xml:"published"`
	Updated  string `xml:"updated"`
	Author   struct {
		Name string `xml:"name"`
	} `xml:"author"`
	ID string `xml:"id"`
}

func (s *RSSSource) Fetch(ctx context.Context, maxItems int) ([]hotspot.Topic, error) {
	if !s.enabled {
		return nil, fmt.Errorf("source %s is disabled", s.id)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", s.feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; PublisherTools/1.0)")
	req.Header.Set("Accept", "application/rss+xml, application/xml, text/xml")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// 尝试解析 RSS
	var topics []hotspot.Topic

	// 先尝试 RSS 格式
	var rssFeed RSSFeed
	if err := xml.NewDecoder(resp.Body).Decode(&rssFeed); err == nil && len(rssFeed.Channel.Items) > 0 {
		topics = s.parseRSSFeed(rssFeed, maxItems)
	} else {
		// 重置响应体并尝试 Atom 格式
		resp.Body.Close()
		resp, err = client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var atomFeed AtomFeed
		if err := xml.NewDecoder(resp.Body).Decode(&atomFeed); err == nil && len(atomFeed.Entries) > 0 {
			topics = s.parseAtomFeed(atomFeed, maxItems)
		} else {
			return nil, fmt.Errorf("failed to parse feed as RSS or Atom")
		}
	}

	logrus.Infof("Fetched %d topics from RSS source %s", len(topics), s.id)
	return topics, nil
}

// parseRSSFeed 解析 RSS Feed
func (s *RSSSource) parseRSSFeed(feed RSSFeed, maxItems int) []hotspot.Topic {
	var topics []hotspot.Topic
	now := time.Now()

	for i, item := range feed.Channel.Items {
		if maxItems > 0 && i >= maxItems {
			break
		}

		// 解析发布时间
		var publishedAt time.Time
		if item.PubDate != "" {
			publishedAt = parseRSSTime(item.PubDate)
		}
		if publishedAt.IsZero() {
			publishedAt = now
		}

		// 生成 ID
		id := item.GUID
		if id == "" {
			id = uuid.New().String()
		}
		id = fmt.Sprintf("%s_%s", s.id, id)

		// 计算热度（基于位置）
		heat := 100 - i*2
		if heat < 30 {
			heat = 30
		}

		// 确定趋势
		trend := hotspot.TrendNew
		if i < 10 {
			trend = hotspot.TrendHot
		} else if i < 30 {
			trend = hotspot.TrendUp
		}

		topic := hotspot.Topic{
			ID:          id,
			Title:       cleanTitle(item.Title),
			Description: cleanDescription(item.Description),
			Category:    s.mapCategory(item.Category),
			Heat:        heat,
			Trend:       trend,
			Source:      s.id,
			SourceURL:   item.Link,
			OriginalURL: item.Link,
			Keywords:    extractKeywords(item.Title),
			PublishedAt: publishedAt,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		topics = append(topics, topic)
	}

	return topics
}

// parseAtomFeed 解析 Atom Feed
func (s *RSSSource) parseAtomFeed(feed AtomFeed, maxItems int) []hotspot.Topic {
	var topics []hotspot.Topic
	now := time.Now()

	for i, entry := range feed.Entries {
		if maxItems > 0 && i >= maxItems {
			break
		}

		// 解析发布时间
		var publishedAt time.Time
		if entry.Published != "" {
			publishedAt, _ = time.Parse(time.RFC3339, entry.Published)
		} else if entry.Updated != "" {
			publishedAt, _ = time.Parse(time.RFC3339, entry.Updated)
		}
		if publishedAt.IsZero() {
			publishedAt = now
		}

		// 生成 ID
		id := entry.ID
		if id == "" {
			id = uuid.New().String()
		}
		id = fmt.Sprintf("%s_%s", s.id, id)

		// 计算热度
		heat := 100 - i*2
		if heat < 30 {
			heat = 30
		}

		trend := hotspot.TrendNew
		if i < 10 {
			trend = hotspot.TrendHot
		} else if i < 30 {
			trend = hotspot.TrendUp
		}

		topic := hotspot.Topic{
			ID:          id,
			Title:       cleanTitle(entry.Title),
			Description: cleanDescription(entry.Summary),
			Category:    hotspot.CategoryOther,
			Heat:        heat,
			Trend:       trend,
			Source:      s.id,
			SourceURL:   entry.Link,
			OriginalURL: entry.Link,
			Keywords:    extractKeywords(entry.Title),
			PublishedAt: publishedAt,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		topics = append(topics, topic)
	}

	return topics
}

// parseRSSTime 解析 RSS 时间格式
func parseRSSTime(timeStr string) time.Time {
	layouts := []string{
		time.RFC1123,
		time.RFC1123Z,
		time.RFC822,
		time.RFC822Z,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, timeStr); err == nil {
			return t
		}
	}

	return time.Time{}
}

// mapCategory 映射分类
func (s *RSSSource) mapCategory(category string) hotspot.Category {
	category = strings.ToLower(category)

	switch {
	case strings.Contains(category, "tech") || strings.Contains(category, "科技"):
		return hotspot.CategoryTech
	case strings.Contains(category, "entertainment") || strings.Contains(category, "娱乐"):
		return hotspot.CategoryEntertainment
	case strings.Contains(category, "finance") || strings.Contains(category, "财经"):
		return hotspot.CategoryFinance
	case strings.Contains(category, "sport") || strings.Contains(category, "体育"):
		return hotspot.CategorySports
	case strings.Contains(category, "society") || strings.Contains(category, "社会"):
		return hotspot.CategorySociety
	case strings.Contains(category, "international") || strings.Contains(category, "国际"):
		return hotspot.CategoryInternational
	default:
		return hotspot.CategoryOther
	}
}

// cleanDescription 清理描述
func cleanDescription(desc string) string {
	desc = strings.TrimSpace(desc)
	// 移除 HTML 标签
	desc = stripHTMLTags(desc)
	if len(desc) > 500 {
		desc = desc[:500] + "..."
	}
	return desc
}

// stripHTMLTags 移除 HTML 标签
func stripHTMLTags(s string) string {
	var result strings.Builder
	inTag := false

	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// DefaultRSSFeeds 默认 RSS 源配置
var DefaultRSSFeeds = []RSSConfig{
	{
		ID:      "hackernews",
		Name:    "Hacker News",
		FeedURL: "https://hnrss.org/frontpage",
		Enabled: true,
	},
	{
		ID:      "techcrunch",
		Name:    "TechCrunch",
		FeedURL: "https://techcrunch.com/feed/",
		Enabled: true,
	},
	{
		ID:      "github_trending",
		Name:    "GitHub Trending",
		FeedURL: "https://mshibanami.github.io/GitHubTrendingRSS/daily.xml",
		Enabled: true,
	},
	{
		ID:      "producthunt",
		Name:    "Product Hunt",
		FeedURL: "https://www.producthunt.com/feed",
		Enabled: true,
	},
	{
		ID:      "reddit_programming",
		Name:    "Reddit Programming",
		FeedURL: "https://www.reddit.com/r/programming/.rss",
		Enabled: true,
	},
}

// CreateDefaultRSSSources 创建默认 RSS 源
func CreateDefaultRSSSources() []hotspot.SourceInterface {
	var sources []hotspot.SourceInterface
	for _, config := range DefaultRSSFeeds {
		if config.Enabled {
			sources = append(sources, NewRSSSource(config))
		}
	}
	return sources
}
