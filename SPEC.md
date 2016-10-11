Spezifikation der Bibliotheksverwaltung
=======================================

Version: Protokollversion 6

0. Index
--------

1. [Projektziele](#1-projektziele)
2. [Einschränkungen](#2-einschränkungen)
3. [Architektur und Implementierungssprachen](#3-architektur-und-implementierungssprachen)
4. [Operationen/Protokoll](#4-operationenprotokoll)
	1. [Allgemeines](#41-allgemeines)
	2. [Befehlsliste](#42-befehlsliste)
5. [Datenstrukturen](#5-datenstrukturen)
6. [Referenzen](#6-referenzen)

1. Projektziele
---------------

Die Bibliotheksverwaltung soll der Lehrmittelbücherei helfen, Ausleihe und
Rückgabe schnell durchzuführen und Beschädigungen zu notieren. Die
Besitzerschaft eines Buchexemplars (sog. *Copy*) soll mitgeloggt werden. Das
Einfügen eines neuen Buches und die Ablösung alter Ausgaben muss auch einfach
sein.

Da es eine Lehrmittelbücherei ist, sollten Informationen wie Jahrgangstufe,
Fach und Zweig pro Buch gespeichert werden können. Das Zusammenstellen einer
Bücherliste für einen Schüler einer bestimmten Jahrgangsstufe und eines
bestimmten Zweigs wäre auch angebracht.

Bücher werden durch aufgeklebte QR-Codes eindeutig identifiziert. Die
Generation von Codes für neue Bücher muss möglich sein, indem z.B. eine
Bilddatei für eine bestimmte ID erzeugt wird.


2. Einschränkungen
------------------

Die App muss auf einem Android-Handy laufen. Der Server ist in Go geschrieben
und statisch kompiliert, sodass eine Binary für Windows ohne Installation oder
irgendwelche Dateien direkt lauffähig ist.


3. Architektur und Implementierungssprachen
-------------------------------------------

Die Bibliotheksverwaltung besteht aus einem Android-Client und einem Server.
Die Android-App verfügt über Fähigkeiten wie

 - QR-Code lesen
 - Infos zum Buch anzeigen
 - Ausleihe, Rückgabe, Beschädigung oder Aussortierung melden

und ist ähnlich einem *dumb terminal*, das nur Daten vom Server fetcht und um
Aktionen bittet. Sie wird in Java geschrieben. Die UI wird semi-dynamisch
generiert; vorgefertigte Layouts werden mithilfe eines `ViewFlipper`s ein- und
ausgeblendet, um der unnötigen Komplexität zu vieler Activities zu entgehen.

4. Operationen/Protokoll
------------------------

### 4.1 Allgemeines

Das Client-Server-Protokoll muss über einen zuverlässigen Stream übertragen
werden. Dies kann durch TCP gewährleistet werden. Zudem sollte er durch TLS
verschlüsselt sein, was jedoch wohl nicht implementiert wird. Die Portnummer
ist 40000.

Das Protokoll ist ein UTF-8-basiertes Textprotokoll. IMHO ist es nicht nötig,
dem Semi-Standard zu folgen und alles über HTTP zu machen; ich preferiere es,
Plaintext-Protokolle in derselben Netzwerkschicht wie HTTP zu kreieren, anstatt
über HTTP zu "tunneln". Aus Prinzip verwende ich kein XML.

Die Datenmengen sind klein und die Performanceanforderungen gering. Die kurze
Schreibweise ist jedoch mnemonisch und einfach genug, um von einem Menschen
verstanden und geschrieben werden zu können.

Generelles Request-Format:

		selector cmd parameters \n

Das Antwortformat des Servers ist von den ausgeführten Kommandos abhängig,
endet jedoch auf jeden Fall mit einer Endmarkierungszeile aus drei
Bindestrichen (`---`).

### 4.2 Befehlsliste

>>> **XXX Antworten und Fehler exakt spezifizieren**

		.       aktuelle Selektion (Dot). Implizit in folgenden nicht selektierenden Kommandos.
		0       Selektiert die leere Menge. Implizit am Anfang jeder neuen Selektionssequenz.
		B       Selektiert alle Bücher.
		C       Selektiert alle Copies, d.h. Buchexemplare.
		U       Selektiert alle Ausleiher (User).
		L       Selektiert alle Labels.
		/isbn/  Selektiert etwas mit dieser ISBN (Bücher, Copies, User)
		/id/    Selektiert etwas mit dieser ID
		/name/  Selektiert etwas mit diesem Namen
		/@tag/  Selektiert etwas mit diesem Tag

Im Selektionsargument (/.../) lassen sich mehrere Kriterien durch ein Komma
kombinieren; ein abschließendes Komma ist erlaubt.

		C/0, 4, 5,/   - selektiert Copies der IDs 0, 4 und 5

Wenn z.B. User selektiert sind, kann man alle filtern, die ein Buch mit einer
bestimmten ISBN haben. Wenn z.B. alle Copies selektiert sind, kann man alle
filtern, die zum selben User gehören, etc.

#### `p` - print

*Synopsis*

		p

*Beschreibung*

Gibt alle Informationen zu jedem Eintrag in *Dot* aus. S-Expressions werden als
Format verwendet. Die folgenden Beispiele dienen als Definitionen:


Copies:

		(copy 594
			(user "Dominik Okwieka")
		 	(book "978-0-201-07981-4"
				(authors "Alfred V. Aho" "Brian W. Kernighan" "Peter J. Weinberger")
				(title "The AWK Programming Language")
			)
			(notes "2016-03-24T11:01+01:00 <- ISO 8601-Date" "...")
		)
Bücher:

		(book "978-0-201-07981-4"
			(authors "Alfred V. Aho" "Brian W. Kernighan" "Peter J. Weinberger")
			(title "The AWK Programming Language")
			(notes "2016-04-26T18:16+02:00 excellent read")
			(copies 594 405 406)
		)
User:

		(user "Dominik Okwieka"
			(notes "2016-04-10T22:26+01:00 dag gummit")
			(copies 594)
		)

Wenn eine Copy nicht an einen User verliehen ist, wird der leere String
verwendet:


		(user "")

#### `r` - return

*Synopsis*

		r

*Beschreibung*

Gibt alle Copies der Selektion zurück.

#### `l` - lend

*Synopsis*

		l user


*Beschreibung*

Leiht alle Bücher der Selektion an den *(L-)*User. Der Username erstreckt sich
bis zum Ende der Zeile und darf nicht gequotet sein. Bei einem Fehler wird ein
String der Form

		can't lend [id]: <error string>

zurückgegeben. Die ID wird insbesondere bei einem internen Fehler nicht
angegeben.

#### `n` - note

*Synopsis*

		n note

*Beschreibung*

Fügt eine Notiz zu allen Objekten der Selektion hinzu. Die Notiz erstreckt sich
bis zum Zeilenende; Anführungszeichen sind nicht erlaubt. Der Zeitpunkt wird im
ISO 8601-Format mitprotokolliert. Die Notizen eines Objekts werden bei einem
`p`-Befehl mitausgegeben.

#### '@' - add or remove label

*Synopsis*

		@ label +|-

*Beschreibung*

Fügt ein Label zu allen Elementen der Selektion hinzu bzw. entfernt es.

#### `R` - retire

*Synopsis*

		R

*Beschreibung*

Zieht alle Copies der Selektion aus dem Verkehr.

#### `d` - delete

*Synopsis*

		d

*Beschreibung*

Löscht Selektion. Bücher mit existierenden Copies können nicht gelöscht
werden, User mit ausgeliehenen Copies auch nicht.

#### `A` - add book

*Synopsis*

		A isbn

*Beschreibung*

Erzeugt ein neues Buch, das diese ISBN hat.

#### `a` - add copy of a book

*Synopsis*

		a book n

*Beschreibung*

Erzeugt `n` Exemplare dieses Buchs.

#### `u` - add user

*Synopsis*

		u name

*Beschreibung*

Erzeugt User. Der Name erstreckt sich bis zum Ende der Zeile.
Er darf nicht gequotet sein.

#### `q` - quit

*Synopsis*

		q [reason]

*Beschreibung*

Schließt die Verbindung. Man kann einen Grund übermitteln.

####  `v` - print version

*Synopsis*

		v pv

*Beschreibung*

Gibt die Protokollversion in der Form

		libmangler proto P

wobei P die Protokollversion ist, zurück. Man muss die eigene Protokollversion
auch übermitteln.


### 4.3 Beispiele

Selektiere alle Copies des Users *Hans*, printe Infos, und gib alle zurück.

		C/Hans/p
		r

Selektiere alle Copies zu diesem Buch und retire sie.

		C/978-0-205-30902-3/R


5. Datenstrukturen
------------------

Die Exemplare eines Buches werden auch als *Copies* bezeichnet, besonders im
Code. (Ursprünglich nannte ich die Exemplare *Bücher* und die "Klasse"
*Buchtypen*, aber das ist eine unglückliche Benennung.)

Der Server hat vollkommene Freiheit, wie diese Struktur im Speicher und auf der
Disk zu repräsentieren ist. Jede Copy hat eine eindeutige ID, jedes Buch wird
durch die ISBN eindeutig identifiziert. Die App sollte nichts groß speichern,
nur evtl. cachen.

Die Ausleiher eines Buches werden durch einen String identifiziert, dessen Form
frei wählbar ist; der Verwalter der Bibliothek tut gut daran, ein einheitliches
Schema zu verwenden.

Gespeichert werden die Daten in drei Dateien: `users`, `books` und `copies`.
Das Format gleicht jeweils dem Output von `Up`, `Bp` und `Cp`, inklusive der
Endmarkierung (`---`) am Dateiende.

6. Referenzen
-------------

>>> **XXX Wie in HTML sichtbar machen?**

[smtp]: https://tools.ietf.org/html/rfc5321 "RFC 5321: Simple Mail Transfer Protocol"
[sam]: http://doc.cat-v.org/plan_9/4th_edition/papers/sam/ "Rob Pike: The Text Editor sam"
[p9p]: https://swtch.com/plan9port/ "Plan 9 from User Space"
