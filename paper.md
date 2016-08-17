Christoph-Scheiner-Gymnasium Ingolstadt
=======================================

Seminararbeit aus dem wissenschaftspropädeutischen Seminar Android-Apps im Fach Informatik

Ein Lernmittelbibliothekssystem mit Fokus auf klar strukturierten Kommunikationsprotokollen

Verfasst von Dominik Okwieka

Reifeprüfungsjahrgang 2017

Kursleiter OStR Pabst

Inhaltsverzeichnis
------------------

1. Einleitung
2. Überblick über das Projekt
3. Gesichtspunkte von Protokollen
4. Das Protokoll
5. Detailbetrachtung des Servers und des Clients
6. Glossar
7. Literaturverzeichnis
8. Eidesstattliche Erklärung

1. Einleitung
-------------

> some quote?

*libmangler* ist ein Verwaltungssystem für Lernmittelbüchereien. Es besteht
aus einem Server und einem Android-Client, die mithilfe eines einfachen
Protokolls Daten austauschen. Dieses Protokoll ist jedoch trotz der Einfachheit
generell und ist vergleichbar mit einer Datenbanksprache wie SQL, jedoch
zugeschnitten auf die Anwendung.

Jedes Buchexemplar hat eine einzigartige Identifikationsnummer, welche als
QR-Code auf diesem befestigt wird. Die App kann diesen Code auslesen; der Nutzer
kann dann das Medium entleihen, zurückgeben, mit Notizen versehen,
kategorisieren, aus dem Verkehr ziehen oder ganz löschen. Gedacht ist die App
für die Leiter der Lernmittelbücherei, für die Lehrer, die Beschädigungen
notieren müssen, sowie für alle, die bei der Buchausgabe und der Rücknahme
beschäftigt sind.

Es gibt eine klare Trennung der Begriffe Buchexemplar (*Copy*) und Buch:
ersteres ist ein physikalisches Medium, letzteres bezieht sich auf eine
Ansammlung von Medien mit derselben ISBN. Zum Beispiel gibt es ein Buch
*Mathematik 8. Klasse*, aber 200 Exemplare, *Copies*, davon.


2. Überblick über das Projekt
-----------------------------

Wie anfangs erwähnt, besteht *libmangler* aus einem Server und einem Client.
Der Server ist in Go geschrieben und speichert alle Bücher, Copies und User
(Ausleiher, also die Schüler) in einem Dateibaum in einem Textformat ab, das
sich leicht von Hand verändern lässt. Verglichen mit dem Client ist der Server
einfach; es werden Gos Stärken ausgespielt, zudem fehlt die Komplexität von
Android.

Der Client ist vergleichbar mit einem Fenster in die Daten des Servers: er
scannt einen QR-Code oder lässt den Nutzer eine Suchanfrage eintippen, fragt
den Server nach dem Gesuchten, speichert nur dieses und erlaubt dann einige
Aktionen bezüglich dieser Daten. Trotz einfacher Anforderungen stellte sich der
Client als schwieriger heraus als der Server, da blockierende
Netzwerkkommunikation in Android nicht möglich ist. Zudem ist es schwer, Daten
lebendig zu halten, weil der User die App pausieren oder rotieren könnte und
auf diese Weise immer einen neuen Prozess (*Activity*) startet.

Bevor wir zu einer genaueren Beschreibung der Komponenten kommen können, muss
das Protokoll verstanden sein. Seine Struktur ist die Struktur des Servers;
seine Struktur prägt auch den Client.


3. Gesichtspunkte von Protokollen
---------------------------------

Computer sind geprägt von formalen Sprachen: die meisten Programme erwarten
ihre Eingabedaten und ihre Konfiguration in einem bestimmten Format und geben
Daten mit einer bestimmten Struktur aus. Wenn man formale Sprachen mit
menschlichen Sprachen ersetzt und Dateien mit Büchern vergleicht, dann sind
Protokolle nichts anderes als Dialoge. Die einzelnen Sätze folgen einer
Grammatik, doch das ist nicht alles: man muss verhindern, dass
Missverständnisse entstehen, indem aneinander vorbeigeredet wird. Man muss
darauf achten, dass die Botschaft unverändert ankommt, dass sie *sicher*
ankommt, ohne überhört worden zu sein. Es kann viel schiefgehen.

In dieser Arbeit wird hauptsächlich von Anwendungsprotokollen die Rede sein,
also Protokollen der siebten Schilcht des OSI-Modells. IP, TCP, ARP, usw., sind
natürlich auch Protokolle, lassen sich aber schwer mit Anwendungsprotokollen
vergleichen.


### 3.1 Anforderungen

Nichts ist wichtiger als die Funktionsfähigkeit des Protokolls: es muss eine
sinnvolle Kommunikation zwischen Hosts erlauben. Dazu gehört natürlich, dass
nichts in falscher Reihenfolge, unvollständig, korrumpiert, oder am falschen
Ziel ankommt. Deshalb verwenden die meisten Protokolle TCP als Unterbau, dass
all diese Dinge garantieren kann. Böswillige Betrachtung und Manipulation der
Daten kann man durch z.B. SSL/TLS verhindern, das leicht integrierbar ist.
Protokolle, die TCP verwenden, sind stream-basiert, das heißt, es scheint für
sie eine bidirektionale Verbindung der Hosts zu bestehen. Solch ein Stream wird
durch den sogenannten Three-Way-Handshake aufgebaut, was zu Beginn dauert.

Doch nicht immer laufen Protokolle über TCP. Prominentes Beispiel ist das
Domain Name System (DNS), das primär der Auflösung von Hostnamen in
IP-Adressen dient. Das DNS-Protokoll verwendet UDP, eine Alternative zu TCP, das
nicht einmal garantiert, dass das Paket ankommt. Es wird aus zwei Gründen
verwendet: es hat geringere Latenzzeit, weil kein TCP-Stream aufgebaut werden
muss; zudem muss der DNS-Server sich nicht um offene Verbindungen sorgen [cit].

Viele Protokolle haben also Performanceanforderungen. Es gibt hier zwei
Größen: Bandbreite und Latenzzeit. Bandbreite ist die Datenmenge pro
Zeiteinheit, die über das Protokoll übertragen wird; Latenzzeit ist die Zeit,
bis die Antwort des entfernten Hosts beim Anfrager eintrifft. Je nachdem, was
die Anwendung ist, ist das eine wichtiger als das andere. Wer Videos übertragen
will, achtet auf Bandbreite. Wer ein Echtzeitmultiplayerspiel hat, den
interessieren möglichst geringe Latenzzeiten.

Die folgenden Attribute sind jedoch am wichtigsten: Robustheit, Testbarkeit,
Verständlichkeit und Portabilität. Ohne Robustheit und Portabilität kann ein
Protokoll im heterogenen Internet nicht überleben: es gibt meist mehrere, immer
leicht falsche Implementierungen, die auf einer Vielzahl unterschiedlicher
Architekturen und einer Vielzahl unterschiedlicher Betriebssysteme laufen. Ohne
Testbarkeit ist es unmöglich, genau diese Robustheit zu testen. Ohne
Verständnis für das Protokoll tappt der Entwickler im Dunkeln herum.

Um das Protokoll korrekt implementieren zu können, muss es einfach sein, denn
einfache Protokolle führen zu einfachen Implementierungen. Einfacher Code
lässt sich vollständiger testen, ist wartbar und portierbar.


### 3.2 Ansätze

Es ist Zeit, mehrere Ansätze zu vergleichen; jeder hat bestimmte Vor- und
Nachteile. Die folgende Liste enthält verschiedenartige Protokolle, manche mehr
und manche weniger bekannt.

 - 9P
 - DNS
 - FTP
 - HTTP
 - IMAP
 - mpmp
 - Protokoll im Protokoll
 - SMTP


