# Build frontend
FROM docker.1ms.run/node:18-alpine AS frontend-builder

WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci --registry=https://registry.npmmirror.com
COPY frontend/ ./
RUN npm run build

# Build backend
FROM docker.1ms.run/golang:1.21-alpine AS backend-builder

WORKDIR /app
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN GOPROXY=https://goproxy.cn,direct go mod download

COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

COPY *.go ./
COPY cache/ ./cache/
COPY config/ ./config/
COPY database/ ./database/
COPY handlers/ ./handlers/
COPY models/ ./models/
COPY services/ ./services/

RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM docker.1ms.run/alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=backend-builder /app/main .

RUN mkdir -p /app/uploads /app/data

EXPOSE 8080

ENV DB_PATH=/app/data/wardrobe.db
ENV UPLOAD_PATH=/app/uploads

CMD ["./main"]
