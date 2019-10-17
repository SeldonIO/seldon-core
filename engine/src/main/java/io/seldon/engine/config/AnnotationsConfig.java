package io.seldon.engine.config;

import java.io.BufferedReader;
import java.io.File;
import java.io.IOException;
import java.nio.charset.Charset;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

/**
 * @author clive Utility class to load annotations from kubernetes file for use by other components
 *     as config during startup
 */
@Component
public class AnnotationsConfig {
  protected static Logger logger = LoggerFactory.getLogger(AnnotationsConfig.class.getName());
  final String ANNOTATIONS_FILE = "/etc/podinfo/annotations";

  final Map<String, String> annotations = new ConcurrentHashMap<>();

  String readFile(String path, Charset encoding) throws IOException {
    byte[] encoded = Files.readAllBytes(Paths.get(path));
    return new String(encoded, encoding);
  }

  public AnnotationsConfig() throws IOException {
    loadAnnotations();
  }

  private void processAnnotation(String line) {
    final String[] parts = line.split("=");
    if (parts.length == 2) {
      final String value =
          parts[1].substring(1, parts[1].length() - 1); // remove start and end quote
      annotations.put(parts[0], value);
    } else logger.warn("Failed to parse annotation {}", line);
  }

  private void loadAnnotations() {
    try {
      File f = new File(ANNOTATIONS_FILE);
      if (f.exists() && !f.isDirectory()) {
        try (BufferedReader r =
            Files.newBufferedReader(Paths.get(ANNOTATIONS_FILE), StandardCharsets.UTF_8)) {
          r.lines().forEach(this::processAnnotation);
        }
      }
    } catch (IOException e) {
      logger.error("Failed to load annotations file {}", ANNOTATIONS_FILE, e);
    }
    logger.info("Annotations {}", annotations);
  }

  public boolean has(String key) {
    return annotations.containsKey(key);
  }

  public String get(String key) {
    return annotations.get(key);
  }
}
