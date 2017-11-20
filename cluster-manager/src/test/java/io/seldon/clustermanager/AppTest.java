package io.seldon.clustermanager;

import java.io.IOException;
import java.nio.charset.Charset;
import java.nio.file.Files;
import java.nio.file.Paths;

public class AppTest 
{
	protected String readFile(String path, Charset encoding) 
			  throws IOException 
	 {
		 byte[] encoded = Files.readAllBytes(Paths.get(path));
		 return new String(encoded, encoding);
	 }	
	
	protected ClusterManagerProperites getProps()
	{
		ClusterManagerProperites c = new ClusterManagerProperites();
		c.setEngineContainerPort(9000);
		c.setEngineContainerImageAndVersion("seldonio/engine:0.1.6");
		c.setPuContainerPortBase(8000);
		return c;
	}	
}
