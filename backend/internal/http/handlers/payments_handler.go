package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"image-play/internal/domain/billing"
)

func PackagesHandler(svc *billing.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		packages, err := svc.GetPackages(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get packages"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"packages": packages})
	}
}

type CreateOrderRequest struct {
	PackageCode string `json:"package_code" binding:"required"`
}

type CreateOrderResponse struct {
	OrderNo     string `json:"order_no"`
	PackageCode string `json:"package_code"`
	Amount      string `json:"amount"`
	PrepayID    string `json:"prepay_id"`
}

func CreateOrderHandler(svc *billing.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		uid, ok := userID.(int64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
			return
		}

		var req CreateOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		order, err := svc.CreateOrder(c.Request.Context(), uid, req.PackageCode)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, CreateOrderResponse{
			OrderNo:     order.OrderNo,
			PackageCode: order.PackageCode,
			Amount:      order.Amount,
			PrepayID:    order.PrepayID,
		})
	}
}

type PaymentCallbackRequest struct {
	WxOrderNo string `json:"wx_order_no" binding:"required"`
}

func PaymentCallbackHandler(svc *billing.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req PaymentCallbackRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := svc.HandlePaymentCallback(c.Request.Context(), req.WxOrderNo); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
