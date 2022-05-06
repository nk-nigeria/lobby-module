FROM heroiclabs/nakama-pluginbuilder:3.11.0 AS builder

ENV GO111MODULE on
ENV CGO_ENABLED 1
#ENV GOPRIVATE "github.com/ciaolink-game-platform/cgp-lobby-module"

WORKDIR /backend
COPY . .

RUN go build --trimpath --mod=readonly --buildmode=plugin -o ./lobby.so

FROM heroiclabs/nakama:3.11.0

COPY --from=builder /backend/lobby.so /nakama/data/modules
COPY --from=builder /backend/bin/chinese-poker.so /nakama/data/modules
COPY --from=builder /backend/local.yml /nakama/data/
