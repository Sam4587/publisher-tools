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

func (s *NewsNowSource) DisplayName() string {
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
		logrus.Warnf("[%s] API request failed: %v", s.id, err)
		return s.getFallbackTopics(maxItems), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logrus.Warnf("[%s] API returned status %d", s.id, resp.StatusCode)
		return s.getFallbackTopics(maxItems), nil
	}

	var result newsNowResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logrus.Warnf("[%s] Failed to parse response: %v", s.id, err)
		return s.getFallbackTopics(maxItems), nil
	}

	if result.Code != 200 {
		logrus.Warnf("[%s] API error: %s", s.id, result.Msg)
		return s.getFallbackTopics(maxItems), nil
	}

	if len(result.Data) == 0 {
		logrus.Warnf("[%s] No data returned from API", s.id)
		return s.getFallbackTopics(maxItems), nil
	}

	var topics []hotspot.Topic
	for i, item := range result.Data {
		if maxItems > 0 && i >= maxItems {
			break
		}
		topic := s.convertItem(item, i+1)
		topics = append(topics, topic)
	}

	logrus.Infof("[%s] fetched %d topics from API", s.id, len(topics))
	return topics, nil
}

func (s *NewsNowSource) getFallbackTopics(maxItems int) []hotspot.Topic {
	now := time.Now()
	fallbackTopics := map[string][]hotspot.Topic{
		"weibo": {
			{ID: "weibo-fallback-1", Title: "微博热搜暂不可用", Source: "weibo", Heat: 100000, SourceURL: "https://s.weibo.com", CreatedAt: now, UpdatedAt: now},
		},
		"douyin": {
			{ID: "douyin-fallback-1", Title: "抖音热点暂不可用", Source: "douyin", Heat: 100000, SourceURL: "https://www.douyin.com", CreatedAt: now, UpdatedAt: now},
		},
		"zhihu": {
			{ID: "zhihu-fallback-1", Title: "知乎热榜暂不可用", Source: "zhihu", Heat: 100000, SourceURL: "https://www.zhihu.com", CreatedAt: now, UpdatedAt: now},
		},
		"toutiao": {
			{ID: "toutiao-fallback-1", Title: "今日头条热点暂不可用", Source: "toutiao", Heat: 100000, SourceURL: "https://www.toutiao.com", CreatedAt: now, UpdatedAt: now},
		},
		"netease": {
			{ID: "netease-fallback-1", Title: "网易新闻暂不可用", Source: "netease", Heat: 100000, SourceURL: "https://news.163.com", CreatedAt: now, UpdatedAt: now},
		},
		"sina": {
			{ID: "sina-fallback-1", Title: "新浪新闻暂不可用", Source: "sina", Heat: 100000, SourceURL: "https://news.sina.com.cn", CreatedAt: now, UpdatedAt: now},
		},
		"qq": {
			{ID: "qq-fallback-1", Title: "腾讯新闻暂不可用", Source: "qq", Heat: 100000, SourceURL: "https://news.qq.com", CreatedAt: now, UpdatedAt: now},
		},
	}

	topics, ok := fallbackTopics[s.sourceID]
	if !ok {
		topics = []hotspot.Topic{
			{ID: fmt.Sprintf("%s-fallback-1", s.sourceID), Title: fmt.Sprintf("%s 热点数据暂不可用", s.name), Source: s.sourceID, Heat: 100000, SourceURL: "", CreatedAt: now, UpdatedAt: now},
		}
	}

	if maxItems > 0 && len(topics) > maxItems {
		topics = topics[:maxItems]
	}

	logrus.Warnf("[%s] using fallback data", s.sourceID)
	return topics
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
	// 使用真实数据源和 NewsNow API 混合模式
	return []hotspot.SourceInterface{
		NewBaiduSource(),           // 百度热搜 - 使用 HTML 解析获取真实数据
		NewNewsNowSource("weibo"),  // 微博热搜 - 使用 NewsNow API
		NewNewsNowSource("douyin"), // 抖音热点 - 使用 NewsNow API
		NewNewsNowSource("zhihu"),  // 知乎热榜 - 使用 NewsNow API
		NewNewsNowSource("toutiao"),// 今日头条 - 使用 NewsNow API
		NewNewsNowSource("netease"),// 网易新闻 - 使用 NewsNow API
		NewNewsNowSource("sina"),   // 新浪新闻 - 使用 NewsNow API
		NewNewsNowSource("qq"),     // 腾讯新闻 - 使用 NewsNow API
	}
}
