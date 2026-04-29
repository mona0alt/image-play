package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"image-play/internal/infrastructure/llm"
)

const faceReadingPrompt = `你是一位精通面相学与命理学的资深相师。请基于中国传统面相学、周易命理学及现代心理学的交叉视角，对图片中人物进行深度解析。分析维度必须包括以下方面：

1. 面相格局与五行属性：分析脸型对应的五行属性及整体格局高低。
2. 五官详解：眉毛、眼睛、鼻子、嘴巴、耳朵分别对应的宫位与运势。
3. 三停比例：上停、中停、下停的均衡度与运势走向。
4. 十二宫位速览：命宫、财帛宫、夫妻宫、疾厄宫等关键宫位。
5. 性格与气质：内在性格、情绪模式与待人接物风格。
6. 情感运势：感情观、桃花运、婚姻稳定性。
7. 事业财运：事业发展潜力、财富积累能力与适合方向。
8. 健康提示：面色、眼神等反映的体质倾向。

要求：请用专业且通俗的语言输出，分点清晰，有理有据。分析仅供参考娱乐，请保持客观理性，避免绝对化断言。`

type FaceReadingRequest struct {
	ImageBase64 string `json:"image_base64" binding:"required"`
}

func FaceReadingHandler(textClient llm.TextClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req FaceReadingRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "image_base64 is required"})
			return
		}

		const maxBase64Len = 7 * 1024 * 1024
		if len(req.ImageBase64) > maxBase64Len {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "image too large"})
			return
		}

		messages := []llm.Message{
			{
				Role: "user",
				Parts: []llm.Part{
					{Type: llm.PartTypeText, Content: faceReadingPrompt},
					{Type: llm.PartTypeImage, Content: req.ImageBase64},
				},
			},
		}

		reader, err := textClient.ChatStream(c.Request.Context(), messages)
		if err != nil {
			fmt.Printf("[face-reading] chat stream error: %v\n", err)
			c.JSON(http.StatusBadGateway, gin.H{"error": "model unavailable"})
			return
		}
		defer reader.Close()

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Writer.WriteHeader(http.StatusOK)

		for {
			chunk, err := reader.Recv()
			if err == io.EOF {
				c.Writer.WriteString("data: [DONE]\n\n")
				c.Writer.Flush()
				break
			}
			if err != nil {
				fmt.Printf("[face-reading] recv error: %v\n", err)
				break
			}
			if chunk.Content != "" {
				out, _ := json.Marshal(map[string]string{"chunk": chunk.Content})
				c.Writer.WriteString("data: " + string(out) + "\n\n")
				c.Writer.Flush()
			}
		}
	}
}
