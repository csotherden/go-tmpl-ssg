# Stage 1: Build the Go binaries
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the binaries
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /go/bin/go-tmpl-generate ./cmd/go-tmpl-generate
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /go/bin/go-tmpl-dev-server ./cmd/go-tmpl-dev-server

# Stage 2: Create the minimal runtime image
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -S appgroup && adduser -S ssger -G appgroup

# Set the working directory
WORKDIR /home/ssger

# Copy binaries from the builder stage
COPY --from=builder /go/bin/go-tmpl-generate /usr/local/bin/
COPY --from=builder /go/bin/go-tmpl-dev-server /usr/local/bin/

# Make binaries executable
RUN chmod +x /usr/local/bin/go-tmpl-generate /usr/local/bin/go-tmpl-dev-server

# Create directories for mounting volumes
RUN mkdir -p /templates /output && \
    chown -R ssger:appgroup /templates /output

# Expose the default port for the development server
EXPOSE 8080

# Set default environment variables
ENV TEMPLATES_DIR=/templates \
    OUTPUT_DIR=/output \
    PORT=8080

# Create entrypoint script
USER root
RUN echo '#!/bin/sh' > /usr/local/bin/entrypoint.sh && \
    echo 'if [ "$1" = "generate" ]; then' >> /usr/local/bin/entrypoint.sh && \
    echo '    exec go-tmpl-generate --template ${TEMPLATES_DIR} --output ${OUTPUT_DIR}' >> /usr/local/bin/entrypoint.sh && \
    echo 'elif [ "$1" = "dev-server" ]; then' >> /usr/local/bin/entrypoint.sh && \
    echo '    exec go-tmpl-dev-server --port ${PORT} --root ${OUTPUT_DIR} --source ${TEMPLATES_DIR}' >> /usr/local/bin/entrypoint.sh && \
    echo 'else' >> /usr/local/bin/entrypoint.sh && \
    echo '    echo "Usage: docker run [options] <image> [generate|dev-server]"' >> /usr/local/bin/entrypoint.sh && \
    echo '    echo ""' >> /usr/local/bin/entrypoint.sh && \
    echo '    echo "Examples:"' >> /usr/local/bin/entrypoint.sh && \
    echo '    echo "  Generate site: docker run -v \$(pwd)/templates:/templates -v \$(pwd)/output:/output <image> generate"' >> /usr/local/bin/entrypoint.sh && \
    echo '    echo "  Run dev server: docker run -p 8080:8080 -v \$(pwd)/templates:/templates -v \$(pwd)/output:/output <image> dev-server"' >> /usr/local/bin/entrypoint.sh && \
    echo 'fi' >> /usr/local/bin/entrypoint.sh && \
    chmod +x /usr/local/bin/entrypoint.sh

# Switch back to non-root user
USER ssger

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"] 