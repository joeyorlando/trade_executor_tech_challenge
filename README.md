# Trade Executor Tech Challenge

Service built using Go 1.18.

## Local development

### Getting Started

This repository makes use [Task](https://taskfile.dev/#/), [golangci-lint](https://golangci-lint.run/).
These may be installed (on Mac) with:

```bash
brew install go-task/tap/go-task
brew install golangci-lint
brew install goenv
```

You must also have `docker-compose` installed, and the Docker daemon must be running on your machine (on Mac this can be installed by following the instructions [here](https://docs.docker.com/desktop/install/mac-install/)).

### Running

The service runs on [air](https://github.com/cosmtrek/air) and supports live-reloading.

```bash
task run
```

### Tests

```bash
task test
```

### Linting

```bash
task lint
```
