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
	private static final int BookInfoLayout = 5;
	private static final int UserInfoLayout = 6;
	private static final int AddBookLayout = 7;

	/** Request code for QR code scanning intent */
	private static final int SCANREQ = 0;

	private static final String SRVADDR = "oquasinus.duckdns.org";
	private static final int PORT = 40000;
	private Connection conn;

	/* currently investigated copy, book, name */
	private long id = -1;
	private String isbn;
	private String name;

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
		initBookInfoLayout();
		initUserInfoLayout();
		initAddBookLayout();
	}

	private void initMainLayout() {
		Button Bscan = (Button) findViewById(R.id.Bscan);
		Button Bsearch = (Button) findViewById(R.id.Bsearch);
		Button Baddbook = (Button) findViewById(R.id.Baddbook);
		Button Badduser = (Button) findViewById(R.id.Badduser);

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

		Bsearch.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				flipView(SearchLayout);
			}
		});

		Baddbook.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				flipView(AddBookLayout);
			}
		});

		Badduser.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				new StringDialog(MainActivity.this, "Neuer Nutzer", "Nutzername:", "",
					new StringDialog.ResultTaker() {
						@Override
						public void take(String res) {
							String err = conn.addUser(res);
							if(err != null && err.equals("")) {
								toast("Neuer Nutzer erstellt");
								name = res;
								userinfo(name);
							} else {
								notice("Fehler", err);
							}
						}
					});
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
							// XXX issue #35: freeze on lend error
								
							if(err.equals("")) {
								toast("Erfolgreich verliehen");
								copyinfo(id);
							} else {
								notice("Fehler beim Verleih", err);
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

	private void initBookInfoLayout() {
		Button Baddcopy = (Button) findViewById(R.id.Baddcopy);
		Button Brmbook = (Button) findViewById(R.id.Brmbook);
		Button Btomain4 = (Button) findViewById(R.id.Btomain4);

		Baddcopy.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				new StringDialog(MainActivity.this, "Exemplare hinzufügen", "Anzahl an neuen Exemplaren", "1",
					new StringDialog.ResultTaker() {
						@Override
						public void take(String res) {
							try {
								int n = Integer.parseInt(res);
								String err = conn.addCopies(isbn, n);
								if(err.equals("")) {
									toast("Exemplare hinzugefügt");
									bookinfo(isbn);
								} else {
									notice("Fehler", err);
								}
							} catch(NumberFormatException nfe) {
								notice("Fehler", "'" + res + "' ist keine Zahl");
							}
						}
					});
			}
		});

		Brmbook.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				String err = conn.deleteBook(isbn);
				if(err.equals(""))
					toast("Buch gelöscht");
				else
					notice("Fehler beim Löschen", err);

				flipView(MainLayout);
			}
		});

		Btomain4.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				flipView(MainLayout);
			}
		});
	}

	private void initUserInfoLayout() {
		Button Brmuser = (Button) findViewById(R.id.Brmuser);
		Button Btomain5 = (Button) findViewById(R.id.Btomain5);

		Brmuser.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				String err = conn.deleteUser(name);

				if(err.equals(""))
					toast("Nutzer gelöscht");
				else
					notice("Fehler beim Löschen", err);

				flipView(MainLayout);
			}
		});

		Btomain5.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				flipView(MainLayout);
			}
		});
	}

	private void initAddBookLayout() {
		Button Bdoaddbook = (Button) findViewById(R.id.Bdoaddbook);
		Button Btomain6 = (Button) findViewById(R.id.Btomain6);

		Bdoaddbook.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				String isbn = ((EditText) findViewById(R.id.Eisbn)).getText().toString();
				String title = ((EditText) findViewById(R.id.Etitle)).getText().toString();
				String sauthors = ((EditText) findViewById(R.id.Eauthors)).getText().toString();
				String authors[] = sauthors.split(",");

				String err = conn.addBook(isbn, title, authors);
				if(err.equals(""))
					toast("Buch erstellt");
				else
					notice("Fehler beim Erstellen", err);

				bookinfo(isbn);
			}
		});

		Btomain6.setOnClickListener(new OnClickListener() {
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

		this.id = id;
		this.isbn = null;
		this.name = null;

		TextView Tcopyinfo = (TextView) findViewById(R.id.Tcopyinfo);
		Tcopyinfo.setText("The server says: " + conn.printCopy(id));
		flipView(CopyInfoLayout);
	}

	/**
	 * Fetch info about a Book and show it in the BookInfoLayout.
	 */
	private void bookinfo(String isbn) {
		this.id = -1;
		this.isbn = isbn;
		this.name = null;


		TextView Tbookinfo = (TextView) findViewById(R.id.Tbookinfo);
		Tbookinfo.setText("The server says: " + conn.printBook(isbn));
		flipView(BookInfoLayout);
	}

	/**
	 * Fetch info about a User and show it in the UserInfoLayout.
	 */
	private void userinfo(String name) {
		this.id = -1;
		this.isbn = null;
		this.name = name;

		TextView Tuserinfo = (TextView) findViewById(R.id.Tuserinfo);
		Tuserinfo.setText("The server says: " + conn.printUser(name));
		flipView(UserInfoLayout);
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
