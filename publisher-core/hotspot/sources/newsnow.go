package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"publisher-core/hotspot"
	"github.com/sirupsen/logrus"
)

type NewsNowSource struct {
	id       string
	name     string
	enabled  bool
	baseURL  string
	sourceID string
}

type newsNowItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	MobileURL   string `json:"mobileUrl"`
	Source      string `json:"source"`
	SourceURL   string `json:"sourceUrl"`
	SourceID    string `json:"sourceId"`
	PublishTime string `json:"publishTime"`
	Extra       struct {
		HotValue    *int64  `json:"hotValue"`
		OriginTitle *string `json:"originTitle"`
	} `json:"extra"`
}

type newsNowResponse struct {
	Code int           `json:"code"`
	Data []newsNowItem `json:"data"`
	Msg  string        `json:"msg"`
}

var sourceNames = map[string]string{
	"weibo":   "微博热搜",
	"douyin":  "抖音热点",
	"zhihu":   "知乎热榜",
	"baidu":   "百度热搜",
	"toutiao": "今日头条",
	"netease": "网易新闻",
	"sina":    "新浪热点",
	"qq":      "腾讯新闻",
}

func NewNewsNowSource(sourceID string) *NewsNowSource {
	name := sourceNames[sourceID]
	if name == "" {
		name = sourceID
	}

	return &NewsNowSource{
		id:       sourceID,
		name:     name,
		enabled:  true,
		baseURL:  "https://api.oioweb.cn/api/newsnow",
		sourceID: sourceID,
	}
}

func (s *NewsNowSource) ID() string {
	return s.id
}

func (s *NewsNowSource) Name() string {
	return s.name
}

func (s *NewsNowSource) IsEnabled() bool {
	return s.enabled
}

func (s *NewsNowSource) SetEnabled(enabled bool) {
	s.enabled = enabled
}

func (s *NewsNowSource) Fetch(ctx context.Context, maxItems int) ([]hotspot.Topic, error) {
	if !s.enabled {
		return nil, fmt.Errorf("source %s is disabled", s.id)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	url := fmt.Sprintf("%s/%s", s.baseURL, s.sourceID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result newsNowResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("API error: %s", result.Msg)
	}

	var topics []hotspot.Topic
	for i, item := range result.Data {
		if maxItems > 0 && i >= maxItems {
			break
		}
		topic := s.convertItem(item, i+1)
		topics = append(topics, topic)
	}

	logrus.Infof("fetched %d topics from %s", len(topics), s.id)
	return topics, nil
}

func (s *NewsNowSource) convertItem(item newsNowItem, rank int) hotspot.Topic {
	now := time.Now()

	title := item.Title
	if item.Extra.OriginTitle != nil && *item.Extra.OriginTitle != "" {
		title = *item.Extra.OriginTitle
	}

	heat := calculateHeat(rank, item.Extra.HotValue)

	trend := hotspot.TrendNew
	if rank <= 10 {
		trend = hotspot.TrendHot
	} else if rank <= 30 {
		trend = hotspot.TrendUp
	}

	var publishedAt time.Time
	if item.PublishTime != "" {
		if t, err := time.Parse(time.RFC3339, item.PublishTime); err == nil {
			publishedAt = t
		}
	}
	if publishedAt.IsZero() {
		publishedAt = now
	}

	var extra *hotspot.Extra
	if item.Extra.HotValue != nil || item.Extra.OriginTitle != nil {
		extra = &hotspot.Extra{
			HotValue:    item.Extra.HotValue,
			OriginTitle: item.Extra.OriginTitle,
		}
	}

	return hotspot.Topic{
		ID:          generateID(item.ID, s.sourceID),
		Title:       cleanTitle(title),
		Description: "",
		Category:    hotspot.CategoryOther,
		Heat:        heat,
		Trend:       trend,
		Source:      s.sourceID,
		SourceID:    item.SourceID,
		SourceURL:   item.SourceURL,
		OriginalURL: item.URL,
		Keywords:    extractKeywords(title),
		PublishedAt: publishedAt,
		CreatedAt:   now,
		UpdatedAt:   now,
		Extra:       extra,
	}
}

func calculateHeat(rank int, hotValue *int64) int {
	if hotValue != nil && *hotValue > 0 {
		return int(*hotValue / 10000)
	}

	if rank <= 5 {
		return 100 - (rank-1)*5
	} else if rank <= 20 {
		return 80 - (rank-5)*2
	} else if rank <= 50 {
		return 50 - (rank - 20)
	}
	return 30
}

func generateID(originalID, source string) string {
	if originalID != "" {
		return fmt.Sprintf("%s_%s", source, originalID)
	}
	return uuid.New().String()
}

func cleanTitle(title string) string {
	title = strings.TrimSpace(title)
	title = strings.ReplaceAll(title, "\n", " ")
	title = strings.ReplaceAll(title, "\t", " ")
	for strings.Contains(title, "  ") {
		title = strings.ReplaceAll(title, "  ", " ")
	}
	if len(title) > 200 {
		title = title[:200]
	}
	return title
}

func extractKeywords(title string) []string {
	words := strings.Fields(title)
	var keywords []string
	for _, word := range words {
		word = strings.Trim(word, ".,!?;:\"'()[]{}")
		if len(word) >= 2 && len(word) <= 10 {
			keywords = append(keywords, word)
		}
		if len(keywords) >= 5 {
			break
		}
	}
	return keywords
}

func GetAllSourceIDs() []string {
	return []string{"weibo", "douyin", "zhihu", "baidu", "toutiao", "netease", "sina", "qq"}
}

func CreateAllSources() []hotspot.SourceInterface {
	var result []hotspot.SourceInterface
	for _, id := range GetAllSourceIDs() {
		result = append(result, NewNewsNowSource(id))
	}
	return result
}
