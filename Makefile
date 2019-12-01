export GO111MODULE=on

default: nomad-bundle.tgz

clean:
	rm -rf bin nomad-bundle.tgz*

nomad-bundle.tgz: bin/dysweb_linux_amd64 bin/dysweb_linux_arm64 bin/dysweb_linux_arm
	tar -czvf $@ bin

publish: nomad-bundle.tgz
	OUT=$(shell sha1sum $< | awk '{print $$1 ".tgz"}'); \
	mv $< $$OUT && rclone copy $$OUT b2:blx-public/nomad/dysweb/ && \
	rm $$OUT

bin/dysweb_%:
	CGO_ENABLED=0 GOOS=linux GOARCH=$(subst bin/dysweb_linux_,,$@) go build -o $@ cmd/dysweb.go
