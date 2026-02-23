package middleware

import (
	"net/http"
	"strconv"
)

const (
	// DefaultLimit 默认查询限制
	DefaultLimit = 50
	// MaxLimit 最大查询限制
	MaxLimit = 1000
	// DefaultOffset 默认偏移量
	DefaultOffset = 0
)

// QueryLimitMiddleware 查询参数限制中间件
func QueryLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		// 处理limit参数
		if limitStr := query.Get("limit"); limitStr != "" {
			limit, err := strconv.Atoi(limitStr)
			if err != nil || limit < 0 {
				// 如果limit无效，使用默认值
				query.Set("limit", strconv.Itoa(DefaultLimit))
			} else if limit > MaxLimit {
				// 如果limit超过最大值，使用最大值
				query.Set("limit", strconv.Itoa(MaxLimit))
			}
		} else {
			// 如果没有limit参数，设置默认值
			query.Set("limit", strconv.Itoa(DefaultLimit))
		}

		// 处理offset参数
		if offsetStr := query.Get("offset"); offsetStr != "" {
			offset, err := strconv.Atoi(offsetStr)
			if err != nil || offset < 0 {
				// 如果offset无效，使用默认值
				query.Set("offset", strconv.Itoa(DefaultOffset))
			}
		} else {
			// 如果没有offset参数，设置默认值
			query.Set("offset", strconv.Itoa(DefaultOffset))
		}

		// 更新请求的查询参数
		r.URL.RawQuery = query.Encode()

		next.ServeHTTP(w, r)
	})
}

// ParseQueryLimit 解析查询限制参数
func ParseQueryLimit(query map[string][]string) (limit, offset int) {
	limit = DefaultLimit
	offset = DefaultOffset

	if limitStrs, ok := query["limit"]; ok && len(limitStrs) > 0 {
		if l, err := strconv.Atoi(limitStrs[0]); err == nil && l >= 0 {
			limit = l
			if limit > MaxLimit {
				limit = MaxLimit
			}
		}
	}

	if offsetStrs, ok := query["offset"]; ok && len(offsetStrs) > 0 {
		if o, err := strconv.Atoi(offsetStrs[0]); err == nil && o >= 0 {
			offset = o
		}
	}

	return limit, offset
}
