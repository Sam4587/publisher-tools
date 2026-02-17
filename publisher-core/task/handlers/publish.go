package handlers

import (
	"context"
	"fmt"

	"publisher-core/adapters"
	publisher "publisher-core/interfaces"
	"publisher-core/task"
	"github.com/sirupsen/logrus"
)

type PublishHandler struct {
	factory *adapters.PublisherFactory
}

func NewPublishHandler(factory *adapters.PublisherFactory) *PublishHandler {
	return &PublishHandler{factory: factory}
}

func (h *PublishHandler) Handle(ctx context.Context, t *task.Task) error {
	logrus.Infof("Starting publish task: %s, platform: %s", t.ID, t.Platform)

	platform, ok := t.Payload["platform"].(string)
	if !ok {
		return fmt.Errorf("invalid platform in payload")
	}

	title, _ := t.Payload["title"].(string)
	content, _ := t.Payload["content"].(string)
	contentType, _ := t.Payload["type"].(string)

	var images []string
	if imgs, ok := t.Payload["images"].([]interface{}); ok {
		for _, img := range imgs {
			if s, ok := img.(string); ok {
				images = append(images, s)
			}
		}
	}

	video, _ := t.Payload["video"].(string)

	var tags []string
	if ts, ok := t.Payload["tags"].([]interface{}); ok {
		for _, tag := range ts {
			if s, ok := tag.(string); ok {
				tags = append(tags, s)
			}
		}
	}

	logrus.Infof("Publish content: platform=%s, type=%s, title=%s, content_len=%d, images=%d, video=%s, tags=%d",
		platform, contentType, title, len(content), len(images), video, len(tags))

	pub, err := h.factory.Create(platform, publisher.DefaultOptions())
	if err != nil {
		logrus.Errorf("Create publisher failed: %v", err)
		return fmt.Errorf("create publisher failed: %w", err)
	}

	publishContent := &publisher.Content{
		Type:       publisher.ContentType(contentType),
		Title:      title,
		Body:       content,
		ImagePaths: images,
		VideoPath:  video,
		Tags:       tags,
	}

	result, err := pub.Publish(ctx, publishContent)
	if err != nil {
		logrus.Errorf("Publish failed: %v", err)
		t.Result = map[string]interface{}{
			"platform": platform,
			"title":    title,
			"status":   "failed",
			"error":    err.Error(),
		}
		return err
	}

	t.Result = map[string]interface{}{
		"platform":   platform,
		"title":      title,
		"type":       contentType,
		"task_id":    result.TaskID,
		"status":     string(result.Status),
		"post_id":    result.PostID,
		"post_url":   result.PostURL,
		"created_at": result.CreatedAt,
	}

	if result.FinishedAt != nil {
		t.Result["finished_at"] = result.FinishedAt
	}

	logrus.Infof("Publish task completed: %s, status: %s", t.ID, result.Status)
	return nil
}
