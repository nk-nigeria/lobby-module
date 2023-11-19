PROJECT_NAME=github.com/ciaolink-game-platform/cgb-lobby-module
APP_NAME=lobby.so
APP_PATH=$(PWD)
NAKAMA_VER=3.19.0

update-submodule-dev:
	git checkout develop && git pull
	git submodule update --init
	git submodule update --remote
	cd ./cgp-common && git checkout develop && git pull && cd ..
	go get github.com/ciaolink-game-platform/cgp-common@develop
update-submodule-stg:
	git checkout staging && git pull
	git submodule update --init
	git submodule update --remote
	cd ./cgp-common && git checkout staging && git pull && cd ..
	go get github.com/ciaolink-game-platform/cgp-common@staging

build:
	# ./sync_pkg_3.11.sh
	go mod tidy
	go mod vendor
	docker run --rm -w "/app" -v "${APP_PATH}:/app" "heroiclabs/nakama-pluginbuilder:${NAKAMA_VER}" build -buildvcs=false --trimpath --buildmode=plugin -o ./bin/${APP_NAME}

syncdev:
	rsync -aurv --delete ./bin/${APP_NAME} root@cgpdev:/root/cgp-server-dev/dist/data/modules/bin/
	ssh root@cgpdev 'cd /root/cgp-server-dev && docker restart nakama_dev'

syncstg:
	rsync -aurv --delete ./bin/${APP_NAME} root@cgpdev:/root/cgp-server/dist/data/modules/
	ssh root@cgpdev 'cd /root/cgp-server && docker restart nakama'

dev: update-submodule-dev build

stg: update-submodule-stg build

3.19: 
	git submodule update --init
	git submodule update --remote
	cd ./cgp-common && git checkout v3.19.0 && git pull && cd ..
	go get github.com/ciaolink-game-platform/cgp-common@v3.19.0
	go mod tidy
	go mod vendor
	### build for deploy
	# docker run --rm -w "/app" -v "${APP_PATH}:/app" "heroiclabs/nakama-pluginbuilder:${NAKAMA_VER}" build -buildvcs=false --trimpath --buildmode=plugin -o ./bin/${APP_NAME}
	### build for using local 
	go build -buildvcs=false --trimpath --mod=vendor --buildmode=plugin -o ./bin/${APP_NAME}

run-dev:
	docker-compose up -d --build nakama && docker logs -f lobby
dev-first-run:
	docker-compose up --build nakama && docker restart lobby

proto:
	protoc -I ./ --go_out=$(pwd)/proto  ./proto/common_api.proto

local:
	# git submodule update --init
	# git submodule update --remote
	# go get github.com/ciaolink-game-platform/cgp-common@main
	./sync_pkg_3.11.sh
	go mod tidy
	go mod vendor
	rm ./bin/* || true
	go build -buildvcs=false --trimpath --mod=vendor --buildmode=plugin -o ./bin/${APP_NAME}
