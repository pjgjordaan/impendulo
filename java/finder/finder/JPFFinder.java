//Copyright (C) 2013  The Impendulo Authors
//
//This library is free software; you can redistribute it and/or
//modify it under the terms of the GNU Lesser General Public
//License as published by the Free Software Foundation; either
//version 2.1 of the License, or (at your option) any later version.
//
//This library is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
//Lesser General Public License for more details.
//
//You should have received a copy of the GNU Lesser General Public
//License along with this library; if not, write to the Free Software
//Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA

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

/**
 * JPFFinder is used to retrieve a list of concrete classes in the same package
 * with a common parent class. These classes are typically then converted to a
 * Json array and written to an output file. See
 * http://godoc.org/github.com/godfried/impendulo/tool/jpf#GetClasses for more
 * information.
 * 
 * @author godfried
 * 
 */
public class JPFFinder {

	/**
	 * Searches for all class in a package with a certain parent class.
	 * 
	 * @param pkg
	 * @param parentName
	 * @return a List of classes in package pkg with parent class parentName.
	 */
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

	/**
	 * Checks whether a class is a concrete subclass of another.
	 * 
	 * @param c
	 * @param parent
	 * @return
	 */
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

	/**
	 * Retrieves all JPF Listeners.
	 * 
	 * @return
	 */
	@SuppressWarnings("rawtypes")
	public static List<Class> findListeners() {
		return findClasses("gov.nasa.jpf.listener", "gov.nasa.jpf.JPFListener");
	}

	/**
	 * Retreieves all JPF Search classes.
	 * 
	 * @return
	 */
	@SuppressWarnings("rawtypes")
	public static List<Class> findSearches() {
		return findClasses("gov.nasa.jpf.search", "gov.nasa.jpf.search.Search");
	}

	/**
	 * Retrieves the specified list of classes and writes them as Json to an
	 * output file.
	 * 
	 * @param args
	 */
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
