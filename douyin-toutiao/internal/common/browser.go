package common

import (
	"context"
	"math/rand"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	DefaultDelayMin   = 300 * time.Millisecond
	DefaultDelayMax   = 2000 * time.Millisecond
	DOMStableWait     = 1 * time.Second
	DOMStableInterval = 100 * time.Millisecond
)

func WaitRandomDelay() {
	delay := DefaultDelayMin + time.Duration(rand.Int63n(int64(DefaultDelayMax-DefaultDelayMin)))
	logrus.Debugf("随机延迟: %v", delay)
	time.Sleep(delay)
}

func WaitDOMStable(page *rod.Page) error {
	timeout := 30 * time.Second
	interval := DOMStableInterval
	start := time.Now()

	for time.Since(start) < timeout {
		isStable, err := page.Eval(`(() => {
			const elements = document.querySelectorAll('*');
			return elements.length === document.body.querySelectorAll('*').length;
		})()`)

		if err == nil {
			if isStable.Value.Bool() {
				logrus.Debug("DOM 已稳定")
				return nil
			}
		}

		time.Sleep(interval)
	}

	return errors.New("等待 DOM 稳定超时")
}

func InputWithRandomDelay(page *rod.Page, elem *rod.Element, text string) error {
	if err := elem.Input(text); err != nil {
		return errors.Wrap(err, "输入失败")
	}

	WaitRandomDelay()
	return nil
}

func ClickWithRandomDelay(page *rod.Page, selector string) error {
	elem, err := page.Element(selector)
	if err != nil {
		return errors.Wrap(err, "查找元素失败: "+selector)
	}

	if err := elem.Click("left", 1); err != nil {
		return errors.Wrap(err, "点击元素失败: "+selector)
	}

	WaitRandomDelay()
	return nil
}

func WaitForElement(page *rod.Page, selector string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return errors.New("等待元素超时")
		default:
			if _, err := page.Element(selector); err == nil {
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func SafeNavigate(page *rod.Page, url string, timeout time.Duration) error {
	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := page.Navigate(url); err != nil {
		return errors.Wrap(err, "导航到页面失败")
	}

	page.Timeout(timeout).MustWaitLoad()
	WaitRandomDelay()

	return nil
}

// CheckLoginWithRetry 带重试的登录检查
func CheckLoginWithRetry(ctx context.Context, checker func() (bool, error), maxRetries int) (bool, error) {
	for i := 0; i < maxRetries; i++ {
		loggedIn, err := checker()
		if err != nil {
			logrus.Warnf("检查登录状态失败 (尝试 %d/%d): %v", i+1, maxRetries, err)
			if i < maxRetries-1 {
				time.Sleep(time.Duration(2+i) * time.Second)
				continue
			}
			return false, err
		}
		return loggedIn, nil
	}
	return false, errors.New("重试次数用尽")
}

// ExtractCookiesWithRetry 带重试的 Cookie 提取
func ExtractCookiesWithRetry(ctx context.Context, extractor func() (map[string]string, error), maxRetries int) (map[string]string, error) {
	for i := 0; i < maxRetries; i++ {
		cookies, err := extractor()
		if err != nil {
			logrus.Warnf("提取 cookies 失败 (尝试 %d/%d): %v", i+1, maxRetries, err)
			if i < maxRetries-1 {
				time.Sleep(time.Duration(2+i) * time.Second)
				continue
			}
			return nil, err
		}
		return cookies, nil
	}
	return nil, errors.New("重试次数用尽")
}

// ValidateImagePaths 验证图片路径
func ValidateImagePaths(paths []string) error {
	if len(paths) == 0 {
		return errors.New("图片路径列表不能为空")
	}

	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return errors.Wrapf(err, "图片文件不存在: %s", path)
		}
	}

	return nil
}

// ValidateVideoPath 验证视频路径
func ValidateVideoPath(path string) error {
	if path == "" {
		return errors.New("视频路径不能为空")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.Wrapf(err, "视频文件不存在: %s", path)
	}

	return nil
}
