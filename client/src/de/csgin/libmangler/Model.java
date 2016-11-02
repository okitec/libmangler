package de.csgin.libmangler;

import java.util.ArrayList;
import java.util.Arrays;

/**
 * In an ideal world, the model for the app should be the same as the one
 * for the server. The go bind system allows that and generates .AAR (Android
 * Archive) files. However, ant can't use them, gradle doesn't quite work at all,
 * and time is short.
 *
 * This is a translation of the code of manglersrv/load.go and sexps/*
 * To see the model at one glance, the individual classes are mere subclasses
 * instead of residing in their own package in separate files.
 *
 * Go allows functions to return an arbitary number of values. To emulate this
 * behaviour in Java, there are simple public-attribute classes like Parser.TokTail
 * and Parser.State.
 */
public class Model {
	/** TRANSLATION OF PARTS OF manglersrv/load.go **/

	public static class Elem {
		public String[] notes;
		public String[] tags;

		protected int state;
		protected int skip;
		protected boolean notesFilled;
		protected boolean tagsFilled;
	}

	public static class Book extends Elem {
		public String   isbn;
		public String[] authors;
		public String   title;
		public long[]   copies;

		private boolean isbnFilled;
		private boolean authorsFilled;
		private boolean titleFilled;
		private boolean copiesFilled;

		public String toString() {
			return "isbn: " + isbn + "\ntitle: " + title + "\n";
		}
	}

	public static class Copy extends Elem {
		public long     id;
		public String   user;
		public String   isbn;
		public String[] authors;
		public String   title;

		private boolean idFilled;
		private boolean userFilled;
		private boolean isbnFilled;
		private boolean authorsFilled;
		private boolean titleFilled;

		public String toString() {
			return "id: " + id + "\nuser: " + user + "\ntitle: " + title + "\n";
		}
	}

	public static class User extends Elem {
		public String   name;
		public long[]   copies;

		private boolean nameFilled;
		private boolean copiesFilled;

		public String toString() {
			return "name: " + name;
		}
	}

	private static class BookHandler implements AppliedFn {
		@Override
		public void fn(Sexp atom, Sexp parent, Object data) {
			Book b = (Book) data;

			if(b.skip > 0) {
				b.skip--;
				return;
			}

			switch(b.state) {
			case 0:
				if(b.isbnFilled && b.authorsFilled && b.titleFilled && b.notesFilled && b.tagsFilled && b.copiesFilled) {
					b.state = 1000;
					break;
				}

				String s = atom.toString();

				if(s.equals("book")) {
					b.state = 1;
				} else if(s.equals("authors")) {
					b.authors = list(parent.cdr()).toArray(new String[]{""});
					b.skip = b.authors.length;
					b.authorsFilled = true;
				} else if(s.equals("title")) {
					b.state = 2;
				} else if(s.equals("notes")) {
					b.notes = list(parent.cdr()).toArray(new String[]{""});
					b.skip = b.notes.length;
					b.notesFilled = true;
				} else if(s.equals("tags")) {
					b.tags = list(parent.cdr()).toArray(new String[]{""});
					b.skip = b.tags.length;
					b.tagsFilled = true;
				} else if(s.equals("copies")) {
					CopyIDError cerr = getCopyIDs(parent.cdr());
					b.copies = cerr.copies;
					if(cerr.err != null) {
						b.state = -1;
					}
					b.skip = b.copies.length;
					b.copiesFilled = true;
				} else {
					b.state = -1;
				}
				break;
			case 1:
				b.isbn = atom.toString();
				b.isbnFilled = true;
				b.state = 0;
				break;

			case 2:
				b.title = atom.toString();
				b.titleFilled = true;
				b.state = 0;

			case 1000:
				return;

			case -1:
			default:
				break;
			}
		}
	}

	private static CopyIDError getCopyIDs(Sexp sexp) {
		ArrayList<String> scopies = list(sexp);
		int size = 1;  // small start value so that the growing code is exercised
		int nused = 0;
		long[] ls = new long[size];

		for(String s : scopies) {
			long i;
			try {
				if(nused >= ls.length) {
					size *= 2;
					ls = Arrays.copyOf(ls, size);
				}

				ls[nused] = Integer.parseInt(s);
				nused++;
			} catch(NumberFormatException nfe) {
				return new CopyIDError(ls, "not a number");
			}
		}
	
		return new CopyIDError(ls, null);
	}

	/** Return tuple of getCopyIDs */
	private static class CopyIDError {
		public long[] copies;
		public String err;

		public CopyIDError(long[] copies, String err) {
			this.copies = copies;
			this.err = err;
		}
	}

	private static class CopyHandler implements AppliedFn {
		@Override
		public void fn(Sexp atom, Sexp parent, Object data) {
			
		}
	}

	private static class UserHandler implements AppliedFn {
		@Override
		public void fn(Sexp atom, Sexp parent, Object data) {
			
		}
	}

	/** Read all books in s */
	public static ArrayList<Book> getBooks(String s) {
		String tail = s;
		ArrayList<Book> books = new ArrayList<Book>();

		while(tail.length() > 1) { // lonely newline 
			Parser.State st = Parser.Parse(s);
			//if(st.err != null)
			//	break;

			Book b = new Book();
			apply(st.sexp, new BookHandler(), b);
			books.add(b);
		}

		return books;
	}

	public static ArrayList<Copy> getCopies(String s) {
		String tail = s;
		ArrayList<Copy> copies = new ArrayList<Copy>();

		while(tail.length() > 1) { // lonely newline 
			Parser.State st = Parser.Parse(s);
			if(st.err != null)
				break;

			Copy c = new Copy();
			apply(st.sexp, new CopyHandler(), c);
			copies.add(c);
		}

		return copies;
	}

	public static ArrayList<User> getUsers(String s) {
		String tail = s;
		ArrayList<User> users = new ArrayList<User>();

		while(tail.length() > 1) { // lonely newline 
			Parser.State st = Parser.Parse(s);
			if(st.err != null)
				break;

			User u = new User();
			apply(st.sexp, new UserHandler(), u);
			users.add(u);
		}

		return users;
	}

	/** TRANSLATION OF sexps/sexps.go **/

	private static interface Sexp {
		public String  toString();
		public Sexp    car();
		public Sexp    cdr();
		public boolean isAtom();
	}

	private static class Cell implements Sexp {
		private Sexp car;
		private Sexp cdr;

		public Cell(Sexp car, Sexp cdr) {
			this.car = car;
			this.cdr = cdr;
		}

		public String toString() {
			String scar, scdr;
			if(car == null)
				scar = "()";
			else
				scar = car.toString();

			if(cdr == null)
				scdr = "()";
			else
				scdr = cdr.toString();

			return "(" + scar + " . " + scdr + ")";
		}

		public Sexp car() {
			return car;
		}

		public Sexp cdr() {
			return cdr;
		}

		public boolean isAtom() {
			return false;
		}
	}

	private static class Atom implements Sexp {
		private String value;

		public Atom(String s) {
			value = s;
		}

		public String toString() {
			return value;
		}

		public Sexp car() {
			return null;
		}

		public Sexp cdr() {
			return null;
		}

		public boolean isAtom() {
			return true;
		}
	}

	private static interface AppliedFn {
		public void fn(Sexp atom, Sexp parent, Object data);
	}

	private static void apply(Sexp sexp, AppliedFn fn, Object data) {
		preorder(sexp, null, fn, data);
	}

	private static void preorder(Sexp sexp, Sexp parent, AppliedFn fn, Object data) {
		if(sexp.isAtom()) {
			fn.fn(sexp, parent, data);
		}

		if(sexp.car() != null)
			preorder(sexp.car(), sexp, fn, data);

		if(sexp.cdr() != null)
			preorder(sexp.cdr(), sexp, fn, data); 
	}

	private static ArrayList<String> list(Sexp sexp) {
		ArrayList<String> ls = new ArrayList<String>();

		if(sexp == null)	
			return null;

		if(sexp.isAtom()) {
			ls.add(sexp.toString());
			return ls;
		}

		for(;;) {
			if(sexp.car() == null)
				ls.add("");
			else
				ls.add(sexp.car().toString());

			sexp = sexp.cdr();
			if(sexp == null)
				return ls;				
		}
	}

	private static class Parser {
		public static class State {
			public Sexp sexp;
			public String tail;
			public String err;

			public State(Sexp sexp, String tail, String err) {
				this.sexp = sexp;
				this.tail = tail;
				this.err = err;
			}
		}

		private static class TokTail {
			public String tok;
			public String tail;

			public TokTail(String tok, String tail) {
				this.tok = tok;
				this.tail = tail;
			}
		}

		public static State Parse(String s) {
			return sexpr(s);
		}

		private static State sexpr(String s) {
			TokTail tt = tok(s);

			if(tt.tok.equals("(")) {
				State st = sexprlist(tt.tail);
				if(st.err != null)
					return st;

				tt = tok(tt.tail);
				if(tt.tok != ")" && tt.tok != "")
					return new State(st.sexp, tt.tail, "sexpr: missing ')'");
			}

			if(tt.tok.equals(")"))
				return new State(null, tt.tail, "sexpr: unexpected ')'");

			return new State(new Atom(tt.tok), tt.tail, null);
		}

		private static State sexprlist(String s) {
			TokTail tt = tok(s);

			if(tt.tok.equals(""))
				return new State(new Atom(tt.tok), tt.tail, null);

			if (tt.tok.equals(")"))
				return new State(null, untok(tt), null);

			State stcar = sexpr(untok(tt));
			if(stcar.err != null)
				return stcar;

			State stcdr = sexprlist(stcar.tail);
			return new State(new Cell(stcar.sexp, stcdr.sexp), stcdr.tail, stcdr.err);
		}

		private static TokTail tok(String s) {
			int start = -1;
			boolean str = false;

			int r = s.codePointAt(0);
			for(int i = 0; i < s.length(); i += Character.charCount(r)) {
				r = s.codePointAt(i);

				switch(r) {
				case '(':
					return new TokTail("(", s.substring(i+1));

				case ')':
					if(start >= 0)
						return new TokTail(s.substring(start, i), s.substring(i));

					return new TokTail(")", s.substring(i+1));

				case '"':
					if(!str && start < 0) {
						str = true;
						start = i + 1;
					} else if(str) {
						return new TokTail(s.substring(start, i), s.substring(i+1));
					}

					break;

				case ' ':
				case '\t':
					if(!str && start >= 0)
						return new TokTail(s.substring(start, i), s.substring(i));

					break;

				default:
					if(start < 0) {
						start = i;
					}
				}
			}

			if(start < 0)
				return new TokTail("", s);

			return new TokTail(s.substring(start), "");
		}

		private static String untok(TokTail tt) {
			if((tt.tok.contains(" ") || tt.tok.contains("\t")) &&
			  !(tt.tok.contains("(") || tt.tok.contains(")"))) {
				return "\"" + tt.tok + "\"" + tt.tail;
			}
			return tt.tok + tt.tail;
		}
	}
}
