all:V:
	mk spec.html
	mk server
	mk client

spec.html: SPEC.md
	echo '<!DOCTYPE html>' >spec.html
	echo '<meta charset="utf8">' >>spec.html
	markdown $prereq >>spec.html

server:V:
	cd manglersrv
	go install

client:V:
	cd client
	ant debug
	adb uninstall de.csgin.libmangler
	adb install bin/libmangler-debug.apk

clean:V:
	cd client
	ant clean
