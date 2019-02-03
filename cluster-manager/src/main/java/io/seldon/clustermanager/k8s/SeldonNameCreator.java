package io.seldon.clustermanager.k8s;

import org.apache.commons.codec.digest.DigestUtils;

import io.kubernetes.client.proto.V1;
import io.seldon.protos.DeploymentProtos.PredictorSpec;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

public class SeldonNameCreator {

	private String hash(String key)  {
	    return DigestUtils.md5Hex(key).toLowerCase().substring(0, 7);
	}
	
	private String createContainerHash(V1.PodTemplateSpec spec)
	{
		StringBuffer sb = new StringBuffer();
		for(V1.Container c : spec.getSpec().getContainersList())
		{
			sb.append(c.getName());
			sb.append(":");
			sb.append(c.getImage());
			sb.append(";");
		}
		return hash(sb.toString());
	}
	
	public String getSeldonDeploymentName(SeldonDeployment dep,PredictorSpec pred,V1.PodTemplateSpec spec) {
		if (spec.getMetadata().hasName())
			return spec.getMetadata().getName();
		else
		{
			String svcName =  dep.getSpec().getName() + "-" + pred.getName()+"-"+createContainerHash(spec);
			if (svcName.length() > 63)
				return "seldon-"+hash(svcName);
			else
				return svcName;
		}
	}
	
	protected static String cleanContainerImageName(String name)
	{
		return name.toLowerCase().replaceAll("[^-a-z0-9]", "-");
	}
	
	public String getServiceOrchestratorName(SeldonDeployment dep,PredictorSpec pred)
	{
		String svcName =  dep.getSpec().getName() + "-" + pred.getName()+"-svc-orch";
		if (svcName.length() > 63)
			return "seldon-"+hash(svcName);
		else
			return svcName;
	}
	

	public String getSeldonServiceName(SeldonDeployment dep,PredictorSpec pred,V1.Container container) {
		String containerImage = cleanContainerImageName(container.getImage());
		String svcName =  dep.getSpec().getName() + "-" + pred.getName()+"-"+container.getName()+"-"+containerImage;
		if (svcName.length() > 63)
		{
			svcName = "seldon-"+containerImage+"-"+hash(svcName);
			if (svcName.length() > 63)
				return "seldon-"+hash(svcName);
			else
				return svcName;
		}
		else
			return svcName;
	}
	
	public String getSeldonId(SeldonDeployment mlDep)
	{
		return mlDep.getSpec().getName() + "-" + mlDep.getMetadata().getName();
	}

/*	
	public String getSeldonServiceName(SeldonDeployment dep,PredictorSpec pred,String containerName) {
		String svcName =  dep.getSpec().getName() + "-" + pred.getName()+"-"+containerName;
		if (svcName.length() > 63)
			return "seldon-"+hash(svcName);
		else
			return svcName;
	}
	*/
}
