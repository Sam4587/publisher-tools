// Package pipeline 提供高级流水线步骤处理器
package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"publisher-core/ai"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// OutlineGenerator 内容大纲生成处理器
type OutlineGenerator struct {
	aiService *ai.Service
}

// NewOutlineGenerator 创建大纲生成器
func NewOutlineGenerator(aiService *ai.Service) *OutlineGenerator {
	return &OutlineGenerator{aiService: aiService}
}

func (h *OutlineGenerator) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	topic, ok := input["topic"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少 topic 参数")
	}

	targetLength, _ := config["target_length"].(int)
	if targetLength == 0 {
		targetLength = 1000
	}

	style, _ := config["style"].(string)
	if style == "" {
		style = "professional"
	}

	// 构建大纲生成提示词
	prompt := fmt.Sprintf(`请为以下主题生成一个详细的内容大纲：

主题: %s
目标字数: %d字
风格: %s

要求:
1. 生成3-5个主要章节
2. 每个章节包含2-4个子要点
3. 标注每个章节的预估字数
4. 确保逻辑连贯、层次分明

请以JSON格式返回大纲，格式如下:
{
  "title": "文章标题",
  "sections": [
    {
      "title": "章节标题",
      "estimated_words": 200,
      "points": ["要点1", "要点2"]
    }
  ],
  "total_estimated_words": 1000,
  "keywords": ["关键词1", "关键词2"]
}`, topic, targetLength, style)

	// 调用 AI 服务
	result, err := h.aiService.Generate(ctx, &ai.GenerateOptions{
		Model:       "deepseek-chat",
		Prompt:      prompt,
		MaxTokens:   1500,
		Temperature: 0.7,
	})
	if err != nil {
		return nil, fmt.Errorf("大纲生成失败: %w", err)
	}

	// 解析大纲
	var outline map[string]interface{}
	if err := json.Unmarshal([]byte(result.Content), &outline); err != nil {
		// 如果解析失败，返回原始结果
		outline = map[string]interface{}{
			"raw_outline": result.Content,
		}
	}

	return map[string]interface{}{
		"outline":      outline,
		"tokens_used":  result.TokensUsed,
		"generated_at": time.Now().Format(time.RFC3339),
	}, nil
}

// ContentClusterer 内容聚类处理器
type ContentClusterer struct {
	aiService *ai.Service
}

// NewContentClusterer 创建内容聚类器
func NewContentClusterer(aiService *ai.Service) *ContentClusterer {
	return &ContentClusterer{aiService: aiService}
}

func (h *ContentClusterer) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	contents, ok := input["contents"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("缺少 contents 参数")
	}

	clusterCount, _ := config["cluster_count"].(int)
	if clusterCount == 0 {
		clusterCount = 5
	}

	// 提取内容文本
	var contentTexts []string
	for _, c := range contents {
		if content, ok := c.(string); ok {
			contentTexts = append(contentTexts, content)
		} else if contentMap, ok := c.(map[string]interface{}); ok {
			if text, ok := contentMap["content"].(string); ok {
				contentTexts = append(contentTexts, text)
			}
		}
	}

	// 构建聚类提示词
	prompt := fmt.Sprintf(`请将以下内容进行聚类分析，分成%d个类别：

内容列表:
%s

要求:
1. 根据内容主题进行分类
2. 每个类别给出一个描述性名称
3. 列出每个类别包含的内容索引
4. 提取每个类别的关键词

请以JSON格式返回聚类结果:
{
  "clusters": [
    {
      "name": "类别名称",
      "description": "类别描述",
      "indices": [0, 2, 5],
      "keywords": ["关键词1", "关键词2"]
    }
  ]
}`, clusterCount, strings.Join(contentTexts, "\n---\n"))

	// 调用 AI 服务
	result, err := h.aiService.Generate(ctx, &ai.GenerateOptions{
		Model:       "deepseek-chat",
		Prompt:      prompt,
		MaxTokens:   2000,
		Temperature: 0.3,
	})
	if err != nil {
		return nil, fmt.Errorf("内容聚类失败: %w", err)
	}

	// 解析聚类结果
	var clusterResult map[string]interface{}
	if err := json.Unmarshal([]byte(result.Content), &clusterResult); err != nil {
		clusterResult = map[string]interface{}{
			"raw_result": result.Content,
		}
	}

	return map[string]interface{}{
		"clusters":     clusterResult,
		"tokens_used":  result.TokensUsed,
		"clustered_at": time.Now().Format(time.RFC3339),
	}, nil
}

// ContentClassifier 内容分类处理器
type ContentClassifier struct {
	aiService *ai.Service
}

// NewContentClassifier 创建内容分类器
func NewContentClassifier(aiService *ai.Service) *ContentClassifier {
	return &ContentClassifier{aiService: aiService}
}

func (h *ContentClassifier) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	content, ok := input["content"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少 content 参数")
	}

	categories, ok := config["categories"].([]string)
	if !ok {
		categories = []string{"科技", "娱乐", "教育", "生活", "财经", "体育", "健康", "其他"}
	}

	// 构建分类提示词
	prompt := fmt.Sprintf(`请对以下内容进行分类：

内容:
%s

可选类别: %s

要求:
1. 选择最匹配的类别
2. 给出置信度（0-1）
3. 提取内容的关键词
4. 判断内容情感倾向

请以JSON格式返回分类结果:
{
  "category": "科技",
  "confidence": 0.95,
  "keywords": ["人工智能", "机器学习"],
  "sentiment": "positive",
  "sentiment_score": 0.8
}`, content, strings.Join(categories, ", "))

	// 调用 AI 服务
	result, err := h.aiService.Generate(ctx, &ai.GenerateOptions{
		Model:       "deepseek-chat",
		Prompt:      prompt,
		MaxTokens:   500,
		Temperature: 0.2,
	})
	if err != nil {
		return nil, fmt.Errorf("内容分类失败: %w", err)
	}

	// 解析分类结果
	var classifyResult map[string]interface{}
	if err := json.Unmarshal([]byte(result.Content), &classifyResult); err != nil {
		classifyResult = map[string]interface{}{
			"raw_result": result.Content,
		}
	}

	return map[string]interface{}{
		"classification": classifyResult,
		"tokens_used":    result.TokensUsed,
		"classified_at":  time.Now().Format(time.RFC3339),
	}, nil
}

// TitleGenerator 标题生成处理器
type TitleGenerator struct {
	aiService *ai.Service
}

// NewTitleGenerator 创建标题生成器
func NewTitleGenerator(aiService *ai.Service) *TitleGenerator {
	return &TitleGenerator{aiService: aiService}
}

func (h *TitleGenerator) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	content, ok := input["content"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少 content 参数")
	}

	count, _ := config["count"].(int)
	if count == 0 {
		count = 5
	}

	maxLength, _ := config["max_length"].(int)
	if maxLength == 0 {
		maxLength = 30
	}

	// 构建标题生成提示词
	prompt := fmt.Sprintf(`请为以下内容生成%d个吸引人的标题：

内容:
%s

要求:
1. 标题长度不超过%d字
2. 标题要有吸引力
3. 适合社交媒体传播
4. 包含适当的emoji表情
5. 风格多样化

请以JSON格式返回标题列表:
{
  "titles": [
    {
      "title": "标题1",
      "style": "question",
      "score": 0.95
    }
  ]
}`, count, content, maxLength)

	// 调用 AI 服务
	result, err := h.aiService.Generate(ctx, &ai.GenerateOptions{
		Model:       "deepseek-chat",
		Prompt:      prompt,
		MaxTokens:   1000,
		Temperature: 0.8,
	})
	if err != nil {
		return nil, fmt.Errorf("标题生成失败: %w", err)
	}

	// 解析标题结果
	var titleResult map[string]interface{}
	if err := json.Unmarshal([]byte(result.Content), &titleResult); err != nil {
		titleResult = map[string]interface{}{
			"raw_result": result.Content,
		}
	}

	return map[string]interface{}{
		"titles":       titleResult,
		"tokens_used":  result.TokensUsed,
		"generated_at": time.Now().Format(time.RFC3339),
	}, nil
}

// DescriptionGenerator 描述生成处理器
type DescriptionGenerator struct {
	aiService *ai.Service
}

// NewDescriptionGenerator 创建描述生成器
func NewDescriptionGenerator(aiService *ai.Service) *DescriptionGenerator {
	return &DescriptionGenerator{aiService: aiService}
}

func (h *DescriptionGenerator) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	content, ok := input["content"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少 content 参数")
	}

	maxLength, _ := config["max_length"].(int)
	if maxLength == 0 {
		maxLength = 200
	}

	includeKeywords, _ := config["include_keywords"].(bool)

	// 构建描述生成提示词
	prompt := fmt.Sprintf(`请为以下内容生成一个简洁的描述：

内容:
%s

要求:
1. 描述长度不超过%d字
2. 概括内容核心要点
3. 吸引读者阅读
%s
请直接输出描述内容。`, content, maxLength, map[bool]string{true: "4. 包含关键词\n"}[includeKeywords])

	// 调用 AI 服务
	result, err := h.aiService.Generate(ctx, &ai.GenerateOptions{
		Model:       "deepseek-chat",
		Prompt:      prompt,
		MaxTokens:   300,
		Temperature: 0.6,
	})
	if err != nil {
		return nil, fmt.Errorf("描述生成失败: %w", err)
	}

	return map[string]interface{}{
		"description":  result.Content,
		"tokens_used":  result.TokensUsed,
		"generated_at": time.Now().Format(time.RFC3339),
	}, nil
}

// ContentFilter 内容筛选处理器
type ContentFilter struct {
	aiService *ai.Service
}

// NewContentFilter 创建内容筛选器
func NewContentFilter(aiService *ai.Service) *ContentFilter {
	return &ContentFilter{aiService: aiService}
}

func (h *ContentFilter) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	contents, ok := input["contents"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("缺少 contents 参数")
	}

	minScore, _ := config["min_score"].(float64)
	if minScore == 0 {
		minScore = 0.7
	}

	filterCriteria, _ := config["filter_criteria"].([]string)
	if len(filterCriteria) == 0 {
		filterCriteria = []string{"quality", "relevance", "originality"}
	}

	// 对每个内容进行评分
	var filteredContents []map[string]interface{}
	var scores []map[string]interface{}

	for i, c := range contents {
		var contentText string
		if content, ok := c.(string); ok {
			contentText = content
		} else if contentMap, ok := c.(map[string]interface{}); ok {
			if text, ok := contentMap["content"].(string); ok {
				contentText = text
			}
		}

		// 构建评分提示词
		prompt := fmt.Sprintf(`请对以下内容进行评分：

内容:
%s

评分维度: %s

请以JSON格式返回评分结果:
{
  "overall_score": 0.85,
  "quality": 0.9,
  "relevance": 0.85,
  "originality": 0.8,
  "reasoning": "评分理由"
}`, contentText, strings.Join(filterCriteria, ", "))

		// 调用 AI 服务
		result, err := h.aiService.Generate(ctx, &ai.GenerateOptions{
			Model:       "deepseek-chat",
			Prompt:      prompt,
			MaxTokens:   300,
			Temperature: 0.2,
		})
		if err != nil {
			logrus.Warnf("内容 %d 评分失败: %v", i, err)
			continue
		}

		// 解析评分结果
		var scoreResult map[string]interface{}
		if err := json.Unmarshal([]byte(result.Content), &scoreResult); err != nil {
			continue
		}

		overallScore, _ := scoreResult["overall_score"].(float64)
		scores = append(scores, map[string]interface{}{
			"index": i,
			"score": overallScore,
			"details": scoreResult,
		})

		// 筛选符合条件的内容
		if overallScore >= minScore {
			filteredContents = append(filteredContents, map[string]interface{}{
				"index":   i,
				"content": c,
				"score":   overallScore,
			})
		}
	}

	// 按分数排序
	sort.Slice(filteredContents, func(i, j int) bool {
		scoreI, _ := filteredContents[i]["score"].(float64)
		scoreJ, _ := filteredContents[j]["score"].(float64)
		return scoreI > scoreJ
	})

	return map[string]interface{}{
		"filtered_contents": filteredContents,
		"scores":            scores,
		"total_count":       len(contents),
		"filtered_count":    len(filteredContents),
		"filtered_at":       time.Now().Format(time.RFC3339),
	}, nil
}

// ConditionalExecutor 条件执行处理器
type ConditionalExecutor struct {
	orchestrator *PipelineOrchestrator
}

// NewConditionalExecutor 创建条件执行器
func NewConditionalExecutor(orchestrator *PipelineOrchestrator) *ConditionalExecutor {
	return &ConditionalExecutor{orchestrator: orchestrator}
}

func (h *ConditionalExecutor) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	condition, ok := config["condition"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少 condition 参数")
	}

	trueStep, hasTrueStep := config["true_step"].(map[string]interface{})
	falseStep, hasFalseStep := config["false_step"].(map[string]interface{})

	// 评估条件
	result, err := h.evaluateCondition(condition, input)
	if err != nil {
		return nil, fmt.Errorf("条件评估失败: %w", err)
	}

	var executedStep string
	var stepOutput map[string]interface{}

	if result {
		if hasTrueStep {
			executedStep = "true_step"
			stepOutput, err = h.executeStep(ctx, trueStep, input)
		}
	} else {
		if hasFalseStep {
			executedStep = "false_step"
			stepOutput, err = h.executeStep(ctx, falseStep, input)
		}
	}

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"condition_result": result,
		"executed_step":    executedStep,
		"step_output":      stepOutput,
		"executed_at":      time.Now().Format(time.RFC3339),
	}, nil
}

func (h *ConditionalExecutor) evaluateCondition(condition string, input map[string]interface{}) (bool, error) {
	// 简单的条件评估逻辑
	// 支持: field == value, field != value, field > value, field < value
	// 支持: &&, || 逻辑运算

	// 这里实现一个简单的条件解析器
	// 实际项目中可以使用更完善的表达式引擎

	// 示例: "score >= 0.7"
	if strings.Contains(condition, ">=") {
		parts := strings.Split(condition, ">=")
		if len(parts) == 2 {
			field := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			fieldValue, ok := input[field]
			if !ok {
				return false, nil
			}

			// 尝试数值比较
			if fieldNum, ok := fieldValue.(float64); ok {
				var valueNum float64
				fmt.Sscanf(value, "%f", &valueNum)
				return fieldNum >= valueNum, nil
			}
		}
	}

	// 默认返回 true
	return true, nil
}

func (h *ConditionalExecutor) executeStep(ctx context.Context, stepConfig map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	handlerName, ok := stepConfig["handler"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少 handler 参数")
	}

	stepConfigMap, ok := stepConfig["config"].(map[string]interface{})
	if !ok {
		stepConfigMap = make(map[string]interface{})
	}

	handler, exists := h.orchestrator.stepHandlers[handlerName]
	if !exists {
		return nil, fmt.Errorf("未找到步骤处理器: %s", handlerName)
	}

	return handler.Execute(ctx, stepConfigMap, input)
}

// ParallelExecutor 并行执行处理器
type ParallelExecutor struct {
	orchestrator *PipelineOrchestrator
}

// NewParallelExecutor 创建并行执行器
func NewParallelExecutor(orchestrator *PipelineOrchestrator) *ParallelExecutor {
	return &ParallelExecutor{orchestrator: orchestrator}
}

func (h *ParallelExecutor) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	steps, ok := config["steps"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("缺少 steps 参数")
	}

	maxParallel, _ := config["max_parallel"].(int)
	if maxParallel == 0 {
		maxParallel = 5
	}

	// 创建结果通道
	type stepResult struct {
		index  int
		output map[string]interface{}
		err    error
	}
	resultChan := make(chan stepResult, len(steps))

	// 使用信号量控制并发数
	sem := make(chan struct{}, maxParallel)

	// 并行执行步骤
	for i, step := range steps {
		stepConfig, ok := step.(map[string]interface{})
		if !ok {
			continue
		}

		go func(index int, cfg map[string]interface{}) {
			sem <- struct{}{}
			defer func() { <-sem }()

			handlerName, ok := cfg["handler"].(string)
			if !ok {
				resultChan <- stepResult{index: index, err: fmt.Errorf("缺少 handler 参数")}
				return
			}

			stepConfigMap, ok := cfg["config"].(map[string]interface{})
			if !ok {
				stepConfigMap = make(map[string]interface{})
			}

			handler, exists := h.orchestrator.stepHandlers[handlerName]
			if !exists {
				resultChan <- stepResult{index: index, err: fmt.Errorf("未找到步骤处理器: %s", handlerName)}
				return
			}

			output, err := handler.Execute(ctx, stepConfigMap, input)
			resultChan <- stepResult{index: index, output: output, err: err}
		}(i, stepConfig)
	}

	// 收集结果
	results := make([]map[string]interface{}, len(steps))
	errors := make([]error, 0)

	for i := 0; i < len(steps); i++ {
		result := <-resultChan
		if result.err != nil {
			errors = append(errors, result.err)
			results[result.index] = map[string]interface{}{
				"error": result.err.Error(),
			}
		} else {
			results[result.index] = result.output
		}
	}

	return map[string]interface{}{
		"results":     results,
		"errors":      errors,
		"total_count": len(steps),
		"success_count": len(steps) - len(errors),
		"executed_at": time.Now().Format(time.RFC3339),
	}, nil
}
