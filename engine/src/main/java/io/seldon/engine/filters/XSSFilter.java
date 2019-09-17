package io.seldon.engine.filters;

import java.io.IOException;

import org.springframework.web.filter.OncePerRequestFilter;
import org.springframework.stereotype.Component;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import javax.servlet.FilterChain;
import javax.servlet.ServletException;

@Component
public class XSSFilter extends OncePerRequestFilter {
    @Override
    protected void doFilterInternal(
        HttpServletRequest request,
        HttpServletResponse response, 
        FilterChain filterChain)  throws ServletException, IOException {
      // Add nosniff option to avoid content sniffing by the browser
      response.addHeader("X-Content-Type-Options", "nosniff");

      filterChain.doFilter(request, response);
    }
}
