local k = import 'k.libsonnet';
local deployment = k.extensions.v1beta1.deployment;

local podTemplateValidation = import 'pod-template-spec-validation.json';

{
  parts(namespace):: {

  apife(apifeImage)::
      baseApife(apifeImage),

  apifeWithRbac(apifeImage)::
      baseApife(apifeImage) +
      deployment.mixin.spec.template.spec.withServiceAccountName("seldon"),


  local baseApife(apifeImage) =
        {
            apiVersion: "extensions/v1beta1",
            kind: "Deployment",
            metadata: {
                name: "seldon-apiserver",
            },
            spec: {
                replicas: 1,
                template: {
                    metadata: {
                        annotations: {
                            "prometheus.io/path": "/prometheus",
                            "prometheus.io/port": "8080",
                            "prometheus.io/scrape": "true",
                        },
                        labels: {
                            app: "seldon-apiserver-container-app",
                            version: "1",
                        },
                    },
                    spec: {
                        containers: [
                            {
                                env: [
                                    {
                                        name: "SELDON_ENGINE_KAFKA_SERVER",
                                        value: "kafka:9092",
                                    },
                                    {
                                        name: "SELDON_CLUSTER_MANAGER_REDIS_HOST",
                                        value: "redis",
                                    },
                                ],
                                image: apifeImage,
                                imagePullPolicy: "IfNotPresent",
                                name: "seldon-apiserver-container",
                                ports: [
                                    {
                                        containerPort: 8080,
                                        protocol: "TCP",
                                    },
                                    {
                                        containerPort: 5000,
                                        protocol: "TCP",
                                    },
                                ],
                            },
                        ],
                    },
                },
            },
        },

 apifeService(serviceType)::
        {
            apiVersion: "v1",
            kind: "Service",
            metadata: {
                labels: {
                    app: "seldon-apiserver-container-app",
                },
                name: "seldon-apiserver",
            },
            spec: {
                ports: [
                    {
                        name: "http",
                        nodePort: 30032,
                        port: 8080,
                        protocol: "TCP",
                        targetPort: 8080,
                    },
                    {
                        name: "grpc",
                        nodePort: 30033,
                        port: 5000,
                        protocol: "TCP",
                        targetPort: 5000,
                    },
                ],
                selector: {
                    app: "seldon-apiserver-container-app",
                },
                sessionAffinity: "None",
                type: serviceType,
            },
            status: {
                loadBalancer: {},
            },
        },

    deploymentOperator(engineImage, clusterManagerImage, springOpts, javaOpts):
{
    kind: "Deployment",
    apiVersion: "extensions/v1beta1",
    metadata: {
        name: "seldon-cluster-manager",
        labels: {
            app: "seldon-cluster-manager-server",
        },
    },
    spec: {
        replicas: 1,
        selector: {
            matchLabels: {
                app: "seldon-cluster-manager-server",
            },
        },
        template: {
            metadata: {
                labels: {
                    app: "seldon-cluster-manager-server",
                },
            },
            spec: {
                containers: [
                    {
                        name: "seldon-cluster-manager-container",
                        image: clusterManagerImage,
                        ports: [
                            {
                                containerPort: 8080,
                                protocol: "TCP",
                            },
                        ],
                        env: [
                            {
                              name: "JAVA_OPTS",
                              value: javaOpts,
                            },
                            {
                              name: "SPRING_OPTS",
                              value: springOpts,
                            },
                            {
                                name: "SELDON_CLUSTER_MANAGER_REDIS_HOST",
                                value: "redis",
                            },
                            {
                                name: "ENGINE_CONTAINER_IMAGE_AND_VERSION",
                                value: engineImage,
                            },
                            {
                                name: "SELDON_CLUSTER_MANAGER_POD_NAMESPACE",
                                valueFrom: {
                                    fieldRef: {
                                        apiVersion: "v1",
                                        fieldPath: "metadata.namespace",
                                    },
                                },
                            },
                        ],
                        resources: {},
                        terminationMessagePath: "/dev/termination-log",
                        terminationMessagePolicy: "File",
                        imagePullPolicy: "IfNotPresent",
                    },
                ],
                restartPolicy: "Always",
                terminationGracePeriodSeconds: 30,
                dnsPolicy: "ClusterFirst",
                securityContext: {},
                schedulerName: "default-scheduler",
            },
        },
        strategy: {
            type: "RollingUpdate",
            rollingUpdate: {
                maxUnavailable: 1,
                maxSurge: 1,
            },
        },
    },
},


    redisDeployment():
        {
            kind: "Deployment",
            apiVersion: "apps/v1beta1",
            metadata: {
                name: "redis",
                creationTimestamp: null,
                labels: {
                    app: "redis-app",
                },
            },
            spec: {
                replicas: 1,
                selector: {
                    matchLabels: {
                        app: "redis-app",
                    },
                },
                template: {
                    metadata: {
                        creationTimestamp: null,
                        labels: {
                            app: "redis-app",
                        },
                    },
                    spec: {
                        containers: [
                            {
                                name: "redis-container",
                                image: "redis:4.0.1",
                                ports: [
                                    {
                                        containerPort: 6379,
                                        protocol: "TCP",
                                    },
                                ],
                                resources: {},
                                terminationMessagePath: "/dev/termination-log",
                                terminationMessagePolicy: "File",
                                imagePullPolicy: "IfNotPresent",
                            },
                        ],
                        restartPolicy: "Always",
                        terminationGracePeriodSeconds: 30,
                        dnsPolicy: "ClusterFirst",
                        securityContext: {},
                        schedulerName: "default-scheduler",
                    },
                },
                strategy: {
                    type: "RollingUpdate",
                    rollingUpdate: {
                        maxUnavailable: 1,
                        maxSurge: 1,
                    },
                },
            },
            status: {},
        },

    redisService():
        {
            kind: "Service",
            apiVersion: "v1",
            metadata: {
                name: "redis",
                creationTimestamp: null,
            },
            spec: {
                ports: [
                    {
                        protocol: "TCP",
                        port: 6379,
                        targetPort: 6379,
                    },
                ],
                selector: {
                    app: "redis-app",
                },
                type: "ClusterIP",
                sessionAffinity: "None",
            },
            status: {
                loadBalancer: {},
            },
        },

    rbacServiceAccount():
        {
            kind: "ServiceAccount",
            apiVersion: "v1",
            metadata: {
                name: "seldon",
                namespace: "default",
            },
        },

    rbacClusterRoleBinding():
        {
            kind: "ClusterRoleBinding",
            apiVersion: "rbac.authorization.k8s.io/v1",
            metadata: {
                name: "seldon",
            },
            subjects: [
                {
                    kind: "ServiceAccount",
                    name: "seldon",
                    namespace: "default",
                },
            ],
            roleRef: {
                apiGroup: "rbac.authorization.k8s.io",
                kind: "ClusterRole",
                name: "cluster-admin",
            },
        },

    crd():
{
    apiVersion: "apiextensions.k8s.io/v1beta1",
    kind: "CustomResourceDefinition",
    metadata: {
        name: "seldondeployments.machinelearning.seldon.io",
    },
    spec: {
        group: "machinelearning.seldon.io",
        names: {
            kind: "SeldonDeployment",
            plural: "seldondeployments",
            shortNames: [
                "sdep",
            ],
            singular: "seldondeployment",
        },
        scope: "Namespaced",
        validation: {
            openAPIV3Schema: {
                properties: {
                    spec: {
                        properties: {
                            annotations: {
                                description: "The annotations to be updated to a deployment",
                                type: "object",
                            },
                            name: {
                                type: "string",
                            },
                            oauth_key: {
                                type: "string",
                            },
                            oauth_secret: {
                                type: "string",
                            },
                            predictors: {
                                description: "List of predictors belonging to the deployment",
                                items: {
                                    properties: {
                                        annotations: {
                                            description: "The annotations to be updated to a predictor",
                                            type: "object",
                                        },
                                        graph: {
                                            properties: {
                                                children: {
                                                    items: {
                                                        properties: {
                                                            children: {
                                                                items: {
                                                                    properties: {
                                                                        children: {
                                                                            items: {},
                                                                            type: "array",
                                                                        },
                                                                        endpoint: {
                                                                            properties: {
                                                                                service_host: {
                                                                                    type: "string",
                                                                                },
                                                                                service_port: {
                                                                                    type: "integer",
                                                                                },
                                                                                type: {
                                                                                    enum: [
                                                                                        "REST",
                                                                                        "GRPC",
                                                                                    ],
                                                                                    type: "string",
                                                                                },
                                                                            },
                                                                        },
                                                                        name: {
                                                                            type: "string",
                                                                        },
                                                                        implementation: {
                                                                            enum: [
                                                                                "UNKNOWN_IMPLEMENTATION",
                                                                                "SIMPLE_MODEL",
                                                                                "SIMPLE_ROUTER",
                                                                                "RANDOM_ABTEST",
                                                                                "AVERAGE_COMBINER",
                                                                            ],
                                                                            type: "string",
                                                                        },
                                                                        type: {
                                                                            enum: [
                                                                                "UNKNOWN_TYPE",
                                                                                "ROUTER",
                                                                                "COMBINER",
                                                                                "MODEL",
                                                                                "TRANSFORMER",
                                                                                "OUTPUT_TRANSFORMER",
                                                                            ],
                                                                            type: "string",
                                                                        },
                                                                        methods: {
                                                                            type: "array",
                                                                            items: {
                                                                                enum: [
                                                                                    "TRANSFORM_INPUT",
                                                                                    "TRANSFORM_OUTPUT",
                                                                                    "ROUTE",
                                                                                    "AGGREGATE",
                                                                                    "SEND_FEEDBACK",
],
                                                                                type: "string",
                                                                            },
                                                                        },
                                                                    },
                                                                },
                                                                type: "array",
                                                            },
                                                            endpoint: {
                                                                properties: {
                                                                    service_host: {
                                                                        type: "string",
                                                                    },
                                                                    service_port: {
                                                                        type: "integer",
                                                                    },
                                                                    type: {
                                                                        enum: [
                                                                            "REST",
                                                                            "GRPC",
                                                                        ],
                                                                        type: "string",
                                                                    },
                                                                },
                                                            },
                                                            name: {
                                                                type: "string",
                                                            },
                                                            implementation: {
                                                                enum: [
                                                                    "UNKNOWN_IMPLEMENTATION",
                                                                    "SIMPLE_MODEL",
                                                                    "SIMPLE_ROUTER",
                                                                    "RANDOM_ABTEST",
                                                                    "AVERAGE_COMBINER",
                                                                ],
                                                                type: "string",
                                                            },
                                                            type: {
                                                                enum: [
                                                                    "UNKNOWN_TYPE",
                                                                    "ROUTER",
                                                                    "COMBINER",
                                                                    "MODEL",
                                                                    "TRANSFORMER",
                                                                    "OUTPUT_TRANSFORMER",
                                                                ],
                                                                type: "string",
                                                            },
                                                            methods: {
                                                                type: "array",
                                                                items: {
                                                                    enum: [
                                                                        "TRANSFORM_INPUT",
                                                                        "TRANSFORM_OUTPUT",
                                                                        "ROUTE",
                                                                        "AGGREGATE",
                                                                        "SEND_FEEDBACK",
],
                                                                    type: "string",
                                                                },
                                                            },
                                                        },
                                                    },
                                                    type: "array",
                                                },
                                                endpoint: {
                                                    properties: {
                                                        service_host: {
                                                            type: "string",
                                                        },
                                                        service_port: {
                                                            type: "integer",
                                                        },
                                                        type: {
                                                            enum: [
                                                                "REST",
                                                                "GRPC",
                                                            ],
                                                            type: "string",
                                                        },
                                                    },
                                                },
                                                name: {
                                                    type: "string",
                                                },
                                                implementation: {
                                                    enum: [
                                                        "UNKNOWN_IMPLEMENTATION",
                                                        "SIMPLE_MODEL",
                                                        "SIMPLE_ROUTER",
                                                        "RANDOM_ABTEST",
                                                        "AVERAGE_COMBINER",
                                                    ],
                                                    type: "string",
                                                },
                                                type: {
                                                    enum: [
                                                        "UNKNOWN_TYPE",
                                                        "ROUTER",
                                                        "COMBINER",
                                                        "MODEL",
                                                        "TRANSFORMER",
                                                        "OUTPUT_TRANSFORMER",
                                                    ],
                                                    type: "string",
                                                },
                                                methods: {
                                                    type: "array",
                                                    items: {
                                                        enum: [
                                                            "TRANSFORM_INPUT",
                                                            "TRANSFORM_OUTPUT",
                                                            "ROUTE",
                                                            "AGGREGATE",
                                                            "SEND_FEEDBACK",
],
                                                        type: "string",
                                                    },
                                                },
                                            },
                                        },
                                        name: {
                                            type: "string",
                                        },
                                        replicas: {
                                            type: "integer",
                                        },
                                    },
                                },
                                type: "array",
                            },
                            componentSpec: podTemplateValidation,

                        },
                    },
                },
            },
        },
        version: "v1alpha1",
    },
},


  },  // parts
}
