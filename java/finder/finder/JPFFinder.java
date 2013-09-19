package finder;

import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.lang.reflect.Modifier;
import java.util.ArrayList;
import java.util.List;

import util.ClassFinder;

import com.google.gson.JsonArray;
import com.google.gson.JsonObject;

public class JPFFinder {

	@SuppressWarnings("rawtypes")
	public static List<Class> findClasses(String pkg, String parentName) {
		Class<?> parent;
		Class<?>[] found;
		try {
			found = ClassFinder.getClasses(pkg);
			parent = Class.forName(parentName);
		} catch (ClassNotFoundException | IOException e) {
			return null;
		}
		List<Class> classes = new ArrayList<Class>();
		for (Class<?> c : found) {
			if (isSubclass(c, parent)) {
				classes.add(c);
			}
		}
		return classes;
	}

	public static boolean isSubclass(Class<?> c, Class<?> parent) {
		try {
			if (!Modifier.isAbstract(c.getModifiers())
					&& !Modifier.isInterface(c.getModifiers())
					&& !c.equals(parent) && c.asSubclass(parent) != null)
				return true;
		} catch (ClassCastException e) {
		}
		return false;
	}

	@SuppressWarnings("rawtypes")
	public static List<Class> findListeners() {
		return findClasses("gov.nasa.jpf.listener", "gov.nasa.jpf.JPFListener");
	}

	@SuppressWarnings("rawtypes")
	public static List<Class> findSearches() {
		return findClasses("gov.nasa.jpf.search", "gov.nasa.jpf.search.Search");
	}

	@SuppressWarnings("rawtypes")
	public static void main(String[] args) {
		if (args.length < 2) {
			System.err.println("Insufficient arguments.");
		}
		String type = args[0];
		String outFilename = args[1];
		List<Class> classes = null;
		if (type.equals("listeners")) {
			classes = findListeners();
		} else if (type.equals("searches")) {
			classes = findSearches();
		} else {
			System.err.printf("Unknown type %s.%n", type);
			return;
		}
		JsonArray vals = new JsonArray();
		for (Class clazz : classes) {
			JsonObject obj = new JsonObject();
			obj.addProperty("Name", clazz.getSimpleName());
			obj.addProperty("Package", clazz.getPackage().getName());
			vals.add(obj);
		}
		FileOutputStream out = null;
		try {
			File outFile = new File(outFilename);
			out = new FileOutputStream(outFile);
			out.write(vals.toString().getBytes());
		} catch (IOException e) {
			System.err.println(e.getMessage());
		} finally {
			if (out != null) {
				try {
					out.close();
				} catch (IOException e) {
					System.err.println(e.getMessage());
				}
			}
		}
	}
}
