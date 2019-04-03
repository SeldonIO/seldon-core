package io.seldon.clustermanager.k8s.tasks;

/**
 * Unique key for a Seldon Deployment
 * 
 * @author clive
 *
 */
public class SeldonDeploymentTaskKey  {
	
	private String name;
	private String version;
	private String namespace;
	public SeldonDeploymentTaskKey(String name, String version, String namespace) {
		super();
		this.name = name;
		this.version = version;
		this.namespace = namespace;
	}
	@Override
	public int hashCode() {
		final int prime = 31;
		int result = 1;
		result = prime * result + ((name == null) ? 0 : name.hashCode());
		result = prime * result + ((namespace == null) ? 0 : namespace.hashCode());
		result = prime * result + ((version == null) ? 0 : version.hashCode());
		return result;
	}
	@Override
	public boolean equals(Object obj) {
		if (this == obj)
			return true;
		if (obj == null)
			return false;
		if (getClass() != obj.getClass())
			return false;
		SeldonDeploymentTaskKey other = (SeldonDeploymentTaskKey) obj;
		if (name == null) {
			if (other.name != null)
				return false;
		} else if (!name.equals(other.name))
			return false;
		if (namespace == null) {
			if (other.namespace != null)
				return false;
		} else if (!namespace.equals(other.namespace))
			return false;
		if (version == null) {
			if (other.version != null)
				return false;
		} else if (!version.equals(other.version))
			return false;
		return true;
	}
	

}
