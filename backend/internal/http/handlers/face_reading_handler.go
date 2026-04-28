package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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

type FaceReadingResponse struct {
	Result string `json:"result"`
}

type dmxMessage struct {
	Role    string       `json:"role"`
	Content []dmxContent `json:"content"`
}

type dmxContent struct {
	Type     string            `json:"type"`
	Text     string            `json:"text,omitempty"`
	ImageURL map[string]string `json:"image_url,omitempty"`
}

type dmxRequest struct {
	Model    string       `json:"model"`
	Messages []dmxMessage `json:"messages"`
}

type dmxResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func FaceReadingHandler(apiKey, baseURL, model string) gin.HandlerFunc {
	client := &http.Client{Timeout: 60 * time.Second}

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

		dmxReq := dmxRequest{
			Model: model,
			Messages: []dmxMessage{
				{
					Role: "user",
					Content: []dmxContent{
						{Type: "text", Text: faceReadingPrompt},
						{Type: "image_url", ImageURL: map[string]string{"url": req.ImageBase64}},
					},
				},
			},
		}

		bodyBytes, err := json.Marshal(dmxReq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build request"})
			return
		}

		httpReq, err := http.NewRequestWithContext(c.Request.Context(), "POST", baseURL+"/chat/completions", bytes.NewReader(bodyBytes))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build request"})
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err := client.Do(httpReq)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "dmx api unreachable"})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			fmt.Fprintf(gin.DefaultWriter, "DMXAPI error: status=%d body=%s\n", resp.StatusCode, string(body))
			c.JSON(http.StatusBadGateway, gin.H{"error": "dmx api error"})
			return
		}

		var dmxResp dmxResponse
		if err := json.NewDecoder(resp.Body).Decode(&dmxResp); err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "invalid dmx response"})
			return
		}

		if len(dmxResp.Choices) == 0 {
			c.JSON(http.StatusBadGateway, gin.H{"error": "empty dmx response"})
			return
		}

		c.JSON(http.StatusOK, FaceReadingResponse{Result: dmxResp.Choices[0].Message.Content})
	}
}
