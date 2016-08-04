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

import java.io.IOException;
import java.net.UnknownHostException;

public class MainActivity extends Activity {
	private static final int SCANREQ = 0;
	private static final String SRVADDR = "oquasinus-pc";
	private Connection conn;
	private long id = -1;            /* id of last scanned copy */

	@Override
	protected void onCreate(Bundle savedInstanceState) {
		super.onCreate(savedInstanceState);
		setContentView(R.layout.activity_main);

		initbuttons();

		try {
			conn = new Connection(SRVADDR, this);
		} catch(UnknownHostException uhe) {
			conn.panic("Server nicht gefunden - Netzwerkfehler?");
		} catch(IOException ioe) {
			conn.panic("Verbindungsfehler");
		}
	}

	private void initbuttons() {
		Button Bscan = (Button) findViewById(R.id.Bscan);
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
		// cf. http://stackoverflow.com/questions/8831050/android-how-to-read-qr-code-in-my-application
		if (req == SCANREQ) {
			if (ans == RESULT_OK) {
				String s = data.getStringExtra("SCAN_RESULT");
				try {
					id = Long.parseLong(s);
					conn.print(id);
				} catch(NumberFormatException nfe) {
					Toast.makeText(getBaseContext(), "QR code is not a valid copy ID", Toast.LENGTH_LONG).show();
				}
			}
			/* don't do anything on failure */
		}
	}

	/* dispinfo: ATM, just display the string */
	public void dispinfo(String s) {
		TextView Tinfo = (TextView) findViewById(R.id.Tinfo);
		Tinfo.setText("Fetched from server: " + s);

		ViewFlipper vf = (ViewFlipper) findViewById(R.id.flipper);
		vf.showNext();
	}
}
