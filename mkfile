all:
	mk spec.html
	mk server
	mk client

spec.html: SPEC.md
	echo '<!DOCTYPE html>' >spec.html
	echo '<meta charset="utf8">' >>spec.html
	markdown $prereq >>spec.html

server:
	cd manglersrv
	go install

client:
	cd client
	ant debug
	adb install bin/libmangler-debug.apk
