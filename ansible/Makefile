install_deps_stable:
	pip install -r requirements.txt
	ansible-galaxy install -r requirements.yml

install_deps_edge:
	pip install ansible openshift kubernetes docker
	ansible-galaxy collection install kubernetes.core --force
