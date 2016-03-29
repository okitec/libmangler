package de.csgin.libmangler;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.PrintWriter;
import java.net.Socket;
import java.net.UnknownHostException;

import android.app.IntentService;
import android.content.Intent;

public class Connection {
	private static final int PORT = 40000;
	private static final int VERS = 2;

	/* protocol error strings */
	private static final String LENDERR = "can't lend";

	private Socket socket;
	private BufferedReader in;
	private PrintWriter out;
	private String result;

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
		int pos, pvers;
		String s;

		/* format: libmangler proto P build B */
		s = transact("v " + VERS);
		pos = s.indexOf("proto");
		pos += "proto ".length();
		pvers = Integer.parseInt(s.substring(pos));
		return pvers;
	}

	/* transact: send request, return answer; *exit* on error */
	private String transact(String req) {
	 	TransactTask tt = new TransactTask();
		tt.execute(req);
		
		return result;
	}

	/* mksel: generate selection string for a list of IDs */
	private String mksel(long... id) {
		String sel;

		sel = "";
		for(int i = 0; i < id.length; i++)
			sel += "id, ";    /* trailing comma is fine */

		return sel;
	}

	private class TransactService extends IntentService {
		public TransactService() {
		}

		protected void onHandleIntent(Intent i) {
			try {
				String req = i.getStringExtra("req");
				out.println(req);
				i.putExtra("ans", in.readLine());
			} catch (IOException ioe) {
				System.exit(1);
			}
		}
	}
}
