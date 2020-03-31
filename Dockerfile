############################
# STEP 1 build executable binary
############################
FROM golang:1.14.1-alpine as builder

# Install SSL ca certificates.
# Ca-certificates is required to call HTTPS endpoints.
RUN apk update && apk add --no-cache ca-certificates pkgconfig


COPY . $GOPATH/src/github.com/companyname/kube-webhook/
WORKDIR $GOPATH/src/github.com/companyname/kube-webhook/

# Using go mod.
# RUN go mod download
# Build the binary

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o /go/bin/svc


############################
# STEP 2 build a small image
############################
FROM scratch

# Import from builder.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy our static executable
COPY --from=builder /go/bin/svc /svc
COPY --from=builder /go/src/github.com/companyname/kube-webhook/version /

# Port on which the service will be exposed.
EXPOSE 8080
EXPOSE 8888

# Run the svc binary.
CMD ["./svc"]
