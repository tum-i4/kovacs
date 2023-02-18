FROM golang:1.17.8-alpine3.15

# Copy Geth config
COPY geth /build/geth
RUN chmod +x /build/geth/startGethClient

# Install geth
RUN apk add --no-cache geth
# Install vim for better dev in container
RUN apk add --no-cache vim
# Install gcc and alpine-sdk for SQLite
RUN apk add --no-cache gcc alpine-sdk

# Node
COPY node /build/node
WORKDIR /build/node/
RUN go mod download

# Query
COPY query /build/query
WORKDIR /build/query/
RUN chmod +x benchmark
RUN go mod download
RUN go build -o query

# Verifier
COPY verifier /build/verifier
WORKDIR /build/verifier/
RUN go mod download
RUN go build -o verifier

# Test password generation time
COPY passwordGenerationTimer /build/passwordGenerationTimer/
WORKDIR /build/passwordGenerationTimer/
RUN go mod download
RUN go build -o timer

# Listener
COPY listener /build/listener
WORKDIR /build/listener/
RUN go mod download
RUN go build -o listener

# Requester
COPY requester /build/requester
WORKDIR /build/requester/
RUN chmod +x benchmark
RUN go mod download
RUN go build -o requester

# MeasureSize
COPY measureStorage /build/measureStorage
WORKDIR /build/measureStorage
RUN go mod download
RUN go build -o measure

# Required by geth
EXPOSE 30401
# Required by the listener
EXPOSE 40000
# Required by the requester
EXPOSE 41000

# Start geth & listener
WORKDIR /build/
CMD cd /build/geth/ && /bin/sh startGethClient && cd /build/listener && ./listener
