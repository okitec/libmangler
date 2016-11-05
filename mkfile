all:V:
	mk spec.html
	mk server
	mk client

spec.html: SPEC.md
	echo '<!DOCTYPE html>' >spec.html
	echo '<meta charset="utf8">' >>spec.html
	markdown $prereq >>spec.html

paper.odt: paper.md local.bib metadata.yaml
	pandoc --filter pandoc-citeproc --biblatex -o $target paper.md metadata.yaml

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
