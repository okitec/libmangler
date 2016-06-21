package de.csgin.libmangler;

import android.app.Activity;
import android.content.Intent;
import android.os.Bundle;
import android.view.Menu;
import android.view.MenuItem;
import android.view.View;
import android.view.View.OnClickListener;
import android.widget.Button;
import android.widget.TextView;
import android.widget.Toast;
import android.widget.ViewFlipper;

public class MainActivity extends Activity {
	private static final int SCANREQ = 0;
	private static final String SRVADDR = "okwieka-9pi2";
	private Connection conn;
	private long id = -1;            /* id of last scanned copy */

	@Override
	protected void onCreate(Bundle savedInstanceState) {
		super.onCreate(savedInstanceState);
		setContentView(R.layout.activity_main);

		initbuttons();
	/*
		try {
			conn = new Connection(SRVADDR);
		} catch(UnknownHostException uhe) {
			Toast.makeText(this, "Server nicht gefunden - Netzwerkfehler?", Toast.LENGTH_LONG).show();
			Log.e("srv", "can't locate server at " + SRVADDR);
			System.exit(1);
		} catch(IOException ioe) {
			Toast.makeText(this, "Verbindungsfehler", Toast.LENGTH_LONG).show();
			Log.e("srv", "can't open socket to server " + SRVADDR);
			System.exit(1);
		}
	*/
	}

	private void initbuttons() {
		Button Bscan = (Button) findViewById(R.id.Bscan);
		Bscan.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				try {
					Intent i = new Intent("com.google.zxing.client.android.SCAN");
					i.putExtra("SCAN_MODE", "QR_CODE_MODE");
					i.putExtra("SAVE_HISTORY", false);
					startActivityForResult(i, SCANREQ);
				} catch(Exception e) {
					// XXX handle specific exception
					// XXX localisations
					Toast.makeText(getBaseContext(), "Please install the ZXing Barcode scanner app.", Toast.LENGTH_LONG).show();
					finish();
				}
			}
		});

		Button Btomain = (Button) findViewById(R.id.Btomain);
		Btomain.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				ViewFlipper vf = (ViewFlipper) findViewById(R.id.flipper);
				vf.showNext();
			}
		});
		
		Button Blend = (Button) findViewById(R.id.Blend);
		// XXX add listener when networking code works
		Button Bretire = (Button) findViewById(R.id.Bretire);
		// XXX add listener when networking code works
		Button Bnote = (Button) findViewById(R.id.Bnote);
		// XXX add listener when networking code works
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
		if (req == SCANREQ) {
			if (ans == RESULT_OK) {
				String s = data.getStringExtra("SCAN_RESULT");
				try {
					id = Long.parseLong(s);
					dispinfo(id);
				} catch(NumberFormatException nfe) {
					Toast.makeText(getBaseContext(), "QR code is not a valid copy ID", Toast.LENGTH_LONG).show();
				}
			}
			/* don't do anything on failure */
		}
	}

	/* dispinfo: go into detailed info layout for a copy of that id */
	private void dispinfo(long id) {
		TextView Tinfo = (TextView) findViewById(R.id.Tinfo);
		Tinfo.setText("Copy ID: " + id);

		ViewFlipper vf = (ViewFlipper) findViewById(R.id.flipper);
		vf.showNext();
	}
}
