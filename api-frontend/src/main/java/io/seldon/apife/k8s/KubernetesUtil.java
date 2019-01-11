package io.seldon.apife.k8s;

import org.apache.commons.lang3.StringUtils;

import io.seldon.protos.DeploymentProtos.SeldonDeployment;

public class KubernetesUtil {
	
	public String getNamespace(SeldonDeployment d)
	 {
		 if (StringUtils.isEmpty(d.getMetadata().getNamespace()))
			 return "default";
		 else
			 return d.getMetadata().getNamespace();
	 }

}
