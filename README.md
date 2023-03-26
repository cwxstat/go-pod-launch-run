# go-pod-launch-run
From client-go launch a pod, wait until stable, then run command from outside. Finally, delete pod

```bash
# Run the following

go get k8s.io/client-go@v0.26.3
```

## Usage

```bash
$ make
 make help                 -> display make targets
 make build-all            -> build: go get, tidy, fmt, go build -o gplr

```