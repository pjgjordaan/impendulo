package runner;

import gov.nasa.jpf.Config;

public class JPFExample {
	public static void main(String[] args) {
		Config config = JPFRunner
				.createConfig(
						"/home/godfried/dev/go/src/github.com/godfried/impendulo/tool/jpf/example/BoundedBuffer.jpf",
						"BoundedBuffer",
						"/home/godfried/dev/go/src/github.com/godfried/impendulo/tool/jpf/bin/");
		System.out.println(JPFRunner.run(config));
	}
}
