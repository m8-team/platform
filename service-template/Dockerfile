# Build from the monorepo root: docker build -f service-template/Dockerfile .
FROM golang:1.27.1-bookworm AS build
WORKDIR /src
COPY platform-sdk/ platform-sdk/
COPY service-template/ service-template/
WORKDIR /src/service-template
RUN CGO_ENABLED=0 go build -trimpath -o /product-service ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /product-service /product-service
USER 65532:65532
EXPOSE 8080 8081
ENTRYPOINT ["/product-service"]
