Install Seldon-Core
===================

Pre-requisites:
---------------

-  Kubernetes cluster >= 1.23

   -  For Openshift it requires version >= 4.10

-  Installer method

   -  Helm version >= 3.0
   -  Kustomize version >= 0.1.0

-  Ingress

   -  Istio : we recommend >= 1.16
   -  Ambassador v1 and v2


Kubernetes Compatibility Matrix
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Seldon Core 1.16 bumps minimum Kubernetes version to 1.23.
This is because as part of making Seldon Core compatible with Kubernetes 1.25 we moved from autoscaling/v2beta1 apiVersion of HorizontalPodAutoscaler to autoscaling/v2 (see this `PR <https://github.com/SeldonIO/seldon-core/pull/4172>`__ for further details).

Following table provides a summary of Seldon Core / Kubernetes version compatibility for recent version of Seldon Core.

+----------------------------+------+------+------+------+------+------+------+
| Core Version \ K8s Version | 1.21 | 1.22 | 1.23 | 1.24 | 1.25 | 1.26 | 1.27 |
+============================+======+======+======+======+======+======+======+
| 1.11                       | ✓    |      |      |      |      |      |      |
+----------------------------+------+------+------+------+------+------+------+
| 1.12                       | ✓    | ✓    | ✓    | ✓    |      |      |      |
+----------------------------+------+------+------+------+------+------+------+
| 1.13                       | ✓    | ✓    | ✓    | ✓    |      |      |      |
+----------------------------+------+------+------+------+------+------+------+
| 1.14                       | ✓    | ✓    | ✓    | ✓    |      |      |      |
+----------------------------+------+------+------+------+------+------+------+
| 1.15                       | ✓    | ✓    | ✓    | ✓    |      |      |      |
+----------------------------+------+------+------+------+------+------+------+
| 1.16                       |      |      | ✓    | ✓    | ✓    | ✓    | ✓    |
+----------------------------+------+------+------+------+------+------+------+
| 1.17                       |      |      | ✓    | ✓    | ✓    | ✓    | ✓    |
+----------------------------+------+------+------+------+------+------+------+
| 1.18                       |      |      | ✓    | ✓    | ✓    | ✓    | ✓    |
+----------------------------+------+------+------+------+------+------+------+

It is always recommended to first upgrade Seldon Core to the latest supported version on your Kubernetes cluster and then upgrade the Kubernetes cluster.

Running older versions of Seldon Core?
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Make sure you read the `"Upgrading Seldon Core
Guide" <../reference/upgrading.md>`__

-  **Seldon Core will stop supporting versions prior to 1.0 so make sure
   you upgrade.**
-  If you are running an older version of Seldon Core, and will be
   upgading it please make sure you read the `Upgrading Seldon Core
   docs <../reference/upgrading.md>`__ to understand breaking changes
   and best practices for upgrading.
-  Please see `Migrating from Helm v2 to Helm
   v3 <https://helm.sh/docs/topics/v2_v3_migration/>`__ if you
   are already running Seldon Core using Helm v2 and wish to upgrade.

Install Seldon Core with Helm
-----------------------------

First `install Helm 3.x <https://docs.helm.sh/docs/intro/install/>`__.
When helm is installed you can deploy the seldon controller to manage
your Seldon Deployment graphs.

If you want to provide advanced parameters with your installation you
can check the full `Seldon Core Helm Chart
Reference <../reference/helm.html>`__.

The namespace ``seldon-system`` is preferred, so we can create it:

.. code:: bash

    kubectl create namespace seldon-system

Now we can install Seldon Core in the ``seldon-system`` namespace.

.. tab-set::

    .. tab-item:: Istio

        .. code:: bash

            helm install seldon-core seldon-core-operator \
                --repo https://storage.googleapis.com/seldon-charts \
                --set usageMetrics.enabled=true \
                --set istio.enabled=true \
                --namespace seldon-system

    .. tab-item:: Ambassador

        .. code:: bash

            helm install seldon-core seldon-core-operator \
                --repo https://storage.googleapis.com/seldon-charts \
                --set usageMetrics.enabled=true \
                --set ambassador.enabled=true \
                --namespace seldon-system


For full instructions on installation with Istio and Ambassador read the
following pages:

* `Ingress with Istio <../ingress/istio.md>`__
* `Ingress with Ambassador <../ingress/ambassador.md>`__

Install a specific version
~~~~~~~~~~~~~~~~~~~~~~~~~~

In order to install a specific version you can do so by running the same
command above with the ``--version`` flag, followed by the version you
want to run.

Install a SNAPSHOT version
~~~~~~~~~~~~~~~~~~~~~~~~~~

Whenever a new PR was merged to master, we have set up our CI to build a
"SNAPSHOT" version, which would contain the Docker images for that
specific development / master-branch code. Whilst the images are pushed
under SNAPSHOT, they also create a new "dated" SNAPSHOT version entry,
which pushes images with the tag
``"<next-version>-SNAPSHOT_<timestamp>"``. A new branch is also created
with the name ``"v<next-version>-SNAPSHOT_<timestamp>"``, which contains
the respective helm charts, and allows for the specific version (as
outlined by the version in ``version.txt``) to be installed.

This means that you can try out a dev version of master if you want to
try a specific feature before it's released.

For this you would be able to clone the repository, and then checkout
the relevant SNAPSHOT branch.

Once you have done that you can install seldon-core using the following
command:

.. code:: bash

    helm install helm-charts/seldon-core-operator seldon-core-operator

In this case ``helm-charts/seldon-core-operator`` is the folder within
the repository that contains the charts.

Install with cert-manager
~~~~~~~~~~~~~~~~~~~~~~~~~

You can follow `the cert manager documentation to install
it <https://cert-manager.io/docs/installation/>`__.

You can then install seldon-core with:

.. code:: bash

    helm install seldon-core seldon-core-operator \
        --repo https://storage.googleapis.com/seldon-charts \
        --set usageMetrics.enabled=true \
        --namespace seldon-system \
        --set certManager.enabled=true

Seldon Core Kustomize Install
-----------------------------

The `Kustomize <https://github.com/kubernetes-sigs/kustomize>`__
installation can be found in the ``/operator/config`` folder of the
repo. You should copy this template to your own kustomize location for
editing.

To use the template directly, there is a Makefile which has a set of
useful commands:

For kubernetes clusters of version higher than 1.15, make sure you
`comment the patch\_object\_selector
here <https://github.com/SeldonIO/seldon-core/blob/master/operator/config/webhook/kustomization.yaml#L8>`__.

Install cert-manager

.. code:: bash

    make install-cert-manager

Install Seldon using cert-manager to provide certificates.

.. code:: bash

    make deploy

Install Seldon with provided certificates in ``config/cert/``

.. code:: bash

    make deploy-cert

Other Options
-------------

Install Production Integrations
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Now that you have Seldon Core installed, you can set it up with:

Install with Kubeflow
^^^^^^^^^^^^^^^^^^^^^

-  `Install Seldon as part of
   Kubeflow. <https://www.kubeflow.org/docs/external-add-ons/serving/seldon/>`__

GCP MarketPlace
^^^^^^^^^^^^^^^

If you have a Google Cloud Platform account you can install via the `GCP
Marketplace <https://console.cloud.google.com/marketplace/details/seldon-portal/seldon-core>`__.

OpenShift
^^^^^^^^^

You can install Seldon Core via OperatorHub on the OpenShift console UI.

OperatorHub
^^^^^^^^^^^

You can install Seldon Core from `Operator
Hub <https://operatorhub.io/operator/seldon-operator>`__.

Upgrading from Previous Versions
--------------------------------

See our `upgrading notes <../reference/upgrading.md>`__

Advanced Usage
--------------

Install Seldon Core in a single namespace (version >=1.0)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

**You will need a k8s cluster >= 1.15**

Helm
^^^^

You can install the Seldon Core Operator so it only manages resources in
its namespace. An example to install in a namespace ``seldon-ns1`` is
shown below:

.. code:: bash

    kubectl create namespace seldon-ns1
    kubectl label namespace seldon-ns1 seldon.io/controller-id=seldon-ns1

We label the namespace with ``seldon.io/controller-id=<namespace>`` to
ensure if there is a clusterwide Seldon Core Operator that it should
ignore resources for this namespace.

Install the Operator into the namespace:

.. code:: bash

    helm install seldon-namespaced seldon-core-operator  --repo https://storage.googleapis.com/seldon-charts  \
        --set singleNamespace=true \
        --set image.pullPolicy=IfNotPresent \
        --set usageMetrics.enabled=false \
        --set crd.create=true \
        --namespace seldon-ns1

We set ``crd.create=true`` to create the CRD. If you are installing a
Seldon Core Operator after you have installed a previous Seldon Core
Operator on the same cluster you will need to set ``crd.create=false``.

Kustomize
^^^^^^^^^

An example install is provided in the Makefile in the Operator folder:

.. code:: bash

    make deploy-namespaced1

See the `multiple server example
notebook <../examples/multiple_operators.html>`__.

Label focused Seldon Core Operator (version >=1.0)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

**You will need a k8s cluster >= 1.15**

You can install the Seldon Core Operator so it manages only
SeldonDeployments with the label ``seldon.io/controller-id`` where the
value of the label matches the controller-id of the running operator. An
example for a namespace ``seldon-id1`` is shown below:

Helm
^^^^

.. code:: bash

    kubectl create namespace seldon-id1

To install the Operator run:

.. code:: bash

    helm install seldon-controllerid seldon-core-operator  --repo https://storage.googleapis.com/seldon-charts  \
        --set singleNamespace=false \
        --set image.pullPolicy=IfNotPresent \
        --set usageMetrics.enabled=false \
        --set crd.create=true \
        --set controllerId=seldon-id1 \
        --namespace seldon-id1

We set ``crd.create=true`` to create the CRD. If you are installing a
Seldon Core Operator after you have installed a previous Seldon Core
Operator on the same cluster you will need to set ``crd.create=false``.

For kustomize you will need to `uncomment the patch\_object\_selector
here <https://github.com/SeldonIO/seldon-core/blob/master/operator/config/webhook/kustomization.yaml>`__

Kustomize
^^^^^^^^^

An example install is provided in the Makefile in the Operator folder:

.. code:: bash

    make deploy-controllerid

See the `multiple server example
notebook <../examples/multiple_operators.html>`__.

Install behind a proxy
~~~~~~~~~~~~~~~~~~~~~~

When your kubernetes cluster is behind a proxy, the ``kube-apiserver``
typically inherits the system proxy variables. This can block the
``kube-apiserver`` from reaching the webhooks needed to create Seldon
resources.

You could see this error:

.. code:: bash

    Internal error occurred: failed calling webhook "v1.vseldondeployment.kb.io": Post https://seldon-webhook-service.seldon-system.svc:443/validate-machinelearning-seldon-io-v1-seldondeployment?timeout=30s: Service Unavailable

To fix this, ensure the ``no_proxy`` environment variable for the
``kube-apiserver`` includes ``.svc,.svc.cluster.local``. See `this
Github Issue
Comment <https://github.com/cert-manager/cert-manager/issues/2640#issuecomment-601872165>`__
for reference. As described there, the error could also occur for the
``cert-manager-webhook``.
