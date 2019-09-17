package io.seldon.engine;

import java.io.IOException;

import org.junit.Assert;
import org.junit.Test;

import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;

import javax.servlet.ServletException;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import javax.servlet.FilterChain;

public class TestXSSFilter {
  @Test
  public void testSecurityHeaders() throws ServletException, IOException {
    HttpServletRequest request = mock(HttpServletRequest.class);
    HttpServletResponse response = mock(HttpServletResponse.class);
    FilterChain chain = mock(FilterChain.class);

    XSSFilter filter = new XSSFilter();
    filter.doFilter(request, response, chain);

    verify(response).addHeader("X-Content-Type-Options", "nosniff");
    verify(chain).doFilter(request, response);
  }
}
