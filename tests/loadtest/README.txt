
Ensure helm installed:
       helm init
Label nodes as role=locust, e.g. if you have done "minikube profile deploy" then:

     kubectl label nodes deploy role=locust

For test deployment run from tests folder:

     make cm_create_deployment 

Then can start default load test as the oauth key/secret matches using helm from helm_scripts folder:

     helm install --name loadtest seldon-core-loadtesting

Remove loadtest

     helm delete loadtest --purge
