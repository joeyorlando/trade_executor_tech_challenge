package binance

import (
	"fmt"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
)

type Binance struct {
	// number of seconds that the executor will try to fulfill an order before giving up
	OrderExecutionTimeoutSeconds int
}

type LimitOrder struct {
	Symbol   string
	Quantity float64
	Price    float64
}

type OrderSplit struct {
	UpdateId    int     `json:"update_id"`
	BidPrice    float64 `json:"bid_price"`
	BidQuantity float64 `json:"bid_quantity"`
}

func NewBinance(orderExecutionTimeoutSeconds int) Binance {
	return Binance{
		OrderExecutionTimeoutSeconds: orderExecutionTimeoutSeconds,
	}
}

func calculateOrderSplitsQuantity(orderSplits []OrderSplit) float64 {
	quantity := 0.00

	for _, orderSplit := range orderSplits {
		quantity += orderSplit.BidQuantity
	}

	return quantity
}

func orderIsFulfilled(order LimitOrder, orderSplits []OrderSplit) bool {
	currentOrderSplitsQuantity := calculateOrderSplitsQuantity(orderSplits)
	return currentOrderSplitsQuantity == order.Quantity
}

func bidHasAcceptablePrice(order LimitOrder, bid binance.Bid) bool {
	if bidPrice, bidPriceErr := strconv.ParseFloat(bid.Price, 64); bidPriceErr == nil {
		return bidPrice >= order.Price
	}
	return false
}

func bidQuantityToTake(currentOrderSplits []OrderSplit, orderSize, bidQuantity float64) float64 {
	currentQuantity := calculateOrderSplitsQuantity(currentOrderSplits)
	remainingQuantityToFill := orderSize - currentQuantity

	if bidQuantity >= remainingQuantityToFill {
		return remainingQuantityToFill
	}
	return bidQuantity
}

func (b *Binance) FulfillLimitOrder(order LimitOrder) (orderSplits []OrderSplit, fulfilled bool, err error) {
	wsDepthHandler := func(event *binance.WsDepthEvent) {
		for _, bid := range event.Bids {
			if bidHasAcceptablePrice(order, bid) && !orderIsFulfilled(order, orderSplits) {
				bidPrice, _ := strconv.ParseFloat(bid.Price, 64)
				bidQuantity, _ := strconv.ParseFloat(bid.Quantity, 64)

				orderSplits = append(orderSplits, OrderSplit{
					UpdateId:    int(event.LastUpdateID),
					BidPrice:    bidPrice,
					BidQuantity: bidQuantityToTake(orderSplits, order.Quantity, bidQuantity),
				})
			}
		}
	}

	// TODO: should maybe validate that the symbol passed in is actually legit
	doneC, stopC, err := binance.WsDepthServe(order.Symbol, wsDepthHandler, func(err error) {
		fmt.Println(err)
	})

	if err != nil {
		return orderSplits, false, err
	}

	go func() {
		timeout := 0

		for !orderIsFulfilled(order, orderSplits) {
			if orderIsFulfilled(order, orderSplits) || timeout > b.OrderExecutionTimeoutSeconds {
				stopC <- struct{}{}
			}

			time.Sleep(1 * time.Second)
			timeout += 1
		}
	}()

	// block until either the order is fulfilled or it times out
	<-doneC

	// TODO: at this point we would have the order ready to execute
	// this is where we would add logic to actually place the order if we wanted to

	return orderSplits, orderIsFulfilled(order, orderSplits), err
}
