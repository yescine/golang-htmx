# Installation 

```bash
docker build -f Dockerfile.dev -t go-dev .

docker run --name go-dev -v .:/app go-dev
```
Or use VsCode devcontainer

# Init 

```bash
go run main.go

```