# Build plugin bằng Go 1.24.3
FROM heroiclabs/nakama-pluginbuilder:3.27.0 AS builder

ENV GO111MODULE on
ENV CGO_ENABLED 1

WORKDIR /backend
COPY . .

RUN go mod tidy
RUN go build -buildvcs=false --trimpath --buildmode=plugin -o /plugin/lobby_plugin.so

# Runtime Nakama 3.27.0
# FROM heroiclabs/nakama:3.27.0

# COPY --from=builder /backend/lobby_plugin.so /nakama/data/modules
# COPY --from=builder /backend/local.yml /nakama/data/
FROM scratch
COPY --from=builder /plugin/lobby_plugin.so /plugin/lobby_plugin.so
# CMD ["sleep", "3600"]
CMD ["true"]