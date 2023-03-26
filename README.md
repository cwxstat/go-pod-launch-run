[![Codefresh build status]( https://g.codefresh.io/api/badges/pipeline/cwxstat-cicd/AWS_pkg2-k8s-aws%2Fgo-pod-launch-run?type=cf-1&key=eyJhbGciOiJIUzI1NiJ9.NjM0YWVkMzM0NzI2YzFhMzBlNWVkOTVh.ZfjYCotys9Ry041iPk25bHpvIClqKVV0FlncXV1wtRk)]( https://g.codefresh.io/pipelines/edit/new/builds?id=64209c6e5977e4568a084e4a&pipeline=go-pod-launch-run&projects=AWS_pkg2-k8s-aws&projectId=6419e306b7daeb096e312446)
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