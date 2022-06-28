PROJECT_NAME=github.com/ciaolink-game-platform/cgp-lobby-module
APP_NAME=lobby.so
APP_PATH=$(PWD)

build:
	docker run --rm -w "/app" -v "${APP_PATH}:/app" heroiclabs/nakama-pluginbuilder:3.11.0 build --trimpath --buildmode=plugin -o ./bin/${APP_NAME}
	
sync:
	rsync -aurv --delete ./bin/${APP_NAME} root@cgpdev:/root/cgp-server/data/modules/
	ssh root@cgpdev 'cd /root/cgp-server && docker restart nakama'

dev:
	docker-compose up -d --build nakama && docker logs -f lobby
dev-first-run:
	docker-compose up --build nakama && docker restart lobby