package prompts

import (
	"fmt"
	"strings"

	"publisher-core/ai/provider"
)

const (
	RoleContentCreator = "You are a professional content creator skilled in writing engaging articles and social media content."
	RoleHotspotAnalyst = "You are a hotspot analysis expert skilled in analyzing news hotspots, extracting key information, and determining trend directions."
	RoleContentAuditor = "You are a content review expert skilled in identifying sensitive information, violations, and potential risks in content."
	RoleSEOExpert      = "You are an SEO optimization expert skilled in optimizing content to improve search engine rankings and social media exposure."
)

type PromptTemplate struct {
	System string
	User   string
}

var Templates = map[string]PromptTemplate{
	"generate_content": {
		System: RoleContentCreator,
		User: `Generate content based on the following requirements:

Topic: {{.Topic}}
Platform: {{.Platform}}
Style: {{.Style}}
Word count: Around {{.Length}} words

Generate content suitable for {{.Platform}} platform in {{.Style}} style, including title and body.
Format:
Title: [Title content]
Body:
[Body content]`,
	},

	"rewrite_content": {
		System: RoleContentCreator,
		User: `Rewrite the following content in {{.Style}} style, suitable for {{.Platform}} platform:

Original:
{{.Content}}

Requirements:
1. Keep the core meaning unchanged
2. Change expression and language style
3. Word count around {{.Length}} words
4. Comply with {{.Platform}} platform content guidelines

Output the rewritten content directly.`,
	},

	"expand_content": {
		System: RoleContentCreator,
		User: `Expand the following content to make it richer and more detailed:

Original:
{{.Content}}

Requirements:
1. Keep the core meaning and style
2. Add details, examples or arguments
3. Expanded word count around {{.Length}} words
4. Suitable for {{.Platform}} platform

Output the complete expanded content.`,
	},

	"summarize_content": {
		System: RoleContentCreator,
		User: `Summarize the following content:

{{.Content}}

Requirements:
1. Extract core points
2. Summary word count within {{.Length}} words
3. Concise and clear language

Output the summary content.`,
	},

	"analyze_hotspot": {
		System: RoleHotspotAnalyst,
		User: `Analyze the following hotspot topic:

Title: {{.Title}}
Content: {{.Content}}

Analyze from the following dimensions:
1. Event summary (within 50 words)
2. Key points (3-5 points)
3. Sentiment (positive/negative/neutral)
4. Relevance score (1-10, indicating relevance to general users)
5. Content creation suggestions (2-3 suggestions)
6. Recommended tags (3-5 tags)

Output in JSON format:
{
  "summary": "Event summary",
  "key_points": ["Point 1", "Point 2"],
  "sentiment": "Sentiment",
  "relevance": Score number,
  "suggestions": ["Suggestion 1", "Suggestion 2"],
  "tags": ["Tag 1", "Tag 2"]
}`,
	},

	"audit_content": {
		System: RoleContentAuditor,
		User: `Review the following content for issues:

{{.Content}}

Check the following aspects:
1. Whether it contains sensitive words or violations
2. Whether there are factual errors
3. Whether there are inappropriate expressions
4. Whether it is suitable for public platform publishing

Output in JSON format:
{
  "passed": true/false,
  "issues": ["Issue 1", "Issue 2"],
  "suggestions": ["Suggestion 1", "Suggestion 2"],
  "score": Compliance score (0-100)
}`,
	},

	"extract_keywords": {
		System: RoleSEOExpert,
		User: `Extract keywords from the following content:

{{.Content}}

Requirements:
1. Extract 5-10 core keywords
2. Keywords should have search value
3. Suitable for use as tags

Output keywords list in JSON array format.`,
	},

	"generate_title": {
		System: RoleSEOExpert,
		User: `Generate 3 attractive titles for the following content:

{{.Content}}

Platform: {{.Platform}}

Requirements:
1. Eye-catching but not clickbait
2. Match {{.Platform}} platform characteristics
3. Each title no more than 30 characters

Output title list in JSON array format.`,
	},
}

func GetTemplate(name string) (PromptTemplate, bool) {
	t, ok := Templates[name]
	return t, ok
}

func BuildPrompt(templateName string, vars map[string]string) ([]provider.Message, error) {
	tmpl, ok := Templates[templateName]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", templateName)
	}

	user := tmpl.User

	for k, v := range vars {
		placeholder := fmt.Sprintf("{{.%s}}", k)
		user = strings.ReplaceAll(user, placeholder, v)
	}

	return []provider.Message{
		{Role: provider.RoleSystem, Content: tmpl.System},
		{Role: provider.RoleUser, Content: user},
	}, nil
}
