test:
	GO111MODULE=on go test github.com/bokuweb/gopher-boy/...

reg:
	reg-cli ./test/actual ./test/expect ./test/diff

reg-update:
	reg-cli ./test/actual ./test/expect ./test/diff -U