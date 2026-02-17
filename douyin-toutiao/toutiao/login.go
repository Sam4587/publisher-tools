package toutiao

import (
	"context"
	"time"

	"github.com/go-rod/rod"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	LoginURL = "https://mp.toutiao.com/"
)

type LoginAction struct {
	page *rod.Page
}

func NewLogin(page *rod.Page) *LoginAction {
	return &LoginAction{page: page}
}

func (a *LoginAction) CheckLoginStatus(ctx context.Context) (bool, error) {
	pp := a.page.Context(ctx)

	pp.MustNavigate(LoginURL).MustWaitLoad()
	time.Sleep(2 * time.Second)

	exists, _, err := pp.Has(".user-avatar")
	if err != nil {
		return false, errors.Wrap(err, "检查登录状态失败")
	}

	return exists, nil
}

func (a *LoginAction) FetchQrcodeImage(ctx context.Context) (string, bool, error) {
	pp := a.page.Context(ctx)

	pp.MustNavigate(LoginURL).MustWaitLoad()
	time.Sleep(2 * time.Second)

	if exists, _, _ := pp.Has(".user-avatar"); exists {
		return "", true, nil
	}

	qrcodeElem, err := pp.Element(".qrcode-img, .qr-code")
	if err != nil {
		return "", false, errors.Wrap(err, "查找二维码元素失败")
	}

	qrcodeSrc, err := qrcodeElem.Attribute("src")
	if err != nil {
		return "", false, errors.Wrap(err, "获取二维码链接失败")
	}

	if qrcodeSrc == nil || *qrcodeSrc == "" {
		return "", false, errors.New("二维码链接为空")
	}

	return *qrcodeSrc, false, nil
}

func (a *LoginAction) WaitForLogin(ctx context.Context) bool {
	pp := a.page.Context(ctx)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			el, err := pp.Element(".user-avatar")
			if err == nil && el != nil {
				return true
			}
		}
	}
}

func (a *LoginAction) ExtractCookies(ctx context.Context) (map[string]string, error) {
	pp := a.page.Context(ctx)

	cookiesMap := make(map[string]string)

	cookies, err := pp.Cookies([]string{})
	if err != nil {
		return nil, errors.Wrap(err, "获取 cookies 失败")
	}

	for _, cookie := range cookies {
		switch cookie.Name {
		case "sessionid", "passport_auth", "tt_token", "tt_webid":
			cookiesMap[cookie.Name] = cookie.Value
		}
	}

	if len(cookiesMap) == 0 {
		return nil, errors.New("未找到有效的 cookies")
	}

	logrus.Infof("提取到 %d 个 cookies", len(cookiesMap))
	return cookiesMap, nil
}
