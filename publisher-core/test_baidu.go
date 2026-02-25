package main

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strings"
	"time"
)

func main() {
	ctx := context.Background()
	
	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", "https://top.baidu.com/board?tab=realtime", nil)
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("HTTP 错误: %d\n", resp.StatusCode)
		return
	}

	// 读取 HTML 内容
	buf := make([]byte, 1024*1024) // 1MB buffer
	n, err := resp.Body.Read(buf)
	if err != nil && err.Error() != "EOF" {
		fmt.Printf("读取响应失败: %v\n", err)
		return
	}
	htmlContent := string(buf[:n])

	fmt.Printf("HTML 内容长度: %d 字节\n", len(htmlContent))

	// 解析 HTML 提取热点数据
	re := regexp.MustCompile(`<div class="c-single-text-ellipsis"[^>]*>\s*([^<]+)\s*</div>`)
	matches := re.FindAllStringSubmatch(htmlContent, -1)

	fmt.Printf("找到 %d 个匹配项\n", len(matches))

	now := time.Now()
	var titles []string
	for i, match := range matches {
		if i >= 10 { // 只显示前10个
			break
		}

		if len(match) < 2 {
			continue
		}

		title := html.UnescapeString(strings.TrimSpace(match[1]))
		if title == "" {
			continue
		}

		titles = append(titles, title)
		fmt.Printf("%d. %s\n", i+1, title)
	}

	if len(titles) > 0 {
		fmt.Printf("\n成功获取到 %d 个真实热点话题！\n", len(titles))
	} else {
		fmt.Println("\n未找到任何热点话题")
	}
}
