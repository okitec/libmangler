all:V:
	mk spec.html
	mk server
	mk client

spec.html: SPEC.md
	echo '<!DOCTYPE html>' >spec.html
	echo '<meta charset="utf8">' >>spec.html
	markdown $prereq >>spec.html

paper.odt: paper.md local.bib metadata.yaml ref.odt
	mk note
	cat paper.md note metadata.yaml | pandoc --filter pandoc-citeproc --biblatex -o paper.odt --reference-odt ref.odt
	rm note

note:V:
	echo >note
	echo 'Diese ODT-Datei wurde aus `paper.md` um `' >>note
	date >>note
	echo '` durch den Befehl' >>note
	echo '`pandoc --filter pandoc-citeproc --biblatex -o paper.odt paper.md metadata.yaml --reference-odt ref.odt`' >>note
	echo 'generiert.' >>note
	echo >>note
	echo '### Quellen' >>note
	echo >>note
	
server:V:
	cd manglersrv
	go install

client:V:
	cd client
	ant debug

install:V:
	adb uninstall de.csgin.libmangler
	adb install bin/libmangler-debug.apk

clean:V:
	cd client
	ant clean
