package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Index is a handler for the index.
// @Summary Index
// @Description Index
// @Success 200 {string} Helloworld
// @Router /v1/ [get]
func Index(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": "hello world"})
}
