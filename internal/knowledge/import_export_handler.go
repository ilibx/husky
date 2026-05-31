package knowledge

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/httputil"
)

func (h *Handler) ImportKnowledge(c *gin.Context) {
	format := c.DefaultQuery("format", "json")

	switch format {
	case "csv":
		h.importCSV(c)
	default:
		h.importJSON(c)
	}
}

func (h *Handler) importJSON(c *gin.Context) {
	var items []model.KnowledgeBase
	if err := c.ShouldBindJSON(&items); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	count, err := h.knowledgeService.ImportKnowledge(c.Request.Context(), items)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{"message": "import complete", "imported": count})
}

func (h *Handler) importCSV(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "file is required")
		return
	}

	f, err := file.Open()
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, "failed to open file")
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("failed to close CSV file: %v", err)
		}
	}()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "failed to parse CSV: "+err.Error())
		return
	}

	if len(records) < 2 {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "CSV must have header and at least one row")
		return
	}

	headers := make([]string, len(records[0]))
	for i, h := range records[0] {
		headers[i] = strings.TrimSpace(strings.ToLower(h))
	}

	var items []model.KnowledgeBase
	for _, row := range records[1:] {
		if len(row) != len(headers) {
			continue
		}
		item := model.KnowledgeBase{Status: "active"}
		for i, val := range row {
			val = strings.TrimSpace(val)
			switch headers[i] {
			case "title":
				item.Title = val
			case "content":
				item.Content = val
			case "category":
				item.Category = val
			case "tags":
				if val != "" {
					tags := strings.Split(val, ";")
			tagJSON, err := json.Marshal(tags)
			if err != nil {
				log.Printf("CSV import: failed to marshal tags: %v", err)
			} else {
				item.Tags = tagJSON
			}
				}
			}
		}
		if item.Title != "" {
			items = append(items, item)
		}
	}

	if len(items) == 0 {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "no valid items found in CSV")
		return
	}

	count, err := h.knowledgeService.ImportKnowledge(c.Request.Context(), items)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{"message": "import complete", "imported": count})
}

func (h *Handler) ExportKnowledge(c *gin.Context) {
	format := c.DefaultQuery("format", "json")

	switch format {
	case "csv":
		h.exportCSV(c)
	default:
		h.exportJSON(c)
	}
}

func (h *Handler) exportJSON(c *gin.Context) {
	items, err := h.knowledgeService.ExportKnowledge(c.Request.Context())
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{"data": items, "total": len(items)})
}

func (h *Handler) exportCSV(c *gin.Context) {
	items, err := h.knowledgeService.ExportKnowledge(c.Request.Context())
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=knowledge_export_%d.csv", len(items)))

	writer := csv.NewWriter(c.Writer)
	if err := writer.Write([]string{"title", "content", "category", "tags"}); err != nil {
		log.Printf("CSV export: failed to write header: %v", err)
	}

	for _, item := range items {
		var tags []string
		if item.Tags != nil {
			if err := json.Unmarshal(item.Tags, &tags); err != nil {
				log.Printf("CSV export: failed to unmarshal tags for item %s: %v", item.Title, err)
			}
		}
		if err := writer.Write([]string{
			item.Title,
			item.Content,
			item.Category,
			strings.Join(tags, ";"),
		}); err != nil {
			log.Printf("CSV export: failed to write row for item %s: %v", item.Title, err)
			continue
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		log.Printf("CSV export: flush error: %v", err)
	}
}
