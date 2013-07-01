package runner;

import gov.nasa.jpf.Config;
import gov.nasa.jpf.JPF;
import gov.nasa.jpf.JPFConfigException;
import gov.nasa.jpf.JPFException;

import java.io.File;
import java.io.FileNotFoundException;
import java.io.PrintStream;

import util.NullOutputStream;

import com.google.gson.JsonArray;
import com.google.gson.JsonElement;
import com.google.gson.JsonObject;

public class JPFRunner {

	/**
	 * @param args
	 * @throws FileNotFoundException
	 */
	public static void main(String[] args) throws FileNotFoundException {
		if(args.length != 3){
			throw new IllegalArgumentException("Expected 3 arguments.");
		}
		if(!new File(args[0]).exists()){
			throw new FileNotFoundException("Could not find config file "+args[0]);
		}
		System.out.println(run(createConfig(args[0], args[1], args[2])));
		
	}
	
	public static JsonElement run(Config config){
		PrintStream defaultOut = System.out;
		PrintStream defaultErr = System.err;
		System.setErr(new PrintStream(new NullOutputStream()));
		System.setOut(new PrintStream(new NullOutputStream()));
		JsonArray errs = null;
		try {
			JPF jpf = new JPF(config);
			jpf.run();
			if (jpf.foundErrors()) {
				errs = new JsonArray();
				for (gov.nasa.jpf.Error err : jpf.getSearchErrors()) {
					final JsonObject ret = new JsonObject();
					ret.addProperty("description", err.getDescription());
					ret.addProperty("details", err.getDetails());
					ret.addProperty("id", err.getId());
					errs.add(ret);
				}
			}
		} catch (JPFConfigException cx) {
			JsonObject err = new JsonObject();
			err.addProperty("error", cx.getMessage());
			return err;
		} catch (JPFException jx) {
			JsonObject err = new JsonObject();
			err.addProperty("error", jx.getMessage());
			return err;
		} finally{
			System.setOut(defaultOut);
			System.setErr(defaultErr);
		}
		return errs;
	}

	public static Config createConfig(String configName, String target,
			String targetLocation) {
		Config config = JPF.createConfig(new String[] { configName });
		config.setProperty("target", target);
		config.setProperty("classpath",
				targetLocation + ";" + config.getProperty("classpath"));
		return config;
	}

}
