# task-manager

## Motivation

A simple task manager REST API which stores bash cmds to be consumed by the also bundled worker.

It uses redpanda as a queue system (kafka compatible) and sqlite as a database.

## Usage

Setup:
```
docker compose up -d
rpk topic create tasks
go install github.com/4thel00z/task-manager/cmd/task-manager@latest
```

Start up the server via:
```
task-manager serve
```

Start up the worker via:
```
task-manager worker
```

## License

This project is licensed under the GPL-3 license.
