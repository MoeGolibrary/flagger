FROM golang:1.22-alpine as builder

ARG TARGETPLATFORM
ARG REVISON

WORKDIR /workspace

# copy modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# cache modules
# RUN go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct
RUN go mod download

# copy source code
COPY cmd/ cmd/
COPY pkg/ pkg/

# build
RUN CGO_ENABLED=0 go build \
    -ldflags "-s -w -X github.com/fluxcd/flagger/pkg/version.REVISION=${REVISON}" \
    -a -o flagger ./cmd/flagger

FROM alpine:3.20

RUN apk --no-cache add ca-certificates

USER nobody

COPY --from=builder --chown=nobody:nobody /workspace/flagger .

ENTRYPOINT ["./flagger"]
