package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/souvikree/gitworld/api/internal/store"
)

type GraphHandler struct {
	Store store.GraphReader
	Log   *zap.Logger
}

func (h *GraphHandler) GetGraph(c *gin.Context) {
	reqCtx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	g, err := h.Store.GetGraph(reqCtx, 500)
	if err != nil {
		h.Log.Error("get graph failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch graph"})
		return
	}

	c.JSON(http.StatusOK, g)
}