package binance

import (
	"fmt"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
)

// A Binance represents configuration related to fulfilling orders
// and communicating with the binance API
type Binance struct {
	// number of seconds that the executor will try to fulfill an order before giving up
	OrderExecutionTimeoutSeconds int
}

// A LimitOrder represents a limit order that is looking to be fulfilled
type LimitOrder struct {
	Symbol   string
	Quantity float64
	Price    float64
}

// An OrderSplit represents a piece of a fulfilled order. A fulfilled order may have one or more OrderSplits
type OrderSplit struct {
	UpdateId    int     `json:"update_id"`    // the unique ID for the binance order
	BidPrice    float64 `json:"bid_price"`    // the bid price on the order
	BidQuantity float64 `json:"bid_quantity"` // the bid quantity on the order
}

// NewBinance configures and returns a new Binance
func NewBinance(orderExecutionTimeoutSeconds int) Binance {
	return Binance{
		OrderExecutionTimeoutSeconds: orderExecutionTimeoutSeconds,
	}
}

// calculateOrderSplitsQuantity will sum up the total quantity of shares
// across a list of OrderSplits
func calculateOrderSplitsQuantity(orderSplits []OrderSplit) float64 {
	quantity := 0.00

	for _, orderSplit := range orderSplits {
		quantity += orderSplit.BidQuantity
	}

	return quantity
}

// orderIsFulfilled determines if the order can be considered fulfilled.
// The criteria is simply, does the quantity requested on the order equal the
// total quantity across the OrderSplits?
func orderIsFulfilled(order LimitOrder, orderSplits []OrderSplit) bool {
	currentOrderSplitsQuantity := calculateOrderSplitsQuantity(orderSplits)
	return currentOrderSplitsQuantity == order.Quantity
}

// bidHasAcceptablePrice takes an order and a bid. Given the requested price on the order,
// and the bid price, it determines if the bid price is greater than or equal to the requested order price
func bidHasAcceptablePrice(order LimitOrder, bid binance.Bid) bool {
	if bidPrice, bidPriceErr := strconv.ParseFloat(bid.Price, 64); bidPriceErr == nil {
		return bidPrice >= order.Price
	}
	return false
}

// bidQuantityToTake determines how much of the current bid we can use for fulfilling the current order
// It uses the current OrderSplits to determine what the remaining quantity to be filled is
// and, based on this, takes either the full bid quantity, or the remaining quantity to be filled
func bidQuantityToTake(currentOrderSplits []OrderSplit, orderSize, bidQuantity float64) float64 {
	currentQuantity := calculateOrderSplitsQuantity(currentOrderSplits)
	remainingQuantityToFill := orderSize - currentQuantity

	if bidQuantity >= remainingQuantityToFill {
		return remainingQuantityToFill
	}
	return bidQuantity
}

// FulfillLimitOrder will try to place a LimitOrder with Binance
// We first subscribe to a stream of websocket events for the LimitOrder.Symbol. As order events come in we
// check to see if the bid price on the order is acceptable given our LimitOrder. If an event has an acceptable
// bid price, we store it in a list of OrderSplits and continue listening to market order events until either
// the LimitOrder has been fulfilled, or we have surpassed a configured timeout
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
		timeoutCounter := 0

		for !orderIsFulfilled(order, orderSplits) && timeoutCounter < b.OrderExecutionTimeoutSeconds {
			time.Sleep(1 * time.Second)
			timeoutCounter += 1
		}

		stopC <- struct{}{}
	}()

	// block until either the order is fulfilled or it times out
	<-doneC

	// TODO: at this point we would have the order ready to execute
	// this is where we would add logic to actually place the order if we wanted to

	return orderSplits, orderIsFulfilled(order, orderSplits), err
}
