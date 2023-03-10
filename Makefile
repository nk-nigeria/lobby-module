PROJECT_NAME=github.com/ciaolink-game-platform/cgb-lobby-module
APP_NAME=lobby.so
APP_PATH=$(PWD)

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
	cd ./cgp-common && git checkout main && cd ..
	go get github.com/ciaolink-game-platform/cgp-common@main

build:
	./sync_pkg_3.11.sh
	go mod tidy
	go mod vendor
	docker run --rm -w "/app" -v "${APP_PATH}:/app" heroiclabs/nakama-pluginbuilder:3.11.0 build -buildvcs=false --trimpath --buildmode=plugin -o ./bin/${APP_NAME}

syncdev:
	rsync -aurv --delete ./bin/${APP_NAME} root@cgpdev:/root/cgp-server/dev/data/modules/
	ssh root@cgpdev 'cd /root/cgp-server/dev && docker restart nakama_dev'

syncstg:
	rsync -aurv --delete ./bin/${APP_NAME} root@cgpdev:/root/cgp-server/data/modules/
	ssh root@cgpdev 'cd /root/cgp-server && docker restart nakama'
dev: update-submodule-dev build
stg: update-submodule-stg build
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
	go mod tidy
	go mod vendor
	rm ./bin/* || true
	go build -buildvcs=false --trimpath --mod=vendor --buildmode=plugin -o ./bin/${APP_NAME}
