Impendulo
=========

Server and web app for Impendulo system.

[Documentation](http://godoc.org/github.com/godfried/impendulo "Godoc Documentation")

Currently the only language supported by Impendulo is Java.

The Java Impendulo toolchain currently consists of the following:
- [PMD](http://pmd.sourceforge.net/ "PMD source code analyzer")
- [Checkstyle](http://checkstyle.sourceforge.net/ "Checkstyle")
- [FindBugs](http://findbugs.sourceforge.net/ "FindBugs static analysis tool")
- [Java Pathfinder](http://babelfish.arc.nasa.gov/trac/jpf/ "Java Pathfinder")
- [JUnit 4](http://junit.org/ "JUnit testing framework")
- [Java Compiler](http://openjdk.java.net/groups/compiler/ "OpenJDK Java Compiler")
 
The only editor currently supported is Eclipse. You can install its plugin from http://cs.sun.ac.za/~pjordaan/intlola/.

Installation
------------
Install and setup go:
```
~$ wget http://go.googlecode.com/files/go1.1.2.linux-amd64.tar.gz
~$ sudo tar -C /usr/local -xzf go1.1.2.linux-amd64.tar.gz
~$ export PATH=$PATH:/usr/local/go/bin
~$ mkdir $HOME/go
~$ GOPATH=$HOME/go
~$ export PATH=$PATH:$GOPATH/bin
```
You should add the export commands to something like `/etc/profile` or `$HOME/.profile` so that you don't 
have to repeatedly set them.

Download and install impendulo:

`go get github.com/godfried/impendulo`

Then you need to install the tools you want to use:
First create an Impendulo tools directory. For this example we create it in our home directory:
```
~/somewhere$ cd ~
~$ mkdir impendulo-tools
~$ cd impendulo-tools
```
Now we download and install the tools we want:
- [PMD](https://sourceforge.net/projects/pmd/files/pmd/5.0.5/pmd-bin-5.0.5.zip/download)

```
~/impendulo-tools$ wget https://sourceforge.net/projects/pmd/files/pmd/5.0.5/pmd-bin-5.0.5.zip/download
... lots of stuff ...
2013-09-20 14:52:07 (1,13 MB/s) - `download' saved [19033807/19033807]
~/impendulo-tools$ unzip download
... lots of stuff ...
~/impendulo-tools$ rm download
```

- [Checkstyle](http://sourceforge.net/projects/checkstyle/files/checkstyle/5.6/checkstyle-5.6-bin.tar.gz/download)

```
~/impendulo-tools$ wget http://sourceforge.net/projects/checkstyle/files/checkstyle/5.6/checkstyle-5.6-bin.tar.gz/download
... lots of stuff ...
2013-09-20 14:56:27 (265 KB/s) - `download' saved [4728059/4728059]
~/impendulo-tools$ tar -xvf download
... lots of stuff ...
~/impendulo-tools$ rm download
```

- [FindBugs](http://prdownloads.sourceforge.net/findbugs/findbugs-2.0.2.tar.gz?download)

```
~/impendulo-tools$ wget http://prdownloads.sourceforge.net/findbugs/findbugs-2.0.2.tar.gz?download
... lots of stuff ...
2013-09-20 14:58:47 (316 KB/s) - `findbugs-2.0.2.tar.gz?download' saved [8295637/8295637]
~/impendulo-tools$ tar -xvf findbugs-2.0.2.tar.gz?download 
... lots of stuff ...
~/impendulo-tools$ rm findbugs-2.0.2.tar.gz\?download
```

- Java Pathfinder:

```
~/impendulo-tools$ sudo apt-get install hg
~/impendulo-tools$ hg clone http://babelfish.arc.nasa.gov/hg/jpf/jpf-core
~/impendulo-tools$ cd jpf-core
~/impendulo-tools/jpf-core$ bin/ant test
```

- [JUnit 4](https://github.com/junit-team/junit/wiki/Download-and-Install)

```
~/impendulo-tools$ sudo apt-get install ant
~/impendulo-tools$ mkdir junit
~/impendulo-tools$ cd junit
~/impendulo-tools/junit$ wget http://search.maven.org/remotecontent?filepath=junit/junit/4.11/junit-4.11.jar
... lots of stuff ...
2013-09-20 15:19:05 (98,9 KB/s) - `remotecontent?filepath=junit%2Fjunit%2F4.11%2Fjunit-4.11.jar' saved [245039/245039]
~/impendulo-tools/junit$ mv remotecontent\?filepath=junit%2Fjunit%2F4.11%2Fjunit-4.11.jar junit4.jar
~/impendulo-tools/junit$ wget http://search.maven.org/remotecontent?filepath=org/hamcrest/hamcrest-core/1.3/hamcrest-core-1.3.jar
... lots of stuff ...
2013-09-20 15:21:00 (43,8 KB/s) - `remotecontent?filepath=org%2Fhamcrest%2Fhamcrest-core%2F1.3%2Fhamcrest-core-1.3.jar' saved [45024/45024]
~/impendulo-tools/junit$ mv remotecontent\?filepath=org%2Fhamcrest%2Fhamcrest-core%2F1.3%2Fhamcrest-core-1.3.jar hamcrest-core.jar
```

- [OpenJDK Java Compiler](http://openjdk.java.net/install/)

```
sudo apt-get install openjdk-7-jre
```

Next you need to setup the configuration file `$GOPATH/src/github.com/godfried/impendulo/config/config.json` to point to the correct locations.
It should therefore look something like:
```json
{
"bin": {
	"java": "/usr/bin/java",
	"javac": "/usr/bin/javac", 
	"diff": "/usr/bin/diff"
    },
    "sh":{
	"pmd": "~/impendulo-tools/pmd-bin-5.0.4/bin/run.sh",
	"diff2html": "~/go/src/github.com/godfried/impendulo/scripts/diff2html.sh" 
    },
    "jar": {
	"junit": "~/impendulo-tools/junit/junit4.jar",
	"ant": "/usr/share/java/ant.jar", 
	"ant_junit": "/usr/share/java/ant-junit.jar",
	"findbugs": "~/impendulo-tools/findbugs-2.0.2/lib/findbugs.jar",
	"jpf": "~/impendulo-tools/jpf-core/build/jpf.jar",
	"jpf_run": "~/impendulo-tools/jpf-core/build/RunJPF.jar",
	"checkstyle": "~/impendulo-tools/checkstyle-5.6/checkstyle-5.6-all.jar",
	"gson": "~/go/src/github.com/godfried/impendulo/java/lib/gson-2.2.4.jar"
    },
    "cfg":{
	"checkstyle_cfg": "~/go/src/github.com/godfried/impendulo/config/checkstyle_config.xml",
	"pmd_cfg": "~/go/src/github.com/godfried/impendulo/config/pmd_config.json"
    },
    "dir": {
	"jpf_home": "~/impendulo-tools/jpf-core",
	"jpf_runner": "~/go/src/github.com/godfried/impendulo/java/runner",
	"jpf_finder": "~/go/src/github.com/godfried/impendulo/java/finder",
	"junit_testing": "~/go/src/github.com/godfried/impendulo/java/testing"
    }
}
```

You should now be able to run Impendulo by simply invoking it from the command line:
```
~$ impendulo

```
