ARG GOLANG_VERSION=1

FROM golang:${GOLANG_VERSION} AS build

COPY . /app
WORKDIR /app

RUN go install github.com/jstemmer/go-junit-report/v2@v2.0.0-beta1

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" main.go \
    && go test -v ./... -coverpkg=./internal/... 2>&1 | go-junit-report -set-exit-code -iocopy -out report.xml \
    && go vet ./...

FROM scratch

COPY --from=build /app/main /main
COPY --from=build /etc/ssl/certs /etc/ssl/certs
COPY --from=build /app/api/openapi-v3-spec.yaml /api/openapi-v3-spec.yaml

ENTRYPOINT ["/main"]
