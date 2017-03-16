package io.seldon.engine;

import java.util.ArrayList;
import java.util.List;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.ControllerAdvice;
import org.springframework.web.bind.annotation.ExceptionHandler;

import io.seldon.engine.exception.APIException;


@ControllerAdvice
public class ExceptionControllerAdvice {

	@ExceptionHandler(APIException.class)
	public ResponseEntity<ErrorLog> handleUnauthorizedException(APIException exception) {

		ErrorLog e = new ErrorLog(System.currentTimeMillis(),"API Exception",exception.getHttpResponse(),exception.getError_msg());
		return new ResponseEntity<ErrorLog>(e,HttpStatus.valueOf(exception.getHttpResponse())); 

	}

	public static class ErrorLog
	{
		private long timestamp;
		private int status;
		private String error;
		private List<String> codes;
		public ErrorLog(long timestamp, String error, int status, List<String> codes) {
			super();
			this.timestamp = timestamp;
			this.status = status;
			this.codes = codes;
			this.error = error;
		}
		public ErrorLog(long timestamp, String error, int status, String code) {
			super();
			this.timestamp = timestamp;
			this.status = status;
			this.codes = new ArrayList<>();
			this.codes.add(code);
			this.error = error;
		}
		public long getTimestamp() {
			return timestamp;
		}
		public void setTimestamp(long timestamp) {
			this.timestamp = timestamp;
		}
		public int getStatus() {
			return status;
		}
		public void setStatus(int status) {
			this.status = status;
		}
		public List<String> getCodes() {
			return codes;
		}
		public void setCodes(List<String> codes) {
			this.codes = codes;
		}
		public void setError(String error){
			this.error = error;
		}
		public String getError(){
			return error;
		}
		
		
	}
}