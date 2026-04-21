package handlers

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"choice-matrix-backend/internal/models"
	"choice-matrix-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type AIHandler struct {
	matrixRepo    *repository.MatrixRepository
	workspaceRepo *repository.WorkspaceRepository
}

type GenerateSummaryRequest struct {
	FocusPrompt string `json:"focus_prompt"`
}

type scoredRow struct {
	row       models.Row
	total     *float64
	completed int
	required  int
	best      string
	weakest   string
}

func NewAIHandler(workspaceRepo *repository.WorkspaceRepository, matrixRepo *repository.MatrixRepository) *AIHandler {
	return &AIHandler{
		matrixRepo:    matrixRepo,
		workspaceRepo: workspaceRepo,
	}
}

func (h *AIHandler) GenerateSummary(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok {
		return
	}

	var req GenerateSummaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	userID := c.MustGet("userID").(uint)
	project, err := h.workspaceRepo.GetProjectByID(projectID, userID)
	if err != nil {
		respondError(c, http.StatusForbidden, "Access denied or project not found")
		return
	}

	rows, err := h.matrixRepo.GetRowsByProjectID(projectID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to fetch rows")
		return
	}

	columns, err := h.matrixRepo.GetColumnsByProjectID(projectID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to fetch columns")
		return
	}

	cells, err := h.matrixRepo.GetCellsByProjectID(projectID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to fetch cells")
		return
	}

	summaryMarkdown := buildProjectSummaryMarkdown(project, rows, columns, cells, req.FocusPrompt)
	respondSuccess(c, http.StatusOK, gin.H{
		"summary":          summaryMarkdown,
		"summary_markdown": summaryMarkdown,
	}, "AI summary generated")
}

func buildProjectSummaryMarkdown(project *models.Project, rows []models.Row, columns []models.Column, cells []models.Cell, focusPrompt string) string {
	scoreColumns := make([]models.Column, 0)
	for _, column := range columns {
		if column.Type == models.ColumnTypeScore {
			scoreColumns = append(scoreColumns, column)
		}
	}

	rankedRows := rankRows(rows, scoreColumns, cells)
	filledCells := countFilledCells(cells)
	totalCells := len(rows) * len(columns)
	focus := strings.TrimSpace(focusPrompt)

	var builder strings.Builder
	builder.WriteString("### 综合评价\n")
	if len(rows) == 0 || len(columns) == 0 {
		builder.WriteString(fmt.Sprintf("%s 还没有形成完整矩阵，建议先补充候选项和比较维度。\n\n", project.Title))
	} else if len(scoreColumns) == 0 {
		builder.WriteString(fmt.Sprintf("%s 当前已有 %d 个候选项、%d 个维度，但还没有评分列，暂时无法形成加权排名。\n\n", project.Title, len(rows), len(columns)))
	} else if len(rankedRows) > 0 && rankedRows[0].total != nil {
		builder.WriteString(fmt.Sprintf("%s 当前完成了 %d/%d 个单元格，基于已填写评分，**%s** 暂时领先，综合分为 **%.1f**。\n\n", project.Title, filledCells, totalCells, rankedRows[0].row.Name, *rankedRows[0].total))
	} else {
		builder.WriteString(fmt.Sprintf("%s 当前完成了 %d/%d 个单元格，但评分信息还不足以生成有效排名。\n\n", project.Title, filledCells, totalCells))
	}

	if focus != "" {
		builder.WriteString("### 本次关注重点\n")
		builder.WriteString(fmt.Sprintf("- %s\n\n", focus))
	}

	builder.WriteString("### 排名与依据\n")
	if len(rankedRows) == 0 {
		builder.WriteString("- 暂无候选项。\n")
	} else {
		for index, row := range rankedRows {
			if row.total == nil {
				builder.WriteString(fmt.Sprintf("- %s：尚未填写有效评分，已完成 %d/%d 个评分项。\n", row.row.Name, row.completed, row.required))
				continue
			}

			builder.WriteString(fmt.Sprintf("- #%d **%s**：%.1f 分，已完成 %d/%d 个评分项。", index+1, row.row.Name, *row.total, row.completed, row.required))
			if row.best != "" {
				builder.WriteString(" 优势：" + row.best + "。")
			}
			if row.weakest != "" {
				builder.WriteString(" 风险：" + row.weakest + "。")
			}
			builder.WriteString("\n")
		}
	}

	builder.WriteString("\n### 信息缺口\n")
	gaps := collectGaps(rankedRows, scoreColumns)
	if len(gaps) == 0 {
		builder.WriteString("- 核心评分项已经具备，可以进入复核和最终判断。\n")
	} else {
		for _, gap := range gaps {
			builder.WriteString("- " + gap + "\n")
		}
	}

	builder.WriteString("\n### 建议\n")
	if len(rankedRows) > 0 && rankedRows[0].total != nil {
		builder.WriteString(fmt.Sprintf("- 先把 **%s** 作为当前基准方案，再重点核对低分维度和备注中的风险。\n", rankedRows[0].row.Name))
	} else {
		builder.WriteString("- 先补齐评分列，再生成下一版摘要，避免信息不足时过早下结论。\n")
	}
	builder.WriteString("- 对主观评分差距小于 1 分的方案，建议补充备注或新增一个区分度更高的评分维度。\n")

	return builder.String()
}

func rankRows(rows []models.Row, scoreColumns []models.Column, cells []models.Cell) []scoredRow {
	cellByKey := map[string]models.Cell{}
	for _, cell := range cells {
		cellByKey[cellKey(cell.RowID, cell.ColumnID)] = cell
	}

	ranked := make([]scoredRow, 0, len(rows))
	for _, row := range rows {
		item := scoredRow{
			row:      row,
			required: len(scoreColumns),
		}

		var weighted float64
		var totalWeight float64
		var bestScore *float64
		var weakestScore *float64

		for _, column := range scoreColumns {
			cell, exists := cellByKey[cellKey(row.ID, column.ID)]
			if !exists || cell.ScoreValue == nil {
				continue
			}

			score := *cell.ScoreValue
			item.completed++
			weighted += score * column.Weight
			totalWeight += column.Weight

			if bestScore == nil || score > *bestScore {
				bestScore = &score
				item.best = fmt.Sprintf("%s %.0f 分", column.Title, score)
			}
			if weakestScore == nil || score < *weakestScore {
				weakestScore = &score
				item.weakest = fmt.Sprintf("%s %.0f 分", column.Title, score)
			}
		}

		if totalWeight > 0 {
			total := weighted / totalWeight
			item.total = &total
		}
		ranked = append(ranked, item)
	}

	sort.SliceStable(ranked, func(i, j int) bool {
		left := ranked[i]
		right := ranked[j]
		if left.total == nil && right.total == nil {
			return left.row.SortOrder < right.row.SortOrder
		}
		if left.total == nil {
			return false
		}
		if right.total == nil {
			return true
		}
		if *left.total == *right.total {
			return left.row.SortOrder < right.row.SortOrder
		}
		return *left.total > *right.total
	})

	return ranked
}

func collectGaps(rows []scoredRow, scoreColumns []models.Column) []string {
	gaps := make([]string, 0)
	if len(scoreColumns) == 0 {
		return []string{"缺少评分列，无法计算加权排名。"}
	}

	for _, row := range rows {
		if row.completed < row.required {
			gaps = append(gaps, fmt.Sprintf("%s 还缺少 %d 个评分项。", row.row.Name, row.required-row.completed))
		}
	}

	return gaps
}

func countFilledCells(cells []models.Cell) int {
	count := 0
	for _, cell := range cells {
		if strings.TrimSpace(cell.TextContent) != "" ||
			strings.TrimSpace(cell.Note) != "" ||
			cell.NumericValue != nil ||
			cell.ScoreValue != nil ||
			cell.SelectValue != nil {
			count++
		}
	}
	return count
}

func cellKey(rowID uint, columnID uint) string {
	return fmt.Sprintf("%d:%d", rowID, columnID)
}
