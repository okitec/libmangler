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

Wenngleich moderne Computer zumeist eine batteriebetriebene Echtzeituhr
besitzen, muss diese mit genaueren Uhren synchronisiert werden, damit sie
korrekt bleibt. Schon 1985 hatte das Network Time Protocol eine
Referenzimplementierung und wurde in RFC 958 dokumentiert. In weiterentwickelter
Form wird das Protokoll in fast allen internetfähigen Systemen verwendet.

Die NTP-Hierarchie ist in sogenannte Strata eingeteilt: Stratum 1 bezeichnet die
an genauen Zeitgebern angeschlossenen Computer (primäre Zeitserver). Generell
greifen Stratum n-Rechner jeweils auf Stratum (n-1)-Rechner zu und gleichen sich
zudem untereinander ab. Das System versucht, einen möglichst minimalen Baum an
Verbindungen aufzubauen, um die Latenzzeiten zu Stratum 1 gering zu halten. Der
restliche Fehler wird durch eine auf Statistiken basierenden Formel entfernt.

Je nach Anwendung kommt eine der drei Betriebsmodi zum Einsatz: Client/Server,
bei dem der Client vom Server pullt; der symmetrische Modus, bei dem sich zwei
*Peers* gegenseitig synchronisieren; Broadcast, bei dem der Server an mehrere
Clients Pakete sendet. Mit jedem Paket wird ein *Packet Mode*-Wert übertragen,
der den Modus identifiziert. Es gibt drei Zeitformate: *Short*, *Timestamp* und
*Date*. Wenn möglich, wird das Datumsformat verwendet [RC5905, 6], das aus
einer *Era Number*, einem in Sekunden gemessenen *Era Offset* und einem Bruch
besteht. Die *Era Number* bezeichnet den Bereich, in dem der 32-bit Offset nicht
überläuft. Momentan sind wir in Era 0; ab dem 08. Februar 2036 werden wir in
Era 1 sein. Das im Protokoll verwendete *Timestamp*-Format hat einen 32-bit
Sekundenzähler und einen Bruch; das *Short*-Format ist ähnlich, hat aber nur
16 Bit Präzision.

TCP kann hier nicht verwendet werden, weil es verlorene Pakete wieder
überträgt und dadurch die Zeitstempel in diesen verfälscht [citation needed],
deswegen wird UDP auf Port 123 verwendet. NTP verwendet konventionelle binäre
Pakete mit einem Header und einem aus vier Timestamps bestehenden Payload. Ein
invalider Wert im Header, Stratum 0, initiiert ein *Kiss-o'-Death*-Paket, mit
welchem Kontrollcodes übertragen werden; diese sind Vier-Zeichen-ASCII-Strings
an der Stelle, an der sonst die Referenz-ID des Zeitgebers steht (z.B. "GPS").


XXX mehr Fokus auf Protokoll, weniger auf Umstände?

 - src: RFC 5905, Wikipedia

#### 3.2.3 HTTP – Hypertext Transfer Protocol

Das Web ist die *Killer application* des Internets, so wie die Glühbirne die
*Killer application* der Elektrizität war. Viele Laien können heute die
Begriffe "Web" und "Internet" nicht mehr auseinanderhalten. Das Hypertext
Transfer Protocol ist so erfolgreich, dass es als Transportprotokoll für alles
gebraucht wird, obwohl es nicht auf generelle Kommunikation ausgerichtet ist.

HTTP hat für Internetstandards typische Merkmale: es verwendet Textbefehle, hat
dreistellige Statuscodes mit einem angefügten, menschenverständlichen Text
(z.B. `404 File Not Found`) und hat ein Headerformat, bei dem jedes Feld die
Form 'Feldname: Wert' hat.

Gemeinhin haben Clients die Initiative und senden hauptsächlich `GET`- und
`POST`-Requests. Eine HTTP-Verbindung hat keinen Zustand. Der Server beantwortet
den Request und vergisst dann den Client. In HTTP/1.0 wurde für jeden Request
einen neue Verbindung geöffnet. Persistente Verbindungen, die der Normalzustand
seit HTTP/1.1 sind, erlauben einen Request nach dem anderen in der Verbindung;
die Latenzzeit durch das Verbindungsöffnen entfällt. *Pipelining*, d.h. das
Senden mehrerer Requests auf einmal und das Empfangen der Antworten auf einmal,
optimiert den Prozess weiter. Pipelining erwies sich als Fehlschlag auf
Clientseite; nur Opera hat eine aktivierte und stabile Implementierung [citation needed].

Im Kern ist HTTP ideal für die Aufgabe, nicht interaktive Webseiten und andere
Dateien zu übertragen. Ein Client gibt den Dateipfad auf dem Server an, der
Server sendet die Datei. Kürzer kann man diese Interaktion nicht gestalten;
problematisch wird es, wenn eine Seite viele Assets von vielen Servern, z.B. von
Tracking- und Ad-Servern einbindet, denn persistente Verbindungen helfen auch
hier nicht. Newsseiten sind häufige Übeltäter.

HTTP ist nicht für bidirektionale Kommunikation gedacht, da die
Zustandslosigkeit im Weg steht. Cookies sind ein Hack, um diese zu umgehen, und
niemand mag Cookies. Ein anderer Weg sind Parameter in der URL, die den ganzen
Zustand übertragen und leicht zu manipulieren sind; mitunter wird grob
fahrlässig ein verschlüsseltes Passwort übertragen [Fahrenlernen Max].

Ein neuer binärer Standard, HTTP/2, wurde inzwischen veröffentlicht und wird
von allen weit verwendeten Browsern unterstützt. Neben mehreren anderen
Änderunge kann der Server nun Dateien pushen, für die der Client
wahrscheinlich sowieso eine Anfrage gestellt hätte; z.B. würden beim Aufruf
einer Seite gleich die CSS-Dateien und etwaiger Javascript-Code neben dem
HTML-Text gesendet. Anfang September 2016 verwendeten 9.8% der 10 Millionen
meistbesuchten Websites HTTP/2 (src)[https://w3techs.com/technologies/details/ce-http2/all/all].


#### 3.2.4 IMAP – Internet Message Access Protocol

Mailboxen lassen sich mit dem *Post Office Protocol* (POP), dem *Internet
Message Access Protocol* (IMAP) oder via einem Webmail-Interface im Browser
verwalten, falls man nicht selbst Admin eines Mailservers ist. Bei POP ist es
Konvention, die Nachrichten auf dem Server nach dem Abrufen zu löschen; die
Mails residieren auf dem Client, wie auch der Verzeichnisbaum mit dem
Posteingang, dem Postausgang und nutzererzeugten Ordnern.

IMAP ist eine neuere Entwicklung, um seine Nachrichtenordner auf dem Server zu
verwalten; dadurch kann man von mehreren Geräten auf denselben Baum zugreifen.
Der Client cacht die Mails nur; der Server hat die relevante Kopie.
Verständlicherweise ist IMAP komplexer als POP3.

Ich will IMAP deswegen ansprechen, weil es *Tags* verwendet, wie auch das
*libmangler*-Protokoll, und weil der Server von sich aus senden kann. Das
folgende Exzerpt in [RFC 3501, Sektion 8] soll das nun verdeutlichen. Zeilen mit
einem `*` werden vom Server in Eigeninitiative gesendet (Zeile 1), oder deuten
die Kontinuation des Outputs an. Die vom Client generierten alphanumerischen
Tags, hier `a001` und `a002`, müssen eindeutig sein; Anfrage und Antwort haben
denselben Tag. Bei Antworten folgt dann `OK` (Erfolg), `NO` (Fehlschlag), oder
`BAD` (formaler Fehler).

	S:   * OK IMAP4rev1 Service Ready
	C:   a001 login mrc secret
	S:   a001 OK LOGIN completed
	C:   a002 select inbox
	S:   * 18 EXISTS
	S:   * FLAGS (\Answered \Flagged \Deleted \Seen \Draft)
	S:   * 2 RECENT
	S:   * OK [UNSEEN 17] Message 17 is the first unseen message
	S:   * OK [UIDVALIDITY 3857529045] UIDs valid
	S:   a002 OK [READ-WRITE] SELECT completed

	[...]

#### 3.2.5 mpmp

*mpmp* ist kein Internetstandard. Es ist ein noch unfertiger Monopoly-Klon im
Stil der Weimarer Republik, bei dem man über ein Netzwerk spielen kann. Es ist
das Informatikprojekt der elften Klasse, das einige aus dem Seminar erstellt
haben und nun von mir instandgehalten wird. Das Protokoll ist meine Schöpfung,
weswegen ich über die speziellen Entscheidungen schreiben will, die in das
Protokoll einflossen.

Der Server enthält den Spielzustand; die Clients cachen diesen, stellen ihn dar
und senden Befehle an den Server, der Änderungen des Spielzustands allen
Clients mitteilt. Client und Server werden aus demselben Code kompiliert und
verwenden ein völlig symmetrisches Proptokollsystem. Sowohl Client als auch
Server senden Befehle mit Argumenten aus und quittieren diese jeweils mit
`+JAWOHL` oder `-NEIN`, gefolgt von einem Fehlerstring. Die Befehle, die allen
Clients übermittelt werden, enden per Kobention in `-update`. Das Beispiel
zeigt auch, dass einige der `+JAWOHL`s noch fehlen. Außerdem sieht man den
einzigartigen `clientlist-update`-Befehl, der mehrere Zeilen Payload hat, deren
Anzahl das erste Argument nennt, hier `1`.Durch einen Mitschnitt des Protokolls
ab dem Beginn kann man den gesamten Verlauf des Spiels verfolgen.

	S: +JAWOHL Willkommen, Genosse! Subscriben Sie!
	C: subscribe player #0f0f0f oki
	S: clientlist-update 1
	S: #0F0F0F: Player: oki
	C: +JAWOHL
	C: chat Dies ist Chat!
	S: chat-update (oki) Dies ist Chat!
	C: +JAWOHL
	C: start-game
	S: pos-update 12 oki
	S: turn-update 12 1 oki
	S: start-update
	C: +JAWOHL
	C: end-turn
	S: pos-update 21 oki
	S: turn-update 9 0 oki
	C: +JAWOHL
	C: buy-plot 21 oki
	S: money-update 25600 oki
	S: show-transaction -4400 derp
	S: plot-update 21 0 nohypothec oki
	C: ragequit
	S: clientlist-update 1
	S: #0F0F0F: Spectator: oki
	C: +JAWOHL

