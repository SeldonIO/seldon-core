====================
Install Locally
====================

This guide runs through how to set up and install Seldon Core in a Kubernetes cluster running on your local machine. By the end, you'll have Seldon Core up and running and be ready to start deploying machine learning models.

Prerequisites
-----------------

In order to install Seldon Core locally you'll need the following tools:

Kind
^^^^^^^^^^^^^
`Kind <https://kind.sigs.k8s.io/>`_ is a tool for running Kubernetes clusters locally. We'll use it to create a cluster on your machine so that you can install Seldon Core in to it. If don't already have `kind <https://kind.sigs.k8s.io/>`_ installed on your machine, you'll need to follow their installation guide:

* `Install Kind <https://kind.sigs.k8s.io/docs/user/quick-start/#installation>`_ 

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

Set Up Kind
----------------

Once kind is installed on your system you can create a new Kubernetes cluster by running

.. code-block:: bash

    kind create cluster --name seldon

After ``kind`` has created your cluster, you can configure ``kubectl`` to use the cluster by setting the context:

.. code-block:: bash

    kubectl cluster-info --context kind-seldon

From now on, all commands run using ``kubectl`` will be directed at your ``kind`` cluster. 

.. note:: Kind prefixes your cluster names with ``kind-`` so your cluster context is ``kind-seldon`` and not just ``seldon``

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
    
    For custom configuration and more details on installing seldon core with Istio please see the `Istio Ingress </ingress/istio.md>`_ page.

.. tabbed:: Ambassador

    Instructions for Ambassador

    For custom configuration and more details on installing seldon core with Istio please see the `Ambassador Ingress </ingress/ambassador.md>`_ page.

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

Local Port Forwarding
-------------------------------

Because your kubernetes cluster is running locally, we need to forward a port on your local machine to one in the cluster for us to be able to access it externally. You can do this by running:

.. code-block:: bash

    kubectl port-forward -n istio-system svc/istio-ingressgateway 8080:80

This will forward any traffic from port 8080 on your local machine to port 80 inside your cluster.

You have now successfully installed Seldon Core on a local cluster and are ready to `start deploying models </workflow/github-readme.md>`_ as production microservices.