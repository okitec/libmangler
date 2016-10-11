package de.csgin.libmangler;

import android.app.Activity;
import android.content.Intent;
import android.os.Bundle;
import android.os.StrictMode;
import android.os.StrictMode.ThreadPolicy;
import android.util.Log;
import android.view.Menu;
import android.view.MenuItem;
import android.view.View;
import android.view.View.OnClickListener;
import android.widget.Button;
import android.widget.EditText;
import android.widget.TextView;
import android.widget.Toast;
import android.widget.ViewFlipper;

import java.io.IOException;
import java.net.UnknownHostException;

public class MainActivity extends Activity {
	/* ViewFlipper indexes */
	private static final int MainLayout = 0;
	private static final int InfoLayout = 1;
	private static final int SearchLayout = 2;

	/** Request code for QR code scanning intent */
	private static final int SCANREQ = 0;

	private static final String SRVADDR = "oquasinus.duckdns.org";
	private static final int PORT = 40000;
	private Connection conn;

	/** currently investigated copy */
	private long id = -1;

	@Override
	protected void onCreate(Bundle savedInstanceState) {
		super.onCreate(savedInstanceState);
		setContentView(R.layout.activity_main);

		// Override NetworkOnMainThreadException because I think in synchronous terms.
		// This is a last-ditch measure, but I want to have a reasonably working app.
		// src: http://stackoverflow.com/questions/6343166/how-to-fix-android-os-networkonmainthreadexception#6343299
		StrictMode.ThreadPolicy p = new StrictMode.ThreadPolicy.Builder().permitAll().build();
		StrictMode.setThreadPolicy(p); 

		initLayouts();

		try {
			conn = new Connection(SRVADDR);
		} catch(UnknownHostException uhe) {
			toast("Server nicht gefunden - Netzwerkfehler?");
			Log.e("srv", "can't locate server at " + SRVADDR);
		} catch(IOException ioe) {
			toast("Verbindungsfehler");
			Log.e("srv", "can't open socket to server " + SRVADDR);
		} catch (android.os.NetworkOnMainThreadException netmain) {
			toast("NetworkOnMainThreadException - sollte nicht passieren");
			Log.e("srv", "shouldn't happen - NetworkOnMainThreadException");
		}
	}

	@Override
	public boolean onCreateOptionsMenu(Menu menu) {
		// Inflate the menu; this adds items to the action bar if it is present.
		getMenuInflater().inflate(R.menu.main, menu);
		return true;
	}

	@Override
	public boolean onOptionsItemSelected(MenuItem item) {
		// Handle action bar item clicks here. The action bar will
		// automatically handle clicks on the Home/Up button, so long
		// as you specify a parent activity in AndroidManifest.xml.
		int id = item.getItemId();
		if (id == R.id.action_settings) {
			return true;
		}
		return super.onOptionsItemSelected(item);
	}

	@Override
	protected void onSaveInstanceState(Bundle out) {
		super.onSaveInstanceState(out);
		out.putLong("id", id);
	}

	@Override
	protected void onRestoreInstanceState(Bundle in) {
		super.onRestoreInstanceState(in);
		id = in.getLong("id");
	}

	@Override
	public void onActivityResult(int req, int ans, Intent data) {
		super.onActivityResult(req, ans, data);
		// cf. http://stackoverflow.com/questions/8831050/android-how-to-read-qr-code-in-my-application
		if (req == SCANREQ) {
			if (ans == RESULT_OK) {
				String s = data.getStringExtra("SCAN_RESULT");
				try {
					id = Long.parseLong(s);
					dispinfo(id);
				} catch(NumberFormatException nfe) {
					toast("QR code is not a valid copy ID");
				}
			}
			/* don't do anything on failure */
		}
	}

	/**
	 * Initialise the buttons of the layouts.
	 */
	private void initLayouts() {
		initMainLayout();
		initInfoLayout();
		initSearchLayout();
	}

	private void initMainLayout() {
		Button Bscan;
		Button Bsearch;

		Bscan = (Button) findViewById(R.id.Bscan);
		// I wish lambda expressions were available in Android.
		Bscan.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				// cf. http://stackoverflow.com/questions/8831050/android-how-to-read-qr-code-in-my-application
				try {
					Intent i = new Intent("com.google.zxing.client.android.SCAN");
					i.putExtra("SCAN_MODE", "QR_CODE_MODE");
					i.putExtra("SAVE_HISTORY", false);
					startActivityForResult(i, SCANREQ);
				} catch(Exception e) {
					// XXX handle specific exception
					// XXX localisations
					toast("Please install the ZXing Barcode scanner app.");
					finish();
				}
			}
		});

		Bsearch = (Button) findViewById(R.id.Bsearch);
		Bsearch.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				flipView(SearchLayout);
			}
		});
	}

	private void initInfoLayout() {
		Button Btomain = (Button) findViewById(R.id.Btomain);
		Button Blend = (Button) findViewById(R.id.Blend);
		Button Bretire = (Button) findViewById(R.id.Bretire);
		Button Bnote = (Button) findViewById(R.id.Bnote);

		Btomain.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				flipView(MainLayout);
			}
		});

		Blend.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				if(id == -1)
					return;

				// XXX Query for lendee name
			}
		});

		//Button Bretire = (Button) findViewById(R.id.Bretire);
		//Button Bnote = (Button) findViewById(R.id.Bnote);
	}

	private void initSearchLayout() {
		Button Btomain2 = (Button) findViewById(R.id.Btomain2);
		Button Bdosearch = (Button) findViewById(R.id.Bdosearch);

		Btomain2.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				flipView(MainLayout);
			}
		});

		Bdosearch.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				EditText Esearch = (EditText) findViewById(R.id.Esearch);
				long id;

				// XXX search for more than just id!
				id = Long.parseLong(Esearch.getText().toString());
				dispinfo(id);
			}
		});
	}

	/**
	 * Fetch info about a Copy and show it in the InfoLayout.
	 */
	private void dispinfo(long id) {
		if(id == -1)
			return;

		TextView Tinfo = (TextView) findViewById(R.id.Tinfo);
		Tinfo.setText("The server says: " + conn.print(id));
		flipView(InfoLayout);
	}

	/**
	 * Flip to a linear layout. See indexes at the top of MainActivity.
	 */
	private void flipView(int layout) {
		ViewFlipper vf = (ViewFlipper) findViewById(R.id.flipper);
		vf.setDisplayedChild(layout);
	}

	/**
	 * Makea long toast.
	 */
	private void toast(String s) {
		Toast.makeText(getBaseContext(), s, Toast.LENGTH_LONG).show();
	}
}
