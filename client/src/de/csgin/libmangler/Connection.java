package de.csgin.libmangler;

import android.content.Context;
import android.util.Log;
import android.widget.Toast;

/**
 * Whereas the actual connection is managed in the Req class, the protocol lives
 * in Connection. When one of these routines is called, the proper request will be
 * assembled and sent to the server.
 *
 * XXX let caller decide on RespHandler (removes Context dependency)
 * XXX naming: not really a Connection any longer, more of a protocol
 * XXX could be totally static
 */
public class Connection {
	private static final int VERS = 2;

	/* protocol error strings */
	private static final String LENDERR = "can't lend";

	private Context ctxt;
	private RespHandler nilHandler;

	public Connection(Context ctxt) {
		int vers;

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
		}).send();
	}

	public void delete(long... id) {
		new Req("C/" + mksel(id) + "/d", nilHandler).send();
	}

	public void note(String note, long... id) {
		new Req("C/" + mksel(id) + "/n " + note, nilHandler).send();
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
		}).send();
	}

	public void returncopy(long... id) {
		new Req("C/" + mksel(id) + "/r", nilHandler).send();
	}

	public void retire(long... id) {
		new Req("C/" + mksel(id) + "/R", nilHandler).send();
	}

	public void quit(String reason) {
		new Req("q " + reason, nilHandler).send();
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
		}).send();
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
