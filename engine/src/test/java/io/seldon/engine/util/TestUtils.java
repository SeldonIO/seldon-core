package io.seldon.engine.util;

import java.io.IOException;
import java.nio.charset.Charset;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.Base64;

public class TestUtils {

  public static byte[] readFileBytes(String path) throws IOException {
    return Files.readAllBytes(Paths.get(path));
  }

  public static String readFile(String path, Charset encoding) throws IOException {
    return new String(readFileBytes(path), encoding);
  }

  public static String readFileBase64(String path) throws IOException {
    return Base64.getEncoder().encodeToString(readFileBytes(path));
  }
}
