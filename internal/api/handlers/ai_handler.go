package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AIHandler struct{}

func NewAIHandler() *AIHandler {
	return &AIHandler{}
}

func (h *AIHandler) GenerateSummary(c *gin.Context) {
	// Mock an AI summary response since third-party proxy is not yet specified.
	projectID, ok := parseProjectID(c)
	if !ok {
		return
	}

	// This is where we'd fetch matrix data and call an AI like OpenAI.
	// For now, we return a mock Markdown response.
	mockedMarkdown := "### 综合评价\n各选项表现平衡，但各有侧重点。\n\n" +
		"### 优点分析\n- **选项 A**: 价格最低，性价比高。\n- **选项 B**: 功能最全面，体验最佳。\n\n" +
		"### 缺点分析\n- **选项 A**: 功能略有欠缺。\n- **选项 B**: 价格昂贵。\n\n" +
		"### 购买建议\n若预算充足推荐选项B，追求性价比推荐选项A。\n\n_注：Project ID " + strconv.Itoa(int(projectID)) + "_"

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"summary_markdown": mockedMarkdown,
		},
	})
}
