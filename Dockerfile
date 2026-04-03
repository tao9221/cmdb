# 多阶段构建
# 阶段1：构建前端
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm config set registry https://registry.npmjs.org && npm install --prefer-online
COPY frontend/ ./
RUN npm run build

# 阶段2：构建后端
FROM golang:1.24-alpine AS backend-builder
ENV GOTOOLCHAIN=auto
ENV GOPROXY=https://goproxy.cn,https://goproxy.io,https://proxy.golang.org,direct
ENV GONOSUMDB=*
ENV GOFLAGS=-mod=mod
RUN apk add --no-cache gcc musl-dev
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=1 GOOS=linux go build -o cmdb-backend .

# 阶段3：最终镜像
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=backend-builder /app/backend/cmdb-backend .
COPY --from=frontend-builder /app/frontend/dist ./dist

VOLUME ["/app/data"]
EXPOSE 8088

CMD ["./cmdb-backend"]
