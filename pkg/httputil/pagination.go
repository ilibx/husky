package httputil

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ParsePagination(c *gin.Context) (offset, limit int, ok bool) {
	offsetStr := c.Query("offset")
	limitStr := c.Query("limit")
	if offsetStr == "" && limitStr == "" {
		pageStr := c.DefaultQuery("page", "1")
		pageSizeStr := c.DefaultQuery("page_size", "20")
		page, err := strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page"})
			return 0, 0, false
		}
		limit, err = strconv.Atoi(pageSizeStr)
		if err != nil || limit <= 0 || limit > 200 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page_size (must be 1-200)"})
			return 0, 0, false
		}
		return (page - 1) * limit, limit, true
	}

	if offsetStr == "" {
		offsetStr = "0"
	}
	if limitStr == "" {
		limitStr = "20"
	}

	var err error
	offset, err = strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset"})
		return 0, 0, false
	}

	limit, err = strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit (must be 1-200)"})
		return 0, 0, false
	}

	return offset, limit, true
}
