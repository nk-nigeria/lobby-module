PROJECT_NAME=github.com/ciaolink-game-platform/cgp-lobby-module
APP_NAME=lobby.so
APP_PATH=$(PWD)

build:
	docker run --rm -w "/app" -v "${APP_PATH}:/app" heroiclabs/nakama-pluginbuilder:3.11.0 build --trimpath --buildmode=plugin -o ./bin/${APP_NAME}
	
sync:
	rsync -aurv --delete ./bin/${APP_NAME} cgpdev:/root/cgp-server/data/modules/
	ssh cgpdev 'cd /root/cgp-server && docker restart nakama'

