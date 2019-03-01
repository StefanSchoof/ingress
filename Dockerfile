# syntax=docker/dockerfile:1.0-experimental

############################
# STEP 1 build executable binary
############################
FROM golang:alpine as builder

# Install git + SSL ca certificates.
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
RUN apk update && apk add --no-cache git ca-certificates tzdata alpine-sdk && update-ca-certificates

# Create appuser
RUN adduser -D -g '' appuser

# Populate and persist modules
WORKDIR $GOPATH/src/github.com/andig/ingress
COPY go.mod .
ENV GO111MODULE=on
RUN go mod download

COPY . .

# Generate files
RUN make assets

# Build the binary
ENV CGO_ENABLED=0
ARG GOOS=linux
# RUN --mount=target=/root/.cache,type=cache go build -ldflags="-w -s" -a -installsuffix cgo -o /go/bin/ingress github.com/andig/ingress/cmd/ingress
RUN go build -ldflags="-w -s" -a -installsuffix cgo -o /go/bin/ingress github.com/andig/ingress/cmd/ingress

#############################
## STEP 2 build a small image
#############################
FROM scratch

# Import from builder.
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd

# Copy our static executable
COPY --from=builder /go/bin/ingress /go/bin/ingress

# Use an unprivileged user.
USER appuser

# Run the binary.
ENTRYPOINT ["/go/bin/ingress"]
