package io.seldon.engine.service;

public class PredictionServiceReturnStatus {

	public int code;
	public String info;
	public String reason;
	public int success;
	
	public PredictionServiceReturnStatus()
	{
		this.code = 0;
		this.success = 0;
		this.info = "success";
	}
	
}
