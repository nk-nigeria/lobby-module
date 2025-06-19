PROJECT_NAME=github.com/nk-nigeria/lobby-module
APP_NAME=lobby_plugin.so
APP_PATH=$(PWD)
NAKAMA_VER=3.27.0

update-submodule-dev:
	go get github.com/nk-nigeria/cgp-common@develop
update-submodule-stg:
	git checkout staging && git pull
	git submodule update --init
	git submodule update --remote
	cd ./cgp-common && git checkout staging && git pull && cd ..
	go get github.com/nk-nigeria/cgp-common@staging

cpdev:
	scp ./bin/${APP_NAME} nakama:/root/cgp-server-dev/dist/data/modules/
cplive:
	scp ./bin/${APP_NAME} nakama:/root/cgp-server/dist/data/modules/bin
build:
	go mod tidy
	go mod vendor
	docker run --rm -w "/app" -v "${APP_PATH}:/app" "heroiclabs/nakama-pluginbuilder:${NAKAMA_VER}" build -buildvcs=false --trimpath --buildmode=plugin -o ./bin/${APP_NAME} && cp ./bin/${APP_NAME} ../bin/
build-dev: build cpdev

build-cmd:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 GOPRIVATE=github.com/nakamaFramework go build --trimpath --buildmode=plugin -o ./bin/${APP_NAME}

build-cross:
	./sync_pkg_3.11.sh
	go mod tidy
	go mod vendor
	docker run -it --rm -w "/app" \
	  --platform linux/amd64 \
      -v ${APP_PATH}:/app \
      docker.elastic.co/beats-dev/golang-crossbuild:1.18-main \
      --build-cmd "make build-cmd" \
      -p "linux/amd64"

syncdev:
	rsync -aurv --delete ./bin/${APP_NAME} root@cgpdev:/root/cgp-server-dev/dist/data/modules/bin/
	ssh root@cgpdev 'cd /root/cgp-server-dev && docker restart nakama_dev'

syncstg:
	rsync -aurv --delete ./bin/${APP_NAME} root@cgpdev:/root/cgp-server/dist/data/modules/bin
	ssh root@cgpdev 'cd /root/cgp-server && docker restart nakama'

dev: update-submodule-dev build

stg: update-submodule-stg build

v3.19.0: 
	git submodule update --init
	git submodule update --remote
	cd ./cgp-common && git checkout develop && git pull origin develop && cd ..
	go get github.com/nakamaFramework/cgp-common@develop
	go mod tidy
	go mod vendor
	### build for deploy
	docker run --rm -w "/app" -v "${APP_PATH}:/app" "heroiclabs/nakama-pluginbuilder:${NAKAMA_VER}" build -buildvcs=false --trimpath --buildmode=plugin -o ./bin/${APP_NAME}
	### build for using local 
	# go build -buildvcs=false --trimpath --mod=vendor --buildmode=plugin -o ./bin/${APP_NAME}

run-dev:
	docker-compose up -d --build nakama && docker logs -f lobby
dev-first-run:
	docker-compose up --build nakama && docker restart lobby

proto:
	protoc -I ./ --go_out=$(pwd)/proto  ./proto/common_api.proto

local:
	./sync_pkg_3.11.sh
	go mod tidy
	go mod vendor
	rm ./bin/* || true
	go build -buildvcs=false --trimpath --mod=vendor --buildmode=plugin -o ./bin/${APP_NAME}
