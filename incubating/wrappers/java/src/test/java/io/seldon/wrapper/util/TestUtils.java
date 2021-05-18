package io.seldon.wrapper.util;

import java.io.*;
import java.nio.charset.Charset;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Paths;

public class TestUtils {

  private static final ClassLoader classLoader = TestUtils.class.getClassLoader();

  public static String readFile(String path, Charset encoding) throws IOException {
    byte[] encoded = Files.readAllBytes(Paths.get(path));
    return new String(encoded, encoding);
  }

  /**
   * Will load file from either an absolute path of a relative path from "target/test-classes"
   * @param file  file path (ex: "requests/jsonData.json", "/dev/null")
   * @return
   */
  public static String readFileFromAbsolutePathOrResources(String file) {
    try {
      InputStream is = getInputStreamFromAbsolutePathOrResources(file, classLoader);
      byte[] bytes = is.readAllBytes();
      return new String(bytes, StandardCharsets.UTF_8);
    } catch(Throwable t) {
      System.out.println(t);
      t.printStackTrace();
      // nothing
    }
    return null;
  }

  public static InputStream getInputStreamFromAbsolutePathOrResources(String file, ClassLoader classLoader) {
    InputStream is = null;

    // try loading assuming an absolute path
    try {
      is = new FileInputStream(file);
    } catch ( FileNotFoundException fne ) {
      // Nothing
    }
    if( is == null ) {
      is = classLoader.getResourceAsStream(file);
    }

    return is;
  }
}
