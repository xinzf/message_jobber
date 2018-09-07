package handlers

import "github.com/gin-gonic/gin"

type Home struct {
	Base
}

func (h *Home) Index(c *gin.Context) {
	h.Success(c, "Hello World!")
}
