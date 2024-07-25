%date


Thu 25 Jul 11:00:08 IST 2024

## Install the tools


```bash
go install github.com/rakyll/hey@latest
```

## Init project

```bash
go mod init greenlight.techkunstler.com

```

## Generating Skeleton Directory structure

```bash
mkdir -p bin cmd/api internal migrations remote
touch makefile
touch cmd/api/main.go
```

1. The *bin* directory will contain our compiled application binaries, ready for deployment to a production server.
2. The *cmd/api* directory will contain the application-specific code for our Greenlight API application. This will include the code for running the server, reading and writing HTTP requests, and managing authentication.
3. The *internal* directory will contain various ancillary packages used by our API. It will contain the code for interacting with our database, doing data validation, sending emails and so on. Basically, any code which isn’t application-specific and can potentially be reused will live in here. Our Go code under cmd/api will import the packages in the internal directory (but never the other way around).
4. The *migrations* directory will contain the SQL migration files for our database.
5. The *remote* directory will contain the configuration files and setup scripts for our production server.
6. The go.mod file will declare our project dependencies, versions and module path.
7. The *Makefile* will contain recipes for automating common administrative tasks — like auditing our Go code, building binaries, and executing database migrations.


It’s important to point out that the directory name *internal* carries a special meaning and behavior in Go: any packages which live under this directory can only be imported by code inside the parent of the internal directory. In our case, this means that any packages which live in internal can only be imported by code inside our greenlight project directory.


## Run and test the "/v1/healthcheck" endpoint

```bash
$ curl -i localhost:4000/v1/healthcheck
HTTP/1.1 200 OK
Date: Mon, 05 Apr 2021 17:46:14 GMT
Content-Length: 58
Content-Type: text/plain; charset=utf-8

status: available
environment: development
version: 1.0.0
```

*NOTE*: The -i flag in the command above instructs curl to display the HTTP response headers as well as the response body.


## API Versioning:

There are two common approaches to doing this:

1. By prefixing all URLs with your API version, like /v1/healthcheck or /v2/healthcheck.
2. By using custom Accept and Content-Type headers on requests and responses to convey the API version, like Accept: application/vnd.greenlight-v1.


Throughout this book we’ll version our API by prefixing all the URL paths with /v1/ — just like we did with the /v1/healthcheck endpoint in this chapter.
