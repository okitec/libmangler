package de.csgin.libmangler;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.PrintWriter;
import java.net.Socket;
import java.net.UnknownHostException;

import android.content.Context;
import android.util.Log;
import android.widget.Toast;

public class Connection {
	private static final int PORT = 40000;
	private static final int VERS = 2;

	/* protocol error strings */
	private static final String LENDERR = "can't lend";

	private Socket socket;
	private BufferedReader in;
	private PrintWriter out;
	private Context ctxt;
	private RespHandler nilHandler;

	/* just rethrow, we can't tell the user */
	public Connection(String addr, Context ctxt) throws UnknownHostException, IOException {
		int vers;

		socket = new Socket(addr, PORT);
		in = new BufferedReader(new InputStreamReader(socket.getInputStream()));
		out = new PrintWriter(socket.getOutputStream());
		this.ctxt = ctxt;
		nilHandler = new RespHandler() {
			public void onResponse(Req r, String resp) {}
		};
	}

	public void print(long... id) {
		new Req("C/" + mksel(id) + "/p", new RespHandler() {
			public void onResponse(Req r, String resp) {
				((MainActivity)ctxt).dispinfo(resp);
			}
		}).send(out);
	}

	public void delete(long... id) {
		new Req("C/" + mksel(id) + "/d", nilHandler).send(out);
	}

	public void note(String note, long... id) {
		new Req("C/" + mksel(id) + "/n " + note, nilHandler).send(out);
	}

	/* lend: lend copy to user */
	public void lend(String user, long... id) {

		new Req("C/" + mksel(id) + "/l " + user, new RespHandler() {
			public void onResponse(Req r, String resp) {
				if(resp.contains(LENDERR)) {
					Toast.makeText(ctxt, "error: " + resp, Toast.LENGTH_LONG).show();
					return;
				}

				// XXX show success?
			}
		}).send(out);
	}

	public void returncopy(long... id) {
		new Req("C/" + mksel(id) + "/r", nilHandler).send(out);
	}

	public void retire(long... id) {
		new Req("C/" + mksel(id) + "/R", nilHandler).send(out);
	}

	public void quit(String reason) {
		new Req("q " + reason, nilHandler).send(out);
	}

	/* version: see if server operates on same protocol version */
	private void version() {
		new Req("v " + VERS, new RespHandler() {
			public void onResponse(Req r, String resp) {
				String args[];
				int pv = -42;

				args = resp.split("( |\t)+");
				if(args.length < 3)
					panic("server sends bogus answer");

				try {
					pv = Integer.parseInt(args[2]);
				} catch(NumberFormatException nfe) {
					panic("server sends bogus answer");
				}

				if(pv != VERS)
					panic("protocol version mismatch");
			}
		}).send(out);
	}

	/**
	 * panic: display panic toast, log and quit
	 */
	public void panic(String s) {
		Toast.makeText(ctxt, "panic: " + s, Toast.LENGTH_LONG).show();
		Log.e("panic", s);
		System.exit(1);
	}

	/* mksel: generate selection string for a list of IDs */
	private String mksel(long... id) {
		String sel;

		sel = "";
		for(int i = 0; i < id.length; i++)
			sel += "id, ";    /* trailing comma is fine */

		return sel;
	}
}
