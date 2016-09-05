package de.csgin.libmangler;

import android.app.Activity;
import android.content.Context;
import android.util.Log;
import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.io.IOException;
import java.io.PrintWriter;
import java.net.Socket;
import java.net.UnknownHostException;
import java.util.concurrent.Callable;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.Executor;
import java.util.concurrent.Executors;
import java.util.concurrent.FutureTask;
import java.util.HashMap;
import java.util.Random;

/**
 * A request to the server. You need to provide a RespHandler. The RespHandler is
 * run on the UI thread when the request is answered.
 *
 *     new Req("B/978-0-201-07981-4/p", new RespHandler() {
 *         public void onResponse(Req r, String resp) {
 *             new Toast.makeText(this, "derp: " + resp, Toast.LENGTH_LONG).show();
 *         }
 *     }).send(out);
 */
public class Req {
	private static HashMap<Integer, Req> reqs;
	private static Random r;
	private static Context ctxt;
	private static Executor networker;

	private static String addr;
	private static int port;
	private static Socket socket;
	private static BufferedReader br;
	private static PrintWriter pw;

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

	public void send() {
		reqs.put(new Integer(tag), this);
		networker.execute(new FutureTask<Void>(new Callable<Void>() {
			public Void call() {
				pw.println(tag + " " + s);
				Log.e("derp", "just sent a Req");
				return null;
			}
		}));
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
	
			Log.e("derp", "just fetched a response");

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

	private static void reconnect() {
		try {
			socket = new Socket(addr, port);
			br = new BufferedReader(new InputStreamReader(socket.getInputStream()));
			pw = new PrintWriter(socket.getOutputStream());
			Log.e("derp", "connected");
		} catch(UnknownHostException uhe) {
			// XXX how to call panic here?
			Log.e("derp", "unknown host");
		} catch(IOException ioe) {
			// XXX and here?
			Log.e("derp", "io exception");
		}
	}

	public static void init(Context context, final String addr, final int port) {
		reqs = new HashMap<Integer, Req>();
		r = new Random();
		ctxt = context;
		Req.addr = addr;
		Req.port = port;

		networker = Executors.newFixedThreadPool(1);

		FutureTask<Void> netinit = new FutureTask<Void>(new Callable<Void>() {
			public Void call() {
				reconnect();
				return null;
			}
		});
		networker.execute(netinit);
		try {
			netinit.get(); // block until network is ready
		} catch(InterruptedException ie) {
		} catch(ExecutionException ee) {
		}
		Log.e("derp", "network should be inited now");

		FutureTask<Void> recv = new FutureTask<Void>(new Callable<Void>() {
			public Void call() {
				try {
					recvloop(br);
				} catch(IOException ioe) {
				}

				return null;
			}
		});
		networker.execute(recv);

		Log.e("derp", "all inited");
	}
}
