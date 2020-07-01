############################
# STEP 1 build executable binary
############################
FROM golang:1.14-alpine as builder

# Install SSL ca certificates.
# Ca-certificates is required to call HTTPS endpoints.
RUN apk update && apk add --no-cache ca-certificates pkgconfig

COPY . /src
WORKDIR /src

# Using go mod.
# RUN go mod download
# Build the binary

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o /bin/svc ./cmd/svc


############################
# STEP 2 build a small image
############################
FROM scratch

# Import from builder.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy our static executable
COPY --from=builder /bin/svc /svc
COPY --from=builder /src/version /

# Run the svc binary.
CMD ["./svc"]
