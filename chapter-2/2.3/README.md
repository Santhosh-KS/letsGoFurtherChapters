%date

Thu 25 Jul 12:00:08 IST 2024


## Install "Router"

[BenchMarkResults](https://github.com/julienschmidt/go-http-routing-benchmark#go-http-router-benchmark)

```bash
go get github.com/julienschmidt/httprouter@v1
```

## Commands executed..

```bash
curl localhost:4000/v1/healthcheck
curl localhost:4000/v1/healthcheck -i
curl -X POST localhost:4000/v1/movies
curl localhost:4000/v1/movies/123
curl -X POST localhost:4000/v1/healthcheck 
curl -X POST localhost:4000/v1/healthcheck  -i
curl -X POST localhost:4000/v1/movies -i
curl -X OPTIONS localhost:4000/v1/healthcheck  -i
curl -i localhost:4000/v1/movies
curl -i localhost:4000/v1/movies/abc
curl -i localhost:4000/v1/movies/-11
curl -i localhost:4000/v1/movies/abc
curl -i localhost:4000/v1/movies/1
```
