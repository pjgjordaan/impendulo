package listener;

import java.io.IOException;
import java.lang.annotation.Annotation;
import java.util.ArrayList;
import java.util.List;

import util.ClassFinder;

import com.google.gson.JsonArray;
import com.google.gson.JsonObject;

public class ListenerFinder {
	@SuppressWarnings("rawtypes")
	public static List<Class> findListeners() {
		Class<?> parent;
		Class<?>[] found;
		try {
			found = ClassFinder.getClasses("gov.nasa.jpf.listener");
			parent = Class
					.forName("gov.nasa.jpf.JPFListener");
		} catch (ClassNotFoundException | IOException e) {
			return null;
		}
		List<Class> listeners = new ArrayList<Class>();
		for (Class<?> c : found) {
			if(isListener(c, parent)){
				listeners.add(c);
			}
		}
		return listeners;
	}
	
	public static boolean isListener(Class<?> c, Class<?> parent){
		try {
			if(c.asSubclass(parent) != null)
				return true;
		} catch (ClassCastException e) {
		}
		return false;
	}
	
	@SuppressWarnings("rawtypes")
	public static void main(String[] args){
		JsonArray vals = new JsonArray();
		List<Class> listeners = findListeners();
		for(Class listener : listeners){
			JsonObject obj = new JsonObject();
			Annotation[] annotations = listener.getAnnotations();
			for(Annotation a : annotations){
				System.out.println(a.toString());
			}
			obj.addProperty("Name", listener.getSimpleName());
			obj.addProperty("Package", listener.getPackage().getName());
			vals.add(obj);
		}
		System.out.println(vals.toString());
	}
}
