package de.csgin.libmangler;

import android.app.AlertDialog;
import android.content.Context;
import android.content.DialogInterface;
import android.widget.EditText;
import android.widget.LinearLayout;
import android.widget.TextView;

/**
 * Simple way of opening a dialog that queries a string.
 */
public class StringDialog {
	private static final int ID = 42;
	private final ResultTaker taker;

	/**
	 * @param ctxt  Context of caller
	 * @param title Title of the dialog
	 * @param msg   Message displayed above the EditText
	 * @param def   Default text in the EditText
	 * @param rt    Callback for result
	 * @return The entered string (perhaps the default) or null if aborted.
	 */
	public StringDialog(Context ctxt, String title, String msg, String def, ResultTaker rt) {
		AlertDialog.Builder b = new AlertDialog.Builder(ctxt);
		LinearLayout ll = new LinearLayout(ctxt);

		taker = rt;

		EditText et = new EditText(ctxt);
		et.setText(def, TextView.BufferType.NORMAL);
		et.setId(ID);

		ll.addView(et);

		b.setTitle(title);
		b.setMessage(msg);
		b.setCancelable(true);
		b.setView(ll);

		b.setPositiveButton("OK", new DialogInterface.OnClickListener() {
			@Override
			public void onClick(DialogInterface dialog, int which) {
				EditText input = (EditText) ((AlertDialog) dialog).findViewById(ID);
				taker.take(input.getText().toString());
			}
		});

		b.show();
	}

	public interface ResultTaker {
		public void take(String res);
	}
}
