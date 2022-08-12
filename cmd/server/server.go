package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joeyorlando/trade_executor_tech_challenge/internal/binance"
	"github.com/joeyorlando/trade_executor_tech_challenge/internal/database"
)

// A LimitOrderRequest represents the incoming request that the POST /order/limit
// endpoint should receive
type LimitOrderRequest struct {
	Symbol    string  `json:"symbol"`
	OrderSize float64 `json:"order_size"`
	Price     float64 `json:"price"`
}

// A Server represents an HTTP server
type Server struct {
	Port     string // the port that the http server should listen on
	Binance  binance.Binance
	Database database.Database
}

// fulfillLimitOrder handles requests to the POST /order/limit HTTP endpoint
// given the LimitOrderRequest request, it trys to fulfill an order using the
// requested information. If an order is placed, the order details are persisted
// to the database
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

// Run starts the HTTP server, binding handlers to endpoints and running
// on the configured port
func (s *Server) Run() {
	router := gin.Default()
	router.POST("/order/limit", s.fulfillLimitOrder)
	router.Run(fmt.Sprintf(":%s", s.Port))
}

// NewServer configures and returns a new Server
func NewServer(port string, bin binance.Binance, db database.Database) Server {
	return Server{
		Port:     port,
		Binance:  bin,
		Database: db,
	}
}
