package toutiao

import (
	"context"
	"math/rand"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	PublishURL = "https://mp.toutiao.com/profile_v4/pub_article"
)

type PublishImageContent struct {
	Title        string
	Content      string
	ImagePaths   []string
	Tags         []string
	ScheduleTime *time.Time
}

type PublishVideoContent struct {
	Title        string
	Content      string
	VideoPath    string
	Tags         []string
	ScheduleTime *time.Time
}

type PublishAction struct {
	page *rod.Page
}

func NewPublishAction(page *rod.Page) (*PublishAction, error) {
	pp := page.Timeout(300 * time.Second)

	if err := pp.Navigate(PublishURL); err != nil {
		return nil, errors.Wrap(err, "导航到发布页面失败")
	}

	if err := pp.WaitLoad(); err != nil {
		logrus.Warnf("等待页面加载出现问题: %v，继续尝试", err)
	}
	time.Sleep(2 * time.Second)

	if err := pp.WaitDOMStable(time.Second, 0.1); err != nil {
		logrus.Warnf("等待 DOM 稳定出现问题: %v，继续尝试", err)
	}
	time.Sleep(1 * time.Second)

	return &PublishAction{page: pp}, nil
}

func (p *PublishAction) PublishImages(ctx context.Context, content PublishImageContent) error {
	if len(content.ImagePaths) == 0 {
		return errors.New("图片不能为空")
	}

	page := p.page.Context(ctx)

	if err := uploadImages(page, content.ImagePaths); err != nil {
		return errors.Wrap(err, "上传图片失败")
	}

	if err := fillContent(page, content.Title, content.Content, content.Tags); err != nil {
		return errors.Wrap(err, "填写内容失败")
	}

	if err := submitPublish(page); err != nil {
		return errors.Wrap(err, "提交发布失败")
	}

	return nil
}

func (p *PublishAction) PublishVideo(ctx context.Context, content PublishVideoContent) error {
	if content.VideoPath == "" {
		return errors.New("视频不能为空")
	}

	page := p.page.Context(ctx)

	if err := uploadVideo(page, content.VideoPath); err != nil {
		return errors.Wrap(err, "上传视频失败")
	}

	if err := fillContent(page, content.Title, content.Content, content.Tags); err != nil {
		return errors.Wrap(err, "填写内容失败")
	}

	if err := submitPublish(page); err != nil {
		return errors.Wrap(err, "提交发布失败")
	}

	return nil
}

func uploadImages(page *rod.Page, imagePaths []string) error {
	for i, imagePath := range imagePaths {
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			return errors.Wrapf(err, "图片文件不存在: %s", imagePath)
		}

		logrus.Infof("上传图片 %d/%d: %s", i+1, len(imagePaths), imagePath)

		fileInput, err := page.Element("input[type='file'][accept*='image']")
		if err != nil {
			return errors.Wrap(err, "查找图片上传输入框失败")
		}

		fileInput.MustSetFiles(imagePath)

		waitTime := time.Duration(1+rand.Intn(2)) * time.Second
		time.Sleep(waitTime)
	}

	return nil
}

func uploadVideo(page *rod.Page, videoPath string) error {
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return errors.Wrapf(err, "视频文件不存在: %s", videoPath)
	}

	logrus.Infof("上传视频: %s", videoPath)

	fileInput, err := page.Element("input[type='file'][accept*='video']")
	if err != nil {
		return errors.Wrap(err, "查找视频上传输入框失败")
	}

	fileInput.MustSetFiles(videoPath)

	logrus.Info("等待视频上传完成...")
	time.Sleep(3 * time.Second)

	return nil
}

func fillContent(page *rod.Page, title, content string, tags []string) error {
	titleInput, err := page.Element("input[placeholder*='标题'], input[name*='title']")
	if err != nil {
		return errors.Wrap(err, "查找标题输入框失败")
	}

	if err := titleInput.Input(title); err != nil {
		return errors.Wrap(err, "输入标题失败")
	}

	time.Sleep(time.Duration(500+rand.Intn(1000)) * time.Millisecond)

	contentInput, err := page.Element("textarea[placeholder*='正文'], textarea[name*='content']")
	if err != nil {
		return errors.Wrap(err, "查找正文输入框失败")
	}

	if err := contentInput.Input(content); err != nil {
		return errors.Wrap(err, "输入正文失败")
	}

	time.Sleep(time.Duration(500+rand.Intn(1000)) * time.Millisecond)

	if len(tags) > 0 {
		if err := inputTags(page, tags); err != nil {
			return errors.Wrap(err, "输入标签失败")
		}
	}

	return nil
}

func inputTags(page *rod.Page, tags []string) error {
	for _, tag := range tags {
		tagInput, err := page.Element("input[placeholder*='标签'], input[name*='tag']")
		if err != nil {
			logrus.Warnf("查找标签输入框失败: %v", err)
			continue
		}

		tagInput.MustInput("#" + tag)

		waitTime := time.Duration(300+rand.Intn(700)) * time.Millisecond
		time.Sleep(waitTime)
	}

	return nil
}

func submitPublish(page *rod.Page) error {
	publishBtn, err := page.Element("button[type='submit'], .publish-btn")
	if err != nil {
		return errors.Wrap(err, "查找发布按钮失败")
	}

	vis, err := publishBtn.Visible()
	if err != nil {
		return errors.Wrap(err, "检查按钮可见性失败")
	}

	if !vis {
		return errors.New("发布按钮不可见")
	}

	time.Sleep(time.Duration(1+rand.Intn(2)) * time.Second)

	if err := publishBtn.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return errors.Wrap(err, "点击发布按钮失败")
	}

	time.Sleep(3 * time.Second)

	return nil
}
