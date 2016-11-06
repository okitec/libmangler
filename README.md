W-Seminar Android: Fachbibliotheksverwaltung
============================================

*This is a school project. The documentation and project paper are written in German.*

Dieses Projekt realisiert ein Fachbibliothekssystem mit einem in Java geschriebenen
Android-Client und einem in Go geschriebenen Server. Das Protokoll ist minimalistisch
gehalten (siehe [Spezifikation](SPEC.md)).

Herunterladen und Builden
-------------------------

*libmangler* ist auf GitHub zu finden:

	git clone https://github.com/okitec/libmangler.git

Zum Builden benötigte Software:
 - [Go](https://golang.org/dl/)
 - [Ant](http://ant.apache.org/bindownload.cgi)
 - [Android SDK](https://developer.android.com/studio/index.html#downloads)
 - (nur evtl.) (Unix) mk aus [Plan 9 from User Space](https://github.com/9fans/plan9port)

Auf dem Handy muss zudem [ZXing](https://github.com/zxing/zxing) installiert sein, um
QR-Codes zu lesen.

Nachdem man Go installiert hat, muss man den [GOPATH setzen](https://golang.org/cmd/go/#hdr-GOPATH_environment_variable)
und den `libmangler`-Ordner in `$GOPATH/src/github.com/okitec/` verschieben. Dann wird
der Server, `manglersrv`, nach dem Kompilieren in `$GOPATH/bin` zu finden sein.

Wenn diese Programme installiert wurden, möge man ein Android-Smartphone oder -Tablet
per USB anschließen (Entwicklermodus einschalten und USB-Zugriff erlauben!) und kann
dann im Hauptverzeichnis der Repo tippen:

	mk
	mk install

Wenn man kein Unix hat und damit kein `mk`, kann man das Android-Projekt `client`
in Eclipse öffnen und dann auch kompilieren, oder man importiert es in Android Studio.
In diesem manuellen Fall kompiliert man den Server mit

	cd manglersrv
	go install

Konfigurieren und Verwenden
---------------------------

Man sollte einen Ordner erstellen und den Server, zuvor zu finden unter `$GOPATH/bin/manglersrv`,
dorthin verschieben, da in diesem Ordner die Datenbank, bestehend aus `copies`, `books`
und `users`, abgelegt werden soll. manglersrv liest und schreibt immer aus den Dateien,
die in dem Ordner liegen, in dem er gestartet wird. Deshalb bietet es sich an, ein
kleines Skript zu schreiben:

	cd server-ordner
	./manglersrv

Weitere Konfiguration ist nicht nötig. Das Log des Servers wird nur an stderr gesendet
und sollte evtl. in eine Datei redirectet werden.

	./manglersrv >>log

Attribution des Logos
---------------------

Das *libmangler*-Logo ist eine von Leander Dreier modifizierte Kopie von Rogerborrell
(Own work) [CC BY-SA 4.0 (http://creativecommons.org/licenses/by-sa/4.0)], via Wikimedia
Commons https://commons.wikimedia.org/wiki/File%3ADraw_book.png.
