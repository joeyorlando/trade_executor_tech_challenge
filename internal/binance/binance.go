package binance

type OrderBookOrder struct {
	OrderBookUpdateId int    `json:"u"`
	Symbol            string `json:"s"`
	BestBidPrice      string `json:"b"`
	BestBidQuantity   string `json:"B"`
	BestAskPrice      string `json:"a"`
	BestAskQuantity   string `json:"A"`
}

func readOrderBookTickerStream() (orders []OrderBookOrder, err error) {
	// TODO:
	return orders, err
}
