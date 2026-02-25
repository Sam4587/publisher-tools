package sources

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"publisher-core/hotspot"

	"github.com/sirupsen/logrus"
)

type BaiduSource struct {
	enabled bool
	client  *http.Client
}

func NewBaiduSource() *BaiduSource {
	return &BaiduSource{
		enabled: true,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *BaiduSource) Name() string {
	return "baidu"
}

func (s *BaiduSource) DisplayName() string {
	return "百度热搜"
}

func (s *BaiduSource) ID() string {
	return "baidu"
}

func (s *BaiduSource) IsEnabled() bool {
	return s.enabled
}

func (s *BaiduSource) SetEnabled(enabled bool) {
	s.enabled = enabled
}

func (s *BaiduSource) Fetch(ctx context.Context, limit int) ([]hotspot.Topic, error) {
	if !s.enabled {
		return nil, fmt.Errorf("source is disabled")
	}

	logrus.Infof("[Baidu] Fetching hot topics, limit=%d", limit)

	// 尝试获取真实数据
	topics, err := s.fetchRealData(ctx, limit)
	if err != nil {
		logrus.Warnf("[Baidu] Failed to fetch real data, using fallback: %v", err)
		return s.getFallbackTopics(limit), nil
	}

	logrus.Infof("[Baidu] Fetched %d topics", len(topics))
	return topics, nil
}

func (s *BaiduSource) fetchRealData(ctx context.Context, limit int) ([]hotspot.Topic, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://top.baidu.com/board?tab=realtime", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// 读取 HTML 内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	htmlContent := string(body)

	// 解析 HTML 提取热点数据
	topics, err := s.parseHTML(htmlContent, limit)
	if err != nil {
		return nil, err
	}

	if len(topics) == 0 {
		return nil, fmt.Errorf("no topics found")
	}

	return topics, nil
}

func (s *BaiduSource) parseHTML(htmlContent string, limit int) ([]hotspot.Topic, error) {
	now := time.Now()

	// 使用正则表达式提取热点标题
	// 匹配格式: <div class="c-single-text-ellipsis">  标题 </div>
	re := regexp.MustCompile(`<div class="c-single-text-ellipsis"[^>]*>\s*([^<]+)\s*</div>`)
	matches := re.FindAllStringSubmatch(htmlContent, -1)

	var topics []hotspot.Topic
	for i, match := range matches {
		if limit > 0 && i >= limit {
			break
		}

		if len(match) < 2 {
			continue
		}

		title := html.UnescapeString(strings.TrimSpace(match[1]))
		if title == "" {
			continue
		}

		// 计算热度值（排名越靠前热度越高）
		heat := 1000000 - (i * 50000)
		if heat < 100000 {
			heat = 100000
		}

		topic := hotspot.Topic{
			ID:        fmt.Sprintf("baidu-%d", i+1),
			Title:     title,
			Source:    "baidu",
			Heat:      heat,
			SourceURL: "https://top.baidu.com/board?tab=realtime",
			CreatedAt: now,
			UpdatedAt: now,
		}

		topics = append(topics, topic)
	}

	return topics, nil
}

func (s *BaiduSource) getFallbackTopics(limit int) []hotspot.Topic {
	now := time.Now()
	topics := []hotspot.Topic{
		{
			ID:        "baidu-1",
			Title:     "百度热搜示例 1",
			Source:    "baidu",
			Heat:      1000000,
			SourceURL: "https://top.baidu.com",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "baidu-2",
			Title:     "百度热搜示例 2",
			Source:    "baidu",
			Heat:      900000,
			SourceURL: "https://top.baidu.com",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	if limit > 0 && len(topics) > limit {
		topics = topics[:limit]
	}

	return topics
}
