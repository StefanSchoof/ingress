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
WORKDIR /go/src/github.com/andig/ingress

ENV GO111MODULE=on
COPY go.* ./
RUN go mod download

COPY . .
RUN make build

#############################
## STEP 2 build a small image
#############################
FROM alpine

# Import from builder.
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd

# Copy our static executable
COPY --from=builder /go/src/github.com/andig/ingress /usr/bin/ingress

# Use an unprivileged user.
USER appuser

# Run the binary.
ENTRYPOINT ["/usr/bin/ingress"]
