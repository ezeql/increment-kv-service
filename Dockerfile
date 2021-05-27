FROM golang:1.16.4-alpine3.13 AS go-toolchain
#File is the intended file to be built using go
ARG FILE 

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

# dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download

# app source cde
COPY internal internal
COPY cmd cmd

# build app
FROM go-toolchain as builder
ARG FILE
RUN go build -o ${FILE} ./cmd/${FILE}

# final lean image
FROM scratch as app
ARG FILE
COPY --from=builder /build/${FILE} ./cmd
CMD ["./cmd"]

