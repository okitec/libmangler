package de.csgin.libmangler;

import android.app.Activity;
import android.content.Context;
import java.io.BufferedReader;
import java.io.IOException;
import java.io.PrintWriter;
import java.util.HashMap;
import java.util.Random;

/**
 * A request to the server. You need to provide a RespHandler. The RespHandler is
 * run on the UI thread when the request is answered.
 *
 *     new Req("B/978-0-201-07981-4/p", new RespHandler() {
 *         public void onResponse(Req r, String resp) {
 *             Log.i("derp", resp);
 *         }
 *     });
 */
public class Req {
	private static HashMap<Integer, Req> reqs;
	private static Random r;
	private static Context ctxt;

	public final int tag;
	public final String s;
	public final RespHandler rh;

	public Req(String s, RespHandler rh) {
		int i;

		this.s = s;
		this.rh  = rh;

		/* generate random, unused tag */
		for(;;) {
			i = r.nextInt();

			if(reqs.get(new Integer(i)) == null) {
				this.tag = i;
				break;
			}	
		}
	}

	public void send(PrintWriter pw) {
		reqs.put(new Integer(tag), this);
		pw.print(s + "\n");	
	}

	private static void recvloop(BufferedReader br) throws IOException {
		for(;;) {
			final StringBuilder payload;
			String line;
			String hdr[];
			final Req r;
			int nlines;
			int tag;

			line = br.readLine();
			if(line == null)
				return;
	
			hdr = line.split("( |\t)+");
			if(hdr.length < 2)
				return;
	
			try {
				tag = Integer.parseInt(hdr[0]);
				nlines = Integer.parseInt(hdr[1]);
			} catch(NumberFormatException nfe) {
				return;
			}
	
			r = reqs.get(new Integer(tag));
			if(r == null)
				return;

			payload = new StringBuilder();
			for(int i = 0; i < nlines; i++) {
				if((line = br.readLine()) == null)
					return;
	
				payload.append(line);
			}

			((Activity) ctxt).runOnUiThread(new Runnable() {
				public void run() {
					r.rh.onResponse(r, payload.toString());
				}
			});
		}
	}

	public static void init(Context context, final BufferedReader br) {
		reqs = new HashMap<Integer, Req>();
		r = new Random();
		ctxt = context;

		new Thread(new Runnable() {
			public void run() {
				try {
					recvloop(br);
				} catch(IOException ioe) {
					// XXX ?
				}
			}
		});
	}
}
