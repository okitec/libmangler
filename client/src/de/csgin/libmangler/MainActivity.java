package de.csgin.libmangler;

import android.app.Activity;
import android.content.Intent;
import android.os.Bundle;
import android.view.Menu;
import android.view.MenuItem;
import android.view.View;
import android.view.View.OnClickListener;
import android.widget.Button;
import android.widget.EditText;
import android.widget.TextView;
import android.widget.Toast;
import android.widget.ViewFlipper;

// XXX handle pause/resume, reestablish connection
public class MainActivity extends Activity {
	/* ViewFlipper indexes */
	private static final int MainLayout = 0;
	private static final int InfoLayout = 1;
	private static final int SearchLayout = 2;

	private static final int SCANREQ = 0;
	private static final String SRVADDR = "oquasinus.duckdns.org";
	private static final int PORT = 40000;
	private Connection conn;
	private long id = -1;            /* id of last scanned copy */

	@Override
	protected void onCreate(Bundle savedInstanceState) {
		super.onCreate(savedInstanceState);
		setContentView(R.layout.activity_main);

		initbuttons();

		Req.init(this, SRVADDR, PORT);
		conn = new Connection(this);
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

		Button Bsearch = (Button) findViewById(R.id.Bsearch);
		Bsearch.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				ViewFlipper vf = (ViewFlipper) findViewById(R.id.flipper);
				vf.setDisplayedChild(SearchLayout);
			}
		});

		OnClickListener tomain = new OnClickListener() {
			@Override
			public void onClick(View v) {
				ViewFlipper vf = (ViewFlipper) findViewById(R.id.flipper);
				vf.setDisplayedChild(MainLayout);
			}
		};
		Button Btomain = (Button) findViewById(R.id.Btomain);
		Button Btomain2 = (Button) findViewById(R.id.Btomain2);
		Btomain.setOnClickListener(tomain);
		Btomain2.setOnClickListener(tomain);

		Button Bdosearch = (Button) findViewById(R.id.Bdosearch);
		Bdosearch.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				EditText Esearch = (EditText) findViewById(R.id.Esearch);
				long id;

				// XXX search for more than just id!
				id = Long.parseLong(Esearch.getText().toString());
				conn.print(id);
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
		vf.setDisplayedChild(InfoLayout);
	}
}
