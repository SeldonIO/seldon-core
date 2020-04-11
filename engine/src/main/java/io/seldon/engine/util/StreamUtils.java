package io.seldon.engine.util;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;

public final class StreamUtils {
    private static final int EOF = -1;
    private static final int BUFFER_SIZE = 1024 * 4;

    public static byte[] toByteArray(InputStream inputStream) throws IOException {
        final ByteArrayOutputStream outputStream = new ByteArrayOutputStream();

        copyStream(inputStream, outputStream);

        return outputStream.toByteArray();
    }

    public static void copyStream(final InputStream input, final OutputStream output)
            throws IOException {
        final byte[] buffer = new byte[BUFFER_SIZE];
        int n;

        while (EOF != (n = input.read(buffer))) {
            output.write(buffer, 0, n);
        }
    }
}
