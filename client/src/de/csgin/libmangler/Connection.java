package de.csgin.libmangler;

import android.util.Log;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.PrintWriter;
import java.net.Socket;
import java.net.UnknownHostException;

public class Connection {
	private static final int PORT = 40000;
	private static final int VERS = 4;

	/* protocol error strings */
	private static final String LENDERR = "can't lend";

	private Socket socket;
	private BufferedReader in;
	private PrintWriter out;

	/* just rethrow, we can't tell the user */
	public Connection(String addr) throws UnknownHostException, IOException {
		int vers;

		socket = new Socket(addr, PORT);
		in = new BufferedReader(new InputStreamReader(socket.getInputStream()));
		out = new PrintWriter(socket.getOutputStream());

		vers = version();
		if(vers != VERS) {
			quit("version mismatch (client " + VERS + "; server " + vers + ")");
			System.exit(1);
		}
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

	/* lend: lend copy to user; return false if any lend failed, true otherwise */
	public boolean lend(String user, long... id) {
		String s;

		s = transact("C/" + mksel(id) + "/l " + user);
		if(s.contains(LENDERR))
			return false;

		return true;
	}

	public void returncopy(long... id) {
		transact("C/" + mksel(id) + "/r");
	}

	public void retire(long... id) {
		transact("C/" + mksel(id) + "/R");
	}

	public void quit(String reason) {
		transact("q " + reason);
	}

	/* version: get protocol version number */
	private int version() {
		String s;

		/* format: libmangler proto P */
		s = transact("v " + VERS);

		String args[] = s.split(" ");
		return Integer.parseInt(args[2]);
	}

	/* transact: send request, return answer */
	private String transact(String req) {
		out.println(req);
		out.flush();
		Log.e("libmangler-proto", "[proto->] " + req);

		try {
			String line = in.readLine();  // note the classic Java naming inconsistency
			Log.e("libmangler-proto", "[->proto] " + line);
			return line;
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
