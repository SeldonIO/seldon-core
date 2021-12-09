========================================
Install on Microsoft Azure Cloud
========================================

This guide runs through how to set up and install Seldon Core in a Kubernetes cluster running on Azure Cloud. By the end, you’ll have Seldon Core up and running and be ready to start deploying machine learning models.

Prerequisites
-----------------------------

Azure Cloud CLI
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

You will need the Azure CLI in order to retrieve your cluster authentication credentials. It can also be used to create clusters and other resources for you:

* `Install Azure CLI <https://docs.microsoft.com/en-us/cli/azure/install-azure-cli>`_

Azure Kubernetes Service (AKS) Cluster
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

If you haven't already created a Kubernetes cluster on AKS, you can follow this quickstart guide to get set up with your first cluster:

* `Create AKS Cluster <https://docs.microsoft.com/en-us/azure/aks/tutorial-kubernetes-deploy-cluster?tabs=azure-cli>`_

.. note:: 

    If you are just evaluating Seldon Core and are want to use http rather than https, make sure you select "Enable HTTP application routing" in your networking configuration.

Kubectl
^^^^^^^^^^^^^
`kubectl <https://kubernetes.io/docs/reference/kubectl/overview/>`_ is the Kubernetes command-line tool. It allows you to run commands against Kubernetes clusters, which we'll need to do as part of setting up Seldon Core. 

* `Install kubectl on Linux <https://kubernetes.io/docs/tasks/tools/install-kubectl-linux>`_ 
* `Install kubectl on macOS <https://kubernetes.io/docs/tasks/tools/install-kubectl-macos>`_ 
* `Install kubectl on Windows <https://kubernetes.io/docs/tasks/tools/install-kubectl-windows>`_ 

Helm
^^^^^^^^^^^^^
`Helm <https://helm.sh/>`_ is a package manager that makes it easy to find, share and use software built for Kubernetes. If you don't already have Helm installed locally, you can install it here:

* `Install Helm <https://helm.sh/docs/intro/install/>`_ 

Connect to Your Cluster
------------------------------

You can connect to your cluster by running the following `az` command:

.. code-block:: bash

    az aks get-credentials --resource-group myResourceGroup --name myAKSCluster

This will configure ``kubectl`` to use your Azure kubernetes cluster. Don't forget to replace ``myResourceGroup`` and ``myAKSCluster`` with whatever you called your resource group and cluster are called. If you've forgotten, you can run ``az aks list``.

.. note:: 

    If you get authentication errors while running the command above, try running ``az login`` to check you are correctly logged in.

Install Cluster Ingress
------------------------------

``Ingress`` is a Kubernetes object that provides routing rules for your cluster. It manages the incomming traffic and routes it to the services running inside the cluster.

Seldon Core supports using either `Istio <https://istio.io/>`_ or `Ambassador <https://www.getambassador.io/>`_ to manage incomming traffic. Seldon Core automatically creates the objects and rules required to route traffic to your deployed machine learning models.

.. tabbed:: Istio

    Istio is an open source service mesh. If the term *service mesh* is unfamiliar to you, it's worth reading `a little more about Istio <https://istio.io/latest/about/service-mesh/>`_.

    **Download Istio**

    For Linux and macOS, the easiest way to download Istio is using the following command:

    .. code-block:: bash 

        curl -L https://istio.io/downloadIstio | sh -

    Move to the Istio package directory. For example, if the package is ``istio-1.11.4``:

    .. code-block:: bash

        cd istio-1.11.4

    Add the istioctl client to your path (Linux or macOS):

    .. code-block:: bash

        export PATH=$PWD/bin:$PATH

    **Install Istio**

    Istio provides a command line tool ``istioctl`` to make the installation process easy. The ``demo`` `configuration profile <https://istio.io/latest/docs/setup/additional-setup/config-profiles/>`_ has a good set of defaults that will work on your local cluster.

    .. code-block:: bash

        istioctl install --set profile=demo -y

    The namespace label ``istio-injection=enabled`` instructs Istio to automatically inject proxies alongside anything we deploy in that namespace. We'll set it up for our ``default`` namespace:

    .. code-block:: bash 

        kubectl label namespace default istio-injection=enabled

    **Create Istio Gateway**

    In order for Seldon Core to use Istio's features to manage cluster traffic, we need to create an `Istio Gateway <https://istio.io/latest/docs/tasks/traffic-management/ingress/ingress-control/>`_ by running the following command:

    .. warning:: You will need to copy the entire command from the code block below
    
    .. code-block:: yaml

        kubectl apply -f - << END
        apiVersion: networking.istio.io/v1alpha3
        kind: Gateway
        metadata:
          name: seldon-gateway
          namespace: istio-system
        spec:
          selector:
            istio: ingressgateway # use istio default controller
          servers:
          - port:
              number: 80
              name: http
              protocol: HTTP
            hosts:
            - "*"
        END
    
    For custom configuration and more details on installing seldon core with Istio please see the `Istio Ingress <../ingress/istio.md>`_ page.

.. tabbed:: Ambassador

    `Ambassador <https://www.getambassador.io/>`_ is a Kubernetes ingress controller and API gateway. It routes incomming traffic to the underlying kubernetes workloads through configuration. 

    **Install Ambassador**

    .. note::
        Seldon Core currently only supports the Ambassador V1 APIs. The installation instructions below will install the latest v1 version of emissary ingress.


    First add the datawire helm repository:

    .. code-block:: bash

        helm repo add datawire https://www.getambassador.io
        helm repo update

    Run the following `helm` command to install Ambassador on your GKE cluster:

    .. code-block:: bash

        helm install ambassador datawire/ambassador --set enableAES=false --namespace ambassador --create-namespace
        kubectl rollout status -n ambassador deployment/ambassador -w
        
    Ambassador is now ready to use. For custom configuration and more details on installing seldon core with Ambassador please see the `Ambassador Ingress <../ingress/ambassador.md>`_ page.

Install Seldon Core
----------------------------

Before we install Seldon Core, we'll create a new namespace ``seldon-system`` for the operator to run in:

.. code:: bash

    kubectl create namespace seldon-system

We're now ready to install Seldon Core in our cluster. Run the following command for your choice of Ingress:

.. tabbed:: Istio

    .. code:: bash

        helm install seldon-core seldon-core-operator \
            --repo https://storage.googleapis.com/seldon-charts \
            --set usageMetrics.enabled=true \
            --set istio.enabled=true \
            --namespace seldon-system

.. tabbed:: Ambassador

    .. code:: bash

        helm install seldon-core seldon-core-operator \
            --repo https://storage.googleapis.com/seldon-charts \
            --set usageMetrics.enabled=true \
            --set ambassador.enabled=true \
            --namespace seldon-system

You can check that your Seldon Controller is running by doing:

.. code-block:: bash

    kubectl get pods -n seldon-system

You should see a ``seldon-controller-manager`` pod with ``STATUS=Running``.

Accessing your models
-------------------------

Congratulations! Seldon Core is now fully installed and operational. Before you move on to deploying models, make a note of your cluster IP and port:

.. tabbed:: Istio

    .. code-block:: bash 

        export INGRESS_HOST=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
        export INGRESS_PORT=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="http2")].port}')
        export INGRESS_URL=$INGRESS_HOST:$INGRESS_PORT
        echo $INGRESS_URL

    This is the public address you will use to access models running in your cluster.

.. tabbed:: Ambassador

    .. code-block:: bash

        export INGRESS_HOST=$(kubectl -n ambassador get service ambassador -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
        export INGRESS_PORT=$(kubectl -n ambassador get service ambassador -o jsonpath='{.spec.ports[?(@.name=="http")].port}')
        export INGRESS_URL=$INGRESS_HOST:$INGRESS_PORT
        echo $INGRESS_URL

    This is the public address you will use to access models running in your cluster.

You are now ready to `deploy models to your cluster <../workflow/github-readme.md>`_.
