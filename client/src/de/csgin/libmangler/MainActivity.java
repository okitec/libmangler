package de.csgin.libmangler;

import android.app.Activity;
import android.app.AlertDialog;
import android.content.Intent;
import android.os.Bundle;
import android.os.StrictMode;
import android.os.StrictMode.ThreadPolicy;
import android.util.Log;
import android.view.Menu;
import android.view.MenuItem;
import android.view.View;
import android.view.View.OnClickListener;
import android.widget.AdapterView;
import android.widget.ArrayAdapter;
import android.widget.Button;
import android.widget.EditText;
import android.widget.ListAdapter;
import android.widget.ListView;
import android.widget.TextView;
import android.widget.Toast;
import android.widget.ViewFlipper;

import java.io.IOException;
import java.net.UnknownHostException;
import java.util.ArrayList;

public class MainActivity extends Activity {
	/* ViewFlipper indexes */
	private static final int MainLayout = 0;
	private static final int CopyInfoLayout = 1;
	private static final int SearchLayout = 2;
	private static final int PanicLayout = 3;
	private static final int ElemsLayout = 4;

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
			panic("Server '" + SRVADDR + "' nicht gefunden");
		} catch(IOException ioe) {
			panic("Kann keine Verbindung zu '" + SRVADDR + "' aufbauen");
		} catch (android.os.NetworkOnMainThreadException netmain) {
			panic("NetworkOnMainThreadException - sollte nicht passieren");
		}

		if(conn == null)
			panic("Kann keine Verbindung zu '" + SRVADDR + "' aufbauen");
		if(conn != null && !conn.isProperVersion())
			panic("Inkompatibles Protokoll zwischen Client und Server!");

		// XXX remove this test code again
		ListView Lelems = (ListView) findViewById(R.id.Lelems);
		ArrayAdapter aa = (ArrayAdapter) Lelems.getAdapter();
		aa.add("hello");
		aa.add("world");
		Lelems.setOnItemClickListener(new AdapterView.OnItemClickListener() {
			@Override
			public void onItemClick(AdapterView<?> parent, View view, int pos, long id) {
				ListView lv = (ListView) parent;

				ListAdapter la = lv.getAdapter();
				toast((String) la.getItem(pos));
			}
		});
		flipView(ElemsLayout);
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
					copyinfo(id);
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
		initCopyInfoLayout();
		initSearchLayout();
		// PanicLayout needs no initialisation, as it has no buttons.
		initElemsLayout();
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
					panic("ZXing Barcode scanner app not installed. Please install it to scan QR codes.");
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

	private void initCopyInfoLayout() {
		Button Btomain = (Button) findViewById(R.id.Btomain);
		Button Blend = (Button) findViewById(R.id.Blend);
		Button Breturn = (Button) findViewById(R.id.Breturn);
		Button Bretire = (Button) findViewById(R.id.Bretire);
		Button Bnote = (Button) findViewById(R.id.Bnote);
		Button Baddtag = (Button) findViewById(R.id.Baddtag);
		Button Brmtag = (Button) findViewById(R.id.Brmtag);

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

				new StringDialog(MainActivity.this, "Zu verleihen an ...", "An wen soll das Exemplar verliehen werden?", "",
					new StringDialog.ResultTaker() {
						@Override
						public void take(String res) {
							String err = conn.lend(res, id);
							if(err != null) {
								// XXX issue #35: freeze on lend error
								notice("Fehler beim Verleih", err);
							} else {
								toast("Erfolgreich verliehen");
								copyinfo(id);
							}
						}
					});
			}
		});

		Breturn.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				if(id == -1)
					return;

				conn.returnCopy(id);
				toast("Zurückgegeben");
				copyinfo(id);
			}
		});

		Bretire.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				if(id == -1)
					return;

				conn.retire(id);
				toast("Beiseitegestellt");
				copyinfo(id);
			}
		});

		Bnote.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				if(id == -1)
					return;

				new StringDialog(MainActivity.this, "Notiz hinzufügen", "", "",
					new StringDialog.ResultTaker() {
						@Override
						public void take(String res) {
							conn.note(res, id);
							toast("Notiz hinzugefügt");
							copyinfo(id);
						}
					});
			}
		});

		Baddtag.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				if(id == -1)
					return;

				new StringDialog(MainActivity.this, "Tag hinzufügen", "", "",
					new StringDialog.ResultTaker() {
						@Override
						public void take(String res) {
							conn.addTag(res, id);
							toast("Tag hinzugefügt");
							copyinfo(id);
						}
					});
			}
		});

		Brmtag.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				if(id == -1)
					return;

				new StringDialog(MainActivity.this, "Tag entfernen", "", "",
					new StringDialog.ResultTaker() {
						@Override
						public void take(String res) {
							conn.rmTag(res, id);
							toast("Tag entfernt");
							copyinfo(id);
						}
					});
			}
		});
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
				copyinfo(id);
			}
		});
	}

	private void initElemsLayout() {
		ListView Lelems = (ListView) findViewById(R.id.Lelems);
		Button Btomain3 = (Button) findViewById(R.id.Btomain3);

		ArrayList<String> ls = new ArrayList<String>();
		ArrayAdapter<String> aa = new ArrayAdapter<String>(this, android.R.layout.simple_list_item_1, ls);
		Lelems.setAdapter(aa);

		Btomain3.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				flipView(MainLayout);
			}
		});
	}

	/**
	 * Fetch info about a Copy and show it in the CopyInfoLayout.
	 */
	private void copyinfo(long id) {
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
	 * Make a long toast.
	 */
	private void toast(String s) {
		Toast.makeText(MainActivity.this, s, Toast.LENGTH_LONG).show();
	}

	/**
	 * Go into a layout which displays the error and offers no way back.
	 * Bug: Does return; can't reasonably loop here.
	 */
	private void panic(String s) {
		Log.e("libmangler", "panic: " + s);
		TextView Tpanic = (TextView) findViewById(R.id.Tpanic);
		Tpanic.setText("Fataler Fehler: " + s);
		flipView(PanicLayout);
	}

	/**
	 * Show a notice dialog.
	 */
	private void notice(String title, String msg) {
		AlertDialog.Builder b = new AlertDialog.Builder(MainActivity.this);
		b.setTitle(title).setMessage(msg);
		b.show();
	}
}
