package browser

import (
	"sync"

	"github.com/go-rod/rod"
	"github.com/sirupsen/logrus"
)

var (
	instance *rod.Browser
	once     sync.Once
	headless bool
)

func NewBrowser(isHeadless bool) *rod.Browser {
	once.Do(func() {
		headless = isHeadless
		instance = rod.New()
		if isHeadless {
			instance = instance.MustConnect()
		} else {
			instance = instance.MustConnect()
		}
		logrus.Info("浏览器初始化成功")
	})
	return instance
}

func GetBrowser() *rod.Browser {
	if instance == nil {
		NewBrowser(true)
	}
	return instance
}

func Close() {
	if instance != nil {
		instance.MustClose()
		instance = nil
		logrus.Info("浏览器已关闭")
	}
}
