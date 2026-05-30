package httputil

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ParsePagination(c *gin.Context) (offset, limit int, ok bool) {
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "20")

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
