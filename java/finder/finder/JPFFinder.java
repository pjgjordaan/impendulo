//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, this
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

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
