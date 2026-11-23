FROM node:22-alpine AS web-builder
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci --ignore-scripts
COPY web/ ./
RUN npm run build

FROM golang:1.27-alpine AS go-builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web-builder /src/internal/learning/adapters/httpapi/web/dist/ internal/learning/adapters/httpapi/web/dist/
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/englishcraft ./cmd/englishcraft

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=go-builder /out/englishcraft /englishcraft
EXPOSE 8080
USER nonroot:nonroot
HEALTHCHECK --interval=30s --timeout=3s --start-period=20s --retries=3 CMD ["/englishcraft", "health"]
ENTRYPOINT ["/englishcraft"]
