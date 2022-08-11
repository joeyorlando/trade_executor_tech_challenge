# Trade Executor Tech Challenge

Built using Go 1.18.

This repository houses a service that will "fulfill" (no actual orders placed.. for now) orders with Binance.

It exposes a single HTTP endpoint `POST /order/limit` that allows a user to consume the service. See [Using the Service](#using-the-service) for more details.

## Local development

### Getting Started

This repository makes use [Task](https://taskfile.dev/#/) and [golangci-lint](https://golangci-lint.run/).
These may be installed (on Mac) with:

```bash
brew install go-task/tap/go-task
brew install golangci-lint
brew install goenv
```

You must also have `docker-compose` installed, and the Docker daemon must be running on your machine (on Mac this can be installed by following the instructions [here](https://docs.docker.com/desktop/install/mac-install/)).

### Running the service

```bash
task run
```

### Tests

**Work In Progress**

```bash
task test
```

### Linting

```bash
task lint
```

## Using the service

```bash
$ curl --request POST \
    --url http://localhost:8080/order/limit \
    --data '{"symbol": "LTCBTC", "order_size": 2, "price": 0.0001}'
{"data":[{"update_id":1742566312,"bid_price":0.002574,"bid_quantity":2}],"error":null,"message":"Order successfully executed"}
```

**Note**: note that the exact response you get back will be heavily dependent on several factors:

- current market stream
- symbol, order size, and ask price you chose

## TODO

- [x] create an HTTP endpoint that executes an order (POST endpoint that takes a symbol, order size and price as input)
- [x] connect to the binance order book ticker stream (https://binance-docs.github.io/apidocs/spot/en/#individual-symbol-book-ticker-streams)
- [ ] write binance package unit tests
- [ ] write to sqlite-db: output summary on how the order was split
- [ ] write database package unit tests
- [ ] add timeout to the trade execution
- [ ] add simple UI using next.js
