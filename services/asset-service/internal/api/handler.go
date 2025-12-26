package api

import (
	"net/http"
	"strconv"

	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/Cxiyuan/NTA/services/asset-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AssetHandler struct {
	service *service.AssetService
	logger  *logrus.Logger
}

func NewAssetHandler(service *service.AssetService, logger *logrus.Logger) *AssetHandler {
	return &AssetHandler{
		service: service,
		logger:  logger,
	}
}

func (h *AssetHandler) RegisterRoutes(r *gin.RouterGroup) {
	assets := r.Group("/assets")
	{
		assets.GET("", h.ListAssets)
		assets.GET("/:ip", h.GetAsset)
		assets.POST("", h.CreateAsset)
		assets.PUT("/:ip", h.UpdateAsset)
		assets.DELETE("/:ip", h.DeleteAsset)
	}
}

func (h *AssetHandler) ListAssets(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	filters := make(map[string]interface{})
	if osType := c.Query("os_type"); osType != "" {
		filters["os_type"] = osType
	}
	if risk := c.Query("risk_level"); risk != "" {
		filters["risk_level"] = risk
	}

	assets, total, err := h.service.ListAssets(c.Request.Context(), pageSize, offset, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      assets,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	})
}

func (h *AssetHandler) GetAsset(c *gin.Context) {
	ip := c.Param("ip")

	asset, err := h.service.GetAssetByIP(c.Request.Context(), ip)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}

	c.JSON(http.StatusOK, asset)
}

func (h *AssetHandler) CreateAsset(c *gin.Context) {
	var asset models.Asset
	if err := c.ShouldBindJSON(&asset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreateAsset(c.Request.Context(), &asset); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, asset)
}

func (h *AssetHandler) UpdateAsset(c *gin.Context) {
	ip := c.Param("ip")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateAsset(c.Request.Context(), ip, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "asset updated"})
}

func (h *AssetHandler) DeleteAsset(c *gin.Context) {
	ip := c.Param("ip")

	if err := h.service.DeleteAsset(c.Request.Context(), ip); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "asset deleted"})
}
