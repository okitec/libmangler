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
7. Danksagungen
8. Literaturverzeichnis
9. Eidesstattliche Erklärung

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
also Protokollen der siebten Schicht des OSI-Modells. IP, TCP, ARP, usw., sind
natürlich auch Protokolle, lassen sich aber schwer mit Anwendungsprotokollen
vergleichen.


### 3.1 Anforderungen

Nichts ist wichtiger als die Funktionsfähigkeit des Protokolls: es muss eine
sinnvolle Kommunikation zwischen Hosts erlauben. Dazu gehört natürlich, dass
nichts in falscher Reihenfolge, unvollständig, korrumpiert, oder am falschen
Ziel ankommt. Deshalb verwenden die meisten Protokolle TCP als Unterbau, dass
all diese Dinge garantieren kann. Böswillige Betrachtung und Manipulation der
Daten kann man z.B. durch SSL/TLS verhindern, das leicht integrierbar ist.
Protokolle, die TCP verwenden, sind stream-basiert, das heißt, es scheint für
sie eine bidirektionale Verbindung der Hosts zu bestehen. Solch ein Stream wird
durch den sogenannten Three-Way-Handshake aufgebaut: der Client sendet ein
SYN-Paket, der Server antwortet mit SYN-ACK, der Client antwortet darauf mit
einem ACK.

Doch nicht immer laufen Protokolle über TCP. Prominentes Beispiel ist das
Domain Name System (DNS), das primär der Auflösung von Hostnamen in
IP-Adressen dient. Das DNS-Protokoll verwendet UDP (User Datagram Protocol), die
verbindungslose Alternative zu TCP. Bei UDP werden einzelne Pakete übertragen, von
denen man nicht weiß, ob und in welcher Reihenfolge sie ankommen. Die Pakete werden
auch *Datagramme* genannt, daher der Protokollname. DNS verwendet UDP, um die
Kosten des Three-Way-Handshake zu vermeiden;; zudem muss der DNS-Server sich nicht
um offene Verbindungen sorgen [citation needed].

Viele Protokolle haben also Performanceanforderungen. Es gibt hier zwei
Größen: Bandbreite und Latenzzeit. Bandbreite ist die Datenmenge pro
Zeiteinheit, die über das Protokoll übertragen wird; Latenzzeit ist die Zeit,
bis die Antwort des entfernten Hosts beim Anfrager eintrifft (*Round Trip Time*, RTT,
"Ping"). Je nachdem, was die Anwendung ist, ist das eine wichtiger als das andere.
Wer Videos übertragen will, achtet auf Bandbreite. Wer ein Echtzeitmultiplayerspiel
hat, den interessieren möglichst geringe Latenzzeiten.

Die folgenden Attribute sind jedoch am wichtigsten: Robustheit, Testbarkeit,
Verständlichkeit und Portabilität. Ohne Robustheit und Portabilität kann ein
Protokoll im heterogenen Internet nicht überleben: es gibt meist mehrere, immer
leicht falsche Implementierungen, die auf einer Vielzahl unterschiedlicher
Architekturen und einer Vielzahl unterschiedlicher Betriebssysteme laufen. Ohne
Testbarkeit ist es unmöglich, genau diese Robustheit zu testen. Ohne
Verständnis für das Protokoll tappt der Entwickler im Dunkeln herum.
Dokumentation ist der Mörtel, der diese Tugenden zusammenhält.

Um das Protokoll korrekt implementieren zu können, muss es einfach sein, denn
einfache Protokolle führen zu einfachen Implementierungen. Einfacher Code
lässt sich vollständiger testen, ist wartbar und portierbar.

### 3.2 Ansätze

Es ist Zeit, mehrere Ansätze zu vergleichen; jeder hat bestimmte Vor- und
Nachteile. Die folgende Liste enthält verschiedenartige Protokolle, manche mehr
und manche weniger bekannt. Die ersten zwei Beispiele werden binär codiert,
der Rest ist textbasiert.

 - 9P
 - NTP
 - HTTP
 - IMAP
 - mpmp
 - Protokoll im Protokoll

#### 3.2.1 9P

*9P* ist das Dateisystemprotokoll des Betriebssystems *Plan 9 from Bell Labs*
[plan9]. Plan 9 wurde entwickelt, um das Unix-Prinzip *Everything is a file*
weiterzutreiben: alles – Geräte, Mailboxen, das Netzwerksystem, das
Grafiksystem und viel mehr – wird durch *Dateisysteme* repräsentiert, deren
Daten zumeist on-the-fly generiert werden (vgl. `/proc`). Jeder Prozess hat
einen eigenen sogenannten *Namespace*, der die Ansammlung aller von diesem
Prozess gemounteten Dateisysteme ist. Der Zugriff auf diese findet über 9P
statt; zur Implementierung eines eigenen Dateisystems muss man nur einen
9P-Server schreiben, was durch Hilfsfunktionen sehr einfach ist [citation
needed]. Die 9P-Verbindung wird zumeist über TCP getunnelt, wenn der Server
nicht lokal ist. Man kann das `/proc`-Verzeichnis eines anderen Systems mounten
und dann die dortigen Prozesse debuggen. Man kann den Bildschirm, die Maus und
die Tastatur eines anderen Systems mounten und diesen dann als Terminal
verwenden. Man kann die Zwischenablage eines anderen Systems mounten und so
auslesen oder modifizieren.

Das Protokoll kümmert sich um das Navigieren im Verzeichnisbaum sowie dem
Erstellen, Öffnen, Lesen, Schreiben und Löschen von Dateien. Die Belastung des
Protokolls ist vielseitig: manchmal werden wenige, große Pakete versendet, so
z.B. beim Lesen großer Dateien von einer Festplatte. Meist jedoch werden viele
kleine Pakete versendet, da kurze Strings in die Kontrolldateien von Geräten
geschrieben werden. In diesem Fall kann die Größe der Paket-Header Überhand
nehmen. Die Entwickler haben darauf geachtet, den Header möglichst kurz zu
halten [citation needed].

9P verwendet binär kodierte Header und identifiziert offene Dateien mit
eindeutigen Ganzzahlen, die *Fids* genannt werden, ist also zustandsbasiert. Der
Client beginnt jede "Transaktion" mit einer T-Message (*T* steht für
*transmit*) und der Server antwortet mit einer R-Message (*R* steht für
*reply*). Jede T-Message erhält vom Client einen eindeutigen *Tag*; die Antwort
des Servers hat denselben. *Tags* finden sich auch in IMAP und im
*libmangler*-Protokoll [9p]. Fehler werden gemeldet, indem ein spezielles Paket,
`Rerror`, gesendet wird; dieses enthält einen String, das den Fehler
beschreibt.

#### 3.2.2 NTP – Network Time Protocol

Wenngleich moderne Computer zumeist eine batteriebetriebene Echtzeituhr besitzen,
muss diese mit genaueren Uhren synchronisiert werden, damit sie korrekt bleibt.
Schon 1985 hatte das Network Time Protocol eine Referenzimplementierung und wurde
in RFC 958 dokumentiert. In weiterentwickelter Form wird das Protokoll in fast
allen internetfähigen Systemen verwendet.

Die NTP-Hierarchie ist in sogenannte Strata eingeteilt: Stratum 1 bezeichnet die
an genauen Zeitgebern angeschlossenen Computer (primäre Zeitserver). Generell greifen
Stratum n-Rechner jeweils auf Stratum (n-1)-Rechner zu und gleichen sich zudem
untereinander ab. Das System versucht, einen möglichst minimalen Baum an
Verbindungen aufzubauen, um die Latenzzeiten zu Stratum 1 gering zu halten. Der
restliche Fehler wird durch eine auf Statistiken basierenden Formel entfernt.

Je nach Anwendung kommt eine der drei Betriebsmodi zum Einsatz: Client/Server,
bei dem der Client vom Server pullt; der symmetrische Modus, bei dem sich zwei
*Peers* gegenseitig synchronisieren; Broadcast, bei dem der Server an mehrere
Clients Pakete sendet. Mit jedem Paket wird ein *Packet Mode*-Wert übertragen,
der den Modus identifiziert. Es gibt drei Zeitformate: *Short*, *Timestamp*
und *Date*. Wenn möglich, wird das Datumsformat verwendet [RC5905, 6], das aus
einer *Era Number*, einem in Sekunden gemessenen *Era Offset* und einem Bruch
besteht. Die *Era Number* bezeichnet den Bereich, in dem der 32-bit Offset nicht
überläuft. Momentan sind wir in Era 0; ab dem 08. Februar 2036 werden wir in
Era 1 sein. Das im Protokoll verwendete *Timestamp*-Format hat einen 32-bit
Sekundenzähler und einen Bruch; das *Short*-Format ist ähnlich, hat aber nur
16 Bit Präzision.

TCP kann hier nicht verwendet werden, weil es verlorene Pakete wieder überträgt
und dadurch die Zeitstempel in diesen verfälscht [citation needed], deswegen
wird UDP auf Port 123 verwendet. NTP verwendet konventionelle binäre Pakete
mit einem Header und einem aus vier Timestamps bestehenden Payload. Ein invalider
Wert im Header, Stratum 0, initiiert ein *Kiss-o'-Death*-Paket, mit welchem
Kontrollcodes übertragen werden; diese sind Vier-Zeichen-ASCII-Strings an der
Stelle, an der sonst die Referenz-ID des Zeitgebers steht (z.B. "GPS").

XXX mehr Fokus auf Protokoll, weniger auf Umstände?

 - src: RFC 5905, Wikipedia

#### 3.2.3 HTTP/1.1 – Hypertext Transfer Protocol

 - textbasiert
 - XYZ-Fehlercodes
 - eine Verbindung pro Datei, sofern nicht Keep Alive verwendet wird
 - zustandslos
 - octet-counting

#### 3.2.4 IMAP – Internet Message Access Protocol

 - textbasiert
 - verwendet Tags
 - XYZ-Fehlercodes
 - hält einen Verzeichnisbaum mit Nachrichten instand

#### 3.2.5 mpmp


#### 3.2.6 Protokoll im Protokoll

