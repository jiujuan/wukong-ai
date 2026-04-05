package parser

import (
	"context"
	"fmt"
	"strings"
)

// VisionLLM 支持 Vision 能力的 LLM 接口（避免循环导入，用接口隔离）
type VisionLLM interface {
	SupportsVision() bool
	ChatWithImages(ctx context.Context, prompt string, images []string) (string, error)
}

// BuildAttachmentContext 根据 LLM 能力决定图片处理路径，返回注入 Prompt 的上下文字符串。
//
//   - Vision LLM：将 base64 图片数据直接传给 LLM（messages 中含 image_url）
//   - 非 Vision LLM：返回提示文本，告知 LLM 有图片但无法直接查看
//
// 对于文本类附件，直接返回提取的文本内容。
func BuildAttachmentContext(llmProvider interface{}, result *ParseResult, fileName string) string {
	if result == nil {
		return ""
	}

	// 文本类附件：直接返回提取的文本
	if !result.IsImage {
		if result.Text == "" {
			return fmt.Sprintf("[附件 %s：内容提取失败，无法注入上下文]\n", fileName)
		}
		return fmt.Sprintf("[附件 %s 内容]\n%s\n", fileName, result.Text)
	}

	// 图片类：检查 LLM 是否支持 Vision
	if vllm, ok := llmProvider.(VisionLLM); ok && vllm.SupportsVision() {
		// Vision LLM：图片 base64 会在调用时直接传入 messages，此处只标注
		return fmt.Sprintf("[图片附件 %s：已附加图片数据，请结合图片内容回答]\n", fileName)
	}

	// 非 Vision LLM：降级为文字说明
	return fmt.Sprintf("[图片附件 %s：当前 LLM 不支持图片理解，请忽略此附件或要求用户提供文字描述]\n", fileName)
}

// BuildVisionMessages 为 Vision LLM 构造携带图片的消息列表
// 返回格式：[{role:system,...}, {role:user, content:[{type:text,...},{type:image_url,...}]}]
// 调用方可将此结构序列化后直接发给 OpenAI Vision API
func BuildVisionMessages(systemPrompt, userText string, imageB64List []string) []map[string]any {
	content := []map[string]any{
		{"type": "text", "text": userText},
	}
	for _, b64 := range imageB64List {
		// 自动检测图片格式前缀
		dataURI := guessDataURI(b64)
		content = append(content, map[string]any{
			"type": "image_url",
			"image_url": map[string]string{
				"url": dataURI,
			},
		})
	}

	msgs := []map[string]any{}
	if systemPrompt != "" {
		msgs = append(msgs, map[string]any{"role": "system", "content": systemPrompt})
	}
	msgs = append(msgs, map[string]any{"role": "user", "content": content})
	return msgs
}

// guessDataURI 根据 base64 数据前缀猜测图片 MIME 类型，构造 data URI
func guessDataURI(b64 string) string {
	// PNG: iVBORw0KGgo...
	// JPEG: /9j/...
	// WEBP: UklGR...
	// GIF: R0lGOD...
	mime := "image/jpeg" // 默认
	if strings.HasPrefix(b64, "iVBORw0K") {
		mime = "image/png"
	} else if strings.HasPrefix(b64, "R0lGOD") {
		mime = "image/gif"
	} else if strings.HasPrefix(b64, "UklGR") {
		mime = "image/webp"
	}
	return fmt.Sprintf("data:%s;base64,%s", mime, b64)
}
