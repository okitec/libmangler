package de.csgin.libmangler;

import android.content.Context;
import android.content.res.AssetManager;
import android.util.Log;

import java.io.BufferedReader;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.io.IOException;
import java.io.PrintWriter;
import java.net.InetSocketAddress;
import java.net.Socket;
import java.net.SocketTimeoutException;
import java.net.UnknownHostException;
import java.security.KeyManagementException;
import java.security.KeyStore;
import java.security.KeyStoreException;
import java.security.NoSuchAlgorithmException;
import java.security.cert.CertificateException;
import javax.net.ssl.SSLContext;
import javax.net.ssl.SSLSocket;
import javax.net.ssl.SSLSocketFactory;
import javax.net.ssl.TrustManager;
import javax.net.ssl.TrustManagerFactory;

/**
 * Connection provides an interface to the server's RPC library, i.e. provides
 * functions wrapping the protocol commands. The socket has a timeout to prevent
 * freezing; thus the synchronous nature of IO in the main thread is not as bad
 * as it would be.
 *
 * There's a wild assortment of small functions. Each of these is needed somewhere;
 * unneeded functions have not been implemented.
 */
public class Connection {
	private static final int VERS         = 11;
	private static final String ENDMARKER = "---";
	private static final int TIMEOUT      = 3000;  /* in ms */

	private Socket socket;
	private BufferedReader in;
	private PrintWriter out;

	public Connection(Context ctxt, String addr, int port) throws UnknownHostException, IOException, SocketTimeoutException,
		KeyStoreException, NoSuchAlgorithmException, KeyManagementException, CertificateException
	{
		socket = new Socket();
		socket.setSoTimeout(TIMEOUT);
		socket.connect(new InetSocketAddress(addr, port));

		KeyStore ks = KeyStore.getInstance("BKS");
		InputStream fis = null;
		try {
			fis = ctxt.getAssets().open("test.bks");
			ks.load(fis, "derp".toCharArray());       // XXX use a real password
		} finally {
			if(fis != null)
				fis.close();
		}

		// cf. https://developer.android.com/training/articles/security-ssl.html
		String algo = TrustManagerFactory.getDefaultAlgorithm();
		TrustManagerFactory tmf = TrustManagerFactory.getInstance(algo);
		tmf.init(ks);

		SSLContext context = SSLContext.getInstance("TLS");
		context.init(null, tmf.getTrustManagers(), null);

		SSLSocketFactory ssf = context.getSocketFactory();
		socket = ssf.createSocket(socket, addr, port, true);

		in = new BufferedReader(new InputStreamReader(socket.getInputStream()));
		out = new PrintWriter(socket.getOutputStream());
	}

	public String printCopy(long... id) {
		return transact("C/" + mksel(id) + "/p");
	}

	public String printBook(String... isbn) {
		return transact("B/" + mksel(isbn) + "/p");
	}

	public String printBookOfID(long id) {
		return transact("B/" + id + "/p");
	}

	public String printUser(String... name) {
		return transact("U/" + mksel(name) + "/p");
	}

	public String deleteCopy(long... id) {
		return transact("C/" + mksel(id) + "/d");
	}

	public String deleteBook(String... isbn) {
		return transact("B/" + mksel(isbn) + "/d");
	}

	public String deleteUser(String... name) {
		return transact("U/" + mksel(name) + "/d");
	}

	public void noteCopy(String note, long... id) {
		transact("C/" + mksel(id) + "/n " + note);
	}

	public void noteBook(String note, String isbn) {
		transact("B/" + isbn + "/n " + note);
	}

	public void noteUser(String note, String name) {
		transact("U/" + name + "/n " + note);
	}

	public void addTagCopy(String tag, long... id) {
		transact("C/" + mksel(id) + "/t + " + tag);
	}

	public void rmTagCopy(String tag, long... id) {
		transact("C/" + mksel(id) + "/t - " + tag);
	}

	public void addTagBook(String tag, String isbn) {
		transact("B/" + isbn + "/t + " + tag);
	}

	public void rmTagBook(String tag, String isbn) {
		transact("B/" + isbn + "/t - " + tag);
	}

	public void addTagUser(String tag, String name) {
		transact("U/" + name + "/t + " + tag);
	}

	public void rmTagUser(String tag, String name) {
		transact("U/" + name + "/t - " + tag);
	}

	/* lend: lend copy to user; return an error string if any lend failed, null otherwise */
	public String lend(String user, long... id) {
		return transact("C/" + mksel(id) + "/l " + user);
	}

	public void returnCopy(long... id) {
		transact("C/" + mksel(id) + "/r");
	}

	public void returnAll(String user) {
		transact("C/" + user + "/r");
	}

	public void retire(long... id) {
		transact("C/" + mksel(id) + "/R");
	}

	public String addCopies(String isbn, int n) {
		return transact("c " + isbn + " " + n);
	}

	public String addBook(String isbn, String title, String[] authors) {
		String as = "";
		for(String a: authors)
			as += "\"" + a + "\" ";
		as = as.trim();

		String s = String.format("(book %s (authors %s) (title \"%s\"))", isbn, as, title);
		return transact("b " + s);
	}

	public String addUser(String name) {
		return transact("u " + name);
	}

	public String listTags() {
		return transact("T");
	}

	/**
	 * List copies belonging to a book (if the server detects a ISBN) or a user.
	 */
	public String listCopies(String isbnOrName) {
		return transact("C/" + isbnOrName + "/λ");
	}

	public String listBooks() {
		return transact("Bλ");
	}

	public String listUsers() {
		return transact("Uλ");
	}

	public void quit(String reason) {
		transact("q " + reason);
	}

	/**
	 * isProperVersion: test whether the protocol versions of client and server match. 
	 */
	public boolean isProperVersion() {
		int vers = version();
		if(vers != VERS) {
			quit("version mismatch (client " + VERS + "; server " + vers + ")");
			return false;
		}

		return true;
	}

	/* version: get protocol version number */
	private int version() {
		String s;

		/* format: libmangler proto P */
		s = transact("v " + VERS);

		String args[] = s.split("( |\n)");
		return Integer.parseInt(args[2]);
	}

	/**
	 * Send a request line and return a multi-line answer.
	 */
	public String transact(String req) {
		out.println(req);
		out.flush();
		Log.e("libmangler-proto", "[proto->] " + req);

		try {
			StringBuilder answer = new StringBuilder();
			String line;

			// note the classic Java naming inconsistency (readLine vs println)
			while((line = in.readLine()) != null && !line.equals(ENDMARKER)) {
				Log.e("libmangler-proto", "[->proto] " + answer);
				answer.append(line);
				answer.append("\n");
			}

			if(line == null)
				return "END OF FILE\nBitte Verbindung prüfen und App erneut starten.";
			return answer.toString();
		} catch(SocketTimeoutException ste) {
			// XXX tell MainActivity to reestablish connection and show error
			Log.e("libmangler", "Socket timeout");
			return "NETWORK TIMEOUT";
		} catch(IOException ioe) {
			Log.e("libmangler-proto", "IO EXCEPTION");
			return "IO EXCEPTION";
		}
	}

	/* mksel: generate selection string for a list of IDs */
	private String mksel(long... id) {
		String sel;

		sel = "";
		for(int i = 0; i < id.length; i++)
			sel += id[i] + ", ";    /* trailing comma is fine */

		return sel;
	}

	/* mksel: generate selection string for a list of strings */
	private String mksel(String... s) {
		String sel;

		sel = "";
		for(int i = 0; i < s.length; i++)
			sel += s[i] + ", ";    /* trailing comma is fine */

		return sel;
	}
}
