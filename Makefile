validate_version:
ifndef VERSION
	$(error VERSION is undefined)
endif

release: validate_version
	mkdir -p ./releases

	# linux
	GOOS=linux go build -ldflags "-s -w -X main.version=${VERSION}" -o ezf ;\
	tar -zcvf ./releases/ezf_${VERSION}_linux.tar.gz ./ezf ;\

	# macos (arm)
	GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w -X main.version=${VERSION}" -o ezf ;\
	tar -zcvf ./releases/ezf_${VERSION}_macos_arm64.tar.gz ./ezf ;\

	# macos (amd)
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w -X main.version=${VERSION}" -o ezf ;\
	tar -zcvf ./releases/ezf_${VERSION}_macos_amd64.tar.gz ./ezf ;\

	# windows
	GOOS=windows go build -ldflags "-s -w -X main.version=${VERSION}" -o ezf ;\
	tar -zcvf ./releases/ezf_${VERSION}_windows.tar.gz ./ezf ;\

	rm ./ezf