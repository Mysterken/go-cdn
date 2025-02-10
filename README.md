# go-cdn

## Description

This is a simple CDN server written in Go. It serves static files from a directory and caches them in memory.

## Prerequisites

- Go 1.23 or later
- Docker (optional)

## Usage

### Running the server

To run the server, you can use the following command:

```bash
go run main.go
```

alternatively using Docker:

```bash
cd docker
docker compose up --build --wait
```
