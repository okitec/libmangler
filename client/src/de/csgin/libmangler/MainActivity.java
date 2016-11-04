package de.csgin.libmangler;

import android.app.Activity;
import android.app.AlertDialog;
import android.content.Intent;
import android.os.Bundle;
import android.os.StrictMode;
import android.text.method.ScrollingMovementMethod;
import android.util.Log;
import android.view.Menu;
import android.view.View;
import android.view.View.OnClickListener;
import android.widget.AdapterView;
import android.widget.ArrayAdapter;
import android.widget.Button;
import android.widget.EditText;
import android.widget.ListAdapter;
import android.widget.ListView;
import android.widget.Spinner;
import android.widget.TextView;
import android.widget.Toast;
import android.widget.ViewFlipper;

import java.io.IOException;
import java.net.UnknownHostException;
import java.util.ArrayList;
import java.util.Arrays;

/**
 * The only Activity of the libmangler app. There's a central ViewFlipper flipping between
 * seven layouts instead of multiple activities. Much of the code of MainActivity is devoted
 * to the initialisation of said layouts and the button-handlers within them.
 *
 * The user interface is only initialised after a network connection has been opened.
 *
 * There are also a few helper routines like list(), notice(), panic(), toast().
 */
public class MainActivity extends Activity {
	/* ViewFlipper indexes - if this were C, I'd use a enum instead of writing = 0, ... */
	private static final int MainLayout = 0;
	private static final int SearchLayout = 1;
	private static final int PanicLayout = 2;
	private static final int ElemsLayout = 3;
	private static final int AddBookLayout = 4;
	private static final int CopyInfoLayout = 5;
	private static final int BookInfoLayout = 6;
	private static final int UserInfoLayout = 7;

	/** Request code for QR code scanning intent */
	private static final int SCANREQ = 0;

	/* default server and port */
	private static final String SRVADDR = "oquasinus.duckdns.org";
	private static final int PORT = 40000;

	private String srvaddr;
	private int port;
	private Connection conn;

	/* currently investigated copy, book, name */
	private long id = -1;
	private String isbn;
	private String name;

	/**
	 * When this is false, the back button closes the app instead of returning
	 * to the main screen. This is the case when panic has been called, but also
	 * before a connection has been established.
	 */
	private boolean canReturnToMain;

	@Override
	protected void onCreate(Bundle savedInstanceState) {
		super.onCreate(savedInstanceState);
		setContentView(R.layout.activity_main);

		// Override NetworkOnMainThreadException because I think in synchronous terms.
		// This is a last-ditch measure, but I want to have a reasonably working app.
		// src: http://stackoverflow.com/questions/6343166/how-to-fix-android-os-networkonmainthreadexception#6343299
		StrictMode.ThreadPolicy p = new StrictMode.ThreadPolicy.Builder().permitAll().build();
		StrictMode.setThreadPolicy(p); 

		srvaddr = SRVADDR;
		port = PORT;
		canReturnToMain = false;

		// The Panic Layout is used now so that the user doesn't click any buttons
		// that don't have any effect because no connection is open. The layouts
		// are initialised only after the connection is open.
		((TextView) findViewById(R.id.Tpanic)).setText("Keine Verbindung aufgebaut");
		flipView(PanicLayout);

		new StringDialog(this, "Serveradresse", "Format: Adresse[:Port]", SRVADDR, new StringDialog.ResultTaker() {
			@Override
			public void take(String res) {
				String ap[] = res.split(":"); // address:port tuple

				if(ap.length == 1) {
					srvaddr = ap[0];
					port = PORT;
				} else if(ap.length == 2) {
					srvaddr = ap[0];
					try {
						port = Integer.parseInt(ap[1]);
					} catch(NumberFormatException nfe) {
						port = PORT;
					}
				} else {
					srvaddr = SRVADDR;
					port = PORT;
				}

				conn = getConn(srvaddr, port);

				if(conn != null) {
					toast("Verbindung geöffnet");
					// Only now initialise the Buttons, after conn is != null and we know
					// that clicking them won't crash the app.
					initLayouts();
					flipView(MainLayout);
				}
			}
		});
	}

	@Override
	public boolean onCreateOptionsMenu(Menu menu) {
		// Inflate the menu; this adds items to the action bar if it is present.
		getMenuInflater().inflate(R.menu.main, menu);
		return true;
	}

	@Override
	protected void onSaveInstanceState(Bundle out) {
		super.onSaveInstanceState(out);
		out.putLong("id", id);
		out.putString("isbn", isbn);
		out.putString("name", name);
		out.putString("srvaddr", srvaddr);
		out.putInt("port", port);
		out.putBoolean("canReturnToMain", canReturnToMain);
	}

	@Override
	protected void onRestoreInstanceState(Bundle in) {
		super.onRestoreInstanceState(in);
		id = in.getLong("id");
		isbn = in.getString("isbn");
		name = in.getString("name");
		srvaddr = in.getString("srvaddr");
		port = in.getInt("port");
		canReturnToMain = in.getBoolean("canReturnToMain");
	}

	@Override
	public void onBackPressed() {
		if(canReturnToMain)
			flipView(MainLayout);
		else
			finish();
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
					toast("QR-Code ist keine Zahl");
				}
			}
			/* don't do anything on failure */
		}
	}

	/**
	 * Open a connection to a server.
	 */
	private Connection getConn(String srvaddr, int port) {
		Connection conn;

		try {
			conn = new Connection(srvaddr, port);
		} catch(UnknownHostException uhe) {
			panic("Server '" + srvaddr + "' nicht gefunden");
			return null;
		} catch(IOException ioe) {
			panic("Kann keine Verbindung zu '" + srvaddr + "' aufbauen");
			return null;
		} catch (android.os.NetworkOnMainThreadException netmain) {
			panic("NetworkOnMainThreadException - sollte nicht passieren");
			return null;
		}

		if(!conn.isProperVersion()) {
			panic("Inkompatibles Protokoll zwischen Client und Server!");
			return null;
		}

		return conn;
	}

	/**
	 * Initialise the layouts.
	 */
	private void initLayouts() {
		initMainLayout();
		initSearchLayout();
		// PanicLayout needs no initialisation, as it has no buttons.
		initElemsLayout();
		initAddBookLayout();
		initCopyInfoLayout();
		initBookInfoLayout();
		initUserInfoLayout();
	}

	private void initMainLayout() {
		Button Bscan = (Button) findViewById(R.id.Bscan);
		Button Bsearch = (Button) findViewById(R.id.Bsearch);
		Button Baddbook = (Button) findViewById(R.id.Baddbook);
		Button Badduser = (Button) findViewById(R.id.Badduser);
		Button Blisttags = (Button) findViewById(R.id.Blisttags);
		Button Blistbooks = (Button) findViewById(R.id.Blistbooks);
		Button Blistusers = (Button) findViewById(R.id.Blistusers);
		Button Bclose = (Button) findViewById(R.id.Bclose);

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
				} catch(android.content.ActivityNotFoundException anfe) {
					notice("Fehler", "Der ZXing-Barcodescanner ist nicht installiert. Bitte installieren, um QR-Codes lesen zu können.");
					flipView(MainLayout);
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

		Blisttags.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				list(conn.listTags().split("\n"));
			}
		});

		Blistbooks.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				list(conn.listBooks().split("\n"));
			}
		});

		Blistusers.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				list(conn.listUsers().split("\n"));
			}
		});

		Bclose.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				conn.quit("app closing");
				finish();
			}
		});
	}

	private void initSearchLayout() {
		Spinner Selemtype = (Spinner) findViewById(R.id.Selemtype);
		Button Btomain2 = (Button) findViewById(R.id.Btomain2);
		Button Bdosearch = (Button) findViewById(R.id.Bdosearch);

		// cf. https://developer.android.com/guide/topics/ui/controls/spinner.html
		ArrayAdapter<CharSequence> aa = ArrayAdapter.createFromResource(this, R.array.elemtype_array, android.R.layout.simple_spinner_item);
		aa.setDropDownViewResource(android.R.layout.simple_spinner_dropdown_item);
		Selemtype.setAdapter(aa);

		Btomain2.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				flipView(MainLayout);
			}
		});

		Bdosearch.setOnClickListener(new OnClickListener() {
			private static final int BookPos = 0;
			private static final int CopyPos = 1;
			private static final int UserPos = 2;

			@Override
			public void onClick(View v) {
				Spinner Selemtype = (Spinner) findViewById(R.id.Selemtype);
				EditText Esearch = (EditText) findViewById(R.id.Esearch);
				ListView Lelems = (ListView) findViewById(R.id.Lelems);
				String cmd = null;

				String s = Esearch.getText().toString();
				switch(Selemtype.getSelectedItemPosition()) {
				case BookPos:
					cmd = String.format("B/authors:%s, title:%s, notes:%s, tags:%s/λ", s, s, s, s);
					break;
				case CopyPos:
					cmd = String.format("C/notes:%s, tags:%s/λ", s, s, s);
					break;
				case UserPos:
					cmd = String.format("U/name:%s, notes:%s, tags:%s/λ", s, s, s);
					break;
				}

				list(conn.transact(cmd).split("\n"));
			}
		});
	}

	private void initElemsLayout() {
		ListView Lelems = (ListView) findViewById(R.id.Lelems);
		Button Btomain3 = (Button) findViewById(R.id.Btomain3);

		ArrayList<String> ls = new ArrayList<String>();
		ArrayAdapter<String> aa = new ArrayAdapter<String>(this, android.R.layout.simple_list_item_1, ls);
		Lelems.setAdapter(aa);

		Lelems.setOnItemClickListener(new AdapterView.OnItemClickListener() {
			@Override
			public void onItemClick(AdapterView<?> parent, View view, int pos, long id) {
				ListView lv = (ListView) parent;

				ListAdapter la = lv.getAdapter();

				String s = (String) la.getItem(pos);
				String fld[] = s.split(" ");

				if(fld[0].equals("book")) {
					bookinfo(fld[1]);
				} else if(fld[0].equals("copy")) {
					try {
						copyinfo(Long.parseLong(fld[1]));
					} catch(NumberFormatException nfe) {
						notice("Interner Fehler", "'"+fld[1] + "' ist keine Zahl");
					}
				} else if(fld[0].equals("user")) {
					// Of course there's no String.join() before Java 8, of course.
					String name = "";

					for(String f: Arrays.copyOfRange(fld, 1, fld.length))
						name += f + " ";
					name = name.trim();

					userinfo(name);
				} // last case: tags; ignore them
			}
		});

		Btomain3.setOnClickListener(new OnClickListener() {
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

				for(int i = 0; i < authors.length; i++)
					authors[i] = authors[i].trim();

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

	private void initCopyInfoLayout() {
		TextView Tcopyinfo = (TextView) findViewById(R.id.Tcopyinfo);
		final ViewFlipper Fcopyinfo = (ViewFlipper) findViewById(R.id.Fcopyinfo);
		Button Blend = (Button) findViewById(R.id.Blend);
		Button Breturn = (Button) findViewById(R.id.Breturn);
		Button Bretire = (Button) findViewById(R.id.Bretire);
		Button Bflipcopy2 = (Button) findViewById(R.id.Bflipcopy2);
		Button Btomain = (Button) findViewById(R.id.Btomain);
		Button Btobook = (Button) findViewById(R.id.Btobook);
		Button Bnote = (Button) findViewById(R.id.Bnote);
		Button Baddtag = (Button) findViewById(R.id.Baddtag);
		Button Brmtag = (Button) findViewById(R.id.Brmtag);
		Button Bflipcopy1 = (Button) findViewById(R.id.Bflipcopy1);

		// Why can't this be done in XML? Really.
		Tcopyinfo.setMovementMethod(new ScrollingMovementMethod());

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

		Bflipcopy2.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				Fcopyinfo.showNext();
			}
		});

		Btomain.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				flipView(MainLayout);
			}
		});

		Btobook.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				bookinfoOfID(id);
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
							conn.noteCopy(res, id);
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
							conn.addTagCopy(res, id);
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
							conn.rmTagCopy(res, id);
							toast("Tag entfernt");
							copyinfo(id);
						}
					});
			}
		});


		Bflipcopy1.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				Fcopyinfo.showNext();
			}
		});
	}

	private void initBookInfoLayout() {
		TextView Tbookinfo = (TextView) findViewById(R.id.Tbookinfo);
		final ViewFlipper Fbookinfo = (ViewFlipper) findViewById(R.id.Fbookinfo);
		Button Blistcopies = (Button) findViewById(R.id.Blistcopies);
		Button Baddcopy = (Button) findViewById(R.id.Baddcopy);
		Button Brmbook = (Button) findViewById(R.id.Brmbook);
		Button Bflipbook2 = (Button) findViewById(R.id.Bflipbook2);
		Button Btomain4 = (Button) findViewById(R.id.Btomain4);
		Button Bnote2 = (Button) findViewById(R.id.Bnote2);
		Button Baddtag2 = (Button) findViewById(R.id.Baddtag2);
		Button Brmtag2 = (Button) findViewById(R.id.Brmtag2);
		Button Bflipbook1 = (Button) findViewById(R.id.Bflipbook1);

		Tbookinfo.setMovementMethod(new ScrollingMovementMethod());

		Blistcopies.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				list(conn.listCopies(isbn).split("\n"));
			}
		});

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

		Bflipbook2.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				Fbookinfo.showNext();
			}
		});

		Btomain4.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				flipView(MainLayout);
			}
		});

		Bnote2.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				new StringDialog(MainActivity.this, "Notiz hinzufügen", "", "",
					new StringDialog.ResultTaker() {
						@Override
						public void take(String res) {
							conn.noteBook(res, isbn);
							toast("Notiz hinzugefügt");
							bookinfo(isbn);
						}
					});
			}
		});

		Baddtag2.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				new StringDialog(MainActivity.this, "Tag hinzufügen", "", "",
					new StringDialog.ResultTaker() {
						@Override
						public void take(String res) {
							conn.addTagBook(res, isbn);
							toast("Tag hinzugefügt");
							bookinfo(isbn);
						}
					});
			}
		});

		Brmtag2.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				new StringDialog(MainActivity.this, "Tag entfernen", "", "",
					new StringDialog.ResultTaker() {
						@Override
						public void take(String res) {
							conn.rmTagBook(res, isbn);
							toast("Tag entfernt");
							bookinfo(isbn);
						}
					});
			}
		});

		Bflipbook1.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				Fbookinfo.showNext();
			}
		});
	}

	private void initUserInfoLayout() {
		TextView Tuserinfo = (TextView) findViewById(R.id.Tuserinfo);
		final ViewFlipper Fuserinfo = (ViewFlipper) findViewById(R.id.Fuserinfo);
		Button Blistcopies2 = (Button) findViewById(R.id.Blistcopies2);
		Button Breturnall = (Button) findViewById(R.id.Breturnall);
		Button Brmuser = (Button) findViewById(R.id.Brmuser);
		Button Bflipuser2 = (Button) findViewById(R.id.Bflipuser2);
		Button Btomain5 = (Button) findViewById(R.id.Btomain5);
		Button Bnote3 = (Button) findViewById(R.id.Bnote3);
		Button Baddtag3 = (Button) findViewById(R.id.Baddtag3);
		Button Brmtag3 = (Button) findViewById(R.id.Brmtag3);
		Button Bflipuser1 = (Button) findViewById(R.id.Bflipuser1);

		Tuserinfo.setMovementMethod(new ScrollingMovementMethod());

		Blistcopies2.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				list(conn.listCopies(name).split("\n"));
			}
		});

		Breturnall.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				conn.returnAll(name);
				toast("Alle Exemplare zurückgegeben");
				userinfo(name);
			}
		});

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

		Bflipuser2.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				Fuserinfo.showNext();
			}
		});

		Btomain5.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				flipView(MainLayout);
			}
		});

		Bnote3.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				new StringDialog(MainActivity.this, "Notiz hinzufügen", "", "",
					new StringDialog.ResultTaker() {
						@Override
						public void take(String res) {
							conn.noteUser(res, name);
							toast("Notiz hinzugefügt");
							userinfo(name);
						}
					});
			}
		});

		Baddtag3.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				new StringDialog(MainActivity.this, "Tag hinzufügen", "", "",
					new StringDialog.ResultTaker() {
						@Override
						public void take(String res) {
							conn.addTagUser(res, name);
							toast("Tag hinzugefügt");
							userinfo(name);
						}
					});
			}
		});

		Brmtag3.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				new StringDialog(MainActivity.this, "Tag entfernen", "", "",
					new StringDialog.ResultTaker() {
						@Override
						public void take(String res) {
							conn.rmTagBook(res, name);
							toast("Tag entfernt");
							userinfo(name);
						}
					});
			}
		});

		Bflipuser1.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View v) {
				Fuserinfo.showNext();
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

		String info = conn.printCopy(id);
		if(info.equals("")) {
			notice("Fehler", "'" + id + "' existiert nicht");
			flipView(MainLayout);
			return;
		}

		TextView Tcopyinfo = (TextView) findViewById(R.id.Tcopyinfo);
		Tcopyinfo.setText(info);
		flipView(CopyInfoLayout);
	}

	/**
	 * Fetch info about a Book and show it in the BookInfoLayout.
	 */
	private void bookinfo(String isbn) {
		this.id = -1;
		this.isbn = isbn;
		this.name = null;

		String info = conn.printBook(isbn);
		if(info.equals("")) {
			notice("Fehler", "'" + isbn + "' existiert nicht");
			flipView(MainLayout);
			return;
		}

		TextView Tbookinfo = (TextView) findViewById(R.id.Tbookinfo);
		Tbookinfo.setText(info);
		flipView(BookInfoLayout);
	}

	/**
	 * Fetch info about the book this copy belongs to.
	 */
	private void bookinfoOfID(long id) {
		this.id = -1;
		this.isbn = isbn;
		this.name = null;

		String info = conn.printBookOfID(id);
		if(info.equals("")) {
			// Every copy *must* have a book. If we don't find a book,
			// we can assume the copy doesn't exist.
			notice("Fehler", "'" + id + "' existiert nicht");
			flipView(MainLayout);
			return;
		}

		TextView Tbookinfo = (TextView) findViewById(R.id.Tbookinfo);
		Tbookinfo.setText(info);
		flipView(BookInfoLayout);
	}

	/**
	 * Fetch info about a User and show it in the UserInfoLayout.
	 */
	private void userinfo(String name) {
		this.id = -1;
		this.isbn = null;
		this.name = name;

		String info = conn.printUser(name);
		if(info.equals("")) {
			notice("Fehler", "'" + name + "' existiert nicht");
			flipView(MainLayout);
			return;
		}

		TextView Tuserinfo = (TextView) findViewById(R.id.Tuserinfo);
		Tuserinfo.setText(info);
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
		canReturnToMain = false;
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

	/**
	 * Show a list in the list layout.
	 */
	private void list(String ls[]) {
		ListView Lelems = (ListView) findViewById(R.id.Lelems);
		ArrayAdapter<String> aa = (ArrayAdapter<String>) Lelems.getAdapter();
		aa.clear();
		aa.addAll(ls);
		flipView(ElemsLayout);
	}
}
