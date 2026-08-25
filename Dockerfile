FROM golang:1.21-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /technocore-webhooks .

FROM gcr.io/distroless/static-debian12
COPY --from=build /technocore-webhooks /technocore-webhooks
ENTRYPOINT ["/technocore-webhooks"]
