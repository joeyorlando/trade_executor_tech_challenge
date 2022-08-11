package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joeyorlando/trade_executor_tech_challenge/internal/binance"
	"github.com/joeyorlando/trade_executor_tech_challenge/internal/database"
)

type LimitOrderRequest struct {
	Symbol    string  `json:"symbol"`
	OrderSize float64 `json:"order_size"`
	Price     float64 `json:"price"`
}

type Server struct {
	Port     string
	Binance  binance.Binance
	Database database.Database
}

func (s *Server) fulfillLimitOrder(c *gin.Context) {
	var req LimitOrderRequest

	if err := c.BindJSON(&req); err != nil {
		// TODO: better request validation - tell the client what parameter(s) fail request validation
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "There was an error processing your request",
		})
	} else {
		order := binance.LimitOrder{
			Symbol:   req.Symbol,
			Quantity: req.OrderSize,
			Price:    req.Price,
		}

		orderSplits, fulfilled, err := s.Binance.FulfillLimitOrder(order)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "There was an error parsing your request",
			})
		} else if !fulfilled {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Order not fulfilled",
			})
		} else {
			// TODO: should handle this scenario differently
			// this would mean that the ordered was "placed" w/ binance but we were unable to
			// persist the order details to the service's database
			if err := s.Database.PersistFulfilledOrder(order, orderSplits); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "There was an error processing your request",
				})
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "Order successfully fulfilled",
			})
		}

	}
}

func (s *Server) Run() {
	router := gin.Default()
	router.POST("/order/limit", s.fulfillLimitOrder)
	router.Run(fmt.Sprintf(":%s", s.Port))
}

func NewServer(port string, bin binance.Binance, db database.Database) Server {
	return Server{
		Port:     port,
		Binance:  bin,
		Database: db,
	}
}
