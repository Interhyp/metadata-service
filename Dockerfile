ARG GOLANG_VERSION=1-buster
ARG PACT_VERSION=1.88.63

FROM golang:${GOLANG_VERSION} as build
ARG PACT_VERSION

# Install pact contract testing standalone binaries (includes Ruby)
RUN curl -LO https://github.com/pact-foundation/pact-ruby-standalone/releases/download/v${PACT_VERSION}/pact-${PACT_VERSION}-linux-x86_64.tar.gz; \
    tar -C /usr/local -xzf pact-${PACT_VERSION}-linux-x86_64.tar.gz; \
    rm pact-${PACT_VERSION}-linux-x86_64.tar.gz; \
    mkdir -p /app

ENV PATH /usr/local/pact/bin:$PATH

COPY . /app
WORKDIR /app

RUN go install github.com/jstemmer/go-junit-report/v2@v2.0.0-beta1

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" main.go \
    && go test -v ./... -coverpkg=./internal/...,./web/... 2>&1 | go-junit-report -set-exit-code -iocopy -out report.xml \
    && go vet ./...

FROM scratch

COPY --from=build /app/main /main
COPY --from=build /etc/ssl/certs /etc/ssl/certs
COPY --from=build /app/api/openapi-v3-spec.yaml /api/openapi-v3-spec.yaml

ENTRYPOINT ["/main"]