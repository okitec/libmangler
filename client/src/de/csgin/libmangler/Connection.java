package de.csgin.libmangler;

import android.util.Log;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.PrintWriter;
import java.net.Socket;
import java.net.UnknownHostException;

/**
 * Connction provides an interface to the server's RPC library, i.e. provides
 * functions wrapping the protocol commands.
 */
public class Connection {
	private static final int PORT         = 40000;
	private static final int VERS         = 7;
	private static final String ENDMARKER = "---";

	/* protocol error strings */
	private static final String LENDERR = "can't lend";

	private Socket socket;
	private BufferedReader in;
	private PrintWriter out;

	/* just rethrow, we can't tell the user */
	public Connection(String addr) throws UnknownHostException, IOException {
		socket = new Socket(addr, PORT);
		in = new BufferedReader(new InputStreamReader(socket.getInputStream()));
		out = new PrintWriter(socket.getOutputStream());
	}

	public String print(long... id) {
		return transact("C/" + mksel(id) + "/p");
	}

	public void delete(long... id) {
		transact("C/" + mksel(id) + "/d");
	}

	public void note(String note, long... id) {
		transact("C/" + mksel(id) + "/n " + note);
	}

	public void addTag(String tag, long... id) {
		transact("C/" + mksel(id) + "/t + " + tag);
	}

	public void rmTag(String tag, long... id) {
		transact("C/" + mksel(id) + "/t - " + tag);
	}

	/* lend: lend copy to user; return an error string if any lend failed, null otherwise */
	public String lend(String user, long... id) {
		String s;

		s = transact("C/" + mksel(id) + "/l " + user);
		if(s.contains(LENDERR))
			return s;

		return null;
	}

	public void returnCopy(long... id) {
		transact("C/" + mksel(id) + "/r");
	}

	public void retire(long... id) {
		transact("C/" + mksel(id) + "/R");
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

		String args[] = s.split(" ");
		return Integer.parseInt(args[2]);
	}

	/* transact: send request line, return multi-line answer */
	private String transact(String req) {
		out.println(req);
		out.flush();
		Log.e("libmangler-proto", "[proto->] " + req);

		try {
			StringBuilder answer = new StringBuilder();
			String line;

			// note the classic Java naming inconsistency (readLine vs println)
			while(!(line = in.readLine()).equals(ENDMARKER)) {
				Log.e("libmangler-proto", "[->proto] " + answer);
				answer.append(line);
			}

			return answer.toString();
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
}
