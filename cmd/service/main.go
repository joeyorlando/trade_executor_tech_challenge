package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joeyorlando/trade_executor_tech_challenge/internal/binance"
)

type LimitOrderRequest struct {
	Symbol    string  `json:"symbol"`
	OrderSize float64 `json:"order_size"`
	Price     float64 `json:"price"`
}

func executeLimitOrder(c *gin.Context, bin binance.Binance) {
	var req LimitOrderRequest

	if err := c.BindJSON(&req); err != nil {
		// TODO: better request validation - tell the client what parameter(s) fail request validation
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "There was an error parsing your request",
			"message": nil,
			"data":    nil,
		})
	} else {
		orderSplits, fulfilled, err := bin.FulfillLimitOrder(binance.LimitOrder{
			Symbol:   req.Symbol,
			Quantity: req.OrderSize,
			Price:    req.Price,
		})

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "There was an error parsing your request",
				"message": nil,
				"data":    nil,
			})
		} else if !fulfilled {
			c.JSON(http.StatusOK, gin.H{
				"error":   nil,
				"message": "Order not fulfilled",
				"data":    []binance.OrderSplit{},
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"error":   nil,
				"message": "Order successfully fulfilled",
				"data":    orderSplits,
			})
		}

	}
}

func main() {
	httpPort := os.Getenv("HTTP_PORT")
	orderTimeoutSeconds, err := strconv.Atoi(os.Getenv("ORDER_TIMEOUT_SECONDS"))

	if err != nil {
		log.Fatal("ORDER_TIMEOUT_SECONDS must be an integer")
	}

	bin := binance.NewBinance(orderTimeoutSeconds)

	router := gin.Default()
	router.POST("/order/limit", func(c *gin.Context) {
		executeLimitOrder(c, bin)
	})
	router.Run(fmt.Sprintf(":%s", httpPort))
}
