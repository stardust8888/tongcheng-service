# 构建阶段
FROM golang:1.24-alpine AS builder
ARG APP_NAME
ARG GOOS=linux
ARG GOARCH=amd64
WORKDIR /src
COPY  . .
RUN CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -mod=vendor -ldflags="-s -w" -o service/${APP_NAME}/tongcheng-${APP_NAME} service/${APP_NAME}/main.go

# 运行阶段
FROM alpine:3.20 AS runner
ARG APP_NAME
ENV ENV_APP_NAME=${APP_NAME}
WORKDIR /app
COPY --from=builder /src/service/${ENV_APP_NAME}/tongcheng-${ENV_APP_NAME} /app/service/${ENV_APP_NAME}/tongcheng-${ENV_APP_NAME}
ENTRYPOINT ["/bin/sh", "-c", "/app/service/${ENV_APP_NAME}/tongcheng-${ENV_APP_NAME}"]
CMD []