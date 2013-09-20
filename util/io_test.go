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

package util

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCopier(t *testing.T) {
	realDest := filepath.Join(os.TempDir(), "copy")
	fakeDest := "/dummy"
	realSrcFile := "data.txt"
	realSrcDir := "folder/"
	fakeSrc := "dummy"
	tests := map[string]bool{fakeDest + "," + fakeSrc: false,
		fakeDest + "," + realSrcFile: false, fakeDest + "," + realSrcDir: false,
		realDest + "," + fakeSrc: false, realDest + "," + realSrcDir: true,
		realDest + "," + realSrcFile: true,
	}
	os.Mkdir(realDest, os.ModeDir|os.ModePerm)
	os.Mkdir(realSrcDir, os.ModeDir|os.ModePerm)
	os.Create(realSrcFile)
	for test, valid := range tests {
		err := tcopy(test, valid)
		if err != nil {
			t.Error(err)
		}

	}
	os.RemoveAll(realSrcFile)
	os.RemoveAll(realDest)
	os.RemoveAll(realSrcDir)
}

func tcopy(destSrc string, valid bool) error {
	vals := strings.Split(destSrc, ",")
	err := Copy(vals[0]+"/"+vals[1], vals[1])
	if valid {
		return err
	} else if err == nil {
		return fmt.Errorf("Expected error for test " + destSrc)
	} else {
		return nil
	}
}

func TestExists(t *testing.T) {
	if !Exists("io_test.go") {
		t.Error("Should return true.")
	}
	if Exists("io_tester.go") {
		t.Error("Should return false.")
	}
}

func TestSaveFile(t *testing.T) {
	tests := map[string]bool{"/fake/fake/fake": false,
		os.TempDir(): false, filepath.Join(os.TempDir(), "real"): true,
	}
	for test, valid := range tests {
		err := tsave(test, valid)
		if err != nil {
			t.Error(err)
		}

	}
	err := SaveFile(filepath.Join(os.TempDir(), "real"), nil)
	if err != nil {
		t.Error(err)
	}
}

func tsave(name string, valid bool) error {
	err := SaveFile(name, []byte{})
	if valid {
		return err
	} else if err == nil {
		return fmt.Errorf("Expected error for test " + name)
	} else {
		return nil
	}
}

func TestReadBytes(t *testing.T) {
	orig := []byte("bytes")
	buff := bytes.NewBuffer(orig)
	ret := ReadBytes(buff)
	if !bytes.Equal(orig, ret) {
		t.Error(errors.New("Bytes not equal"))
	}

	ret = ReadBytes(nil)
	if !bytes.Equal(ret, []byte{}) {
		t.Error("Expected empty []byte.")
	}
	ret = ReadBytes(new(ErrorReader))
	if !bytes.Equal(ret, []byte{}) {
		t.Error("Expected empty []byte.")
	}
}

type ErrorReader struct{}

func (this *ErrorReader) Read(b []byte) (int, error) {
	return 0, fmt.Errorf("I just give errors")
}

func TestReadData(t *testing.T) {
	tests := map[string][]byte{file1: []byte(file1), file2 + EOT: []byte(file2), file3: []byte(file3)}
	for k, v := range tests {
		data, err := ReadData(strings.NewReader(k))
		if err != nil {
			t.Error(err)
		}
		if !bytes.Equal(v, data) {
			t.Error(fmt.Sprintf("Expected data %q but got %q.", v, data))
		}
	}
	_, err := ReadData(nil)
	if err == nil {
		t.Error("Expected error for nil reader.")
	}
	_, err = ReadData(new(ErrorReader))
	if err == nil {
		t.Error("Expected error for error reader.")
	}
}

func TestGetPackage(t *testing.T) {
	tests := map[string]string{"package pkg.p.s;": "pkg.p.s", file1: "java.io", file2: "za.ac.sun.cs.intlola.file", file3: ""}
	for k, v := range tests {
		pkg := GetPackage(strings.NewReader(k))
		if v != pkg {
			t.Error(fmt.Sprintf("Expected package %q but got %q.", v, pkg))
		}
	}
}

var file1 = `//
// Copyright (C) 2006 United States Government as represented by the
// Administrator of the National Aeronautics and Space Administration
// (NASA).  All Rights Reserved.
// 
// This software is distributed under the NASA Open Source Agreement
// (NOSA), version 1.3.  The NOSA has been approved by the Open Source
// Initiative.  See the file NOSA-1.3-JPF at the top of the distribution
// directory tree for the complete NOSA document.
// 
// THE SUBJECT SOFTWARE IS PROVIDED "AS IS" WITHOUT ANY WARRANTY OF ANY
// KIND, EITHER EXPRESSED, IMPLIED, OR STATUTORY, INCLUDING, BUT NOT
// LIMITED TO, ANY WARRANTY THAT THE SUBJECT SOFTWARE WILL CONFORM TO
// SPECIFICATIONS, ANY IMPLIED WARRANTIES OF MERCHANTABILITY, FITNESS FOR
// A PARTICULAR PURPOSE, OR FREEDOM FROM INFRINGEMENT, ANY WARRANTY THAT
// THE SUBJECT SOFTWARE WILL BE ERROR FREE, OR ANY WARRANTY THAT
// DOCUMENTATION, IF PROVIDED, WILL CONFORM TO THE SUBJECT SOFTWARE.
//
package java.io;

import gov.nasa.jpf.annotation.FilterField;

import java.net.URI;
import java.net.URISyntaxException;
import java.net.URL;


/**
 * MJI model class for java.io.File
 *
 * NOTE - a number of methods are only stubbed out here to make Eclipse compile
 * JPF code that uses java.io.File (there is no way to tell Eclipse to exclude the
 * model classes from ths build-path)
 *
 * @author Owen O'Malley
 */
public class File
{
  public static final String separator = System.getProperty("file.separator");
  public static final char separatorChar = separator.charAt(0);
  public static final String pathSeparator = System.getProperty("path.separator");
  public static final char pathSeparatorChar = pathSeparator.charAt(0);

  @FilterField int id; // link to the real File object
  private String filename;

  public File(String filename) {
    if (filename == null){
      throw new NullPointerException();
    }
    
    this.filename = filename;
  }

  public File (String parent, String child) {
  	filename = parent + separator + child;
  }
  
  public File (File parent, String child) {
    filename = parent.filename + separator + child;
  }
  
  public File(java.net.URI uri) { throw new UnsupportedOperationException(); }
  
  public String getName() {
    int idx = filename.lastIndexOf(separatorChar);
    if (idx >= 0){
      return filename.substring(idx+1);
    } else {
      return filename;
    }
  }

  public String getParent() {
    int idx = filename.lastIndexOf(separatorChar);
    if (idx >= 0){
      return filename.substring(0,idx);
    } else {
      return null;
    }
  }
  
  public int compareTo(File that) {
    return this.filename.compareTo(that.filename);
  }
  
  public boolean equals(Object o) {
    if (o instanceof File){
      File otherFile = (File) o;
      return filename.equals(otherFile.filename);
    } else {
      return false;
    }
  }
  
  public int hashCode() {
    return filename.hashCode();
  }
  
  public String toString()  {
    return filename;
  }
  
  
  //--- native peer intercepted (hopefully)
  
  int getPrefixLength() { return 0; }
  public native File getParentFile();
  
  public String getPath() {
    return filename;
  }

  public native boolean isAbsolute();
  public native String getAbsolutePath();
  public native File getAbsoluteFile();
  public native String getCanonicalPath() throws java.io.IOException;

  public native File getCanonicalFile() throws java.io.IOException;

  private native String getURLSpec();
  public java.net.URL toURL() throws java.net.MalformedURLException {
    return new URL(getURLSpec());
  }

  private native String getURISpec();
  public java.net.URI toURI() {
    try {
      return new URI(getURISpec());
    } catch (URISyntaxException x){
      return null;
    }
  }

  public native boolean canRead();
  public native boolean canWrite();
  public native boolean exists();
  public boolean isDirectory() { return false; }
  public boolean isFile() { return false; }
  public boolean isHidden() { return false; }
  public long lastModified() { return -1L; }
  public long length() { return -1; }
  public native boolean createNewFile() throws java.io.IOException;
  public boolean delete()  { return false; }
  public void deleteOnExit() {}
  public String[] list()  { return null; }
  public String[] list(FilenameFilter fnf)  { return null; }
  public File[] listFiles()  { return null; }
  public File[] listFiles(FilenameFilter fnf)  { return null; }
  public File[] listFiles(FileFilter ff)  { return null; }
  public boolean mkdir()  { return false; }
  public boolean mkdirs() { return false; }
  public boolean renameTo(File f)  { return false; }
  public boolean setLastModified(long t)  { return false; }
  public boolean setReadOnly()  { return false; }
  
  public static native File[] listRoots();
  
  public static File createTempFile(String prefix, String suffix, File dir) throws IOException  {
    if (prefix == null){
      throw new NullPointerException();
    }
    
    String tmpDir;
    if (dir == null){
      tmpDir = System.getProperty("java.io.tmpdir");
      if (tmpDir == null){
        tmpDir = ".";
      }
      if (tmpDir.charAt(tmpDir.length()-1) != separatorChar){
        tmpDir += separatorChar;
      }
      
      if (suffix == null){
        suffix = ".tmp";
      }
    } else {
      tmpDir = dir.getPath();
    }
    
    return new File(tmpDir + prefix + suffix);
  }
  
  public static File createTempFile(String prefix, String suffix) throws IOException  {
    return createTempFile(prefix, suffix, null);
  }
}`

var file2 = `package za.ac.sun.cs.intlola.file;

import com.google.gson.JsonObject;

public class ArchiveFile implements IntlolaFile {
	private final String	path;

	public ArchiveFile(final String path) {
		this.path = path;
	}

	@Override
	public String getPath() {
		return path;
	}

	@Override
	public boolean hasContents() {
		return true;
	}

	@Override
	public JsonObject toJSON() {
		final JsonObject ret = new JsonObject();
		ret.addProperty(Const.TYPE, Const.ARCHIVE);
		ret.addProperty(Const.FTYPE, Const.ZIP);
		return ret;
	}

}`
var file3 = `

import com.google.gson.JsonObject;

public class ArchiveFile implements IntlolaFile {
	private final String	path;

	public ArchiveFile(final String path) {
		this.path = path;
	}

	@Override
	public String getPath() {
		return path;
	}

	@Override
	public boolean hasContents() {
		return true;
	}

	@Override
	public JsonObject toJSON() {
		final JsonObject ret = new JsonObject();
		ret.addProperty(Const.TYPE, Const.ARCHIVE);
		ret.addProperty(Const.FTYPE, Const.ZIP);
		return ret;
	}

}`
