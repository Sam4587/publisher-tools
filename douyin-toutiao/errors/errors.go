package errors

import stderrors "errors"

var (
	ErrNotLoggedIn     = stderrors.New("未登录")
	ErrCookieExpired   = stderrors.New("Cookie 已过期")
	ErrUploadFailed    = stderrors.New("上传失败")
	ErrPublishFailed   = stderrors.New("发布失败")
	ErrInvalidImage    = stderrors.New("无效的图片")
	ErrInvalidVideo    = stderrors.New("无效的视频")
	ErrVideoTooLarge   = stderrors.New("视频过大")
	ErrTitleTooLong    = stderrors.New("标题过长")
	ErrContentTooLong  = stderrors.New("正文过长")
	ErrNetworkError    = stderrors.New("网络错误")
	ErrRateLimited     = stderrors.New("操作过于频繁")
	ErrAntiSpider      = stderrors.New("触发风控")
	ErrElementNotFound = stderrors.New("元素未找到")
	ErrTimeout         = stderrors.New("操作超时")
	ErrInvalidPlatform = stderrors.New("无效的平台")
	ErrLoginFailed     = stderrors.New("登录失败")
)
