{{- define "podSpec" }}
{
    "description": "PodSpec is a description of a pod.",
    "properties": {
        "activeDeadlineSeconds": {
            "description": "Optional duration in seconds the pod may be active on the node relative to StartTime before the system will actively try to mark it failed and kill associated containers. Value must be a positive integer.",
            "format": "int64",
            "type": "integer"
        },
        "affinity": {
            "description": "Affinity is a group of affinity scheduling rules.",
            "properties": {
                "nodeAffinity": {
                    "description": "Node affinity is a group of node affinity scheduling rules.",
                    "properties": {
                        "preferredDuringSchedulingIgnoredDuringExecution": {
                            "description": "The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding \"weight\" to the sum if the node matches the corresponding matchExpressions; the node(s) with the highest sum are the most preferred.",
                            "items": {
                                "description": "An empty preferred scheduling term matches all objects with implicit weight 0 (i.e. it's a no-op). A null preferred scheduling term matches no objects (i.e. is also a no-op).",
                                "properties": {
                                    "preference": {
                                        "description": "A null or empty node selector term matches no objects. The requirements of them are ANDed. The TopologySelectorTerm type implements a subset of the NodeSelectorTerm.",
                                        "properties": {
                                            "matchExpressions": {
                                                "description": "A list of node selector requirements by node's labels.",
                                                "items": {
                                                    "description": "A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.",
                                                    "properties": {
                                                        "key": {
                                                            "description": "The label key that the selector applies to.",
                                                            "type": "string"
                                                        },
                                                        "operator": {
                                                            "description": "Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.",
                                                            "type": "string"
                                                        },
                                                        "values": {
                                                            "description": "An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.",
                                                            "items": {
                                                                "type": "string"
                                                            },
                                                            "type": "array"
                                                        }
                                                    },
                                                    "required": [
                                                        "key",
                                                        "operator"
                                                    ],
                                                    "type": "object"
                                                },
                                                "type": "array"
                                            },
                                            "matchFields": {
                                                "description": "A list of node selector requirements by node's fields.",
                                                "items": {
                                                    "description": "A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.",
                                                    "properties": {
                                                        "key": {
                                                            "description": "The label key that the selector applies to.",
                                                            "type": "string"
                                                        },
                                                        "operator": {
                                                            "description": "Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.",
                                                            "type": "string"
                                                        },
                                                        "values": {
                                                            "description": "An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.",
                                                            "items": {
                                                                "type": "string"
                                                            },
                                                            "type": "array"
                                                        }
                                                    },
                                                    "required": [
                                                        "key",
                                                        "operator"
                                                    ],
                                                    "type": "object"
                                                },
                                                "type": "array"
                                            }
                                        },
                                        "type": "object"
                                    },
                                    "weight": {
                                        "description": "Weight associated with matching the corresponding nodeSelectorTerm, in the range 1-100.",
                                        "format": "int32",
                                        "type": "integer"
                                    }
                                },
                                "required": [
                                    "weight",
                                    "preference"
                                ],
                                "type": "object"
                            },
                            "type": "array"
                        },
                        "requiredDuringSchedulingIgnoredDuringExecution": {
                            "description": "A node selector represents the union of the results of one or more label queries over a set of nodes; that is, it represents the OR of the selectors represented by the node selector terms.",
                            "properties": {
                                "nodeSelectorTerms": {
                                    "description": "Required. A list of node selector terms. The terms are ORed.",
                                    "items": {
                                        "description": "A null or empty node selector term matches no objects. The requirements of them are ANDed. The TopologySelectorTerm type implements a subset of the NodeSelectorTerm.",
                                        "properties": {
                                            "matchExpressions": {
                                                "description": "A list of node selector requirements by node's labels.",
                                                "items": {
                                                    "description": "A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.",
                                                    "properties": {
                                                        "key": {
                                                            "description": "The label key that the selector applies to.",
                                                            "type": "string"
                                                        },
                                                        "operator": {
                                                            "description": "Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.",
                                                            "type": "string"
                                                        },
                                                        "values": {
                                                            "description": "An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.",
                                                            "items": {
                                                                "type": "string"
                                                            },
                                                            "type": "array"
                                                        }
                                                    },
                                                    "required": [
                                                        "key",
                                                        "operator"
                                                    ],
                                                    "type": "object"
                                                },
                                                "type": "array"
                                            },
                                            "matchFields": {
                                                "description": "A list of node selector requirements by node's fields.",
                                                "items": {
                                                    "description": "A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.",
                                                    "properties": {
                                                        "key": {
                                                            "description": "The label key that the selector applies to.",
                                                            "type": "string"
                                                        },
                                                        "operator": {
                                                            "description": "Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.",
                                                            "type": "string"
                                                        },
                                                        "values": {
                                                            "description": "An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.",
                                                            "items": {
                                                                "type": "string"
                                                            },
                                                            "type": "array"
                                                        }
                                                    },
                                                    "required": [
                                                        "key",
                                                        "operator"
                                                    ],
                                                    "type": "object"
                                                },
                                                "type": "array"
                                            }
                                        },
                                        "type": "object"
                                    },
                                    "type": "array"
                                }
                            },
                            "required": [
                                "nodeSelectorTerms"
                            ],
                            "type": "object"
                        }
                    },
                    "type": "object"
                },
                "podAffinity": {
                    "description": "Pod affinity is a group of inter pod affinity scheduling rules.",
                    "properties": {
                        "preferredDuringSchedulingIgnoredDuringExecution": {
                            "description": "The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding \"weight\" to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred.",
                            "items": {
                                "description": "The weights of all of the matched WeightedPodAffinityTerm fields are added per-node to find the most preferred node(s)",
                                "properties": {
                                    "podAffinityTerm": {
                                        "description": "Defines a set of pods (namely those matching the labelSelector relative to the given namespace(s)) that this pod should be co-located (affinity) or not co-located (anti-affinity) with, where co-located is defined as running on a node whose value of the label with key <topologyKey> matches that of any node on which a pod of the set of pods is running",
                                        "properties": {
                                            "labelSelector": {
                                                "description": "A label selector is a label query over a set of resources. The result of matchLabels and matchExpressions are ANDed. An empty label selector matches all objects. A null label selector matches no objects.",
                                                "properties": {
                                                    "matchExpressions": {
                                                        "description": "matchExpressions is a list of label selector requirements. The requirements are ANDed.",
                                                        "items": {
                                                            "description": "A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.",
                                                            "properties": {
                                                                "key": {
                                                                    "description": "key is the label key that the selector applies to.",
                                                                    "type": "string",
                                                                    "x-kubernetes-patch-merge-key": "key",
                                                                    "x-kubernetes-patch-strategy": "merge"
                                                                },
                                                                "operator": {
                                                                    "description": "operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.",
                                                                    "type": "string"
                                                                },
                                                                "values": {
                                                                    "description": "values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.",
                                                                    "items": {
                                                                        "type": "string"
                                                                    },
                                                                    "type": "array"
                                                                }
                                                            },
                                                            "required": [
                                                                "key",
                                                                "operator"
                                                            ],
                                                            "type": "object"
                                                        },
                                                        "type": "array"
                                                    },
                                                    "matchLabels": {
                                                        "additionalProperties": {
                                                            "type": "string"
                                                        },
                                                        "description": "matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is \"key\", the operator is \"In\", and the values array contains only \"value\". The requirements are ANDed.",
                                                        "type": "object"
                                                    }
                                                },
                                                "type": "object"
                                            },
                                            "namespaces": {
                                                "description": "namespaces specifies which namespaces the labelSelector applies to (matches against); null or empty list means \"this pod's namespace\"",
                                                "items": {
                                                    "type": "string"
                                                },
                                                "type": "array"
                                            },
                                            "topologyKey": {
                                                "description": "This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.",
                                                "type": "string"
                                            }
                                        },
                                        "required": [
                                            "topologyKey"
                                        ],
                                        "type": "object"
                                    },
                                    "weight": {
                                        "description": "weight associated with matching the corresponding podAffinityTerm, in the range 1-100.",
                                        "format": "int32",
                                        "type": "integer"
                                    }
                                },
                                "required": [
                                    "weight",
                                    "podAffinityTerm"
                                ],
                                "type": "object"
                            },
                            "type": "array"
                        },
                        "requiredDuringSchedulingIgnoredDuringExecution": {
                            "description": "If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied.",
                            "items": {
                                "description": "Defines a set of pods (namely those matching the labelSelector relative to the given namespace(s)) that this pod should be co-located (affinity) or not co-located (anti-affinity) with, where co-located is defined as running on a node whose value of the label with key <topologyKey> matches that of any node on which a pod of the set of pods is running",
                                "properties": {
                                    "labelSelector": {
                                        "description": "A label selector is a label query over a set of resources. The result of matchLabels and matchExpressions are ANDed. An empty label selector matches all objects. A null label selector matches no objects.",
                                        "properties": {
                                            "matchExpressions": {
                                                "description": "matchExpressions is a list of label selector requirements. The requirements are ANDed.",
                                                "items": {
                                                    "description": "A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.",
                                                    "properties": {
                                                        "key": {
                                                            "description": "key is the label key that the selector applies to.",
                                                            "type": "string",
                                                            "x-kubernetes-patch-merge-key": "key",
                                                            "x-kubernetes-patch-strategy": "merge"
                                                        },
                                                        "operator": {
                                                            "description": "operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.",
                                                            "type": "string"
                                                        },
                                                        "values": {
                                                            "description": "values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.",
                                                            "items": {
                                                                "type": "string"
                                                            },
                                                            "type": "array"
                                                        }
                                                    },
                                                    "required": [
                                                        "key",
                                                        "operator"
                                                    ],
                                                    "type": "object"
                                                },
                                                "type": "array"
                                            },
                                            "matchLabels": {
                                                "additionalProperties": {
                                                    "type": "string"
                                                },
                                                "description": "matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is \"key\", the operator is \"In\", and the values array contains only \"value\". The requirements are ANDed.",
                                                "type": "object"
                                            }
                                        },
                                        "type": "object"
                                    },
                                    "namespaces": {
                                        "description": "namespaces specifies which namespaces the labelSelector applies to (matches against); null or empty list means \"this pod's namespace\"",
                                        "items": {
                                            "type": "string"
                                        },
                                        "type": "array"
                                    },
                                    "topologyKey": {
                                        "description": "This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.",
                                        "type": "string"
                                    }
                                },
                                "required": [
                                    "topologyKey"
                                ],
                                "type": "object"
                            },
                            "type": "array"
                        }
                    },
                    "type": "object"
                },
                "podAntiAffinity": {
                    "description": "Pod anti affinity is a group of inter pod anti affinity scheduling rules.",
                    "properties": {
                        "preferredDuringSchedulingIgnoredDuringExecution": {
                            "description": "The scheduler will prefer to schedule pods to nodes that satisfy the anti-affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling anti-affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding \"weight\" to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred.",
                            "items": {
                                "description": "The weights of all of the matched WeightedPodAffinityTerm fields are added per-node to find the most preferred node(s)",
                                "properties": {
                                    "podAffinityTerm": {
                                        "description": "Defines a set of pods (namely those matching the labelSelector relative to the given namespace(s)) that this pod should be co-located (affinity) or not co-located (anti-affinity) with, where co-located is defined as running on a node whose value of the label with key <topologyKey> matches that of any node on which a pod of the set of pods is running",
                                        "properties": {
                                            "labelSelector": {
                                                "description": "A label selector is a label query over a set of resources. The result of matchLabels and matchExpressions are ANDed. An empty label selector matches all objects. A null label selector matches no objects.",
                                                "properties": {
                                                    "matchExpressions": {
                                                        "description": "matchExpressions is a list of label selector requirements. The requirements are ANDed.",
                                                        "items": {
                                                            "description": "A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.",
                                                            "properties": {
                                                                "key": {
                                                                    "description": "key is the label key that the selector applies to.",
                                                                    "type": "string",
                                                                    "x-kubernetes-patch-merge-key": "key",
                                                                    "x-kubernetes-patch-strategy": "merge"
                                                                },
                                                                "operator": {
                                                                    "description": "operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.",
                                                                    "type": "string"
                                                                },
                                                                "values": {
                                                                    "description": "values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.",
                                                                    "items": {
                                                                        "type": "string"
                                                                    },
                                                                    "type": "array"
                                                                }
                                                            },
                                                            "required": [
                                                                "key",
                                                                "operator"
                                                            ],
                                                            "type": "object"
                                                        },
                                                        "type": "array"
                                                    },
                                                    "matchLabels": {
                                                        "additionalProperties": {
                                                            "type": "string"
                                                        },
                                                        "description": "matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is \"key\", the operator is \"In\", and the values array contains only \"value\". The requirements are ANDed.",
                                                        "type": "object"
                                                    }
                                                },
                                                "type": "object"
                                            },
                                            "namespaces": {
                                                "description": "namespaces specifies which namespaces the labelSelector applies to (matches against); null or empty list means \"this pod's namespace\"",
                                                "items": {
                                                    "type": "string"
                                                },
                                                "type": "array"
                                            },
                                            "topologyKey": {
                                                "description": "This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.",
                                                "type": "string"
                                            }
                                        },
                                        "required": [
                                            "topologyKey"
                                        ],
                                        "type": "object"
                                    },
                                    "weight": {
                                        "description": "weight associated with matching the corresponding podAffinityTerm, in the range 1-100.",
                                        "format": "int32",
                                        "type": "integer"
                                    }
                                },
                                "required": [
                                    "weight",
                                    "podAffinityTerm"
                                ],
                                "type": "object"
                            },
                            "type": "array"
                        },
                        "requiredDuringSchedulingIgnoredDuringExecution": {
                            "description": "If the anti-affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the anti-affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied.",
                            "items": {
                                "description": "Defines a set of pods (namely those matching the labelSelector relative to the given namespace(s)) that this pod should be co-located (affinity) or not co-located (anti-affinity) with, where co-located is defined as running on a node whose value of the label with key <topologyKey> matches that of any node on which a pod of the set of pods is running",
                                "properties": {
                                    "labelSelector": {
                                        "description": "A label selector is a label query over a set of resources. The result of matchLabels and matchExpressions are ANDed. An empty label selector matches all objects. A null label selector matches no objects.",
                                        "properties": {
                                            "matchExpressions": {
                                                "description": "matchExpressions is a list of label selector requirements. The requirements are ANDed.",
                                                "items": {
                                                    "description": "A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.",
                                                    "properties": {
                                                        "key": {
                                                            "description": "key is the label key that the selector applies to.",
                                                            "type": "string",
                                                            "x-kubernetes-patch-merge-key": "key",
                                                            "x-kubernetes-patch-strategy": "merge"
                                                        },
                                                        "operator": {
                                                            "description": "operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.",
                                                            "type": "string"
                                                        },
                                                        "values": {
                                                            "description": "values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.",
                                                            "items": {
                                                                "type": "string"
                                                            },
                                                            "type": "array"
                                                        }
                                                    },
                                                    "required": [
                                                        "key",
                                                        "operator"
                                                    ],
                                                    "type": "object"
                                                },
                                                "type": "array"
                                            },
                                            "matchLabels": {
                                                "additionalProperties": {
                                                    "type": "string"
                                                },
                                                "description": "matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is \"key\", the operator is \"In\", and the values array contains only \"value\". The requirements are ANDed.",
                                                "type": "object"
                                            }
                                        },
                                        "type": "object"
                                    },
                                    "namespaces": {
                                        "description": "namespaces specifies which namespaces the labelSelector applies to (matches against); null or empty list means \"this pod's namespace\"",
                                        "items": {
                                            "type": "string"
                                        },
                                        "type": "array"
                                    },
                                    "topologyKey": {
                                        "description": "This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.",
                                        "type": "string"
                                    }
                                },
                                "required": [
                                    "topologyKey"
                                ],
                                "type": "object"
                            },
                            "type": "array"
                        }
                    },
                    "type": "object"
                }
            },
            "type": "object"
        },
        "automountServiceAccountToken": {
            "description": "AutomountServiceAccountToken indicates whether a service account token should be automatically mounted.",
            "type": "boolean"
        },
        "containers": {
            "description": "List of containers belonging to the pod. Containers cannot currently be added or removed. There must be at least one container in a Pod. Cannot be updated.",
            "items": {
                "description": "A single application container that you want to run within a pod.",
                "properties": {
                    "args": {
                        "description": "Arguments to the entrypoint. The docker image's CMD is used if this is not provided. Variable references $(VAR_NAME) are expanded using the container's environment. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Cannot be updated. More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell",
                        "items": {
                            "type": "string"
                        },
                        "type": "array"
                    },
                    "command": {
                        "description": "Entrypoint array. Not executed within a shell. The docker image's ENTRYPOINT is used if this is not provided. Variable references $(VAR_NAME) are expanded using the container's environment. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Cannot be updated. More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell",
                        "items": {
                            "type": "string"
                        },
                        "type": "array"
                    },
                    "env": {
                        "description": "List of environment variables to set in the container. Cannot be updated.",
                        "items": {
                            "description": "EnvVar represents an environment variable present in a Container.",
                            "properties": {
                                "name": {
                                    "description": "Name of the environment variable. Must be a C_IDENTIFIER.",
                                    "type": "string"
                                },
                                "value": {
                                    "description": "Variable references $(VAR_NAME) are expanded using the previous defined environment variables in the container and any service environment variables. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Defaults to \"\".",
                                    "type": "string"
                                },
                                "valueFrom": {
                                    "description": "EnvVarSource represents a source for the value of an EnvVar.",
                                    "properties": {
                                        "configMapKeyRef": {
                                            "description": "Selects a key from a ConfigMap.",
                                            "properties": {
                                                "key": {
                                                    "description": "The key to select.",
                                                    "type": "string"
                                                },
                                                "name": {
                                                    "description": "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
                                                    "type": "string"
                                                },
                                                "optional": {
                                                    "description": "Specify whether the ConfigMap or it's key must be defined",
                                                    "type": "boolean"
                                                }
                                            },
                                            "required": [
                                                "key"
                                            ],
                                            "type": "object"
                                        },
                                        "fieldRef": {
                                            "description": "ObjectFieldSelector selects an APIVersioned field of an object.",
                                            "properties": {
                                                "apiVersion": {
                                                    "description": "Version of the schema the FieldPath is written in terms of, defaults to \"v1\".",
                                                    "type": "string"
                                                },
                                                "fieldPath": {
                                                    "description": "Path of the field to select in the specified API version.",
                                                    "type": "string"
                                                }
                                            },
                                            "required": [
                                                "fieldPath"
                                            ],
                                            "type": "object"
                                        },
                                        "resourceFieldRef": {
                                            "description": "ResourceFieldSelector represents container resources (cpu, memory) and their output format",
                                            "properties": {
                                                "containerName": {
                                                    "description": "Container name: required for volumes, optional for env vars",
                                                    "type": "string"
                                                },
                                                "divisor": {
                                                    "description": "Quantity is a fixed-point representation of a number. It provides convenient marshaling/unmarshaling in JSON and YAML, in addition to String() and Int64() accessors.\n\nThe serialization format is:\n\n<quantity>        ::= <signedNumber><suffix>\n  (Note that <suffix> may be empty, from the \"\" case in <decimalSI>.)\n<digit>           ::= 0 | 1 | ... | 9 <digits>          ::= <digit> | <digit><digits> <number>          ::= <digits> | <digits>.<digits> | <digits>. | .<digits> <sign>            ::= \"+\" | \"-\" <signedNumber>    ::= <number> | <sign><number> <suffix>          ::= <binarySI> | <decimalExponent> | <decimalSI> <binarySI>        ::= Ki | Mi | Gi | Ti | Pi | Ei\n  (International System of units; See: http://physics.nist.gov/cuu/Units/binary.html)\n<decimalSI>       ::= m | \"\" | k | M | G | T | P | E\n  (Note that 1024 = 1Ki but 1000 = 1k; I didn't choose the capitalization.)\n<decimalExponent> ::= \"e\" <signedNumber> | \"E\" <signedNumber>\n\nNo matter which of the three exponent forms is used, no quantity may represent a number greater than 2^63-1 in magnitude, nor may it have more than 3 decimal places. Numbers larger or more precise will be capped or rounded up. (E.g.: 0.1m will rounded up to 1m.) This may be extended in the future if we require larger or smaller quantities.\n\nWhen a Quantity is parsed from a string, it will remember the type of suffix it had, and will use the same type again when it is serialized.\n\nBefore serializing, Quantity will be put in \"canonical form\". This means that Exponent/suffix will be adjusted up or down (with a corresponding increase or decrease in Mantissa) such that:\n  a. No precision is lost\n  b. No fractional digits will be emitted\n  c. The exponent (or suffix) is as large as possible.\nThe sign will be omitted unless the number is negative.\n\nExamples:\n  1.5 will be serialized as \"1500m\"\n  1.5Gi will be serialized as \"1536Mi\"\n\nNote that the quantity will NEVER be internally represented by a floating point number. That is the whole point of this exercise.\n\nNon-canonical values will still parse as long as they are well formed, but will be re-emitted in their canonical form. (So always use canonical form, or don't diff.)\n\nThis format is intended to make it difficult to use these numbers without writing some sort of special handling code in the hopes that that will cause implementors to also use a fixed point implementation.",
                                                    "type": "string"
                                                },
                                                "resource": {
                                                    "description": "Required: resource to select",
                                                    "type": "string"
                                                }
                                            },
                                            "required": [
                                                "resource"
                                            ],
                                            "type": "object"
                                        },
                                        "secretKeyRef": {
                                            "description": "SecretKeySelector selects a key of a Secret.",
                                            "properties": {
                                                "key": {
                                                    "description": "The key of the secret to select from.  Must be a valid secret key.",
                                                    "type": "string"
                                                },
                                                "name": {
                                                    "description": "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
                                                    "type": "string"
                                                },
                                                "optional": {
                                                    "description": "Specify whether the Secret or it's key must be defined",
                                                    "type": "boolean"
                                                }
                                            },
                                            "required": [
                                                "key"
                                            ],
                                            "type": "object"
                                        }
                                    },
                                    "type": "object"
                                }
                            },
                            "required": [
                                "name"
                            ],
                            "type": "object"
                        },
                        "type": "array",
                        "x-kubernetes-patch-merge-key": "name",
                        "x-kubernetes-patch-strategy": "merge"
                    },
                    "envFrom": {
                        "description": "List of sources to populate environment variables in the container. The keys defined within a source must be a C_IDENTIFIER. All invalid keys will be reported as an event when the container is starting. When a key exists in multiple sources, the value associated with the last source will take precedence. Values defined by an Env with a duplicate key will take precedence. Cannot be updated.",
                        "items": {
                            "description": "EnvFromSource represents the source of a set of ConfigMaps",
                            "properties": {
                                "configMapRef": {
                                    "description": "ConfigMapEnvSource selects a ConfigMap to populate the environment variables with.\n\nThe contents of the target ConfigMap's Data field will represent the key-value pairs as environment variables.",
                                    "properties": {
                                        "name": {
                                            "description": "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
                                            "type": "string"
                                        },
                                        "optional": {
                                            "description": "Specify whether the ConfigMap must be defined",
                                            "type": "boolean"
                                        }
                                    },
                                    "type": "object"
                                },
                                "prefix": {
                                    "description": "An optional identifier to prepend to each key in the ConfigMap. Must be a C_IDENTIFIER.",
                                    "type": "string"
                                },
                                "secretRef": {
                                    "description": "SecretEnvSource selects a Secret to populate the environment variables with.\n\nThe contents of the target Secret's Data field will represent the key-value pairs as environment variables.",
                                    "properties": {
                                        "name": {
                                            "description": "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
                                            "type": "string"
                                        },
                                        "optional": {
                                            "description": "Specify whether the Secret must be defined",
                                            "type": "boolean"
                                        }
                                    },
                                    "type": "object"
                                }
                            },
                            "type": "object"
                        },
                        "type": "array"
                    },
                    "image": {
                        "description": "Docker image name. More info: https://kubernetes.io/docs/concepts/containers/images This field is optional to allow higher level config management to default or override container images in workload controllers like Deployments and StatefulSets.",
                        "type": "string"
                    },
                    "imagePullPolicy": {
                        "description": "Image pull policy. One of Always, Never, IfNotPresent. Defaults to Always if :latest tag is specified, or IfNotPresent otherwise. Cannot be updated. More info: https://kubernetes.io/docs/concepts/containers/images#updating-images",
                        "type": "string"
                    },
                    "lifecycle": {
                        "description": "Lifecycle describes actions that the management system should take in response to container lifecycle events. For the PostStart and PreStop lifecycle handlers, management of the container blocks until the action is complete, unless the container process fails, in which case the handler is aborted.",
                        "properties": {
                            "postStart": {
                                "description": "Handler defines a specific action that should be taken",
                                "properties": {
                                    "exec": {
                                        "description": "ExecAction describes a \"run in container\" action.",
                                        "properties": {
                                            "command": {
                                                "description": "Command is the command line to execute inside the container, the working directory for the command  is root ('/') in the container's filesystem. The command is simply exec'd, it is not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.",
                                                "items": {
                                                    "type": "string"
                                                },
                                                "type": "array"
                                            }
                                        },
                                        "type": "object"
                                    },
                                    "httpGet": {
                                        "description": "HTTPGetAction describes an action based on HTTP Get requests.",
                                        "properties": {
                                            "host": {
                                                "description": "Host name to connect to, defaults to the pod IP. You probably want to set \"Host\" in httpHeaders instead.",
                                                "type": "string"
                                            },
                                            "httpHeaders": {
                                                "description": "Custom headers to set in the request. HTTP allows repeated headers.",
                                                "items": {
                                                    "description": "HTTPHeader describes a custom header to be used in HTTP probes",
                                                    "properties": {
                                                        "name": {
                                                            "description": "The header field name",
                                                            "type": "string"
                                                        },
                                                        "value": {
                                                            "description": "The header field value",
                                                            "type": "string"
                                                        }
                                                    },
                                                    "required": [
                                                        "name",
                                                        "value"
                                                    ],
                                                    "type": "object"
                                                },
                                                "type": "array"
                                            },
                                            "path": {
                                                "description": "Path to access on the HTTP server.",
                                                "type": "string"
                                            },
                                            "port": {
                                                "description": "IntOrString is a type that can hold an int32 or a string.  When used in JSON or YAML marshalling and unmarshalling, it produces or consumes the inner type.  This allows you to have, for example, a JSON field that can accept a name or number.",
                                                "format": "int-or-string",
                                                "type": "string"
                                            },
                                            "scheme": {
                                                "description": "Scheme to use for connecting to the host. Defaults to HTTP.",
                                                "type": "string"
                                            }
                                        },
                                        "required": [
                                            "port"
                                        ],
                                        "type": "object"
                                    },
                                    "tcpSocket": {
                                        "description": "TCPSocketAction describes an action based on opening a socket",
                                        "properties": {
                                            "host": {
                                                "description": "Optional: Host name to connect to, defaults to the pod IP.",
                                                "type": "string"
                                            },
                                            "port": {
                                                "description": "IntOrString is a type that can hold an int32 or a string.  When used in JSON or YAML marshalling and unmarshalling, it produces or consumes the inner type.  This allows you to have, for example, a JSON field that can accept a name or number.",
                                                "format": "int-or-string",
                                                "type": "string"
                                            }
                                        },
                                        "required": [
                                            "port"
                                        ],
                                        "type": "object"
                                    }
                                },
                                "type": "object"
                            },
                            "preStop": {
                                "description": "Handler defines a specific action that should be taken",
                                "properties": {
                                    "exec": {
                                        "description": "ExecAction describes a \"run in container\" action.",
                                        "properties": {
                                            "command": {
                                                "description": "Command is the command line to execute inside the container, the working directory for the command  is root ('/') in the container's filesystem. The command is simply exec'd, it is not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.",
                                                "items": {
                                                    "type": "string"
                                                },
                                                "type": "array"
                                            }
                                        },
                                        "type": "object"
                                    },
                                    "httpGet": {
                                        "description": "HTTPGetAction describes an action based on HTTP Get requests.",
                                        "properties": {
                                            "host": {
                                                "description": "Host name to connect to, defaults to the pod IP. You probably want to set \"Host\" in httpHeaders instead.",
                                                "type": "string"
                                            },
                                            "httpHeaders": {
                                                "description": "Custom headers to set in the request. HTTP allows repeated headers.",
                                                "items": {
                                                    "description": "HTTPHeader describes a custom header to be used in HTTP probes",
                                                    "properties": {
                                                        "name": {
                                                            "description": "The header field name",
                                                            "type": "string"
                                                        },
                                                        "value": {
                                                            "description": "The header field value",
                                                            "type": "string"
                                                        }
                                                    },
                                                    "required": [
                                                        "name",
                                                        "value"
                                                    ],
                                                    "type": "object"
                                                },
                                                "type": "array"
                                            },
                                            "path": {
                                                "description": "Path to access on the HTTP server.",
                                                "type": "string"
                                            },
                                            "port": {
                                                "description": "IntOrString is a type that can hold an int32 or a string.  When used in JSON or YAML marshalling and unmarshalling, it produces or consumes the inner type.  This allows you to have, for example, a JSON field that can accept a name or number.",
                                                "format": "int-or-string",
                                                "type": "string"
                                            },
                                            "scheme": {
                                                "description": "Scheme to use for connecting to the host. Defaults to HTTP.",
                                                "type": "string"
                                            }
                                        },
                                        "required": [
                                            "port"
                                        ],
                                        "type": "object"
                                    },
                                    "tcpSocket": {
                                        "description": "TCPSocketAction describes an action based on opening a socket",
                                        "properties": {
                                            "host": {
                                                "description": "Optional: Host name to connect to, defaults to the pod IP.",
                                                "type": "string"
                                            },
                                            "port": {
                                                "description": "IntOrString is a type that can hold an int32 or a string.  When used in JSON or YAML marshalling and unmarshalling, it produces or consumes the inner type.  This allows you to have, for example, a JSON field that can accept a name or number.",
                                                "format": "int-or-string",
                                                "type": "string"
                                            }
                                        },
                                        "required": [
                                            "port"
                                        ],
                                        "type": "object"
                                    }
                                },
                                "type": "object"
                            }
                        },
                        "type": "object"
                    },
                    "livenessProbe": {
                        "description": "Probe describes a health check to be performed against a container to determine whether it is alive or ready to receive traffic.",
                        "properties": {
                            "exec": {
                                "description": "ExecAction describes a \"run in container\" action.",
                                "properties": {
                                    "command": {
                                        "description": "Command is the command line to execute inside the container, the working directory for the command  is root ('/') in the container's filesystem. The command is simply exec'd, it is not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.",
                                        "items": {
                                            "type": "string"
                                        },
                                        "type": "array"
                                    }
                                },
                                "type": "object"
                            },
                            "failureThreshold": {
                                "description": "Minimum consecutive failures for the probe to be considered failed after having succeeded. Defaults to 3. Minimum value is 1.",
                                "format": "int32",
                                "type": "integer"
                            },
                            "httpGet": {
                                "description": "HTTPGetAction describes an action based on HTTP Get requests.",
                                "properties": {
                                    "host": {
                                        "description": "Host name to connect to, defaults to the pod IP. You probably want to set \"Host\" in httpHeaders instead.",
                                        "type": "string"
                                    },
                                    "httpHeaders": {
                                        "description": "Custom headers to set in the request. HTTP allows repeated headers.",
                                        "items": {
                                            "description": "HTTPHeader describes a custom header to be used in HTTP probes",
                                            "properties": {
                                                "name": {
                                                    "description": "The header field name",
                                                    "type": "string"
                                                },
                                                "value": {
                                                    "description": "The header field value",
                                                    "type": "string"
                                                }
                                            },
                                            "required": [
                                                "name",
                                                "value"
                                            ],
                                            "type": "object"
                                        },
                                        "type": "array"
                                    },
                                    "path": {
                                        "description": "Path to access on the HTTP server.",
                                        "type": "string"
                                    },
                                    "port": {
                                        "description": "IntOrString is a type that can hold an int32 or a string.  When used in JSON or YAML marshalling and unmarshalling, it produces or consumes the inner type.  This allows you to have, for example, a JSON field that can accept a name or number.",
                                        "format": "int-or-string",
                                        "type": "string"
                                    },
                                    "scheme": {
                                        "description": "Scheme to use for connecting to the host. Defaults to HTTP.",
                                        "type": "string"
                                    }
                                },
                                "required": [
                                    "port"
                                ],
                                "type": "object"
                            },
                            "initialDelaySeconds": {
                                "description": "Number of seconds after the container has started before liveness probes are initiated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes",
                                "format": "int32",
                                "type": "integer"
                            },
                            "periodSeconds": {
                                "description": "How often (in seconds) to perform the probe. Default to 10 seconds. Minimum value is 1.",
                                "format": "int32",
                                "type": "integer"
                            },
                            "successThreshold": {
                                "description": "Minimum consecutive successes for the probe to be considered successful after having failed. Defaults to 1. Must be 1 for liveness. Minimum value is 1.",
                                "format": "int32",
                                "type": "integer"
                            },
                            "tcpSocket": {
                                "description": "TCPSocketAction describes an action based on opening a socket",
                                "properties": {
                                    "host": {
                                        "description": "Optional: Host name to connect to, defaults to the pod IP.",
                                        "type": "string"
                                    },
                                    "port": {
                                        "description": "IntOrString is a type that can hold an int32 or a string.  When used in JSON or YAML marshalling and unmarshalling, it produces or consumes the inner type.  This allows you to have, for example, a JSON field that can accept a name or number.",
                                        "format": "int-or-string",
                                        "type": "string"
                                    }
                                },
                                "required": [
                                    "port"
                                ],
                                "type": "object"
                            },
                            "timeoutSeconds": {
                                "description": "Number of seconds after which the probe times out. Defaults to 1 second. Minimum value is 1. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes",
                                "format": "int32",
                                "type": "integer"
                            }
                        },
                        "type": "object"
                    },
                    "name": {
                        "description": "Name of the container specified as a DNS_LABEL. Each container in a pod must have a unique name (DNS_LABEL). Cannot be updated.",
                        "type": "string"
                    },
                    "ports": {
                        "description": "List of ports to expose from the container. Exposing a port here gives the system additional information about the network connections a container uses, but is primarily informational. Not specifying a port here DOES NOT prevent that port from being exposed. Any port which is listening on the default \"0.0.0.0\" address inside a container will be accessible from the network. Cannot be updated.",
                        "items": {
                            "description": "ContainerPort represents a network port in a single container.",
                            "properties": {
                                "containerPort": {
                                    "description": "Number of port to expose on the pod's IP address. This must be a valid port number, 0 < x < 65536.",
                                    "format": "int32",
                                    "type": "integer"
                                },
                                "hostIP": {
                                    "description": "What host IP to bind the external port to.",
                                    "type": "string"
                                },
                                "hostPort": {
                                    "description": "Number of port to expose on the host. If specified, this must be a valid port number, 0 < x < 65536. If HostNetwork is specified, this must match ContainerPort. Most containers do not need this.",
                                    "format": "int32",
                                    "type": "integer"
                                },
                                "name": {
                                    "description": "If specified, this must be an IANA_SVC_NAME and unique within the pod. Each named port in a pod must have a unique name. Name for the port that can be referred to by services.",
                                    "type": "string"
                                },
                                "protocol": {
                                    "description": "Protocol for port. Must be UDP, TCP, or SCTP. Defaults to \"TCP\".",
                                    "type": "string"
                                }
                            },
                            "required": [
                                "containerPort"
                            ],
                            "type": "object"
                        },
                        "type": "array",
                        "x-kubernetes-list-map-keys": [
                            "containerPort",
                            "protocol"
                        ],
                        "x-kubernetes-list-type": "map",
                        "x-kubernetes-patch-merge-key": "containerPort",
                        "x-kubernetes-patch-strategy": "merge"
                    },
                    "readinessProbe": {
                        "description": "Probe describes a health check to be performed against a container to determine whether it is alive or ready to receive traffic.",
                        "properties": {
                            "exec": {
                                "description": "ExecAction describes a \"run in container\" action.",
                                "properties": {
                                    "command": {
                                        "description": "Command is the command line to execute inside the container, the working directory for the command  is root ('/') in the container's filesystem. The command is simply exec'd, it is not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.",
                                        "items": {
                                            "type": "string"
                                        },
                                        "type": "array"
                                    }
                                },
                                "type": "object"
                            },
                            "failureThreshold": {
                                "description": "Minimum consecutive failures for the probe to be considered failed after having succeeded. Defaults to 3. Minimum value is 1.",
                                "format": "int32",
                                "type": "integer"
                            },
                            "httpGet": {
                                "description": "HTTPGetAction describes an action based on HTTP Get requests.",
                                "properties": {
                                    "host": {
                                        "description": "Host name to connect to, defaults to the pod IP. You probably want to set \"Host\" in httpHeaders instead.",
                                        "type": "string"
                                    },
                                    "httpHeaders": {
                                        "description": "Custom headers to set in the request. HTTP allows repeated headers.",
                                        "items": {
                                            "description": "HTTPHeader describes a custom header to be used in HTTP probes",
                                            "properties": {
                                                "name": {
                                                    "description": "The header field name",
                                                    "type": "string"
                                                },
                                                "value": {
                                                    "description": "The header field value",
                                                    "type": "string"
                                                }
                                            },
                                            "required": [
                                                "name",
                                                "value"
                                            ],
                                            "type": "object"
                                        },
                                        "type": "array"
                                    },
                                    "path": {
                                        "description": "Path to access on the HTTP server.",
                                        "type": "string"
                                    },
                                    "port": {
                                        "description": "IntOrString is a type that can hold an int32 or a string.  When used in JSON or YAML marshalling and unmarshalling, it produces or consumes the inner type.  This allows you to have, for example, a JSON field that can accept a name or number.",
                                        "format": "int-or-string",
                                        "type": "string"
                                    },
                                    "scheme": {
                                        "description": "Scheme to use for connecting to the host. Defaults to HTTP.",
                                        "type": "string"
                                    }
                                },
                                "required": [
                                    "port"
                                ],
                                "type": "object"
                            },
                            "initialDelaySeconds": {
                                "description": "Number of seconds after the container has started before liveness probes are initiated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes",
                                "format": "int32",
                                "type": "integer"
                            },
                            "periodSeconds": {
                                "description": "How often (in seconds) to perform the probe. Default to 10 seconds. Minimum value is 1.",
                                "format": "int32",
                                "type": "integer"
                            },
                            "successThreshold": {
                                "description": "Minimum consecutive successes for the probe to be considered successful after having failed. Defaults to 1. Must be 1 for liveness. Minimum value is 1.",
                                "format": "int32",
                                "type": "integer"
                            },
                            "tcpSocket": {
                                "description": "TCPSocketAction describes an action based on opening a socket",
                                "properties": {
                                    "host": {
                                        "description": "Optional: Host name to connect to, defaults to the pod IP.",
                                        "type": "string"
                                    },
                                    "port": {
                                        "description": "IntOrString is a type that can hold an int32 or a string.  When used in JSON or YAML marshalling and unmarshalling, it produces or consumes the inner type.  This allows you to have, for example, a JSON field that can accept a name or number.",
                                        "format": "int-or-string",
                                        "type": "string"
                                    }
                                },
                                "required": [
                                    "port"
                                ],
                                "type": "object"
                            },
                            "timeoutSeconds": {
                                "description": "Number of seconds after which the probe times out. Defaults to 1 second. Minimum value is 1. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes",
                                "format": "int32",
                                "type": "integer"
                            }
                        },
                        "type": "object"
                    },
                    "resources": {
                        "description": "ResourceRequirements describes the compute resource requirements.",
                        "properties": {
                            "limits": {
                                "additionalProperties": true,
                                "description": "Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/",
                                "type": "object"
                            },
                            "requests": {
                                "additionalProperties": true,
                                "description": "Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/",
                                "type": "object"
                            }
                        },
                        "type": "object"
                    },
                    "securityContext": {
                        "description": "SecurityContext holds security configuration that will be applied to a container. Some fields are present in both SecurityContext and PodSecurityContext.  When both are set, the values in SecurityContext take precedence.",
                        "properties": {
                            "allowPrivilegeEscalation": {
                                "description": "AllowPrivilegeEscalation controls whether a process can gain more privileges than its parent process. This bool directly controls if the no_new_privs flag will be set on the container process. AllowPrivilegeEscalation is true always when the container is: 1) run as Privileged 2) has CAP_SYS_ADMIN",
                                "type": "boolean"
                            },
                            "capabilities": {
                                "description": "Adds and removes POSIX capabilities from running containers.",
                                "properties": {
                                    "add": {
                                        "description": "Added capabilities",
                                        "items": {
                                            "type": "string"
                                        },
                                        "type": "array"
                                    },
                                    "drop": {
                                        "description": "Removed capabilities",
                                        "items": {
                                            "type": "string"
                                        },
                                        "type": "array"
                                    }
                                },
                                "type": "object"
                            },
                            "privileged": {
                                "description": "Run container in privileged mode. Processes in privileged containers are essentially equivalent to root on the host. Defaults to false.",
                                "type": "boolean"
                            },
                            "procMount": {
                                "description": "procMount denotes the type of proc mount to use for the containers. The default is DefaultProcMount which uses the container runtime defaults for readonly paths and masked paths. This requires the ProcMountType feature flag to be enabled.",
                                "type": "string"
                            },
                            "readOnlyRootFilesystem": {
                                "description": "Whether this container has a read-only root filesystem. Default is false.",
                                "type": "boolean"
                            },
                            "runAsGroup": {
                                "description": "The GID to run the entrypoint of the container process. Uses runtime default if unset. May also be set in PodSecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence.",
                                "format": "int64",
                                "type": "integer"
                            },
                            "runAsNonRoot": {
                                "description": "Indicates that the container must run as a non-root user. If true, the Kubelet will validate the image at runtime to ensure that it does not run as UID 0 (root) and fail to start the container if it does. If unset or false, no such validation will be performed. May also be set in PodSecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence.",
                                "type": "boolean"
                            },
                            "runAsUser": {
                                "description": "The UID to run the entrypoint of the container process. Defaults to user specified in image metadata if unspecified. May also be set in PodSecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence.",
                                "format": "int64",
                                "type": "integer"
                            },
                            "seLinuxOptions": {
                                "description": "SELinuxOptions are the labels to be applied to the container",
                                "properties": {
                                    "level": {
                                        "description": "Level is SELinux level label that applies to the container.",
                                        "type": "string"
                                    },
                                    "role": {
                                        "description": "Role is a SELinux role label that applies to the container.",
                                        "type": "string"
                                    },
                                    "type": {
                                        "description": "Type is a SELinux type label that applies to the container.",
                                        "type": "string"
                                    },
                                    "user": {
                                        "description": "User is a SELinux user label that applies to the container.",
                                        "type": "string"
                                    }
                                },
                                "type": "object"
                            }
                        },
                        "type": "object"
                    },
                    "stdin": {
                        "description": "Whether this container should allocate a buffer for stdin in the container runtime. If this is not set, reads from stdin in the container will always result in EOF. Default is false.",
                        "type": "boolean"
                    },
                    "stdinOnce": {
                        "description": "Whether the container runtime should close the stdin channel after it has been opened by a single attach. When stdin is true the stdin stream will remain open across multiple attach sessions. If stdinOnce is set to true, stdin is opened on container start, is empty until the first client attaches to stdin, and then remains open and accepts data until the client disconnects, at which time stdin is closed and remains closed until the container is restarted. If this flag is false, a container processes that reads from stdin will never receive an EOF. Default is false",
                        "type": "boolean"
                    },
                    "terminationMessagePath": {
                        "description": "Optional: Path at which the file to which the container's termination message will be written is mounted into the container's filesystem. Message written is intended to be brief final status, such as an assertion failure message. Will be truncated by the node if greater than 4096 bytes. The total message length across all containers will be limited to 12kb. Defaults to /dev/termination-log. Cannot be updated.",
                        "type": "string"
                    },
                    "terminationMessagePolicy": {
                        "description": "Indicate how the termination message should be populated. File will use the contents of terminationMessagePath to populate the container status message on both success and failure. FallbackToLogsOnError will use the last chunk of container log output if the termination message file is empty and the container exited with an error. The log output is limited to 2048 bytes or 80 lines, whichever is smaller. Defaults to File. Cannot be updated.",
                        "type": "string"
                    },
                    "tty": {
                        "description": "Whether this container should allocate a TTY for itself, also requires 'stdin' to be true. Default is false.",
                        "type": "boolean"
                    },
                    "volumeDevices": {
                        "description": "volumeDevices is the list of block devices to be used by the container. This is a beta feature.",
                        "items": {
                            "description": "volumeDevice describes a mapping of a raw block device within a container.",
                            "properties": {
                                "devicePath": {
                                    "description": "devicePath is the path inside of the container that the device will be mapped to.",
                                    "type": "string"
                                },
                                "name": {
                                    "description": "name must match the name of a persistentVolumeClaim in the pod",
                                    "type": "string"
                                }
                            },
                            "required": [
                                "name",
                                "devicePath"
                            ],
                            "type": "object"
                        },
                        "type": "array",
                        "x-kubernetes-patch-merge-key": "devicePath",
                        "x-kubernetes-patch-strategy": "merge"
                    },
                    "volumeMounts": {
                        "description": "Pod volumes to mount into the container's filesystem. Cannot be updated.",
                        "items": {
                            "description": "VolumeMount describes a mounting of a Volume within a container.",
                            "properties": {
                                "mountPath": {
                                    "description": "Path within the container at which the volume should be mounted.  Must not contain ':'.",
                                    "type": "string"
                                },
                                "mountPropagation": {
                                    "description": "mountPropagation determines how mounts are propagated from the host to container and the other way around. When not set, MountPropagationNone is used. This field is beta in 1.10.",
                                    "type": "string"
                                },
                                "name": {
                                    "description": "This must match the Name of a Volume.",
                                    "type": "string"
                                },
                                "readOnly": {
                                    "description": "Mounted read-only if true, read-write otherwise (false or unspecified). Defaults to false.",
                                    "type": "boolean"
                                },
                                "subPath": {
                                    "description": "Path within the volume from which the container's volume should be mounted. Defaults to \"\" (volume's root).",
                                    "type": "string"
                                },
                                "subPathExpr": {
                                    "description": "Expanded path within the volume from which the container's volume should be mounted. Behaves similarly to SubPath but environment variable references $(VAR_NAME) are expanded using the container's environment. Defaults to \"\" (volume's root). SubPathExpr and SubPath are mutually exclusive. This field is alpha in 1.14.",
                                    "type": "string"
                                }
                            },
                            "required": [
                                "name",
                                "mountPath"
                            ],
                            "type": "object"
                        },
                        "type": "array",
                        "x-kubernetes-patch-merge-key": "mountPath",
                        "x-kubernetes-patch-strategy": "merge"
                    },
                    "workingDir": {
                        "description": "Container's working directory. If not specified, the container runtime's default will be used, which might be configured in the container image. Cannot be updated.",
                        "type": "string"
                    }
                },
                "required": [
                    "name"
                ],
                "type": "object"
            },
            "type": "array",
            "x-kubernetes-patch-merge-key": "name",
            "x-kubernetes-patch-strategy": "merge"
        },
        "dnsConfig": {
            "description": "PodDNSConfig defines the DNS parameters of a pod in addition to those generated from DNSPolicy.",
            "properties": {
                "nameservers": {
                    "description": "A list of DNS name server IP addresses. This will be appended to the base nameservers generated from DNSPolicy. Duplicated nameservers will be removed.",
                    "items": {
                        "type": "string"
                    },
                    "type": "array"
                },
                "options": {
                    "description": "A list of DNS resolver options. This will be merged with the base options generated from DNSPolicy. Duplicated entries will be removed. Resolution options given in Options will override those that appear in the base DNSPolicy.",
                    "items": {
                        "description": "PodDNSConfigOption defines DNS resolver options of a pod.",
                        "properties": {
                            "name": {
                                "description": "Required.",
                                "type": "string"
                            },
                            "value": {
                                "type": "string"
                            }
                        },
                        "type": "object"
                    },
                    "type": "array"
                },
                "searches": {
                    "description": "A list of DNS search domains for host-name lookup. This will be appended to the base search paths generated from DNSPolicy. Duplicated search paths will be removed.",
                    "items": {
                        "type": "string"
                    },
                    "type": "array"
                }
            },
            "type": "object"
        },
        "dnsPolicy": {
            "description": "Set DNS policy for the pod. Defaults to \"ClusterFirst\". Valid values are 'ClusterFirstWithHostNet', 'ClusterFirst', 'Default' or 'None'. DNS parameters given in DNSConfig will be merged with the policy selected with DNSPolicy. To have DNS options set along with hostNetwork, you have to specify DNS policy explicitly to 'ClusterFirstWithHostNet'.",
            "type": "string"
        },
        "enableServiceLinks": {
            "description": "EnableServiceLinks indicates whether information about services should be injected into pod's environment variables, matching the syntax of Docker links. Optional: Defaults to true.",
            "type": "boolean"
        },
        "hostAliases": {
            "description": "HostAliases is an optional list of hosts and IPs that will be injected into the pod's hosts file if specified. This is only valid for non-hostNetwork pods.",
            "items": {
                "description": "HostAlias holds the mapping between IP and hostnames that will be injected as an entry in the pod's hosts file.",
                "properties": {
                    "hostnames": {
                        "description": "Hostnames for the above IP address.",
                        "items": {
                            "type": "string"
                        },
                        "type": "array"
                    },
                    "ip": {
                        "description": "IP address of the host file entry.",
                        "type": "string"
                    }
                },
                "type": "object"
            },
            "type": "array",
            "x-kubernetes-patch-merge-key": "ip",
            "x-kubernetes-patch-strategy": "merge"
        },
        "hostIPC": {
            "description": "Use the host's ipc namespace. Optional: Default to false.",
            "type": "boolean"
        },
        "hostNetwork": {
            "description": "Host networking requested for this pod. Use the host's network namespace. If this option is set, the ports that will be used must be specified. Default to false.",
            "type": "boolean"
        },
        "hostPID": {
            "description": "Use the host's pid namespace. Optional: Default to false.",
            "type": "boolean"
        },
        "hostname": {
            "description": "Specifies the hostname of the Pod If not specified, the pod's hostname will be set to a system-defined value.",
            "type": "string"
        },
        "imagePullSecrets": {
            "description": "ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec. If specified, these secrets will be passed to individual puller implementations for them to use. For example, in the case of docker, only DockerConfig type secrets are honored. More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod",
            "items": {
                "description": "LocalObjectReference contains enough information to let you locate the referenced object inside the same namespace.",
                "properties": {
                    "name": {
                        "description": "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
                        "type": "string"
                    }
                },
                "type": "object"
            },
            "type": "array",
            "x-kubernetes-patch-merge-key": "name",
            "x-kubernetes-patch-strategy": "merge"
        },
        "initContainers": {
            "description": "List of initialization containers belonging to the pod. Init containers are executed in order prior to containers being started. If any init container fails, the pod is considered to have failed and is handled according to its restartPolicy. The name for an init container or normal container must be unique among all containers. Init containers may not have Lifecycle actions, Readiness probes, or Liveness probes. The resourceRequirements of an init container are taken into account during scheduling by finding the highest request/limit for each resource type, and then using the max of of that value or the sum of the normal containers. Limits are applied to init containers in a similar fashion. Init containers cannot currently be added or removed. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/",
            "items": {
                "description": "A single application container that you want to run within a pod.",
                "properties": {
                    "args": {
                        "description": "Arguments to the entrypoint. The docker image's CMD is used if this is not provided. Variable references $(VAR_NAME) are expanded using the container's environment. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Cannot be updated. More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell",
                        "items": {
                            "type": "string"
                        },
                        "type": "array"
                    },
                    "command": {
                        "description": "Entrypoint array. Not executed within a shell. The docker image's ENTRYPOINT is used if this is not provided. Variable references $(VAR_NAME) are expanded using the container's environment. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Cannot be updated. More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell",
                        "items": {
                            "type": "string"
                        },
                        "type": "array"
                    },
                    "env": {
                        "description": "List of environment variables to set in the container. Cannot be updated.",
                        "items": {
                            "description": "EnvVar represents an environment variable present in a Container.",
                            "properties": {
                                "name": {
                                    "description": "Name of the environment variable. Must be a C_IDENTIFIER.",
                                    "type": "string"
                                },
                                "value": {
                                    "description": "Variable references $(VAR_NAME) are expanded using the previous defined environment variables in the container and any service environment variables. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Defaults to \"\".",
                                    "type": "string"
                                },
                                "valueFrom": {
                                    "description": "EnvVarSource represents a source for the value of an EnvVar.",
                                    "properties": {
                                        "configMapKeyRef": {
                                            "description": "Selects a key from a ConfigMap.",
                                            "properties": {
                                                "key": {
                                                    "description": "The key to select.",
                                                    "type": "string"
                                                },
                                                "name": {
                                                    "description": "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
                                                    "type": "string"
                                                },
                                                "optional": {
                                                    "description": "Specify whether the ConfigMap or it's key must be defined",
                                                    "type": "boolean"
                                                }
                                            },
                                            "required": [
                                                "key"
                                            ],
                                            "type": "object"
                                        },
                                        "fieldRef": {
                                            "description": "ObjectFieldSelector selects an APIVersioned field of an object.",
                                            "properties": {
                                                "apiVersion": {
                                                    "description": "Version of the schema the FieldPath is written in terms of, defaults to \"v1\".",
                                                    "type": "string"
                                                },
                                                "fieldPath": {
                                                    "description": "Path of the field to select in the specified API version.",
                                                    "type": "string"
                                                }
                                            },
                                            "required": [
                                                "fieldPath"
                                            ],
                                            "type": "object"
                                        },
                                        "resourceFieldRef": {
                                            "description": "ResourceFieldSelector represents container resources (cpu, memory) and their output format",
                                            "properties": {
                                                "containerName": {
                                                    "description": "Container name: required for volumes, optional for env vars",
                                                    "type": "string"
                                                },
                                                "divisor": {
                                                    "description": "Quantity is a fixed-point representation of a number. It provides convenient marshaling/unmarshaling in JSON and YAML, in addition to String() and Int64() accessors.\n\nThe serialization format is:\n\n<quantity>        ::= <signedNumber><suffix>\n  (Note that <suffix> may be empty, from the \"\" case in <decimalSI>.)\n<digit>           ::= 0 | 1 | ... | 9 <digits>          ::= <digit> | <digit><digits> <number>          ::= <digits> | <digits>.<digits> | <digits>. | .<digits> <sign>            ::= \"+\" | \"-\" <signedNumber>    ::= <number> | <sign><number> <suffix>          ::= <binarySI> | <decimalExponent> | <decimalSI> <binarySI>        ::= Ki | Mi | Gi | Ti | Pi | Ei\n  (International System of units; See: http://physics.nist.gov/cuu/Units/binary.html)\n<decimalSI>       ::= m | \"\" | k | M | G | T | P | E\n  (Note that 1024 = 1Ki but 1000 = 1k; I didn't choose the capitalization.)\n<decimalExponent> ::= \"e\" <signedNumber> | \"E\" <signedNumber>\n\nNo matter which of the three exponent forms is used, no quantity may represent a number greater than 2^63-1 in magnitude, nor may it have more than 3 decimal places. Numbers larger or more precise will be capped or rounded up. (E.g.: 0.1m will rounded up to 1m.) This may be extended in the future if we require larger or smaller quantities.\n\nWhen a Quantity is parsed from a string, it will remember the type of suffix it had, and will use the same type again when it is serialized.\n\nBefore serializing, Quantity will be put in \"canonical form\". This means that Exponent/suffix will be adjusted up or down (with a corresponding increase or decrease in Mantissa) such that:\n  a. No precision is lost\n  b. No fractional digits will be emitted\n  c. The exponent (or suffix) is as large as possible.\nThe sign will be omitted unless the number is negative.\n\nExamples:\n  1.5 will be serialized as \"1500m\"\n  1.5Gi will be serialized as \"1536Mi\"\n\nNote that the quantity will NEVER be internally represented by a floating point number. That is the whole point of this exercise.\n\nNon-canonical values will still parse as long as they are well formed, but will be re-emitted in their canonical form. (So always use canonical form, or don't diff.)\n\nThis format is intended to make it difficult to use these numbers without writing some sort of special handling code in the hopes that that will cause implementors to also use a fixed point implementation.",
                                                    "type": "string"
                                                },
                                                "resource": {
                                                    "description": "Required: resource to select",
                                                    "type": "string"
                                                }
                                            },
                                            "required": [
                                                "resource"
                                            ],
                                            "type": "object"
                                        },
                                        "secretKeyRef": {
                                            "description": "SecretKeySelector selects a key of a Secret.",
                                            "properties": {
                                                "key": {
                                                    "description": "The key of the secret to select from.  Must be a valid secret key.",
                                                    "type": "string"
                                                },
                                                "name": {
                                                    "description": "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
                                                    "type": "string"
                                                },
                                                "optional": {
                                                    "description": "Specify whether the Secret or it's key must be defined",
                                                    "type": "boolean"
                                                }
                                            },
                                            "required": [
                                                "key"
                                            ],
                                            "type": "object"
                                        }
                                    },
                                    "type": "object"
                                }
                            },
                            "required": [
                                "name"
                            ],
                            "type": "object"
                        },
                        "type": "array",
                        "x-kubernetes-patch-merge-key": "name",
                        "x-kubernetes-patch-strategy": "merge"
                    },
                    "envFrom": {
                        "description": "List of sources to populate environment variables in the container. The keys defined within a source must be a C_IDENTIFIER. All invalid keys will be reported as an event when the container is starting. When a key exists in multiple sources, the value associated with the last source will take precedence. Values defined by an Env with a duplicate key will take precedence. Cannot be updated.",
                        "items": {
                            "description": "EnvFromSource represents the source of a set of ConfigMaps",
                            "properties": {
                                "configMapRef": {
                                    "description": "ConfigMapEnvSource selects a ConfigMap to populate the environment variables with.\n\nThe contents of the target ConfigMap's Data field will represent the key-value pairs as environment variables.",
                                    "properties": {
                                        "name": {
                                            "description": "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
                                            "type": "string"
                                        },
                                        "optional": {
                                            "description": "Specify whether the ConfigMap must be defined",
                                            "type": "boolean"
                                        }
                                    },
                                    "type": "object"
                                },
                                "prefix": {
                                    "description": "An optional identifier to prepend to each key in the ConfigMap. Must be a C_IDENTIFIER.",
                                    "type": "string"
                                },
                                "secretRef": {
                                    "description": "SecretEnvSource selects a Secret to populate the environment variables with.\n\nThe contents of the target Secret's Data field will represent the key-value pairs as environment variables.",
                                    "properties": {
                                        "name": {
                                            "description": "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
                                            "type": "string"
                                        },
                                        "optional": {
                                            "description": "Specify whether the Secret must be defined",
                                            "type": "boolean"
                                        }
                                    },
                                    "type": "object"
                                }
                            },
                            "type": "object"
                        },
                        "type": "array"
                    },
                    "image": {
                        "description": "Docker image name. More info: https://kubernetes.io/docs/concepts/containers/images This field is optional to allow higher level config management to default or override container images in workload controllers like Deployments and StatefulSets.",
                        "type": "string"
                    },
                    "imagePullPolicy": {
                        "description": "Image pull policy. One of Always, Never, IfNotPresent. Defaults to Always if :latest tag is specified, or IfNotPresent otherwise. Cannot be updated. More info: https://kubernetes.io/docs/concepts/containers/images#updating-images",
                        "type": "string"
                    },
                    "lifecycle": {
                        "description": "Lifecycle describes actions that the management system should take in response to container lifecycle events. For the PostStart and PreStop lifecycle handlers, management of the container blocks until the action is complete, unless the container process fails, in which case the handler is aborted.",
                        "properties": {
                            "postStart": {
                                "description": "Handler defines a specific action that should be taken",
                                "properties": {
                                    "exec": {
                                        "description": "ExecAction describes a \"run in container\" action.",
                                        "properties": {
                                            "command": {
                                                "description": "Command is the command line to execute inside the container, the working directory for the command  is root ('/') in the container's filesystem. The command is simply exec'd, it is not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.",
                                                "items": {
                                                    "type": "string"
                                                },
                                                "type": "array"
                                            }
                                        },
                                        "type": "object"
                                    },
                                    "httpGet": {
                                        "description": "HTTPGetAction describes an action based on HTTP Get requests.",
                                        "properties": {
                                            "host": {
                                                "description": "Host name to connect to, defaults to the pod IP. You probably want to set \"Host\" in httpHeaders instead.",
                                                "type": "string"
                                            },
                                            "httpHeaders": {
                                                "description": "Custom headers to set in the request. HTTP allows repeated headers.",
                                                "items": {
                                                    "description": "HTTPHeader describes a custom header to be used in HTTP probes",
                                                    "properties": {
                                                        "name": {
                                                            "description": "The header field name",
                                                            "type": "string"
                                                        },
                                                        "value": {
                                                            "description": "The header field value",
                                                            "type": "string"
                                                        }
                                                    },
                                                    "required": [
                                                        "name",
                                                        "value"
                                                    ],
                                                    "type": "object"
                                                },
                                                "type": "array"
                                            },
                                            "path": {
                                                "description": "Path to access on the HTTP server.",
                                                "type": "string"
                                            },
                                            "port": {
                                                "description": "IntOrString is a type that can hold an int32 or a string.  When used in JSON or YAML marshalling and unmarshalling, it produces or consumes the inner type.  This allows you to have, for example, a JSON field that can accept a name or number.",
                                                "format": "int-or-string",
                                                "type": "string"
                                            },
                                            "scheme": {
                                                "description": "Scheme to use for connecting to the host. Defaults to HTTP.",
                                                "type": "string"
                                            }
                                        },
                                        "required": [
                                            "port"
                                        ],
                                        "type": "object"
                                    },
                                    "tcpSocket": {
                                        "description": "TCPSocketAction describes an action based on opening a socket",
                                        "properties": {
                                            "host": {
                                                "description": "Optional: Host name to connect to, defaults to the pod IP.",
                                                "type": "string"
                                            },
                                            "port": {
                                                "description": "IntOrString is a type that can hold an int32 or a string.  When used in JSON or YAML marshalling and unmarshalling, it produces or consumes the inner type.  This allows you to have, for example, a JSON field that can accept a name or number.",
                                                "format": "int-or-string",
                                                "type": "string"
                                            }
                                        },
                                        "required": [
                                            "port"
                                        ],
                                        "type": "object"
                                    }
                                },
                                "type": "object"
                            },
                            "preStop": {
                                "description": "Handler defines a specific action that should be taken",
                                "properties": {
                                    "exec": {
                                        "description": "ExecAction describes a \"run in container\" action.",
                                        "properties": {
                                            "command": {
                                                "description": "Command is the command line to execute inside the container, the working directory for the command  is root ('/') in the container's filesystem. The command is simply exec'd, it is not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.",
                                                "items": {
                                                    "type": "string"
                                                },
                                                "type": "array"
                                            }
                                        },
                                        "type": "object"
                                    },
                                    "httpGet": {
                                        "description": "HTTPGetAction describes an action based on HTTP Get requests.",
                                        "properties": {
                                            "host": {
                                                "description": "Host name to connect to, defaults to the pod IP. You probably want to set \"Host\" in httpHeaders instead.",
                                                "type": "string"
                                            },
                                            "httpHeaders": {
                                                "description": "Custom headers to set in the request. HTTP allows repeated headers.",
                                                "items": {
                                                    "description": "HTTPHeader describes a custom header to be used in HTTP probes",
                                                    "properties": {
                                                        "name": {
                                                            "description": "The header field name",
                                                            "type": "string"
                                                        },
                                                        "value": {
                                                            "description": "The header field value",
                                                            "type": "string"
                                                        }
                                                    },
                                                    "required": [
                                                        "name",
                                                        "value"
                                                    ],
                                                    "type": "object"
                                                },
                                                "type": "array"
                                            },
                                            "path": {
                                                "description": "Path to access on the HTTP server.",
                                                "type": "string"
                                            },
                                            "port": {
                                                "description": "IntOrString is a type that can hold an int32 or a string.  When used in JSON or YAML marshalling and unmarshalling, it produces or consumes the inner type.  This allows you to have, for example, a JSON field that can accept a name or number.",
                                                "format": "int-or-string",
                                                "type": "string"
                                            },
                                            "scheme": {
                                                "description": "Scheme to use for connecting to the host. Defaults to HTTP.",
                                                "type": "string"
                                            }
                                        },
                                        "required": [
                                            "port"
                                        ],
                                        "type": "object"
                                    },
                                    "tcpSocket": {
                                        "description": "TCPSocketAction describes an action based on opening a socket",
                                        "properties": {
                                            "host": {
                                                "description": "Optional: Host name to connect to, defaults to the pod IP.",
                                                "type": "string"
                                            },
                                            "port": {
                                                "description": "IntOrString is a type that can hold an int32 or a string.  When used in JSON or YAML marshalling and unmarshalling, it produces or consumes the inner type.  This allows you to have, for example, a JSON field that can accept a name or number.",
                                                "format": "int-or-string",
                                                "type": "string"
                                            }
                                        },
                                        "required": [
                                            "port"
                                        ],
                                        "type": "object"
                                    }
                                },
                                "type": "object"
                            }
                        },
                        "type": "object"
                    },
                    "livenessProbe": {
                        "description": "Probe describes a health check to be performed against a container to determine whether it is alive or ready to receive traffic.",
                        "properties": {
                            "exec": {
                                "description": "ExecAction describes a \"run in container\" action.",
                                "properties": {
                                    "command": {
                                        "description": "Command is the command line to execute inside the container, the working directory for the command  is root ('/') in the container's filesystem. The command is simply exec'd, it is not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.",
                                        "items": {
                                            "type": "string"
                                        },
                                        "type": "array"
                                    }
                                },
                                "type": "object"
                            },
                            "failureThreshold": {
                                "description": "Minimum consecutive failures for the probe to be considered failed after having succeeded. Defaults to 3. Minimum value is 1.",
                                "format": "int32",
                                "type": "integer"
                            },
                            "httpGet": {
                                "description": "HTTPGetAction describes an action based on HTTP Get requests.",
                                "properties": {
                                    "host": {
                                        "description": "Host name to connect to, defaults to the pod IP. You probably want to set \"Host\" in httpHeaders instead.",
                                        "type": "string"
                                    },
                                    "httpHeaders": {
                                        "description": "Custom headers to set in the request. HTTP allows repeated headers.",
                                        "items": {
                                            "description": "HTTPHeader describes a custom header to be used in HTTP probes",
                                            "properties": {
                                                "name": {
                                                    "description": "The header field name",
                                                    "type": "string"
                                                },
                                                "value": {
                                                    "description": "The header field value",
                                                    "type": "string"
                                                }
                                            },
                                            "required": [
                                                "name",
                                                "value"
                                            ],
                                            "type": "object"
                                        },
                                        "type": "array"
                                    },
                                    "path": {
                                        "description": "Path to access on the HTTP server.",
                                        "type": "string"
                                    },
                                    "port": {
                                        "description": "IntOrString is a type that can hold an int32 or a string.  When used in JSON or YAML marshalling and unmarshalling, it produces or consumes the inner type.  This allows you to have, for example, a JSON field that can accept a name or number.",
                                        "format": "int-or-string",
                                        "type": "string"
                                    },
                                    "scheme": {
                                        "description": "Scheme to use for connecting to the host. Defaults to HTTP.",
                                        "type": "string"
                                    }
                                },
                                "required": [
                                    "port"
                                ],
                                "type": "object"
                            },
                            "initialDelaySeconds": {
                                "description": "Number of seconds after the container has started before liveness probes are initiated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes",
                                "format": "int32",
                                "type": "integer"
                            },
                            "periodSeconds": {
                                "description": "How often (in seconds) to perform the probe. Default to 10 seconds. Minimum value is 1.",
                                "format": "int32",
                                "type": "integer"
                            },
                            "successThreshold": {
                                "description": "Minimum consecutive successes for the probe to be considered successful after having failed. Defaults to 1. Must be 1 for liveness. Minimum value is 1.",
                                "format": "int32",
                                "type": "integer"
                            },
                            "tcpSocket": {
                                "description": "TCPSocketAction describes an action based on opening a socket",
                                "properties": {
                                    "host": {
                                        "description": "Optional: Host name to connect to, defaults to the pod IP.",
                                        "type": "string"
                                    },
                                    "port": {
                                        "description": "IntOrString is a type that can hold an int32 or a string.  When used in JSON or YAML marshalling and unmarshalling, it produces or consumes the inner type.  This allows you to have, for example, a JSON field that can accept a name or number.",
                                        "format": "int-or-string",
                                        "type": "string"
                                    }
                                },
                                "required": [
                                    "port"
                                ],
                                "type": "object"
                            },
                            "timeoutSeconds": {
                                "description": "Number of seconds after which the probe times out. Defaults to 1 second. Minimum value is 1. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes",
                                "format": "int32",
                                "type": "integer"
                            }
                        },
                        "type": "object"
                    },
                    "name": {
                        "description": "Name of the container specified as a DNS_LABEL. Each container in a pod must have a unique name (DNS_LABEL). Cannot be updated.",
                        "type": "string"
                    },
                    "ports": {
                        "description": "List of ports to expose from the container. Exposing a port here gives the system additional information about the network connections a container uses, but is primarily informational. Not specifying a port here DOES NOT prevent that port from being exposed. Any port which is listening on the default \"0.0.0.0\" address inside a container will be accessible from the network. Cannot be updated.",
                        "items": {
                            "description": "ContainerPort represents a network port in a single container.",
                            "properties": {
                                "containerPort": {
                                    "description": "Number of port to expose on the pod's IP address. This must be a valid port number, 0 < x < 65536.",
                                    "format": "int32",
                                    "type": "integer"
                                },
                                "hostIP": {
                                    "description": "What host IP to bind the external port to.",
                                    "type": "string"
                                },
                                "hostPort": {
                                    "description": "Number of port to expose on the host. If specified, this must be a valid port number, 0 < x < 65536. If HostNetwork is specified, this must match ContainerPort. Most containers do not need this.",
                                    "format": "int32",
                                    "type": "integer"
                                },
                                "name": {
                                    "description": "If specified, this must be an IANA_SVC_NAME and unique within the pod. Each named port in a pod must have a unique name. Name for the port that can be referred to by services.",
                                    "type": "string"
                                },
                                "protocol": {
                                    "description": "Protocol for port. Must be UDP, TCP, or SCTP. Defaults to \"TCP\".",
                                    "type": "string"
                                }
                            },
                            "required": [
                                "containerPort"
                            ],
                            "type": "object"
                        },
                        "type": "array",
                        "x-kubernetes-list-map-keys": [
                            "containerPort",
                            "protocol"
                        ],
                        "x-kubernetes-list-type": "map",
                        "x-kubernetes-patch-merge-key": "containerPort",
                        "x-kubernetes-patch-strategy": "merge"
                    },
                    "readinessProbe": {
                        "description": "Probe describes a health check to be performed against a container to determine whether it is alive or ready to receive traffic.",
                        "properties": {
                            "exec": {
                                "description": "ExecAction describes a \"run in container\" action.",
                                "properties": {
                                    "command": {
                                        "description": "Command is the command line to execute inside the container, the working directory for the command  is root ('/') in the container's filesystem. The command is simply exec'd, it is not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.",
                                        "items": {
                                            "type": "string"
                                        },
                                        "type": "array"
                                    }
                                },
                                "type": "object"
                            },
                            "failureThreshold": {
                                "description": "Minimum consecutive failures for the probe to be considered failed after having succeeded. Defaults to 3. Minimum value is 1.",
                                "format": "int32",
                                "type": "integer"
                            },
                            "httpGet": {
                                "description": "HTTPGetAction describes an action based on HTTP Get requests.",
                                "properties": {
                                    "host": {
                                        "description": "Host name to connect to, defaults to the pod IP. You probably want to set \"Host\" in httpHeaders instead.",
                                        "type": "string"
                                    },
                                    "httpHeaders": {
                                        "description": "Custom headers to set in the request. HTTP allows repeated headers.",
                                        "items": {
                                            "description": "HTTPHeader describes a custom header to be used in HTTP probes",
                                            "properties": {
                                                "name": {
                                                    "description": "The header field name",
                                                    "type": "string"
                                                },
                                                "value": {
                                                    "description": "The header field value",
                                                    "type": "string"
                                                }
                                            },
                                            "required": [
                                                "name",
                                                "value"
                                            ],
                                            "type": "object"
                                        },
                                        "type": "array"
                                    },
                                    "path": {
                                        "description": "Path to access on the HTTP server.",
                                        "type": "string"
                                    },
                                    "port": {
                                        "description": "IntOrString is a type that can hold an int32 or a string.  When used in JSON or YAML marshalling and unmarshalling, it produces or consumes the inner type.  This allows you to have, for example, a JSON field that can accept a name or number.",
                                        "format": "int-or-string",
                                        "type": "string"
                                    },
                                    "scheme": {
                                        "description": "Scheme to use for connecting to the host. Defaults to HTTP.",
                                        "type": "string"
                                    }
                                },
                                "required": [
                                    "port"
                                ],
                                "type": "object"
                            },
                            "initialDelaySeconds": {
                                "description": "Number of seconds after the container has started before liveness probes are initiated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes",
                                "format": "int32",
                                "type": "integer"
                            },
                            "periodSeconds": {
                                "description": "How often (in seconds) to perform the probe. Default to 10 seconds. Minimum value is 1.",
                                "format": "int32",
                                "type": "integer"
                            },
                            "successThreshold": {
                                "description": "Minimum consecutive successes for the probe to be considered successful after having failed. Defaults to 1. Must be 1 for liveness. Minimum value is 1.",
                                "format": "int32",
                                "type": "integer"
                            },
                            "tcpSocket": {
                                "description": "TCPSocketAction describes an action based on opening a socket",
                                "properties": {
                                    "host": {
                                        "description": "Optional: Host name to connect to, defaults to the pod IP.",
                                        "type": "string"
                                    },
                                    "port": {
                                        "description": "IntOrString is a type that can hold an int32 or a string.  When used in JSON or YAML marshalling and unmarshalling, it produces or consumes the inner type.  This allows you to have, for example, a JSON field that can accept a name or number.",
                                        "format": "int-or-string",
                                        "type": "string"
                                    }
                                },
                                "required": [
                                    "port"
                                ],
                                "type": "object"
                            },
                            "timeoutSeconds": {
                                "description": "Number of seconds after which the probe times out. Defaults to 1 second. Minimum value is 1. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes",
                                "format": "int32",
                                "type": "integer"
                            }
                        },
                        "type": "object"
                    },
                    "resources": {
                        "description": "ResourceRequirements describes the compute resource requirements.",
                        "properties": {
                            "limits": {
                                "additionalProperties": true,
                                "description": "Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/",
                                "type": "object"
                            },
                            "requests": {
                                "additionalProperties": true,
                                "description": "Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/",
                                "type": "object"
                            }
                        },
                        "type": "object"
                    },
                    "securityContext": {
                        "description": "SecurityContext holds security configuration that will be applied to a container. Some fields are present in both SecurityContext and PodSecurityContext.  When both are set, the values in SecurityContext take precedence.",
                        "properties": {
                            "allowPrivilegeEscalation": {
                                "description": "AllowPrivilegeEscalation controls whether a process can gain more privileges than its parent process. This bool directly controls if the no_new_privs flag will be set on the container process. AllowPrivilegeEscalation is true always when the container is: 1) run as Privileged 2) has CAP_SYS_ADMIN",
                                "type": "boolean"
                            },
                            "capabilities": {
                                "description": "Adds and removes POSIX capabilities from running containers.",
                                "properties": {
                                    "add": {
                                        "description": "Added capabilities",
                                        "items": {
                                            "type": "string"
                                        },
                                        "type": "array"
                                    },
                                    "drop": {
                                        "description": "Removed capabilities",
                                        "items": {
                                            "type": "string"
                                        },
                                        "type": "array"
                                    }
                                },
                                "type": "object"
                            },
                            "privileged": {
                                "description": "Run container in privileged mode. Processes in privileged containers are essentially equivalent to root on the host. Defaults to false.",
                                "type": "boolean"
                            },
                            "procMount": {
                                "description": "procMount denotes the type of proc mount to use for the containers. The default is DefaultProcMount which uses the container runtime defaults for readonly paths and masked paths. This requires the ProcMountType feature flag to be enabled.",
                                "type": "string"
                            },
                            "readOnlyRootFilesystem": {
                                "description": "Whether this container has a read-only root filesystem. Default is false.",
                                "type": "boolean"
                            },
                            "runAsGroup": {
                                "description": "The GID to run the entrypoint of the container process. Uses runtime default if unset. May also be set in PodSecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence.",
                                "format": "int64",
                                "type": "integer"
                            },
                            "runAsNonRoot": {
                                "description": "Indicates that the container must run as a non-root user. If true, the Kubelet will validate the image at runtime to ensure that it does not run as UID 0 (root) and fail to start the container if it does. If unset or false, no such validation will be performed. May also be set in PodSecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence.",
                                "type": "boolean"
                            },
                            "runAsUser": {
                                "description": "The UID to run the entrypoint of the container process. Defaults to user specified in image metadata if unspecified. May also be set in PodSecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence.",
                                "format": "int64",
                                "type": "integer"
                            },
                            "seLinuxOptions": {
                                "description": "SELinuxOptions are the labels to be applied to the container",
                                "properties": {
                                    "level": {
                                        "description": "Level is SELinux level label that applies to the container.",
                                        "type": "string"
                                    },
                                    "role": {
                                        "description": "Role is a SELinux role label that applies to the container.",
                                        "type": "string"
                                    },
                                    "type": {
                                        "description": "Type is a SELinux type label that applies to the container.",
                                        "type": "string"
                                    },
                                    "user": {
                                        "description": "User is a SELinux user label that applies to the container.",
                                        "type": "string"
                                    }
                                },
                                "type": "object"
                            }
                        },
                        "type": "object"
                    },
                    "stdin": {
                        "description": "Whether this container should allocate a buffer for stdin in the container runtime. If this is not set, reads from stdin in the container will always result in EOF. Default is false.",
                        "type": "boolean"
                    },
                    "stdinOnce": {
                        "description": "Whether the container runtime should close the stdin channel after it has been opened by a single attach. When stdin is true the stdin stream will remain open across multiple attach sessions. If stdinOnce is set to true, stdin is opened on container start, is empty until the first client attaches to stdin, and then remains open and accepts data until the client disconnects, at which time stdin is closed and remains closed until the container is restarted. If this flag is false, a container processes that reads from stdin will never receive an EOF. Default is false",
                        "type": "boolean"
                    },
                    "terminationMessagePath": {
                        "description": "Optional: Path at which the file to which the container's termination message will be written is mounted into the container's filesystem. Message written is intended to be brief final status, such as an assertion failure message. Will be truncated by the node if greater than 4096 bytes. The total message length across all containers will be limited to 12kb. Defaults to /dev/termination-log. Cannot be updated.",
                        "type": "string"
                    },
                    "terminationMessagePolicy": {
                        "description": "Indicate how the termination message should be populated. File will use the contents of terminationMessagePath to populate the container status message on both success and failure. FallbackToLogsOnError will use the last chunk of container log output if the termination message file is empty and the container exited with an error. The log output is limited to 2048 bytes or 80 lines, whichever is smaller. Defaults to File. Cannot be updated.",
                        "type": "string"
                    },
                    "tty": {
                        "description": "Whether this container should allocate a TTY for itself, also requires 'stdin' to be true. Default is false.",
                        "type": "boolean"
                    },
                    "volumeDevices": {
                        "description": "volumeDevices is the list of block devices to be used by the container. This is a beta feature.",
                        "items": {
                            "description": "volumeDevice describes a mapping of a raw block device within a container.",
                            "properties": {
                                "devicePath": {
                                    "description": "devicePath is the path inside of the container that the device will be mapped to.",
                                    "type": "string"
                                },
                                "name": {
                                    "description": "name must match the name of a persistentVolumeClaim in the pod",
                                    "type": "string"
                                }
                            },
                            "required": [
                                "name",
                                "devicePath"
                            ],
                            "type": "object"
                        },
                        "type": "array",
                        "x-kubernetes-patch-merge-key": "devicePath",
                        "x-kubernetes-patch-strategy": "merge"
                    },
                    "volumeMounts": {
                        "description": "Pod volumes to mount into the container's filesystem. Cannot be updated.",
                        "items": {
                            "description": "VolumeMount describes a mounting of a Volume within a container.",
                            "properties": {
                                "mountPath": {
                                    "description": "Path within the container at which the volume should be mounted.  Must not contain ':'.",
                                    "type": "string"
                                },
                                "mountPropagation": {
                                    "description": "mountPropagation determines how mounts are propagated from the host to container and the other way around. When not set, MountPropagationNone is used. This field is beta in 1.10.",
                                    "type": "string"
                                },
                                "name": {
                                    "description": "This must match the Name of a Volume.",
                                    "type": "string"
                                },
                                "readOnly": {
                                    "description": "Mounted read-only if true, read-write otherwise (false or unspecified). Defaults to false.",
                                    "type": "boolean"
                                },
                                "subPath": {
                                    "description": "Path within the volume from which the container's volume should be mounted. Defaults to \"\" (volume's root).",
                                    "type": "string"
                                },
                                "subPathExpr": {
                                    "description": "Expanded path within the volume from which the container's volume should be mounted. Behaves similarly to SubPath but environment variable references $(VAR_NAME) are expanded using the container's environment. Defaults to \"\" (volume's root). SubPathExpr and SubPath are mutually exclusive. This field is alpha in 1.14.",
                                    "type": "string"
                                }
                            },
                            "required": [
                                "name",
                                "mountPath"
                            ],
                            "type": "object"
                        },
                        "type": "array",
                        "x-kubernetes-patch-merge-key": "mountPath",
                        "x-kubernetes-patch-strategy": "merge"
                    },
                    "workingDir": {
                        "description": "Container's working directory. If not specified, the container runtime's default will be used, which might be configured in the container image. Cannot be updated.",
                        "type": "string"
                    }
                },
                "required": [
                    "name"
                ],
                "type": "object"
            },
            "type": "array",
            "x-kubernetes-patch-merge-key": "name",
            "x-kubernetes-patch-strategy": "merge"
        },
        "nodeName": {
            "description": "NodeName is a request to schedule this pod onto a specific node. If it is non-empty, the scheduler simply schedules this pod onto that node, assuming that it fits resource requirements.",
            "type": "string"
        },
        "nodeSelector": {
            "additionalProperties": {
                "type": "string"
            },
            "description": "NodeSelector is a selector which must be true for the pod to fit on a node. Selector which must match a node's labels for the pod to be scheduled on that node. More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/",
            "type": "object"
        },
        "priority": {
            "description": "The priority value. Various system components use this field to find the priority of the pod. When Priority Admission Controller is enabled, it prevents users from setting this field. The admission controller populates this field from PriorityClassName. The higher the value, the higher the priority.",
            "format": "int32",
            "type": "integer"
        },
        "priorityClassName": {
            "description": "If specified, indicates the pod's priority. \"system-node-critical\" and \"system-cluster-critical\" are two special keywords which indicate the highest priorities with the former being the highest priority. Any other name must be defined by creating a PriorityClass object with that name. If not specified, the pod priority will be default or zero if there is no default.",
            "type": "string"
        },
        "readinessGates": {
            "description": "If specified, all readiness gates will be evaluated for pod readiness. A pod is ready when all its containers are ready AND all conditions specified in the readiness gates have status equal to \"True\" More info: https://git.k8s.io/enhancements/keps/sig-network/0007-pod-ready%2B%2B.md",
            "items": {
                "description": "PodReadinessGate contains the reference to a pod condition",
                "properties": {
                    "conditionType": {
                        "description": "ConditionType refers to a condition in the pod's condition list with matching type.",
                        "type": "string"
                    }
                },
                "required": [
                    "conditionType"
                ],
                "type": "object"
            },
            "type": "array"
        },
        "restartPolicy": {
            "description": "Restart policy for all containers within the pod. One of Always, OnFailure, Never. Default to Always. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#restart-policy",
            "type": "string"
        },
        "runtimeClassName": {
            "description": "RuntimeClassName refers to a RuntimeClass object in the node.k8s.io group, which should be used to run this pod.  If no RuntimeClass resource matches the named class, the pod will not be run. If unset or empty, the \"legacy\" RuntimeClass will be used, which is an implicit class with an empty definition that uses the default runtime handler. More info: https://git.k8s.io/enhancements/keps/sig-node/runtime-class.md This is an alpha feature and may change in the future.",
            "type": "string"
        },
        "schedulerName": {
            "description": "If specified, the pod will be dispatched by specified scheduler. If not specified, the pod will be dispatched by default scheduler.",
            "type": "string"
        },
        "securityContext": {
            "description": "PodSecurityContext holds pod-level security attributes and common container settings. Some fields are also present in container.securityContext.  Field values of container.securityContext take precedence over field values of PodSecurityContext.",
            "properties": {
                "fsGroup": {
                    "description": "A special supplemental group that applies to all containers in a pod. Some volume types allow the Kubelet to change the ownership of that volume to be owned by the pod:\n\n1. The owning GID will be the FSGroup 2. The setgid bit is set (new files created in the volume will be owned by FSGroup) 3. The permission bits are OR'd with rw-rw----\n\nIf unset, the Kubelet will not modify the ownership and permissions of any volume.",
                    "format": "int64",
                    "type": "integer"
                },
                "runAsGroup": {
                    "description": "The GID to run the entrypoint of the container process. Uses runtime default if unset. May also be set in SecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence for that container.",
                    "format": "int64",
                    "type": "integer"
                },
                "runAsNonRoot": {
                    "description": "Indicates that the container must run as a non-root user. If true, the Kubelet will validate the image at runtime to ensure that it does not run as UID 0 (root) and fail to start the container if it does. If unset or false, no such validation will be performed. May also be set in SecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence.",
                    "type": "boolean"
                },
                "runAsUser": {
                    "description": "The UID to run the entrypoint of the container process. Defaults to user specified in image metadata if unspecified. May also be set in SecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence for that container.",
                    "format": "int64",
                    "type": "integer"
                },
                "seLinuxOptions": {
                    "description": "SELinuxOptions are the labels to be applied to the container",
                    "properties": {
                        "level": {
                            "description": "Level is SELinux level label that applies to the container.",
                            "type": "string"
                        },
                        "role": {
                            "description": "Role is a SELinux role label that applies to the container.",
                            "type": "string"
                        },
                        "type": {
                            "description": "Type is a SELinux type label that applies to the container.",
                            "type": "string"
                        },
                        "user": {
                            "description": "User is a SELinux user label that applies to the container.",
                            "type": "string"
                        }
                    },
                    "type": "object"
                },
                "supplementalGroups": {
                    "description": "A list of groups applied to the first process run in each container, in addition to the container's primary GID.  If unspecified, no groups will be added to any container.",
                    "items": {
                        "format": "int64",
                        "type": "integer"
                    },
                    "type": "array"
                },
                "sysctls": {
                    "description": "Sysctls hold a list of namespaced sysctls used for the pod. Pods with unsupported sysctls (by the container runtime) might fail to launch.",
                    "items": {
                        "description": "Sysctl defines a kernel parameter to be set",
                        "properties": {
                            "name": {
                                "description": "Name of a property to set",
                                "type": "string"
                            },
                            "value": {
                                "description": "Value of a property to set",
                                "type": "string"
                            }
                        },
                        "required": [
                            "name",
                            "value"
                        ],
                        "type": "object"
                    },
                    "type": "array"
                }
            },
            "type": "object"
        },
        "serviceAccount": {
            "description": "DeprecatedServiceAccount is a depreciated alias for ServiceAccountName. Deprecated: Use serviceAccountName instead.",
            "type": "string"
        },
        "serviceAccountName": {
            "description": "ServiceAccountName is the name of the ServiceAccount to use to run this pod. More info: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/",
            "type": "string"
        },
        "shareProcessNamespace": {
            "description": "Share a single process namespace between all of the containers in a pod. When this is set containers will be able to view and signal processes from other containers in the same pod, and the first process in each container will not be assigned PID 1. HostPID and ShareProcessNamespace cannot both be set. Optional: Default to false. This field is beta-level and may be disabled with the PodShareProcessNamespace feature.",
            "type": "boolean"
        },
        "subdomain": {
            "description": "If specified, the fully qualified Pod hostname will be \"<hostname>.<subdomain>.<pod namespace>.svc.<cluster domain>\". If not specified, the pod will not have a domainname at all.",
            "type": "string"
        },
        "terminationGracePeriodSeconds": {
            "description": "Optional duration in seconds the pod needs to terminate gracefully. May be decreased in delete request. Value must be non-negative integer. The value zero indicates delete immediately. If this value is nil, the default grace period will be used instead. The grace period is the duration in seconds after the processes running in the pod are sent a termination signal and the time when the processes are forcibly halted with a kill signal. Set this value longer than the expected cleanup time for your process. Defaults to 30 seconds.",
            "format": "int64",
            "type": "integer"
        },
        "tolerations": {
            "description": "If specified, the pod's tolerations.",
            "items": {
                "description": "The pod this Toleration is attached to tolerates any taint that matches the triple <key,value,effect> using the matching operator <operator>.",
                "properties": {
                    "effect": {
                        "description": "Effect indicates the taint effect to match. Empty means match all taint effects. When specified, allowed values are NoSchedule, PreferNoSchedule and NoExecute.",
                        "type": "string"
                    },
                    "key": {
                        "description": "Key is the taint key that the toleration applies to. Empty means match all taint keys. If the key is empty, operator must be Exists; this combination means to match all values and all keys.",
                        "type": "string"
                    },
                    "operator": {
                        "description": "Operator represents a key's relationship to the value. Valid operators are Exists and Equal. Defaults to Equal. Exists is equivalent to wildcard for value, so that a pod can tolerate all taints of a particular category.",
                        "type": "string"
                    },
                    "tolerationSeconds": {
                        "description": "TolerationSeconds represents the period of time the toleration (which must be of effect NoExecute, otherwise this field is ignored) tolerates the taint. By default, it is not set, which means tolerate the taint forever (do not evict). Zero and negative values will be treated as 0 (evict immediately) by the system.",
                        "format": "int64",
                        "type": "integer"
                    },
                    "value": {
                        "description": "Value is the taint value the toleration matches to. If the operator is Exists, the value should be empty, otherwise just a regular string.",
                        "type": "string"
                    }
                },
                "type": "object"
            },
            "type": "array"
        },
        "volumes": {
            "description": "List of volumes that can be mounted by containers belonging to the pod. More info: https://kubernetes.io/docs/concepts/storage/volumes",
            "items": {
                "description": "Volume represents a named volume in a pod that may be accessed by any container in the pod.",
                "properties": {
                    "awsElasticBlockStore": {
                        "description": "Represents a Persistent Disk resource in AWS.\n\nAn AWS EBS disk must exist before mounting to a container. The disk must also be in the same AWS zone as the kubelet. An AWS EBS disk can only be mounted as read/write once. AWS EBS volumes support ownership management and SELinux relabeling.",
                        "properties": {
                            "fsType": {
                                "description": "Filesystem type of the volume that you want to mount. Tip: Ensure that the filesystem type is supported by the host operating system. Examples: \"ext4\", \"xfs\", \"ntfs\". Implicitly inferred to be \"ext4\" if unspecified. More info: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore",
                                "type": "string"
                            },
                            "partition": {
                                "description": "The partition in the volume that you want to mount. If omitted, the default is to mount by volume name. Examples: For volume /dev/sda1, you specify the partition as \"1\". Similarly, the volume partition for /dev/sda is \"0\" (or you can leave the property empty).",
                                "format": "int32",
                                "type": "integer"
                            },
                            "readOnly": {
                                "description": "Specify \"true\" to force and set the ReadOnly property in VolumeMounts to \"true\". If omitted, the default is \"false\". More info: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore",
                                "type": "boolean"
                            },
                            "volumeID": {
                                "description": "Unique ID of the persistent disk resource in AWS (Amazon EBS volume). More info: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore",
                                "type": "string"
                            }
                        },
                        "required": [
                            "volumeID"
                        ],
                        "type": "object"
                    },
                    "azureDisk": {
                        "description": "AzureDisk represents an Azure Data Disk mount on the host and bind mount to the pod.",
                        "properties": {
                            "cachingMode": {
                                "description": "Host Caching mode: None, Read Only, Read Write.",
                                "type": "string"
                            },
                            "diskName": {
                                "description": "The Name of the data disk in the blob storage",
                                "type": "string"
                            },
                            "diskURI": {
                                "description": "The URI the data disk in the blob storage",
                                "type": "string"
                            },
                            "fsType": {
                                "description": "Filesystem type to mount. Must be a filesystem type supported by the host operating system. Ex. \"ext4\", \"xfs\", \"ntfs\". Implicitly inferred to be \"ext4\" if unspecified.",
                                "type": "string"
                            },
                            "kind": {
                                "description": "Expected values Shared: multiple blob disks per storage account  Dedicated: single blob disk per storage account  Managed: azure managed data disk (only in managed availability set). defaults to shared",
                                "type": "string"
                            },
                            "readOnly": {
                                "description": "Defaults to false (read/write). ReadOnly here will force the ReadOnly setting in VolumeMounts.",
                                "type": "boolean"
                            }
                        },
                        "required": [
                            "diskName",
                            "diskURI"
                        ],
                        "type": "object"
                    },
                    "azureFile": {
                        "description": "AzureFile represents an Azure File Service mount on the host and bind mount to the pod.",
                        "properties": {
                            "readOnly": {
                                "description": "Defaults to false (read/write). ReadOnly here will force the ReadOnly setting in VolumeMounts.",
                                "type": "boolean"
                            },
                            "secretName": {
                                "description": "the name of secret that contains Azure Storage Account Name and Key",
                                "type": "string"
                            },
                            "shareName": {
                                "description": "Share Name",
                                "type": "string"
                            }
                        },
                        "required": [
                            "secretName",
                            "shareName"
                        ],
                        "type": "object"
                    },
                    "cephfs": {
                        "description": "Represents a Ceph Filesystem mount that lasts the lifetime of a pod Cephfs volumes do not support ownership management or SELinux relabeling.",
                        "properties": {
                            "monitors": {
                                "description": "Required: Monitors is a collection of Ceph monitors More info: https://releases.k8s.io/HEAD/examples/volumes/cephfs/README.md#how-to-use-it",
                                "items": {
                                    "type": "string"
                                },
                                "type": "array"
                            },
                            "path": {
                                "description": "Optional: Used as the mounted root, rather than the full Ceph tree, default is /",
                                "type": "string"
                            },
                            "readOnly": {
                                "description": "Optional: Defaults to false (read/write). ReadOnly here will force the ReadOnly setting in VolumeMounts. More info: https://releases.k8s.io/HEAD/examples/volumes/cephfs/README.md#how-to-use-it",
                                "type": "boolean"
                            },
                            "secretFile": {
                                "description": "Optional: SecretFile is the path to key ring for User, default is /etc/ceph/user.secret More info: https://releases.k8s.io/HEAD/examples/volumes/cephfs/README.md#how-to-use-it",
                                "type": "string"
                            },
                            "secretRef": {
                                "description": "LocalObjectReference contains enough information to let you locate the referenced object inside the same namespace.",
                                "properties": {
                                    "name": {
                                        "description": "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
                                        "type": "string"
                                    }
                                },
                                "type": "object"
                            },
                            "user": {
                                "description": "Optional: User is the rados user name, default is admin More info: https://releases.k8s.io/HEAD/examples/volumes/cephfs/README.md#how-to-use-it",
                                "type": "string"
                            }
                        },
                        "required": [
                            "monitors"
                        ],
                        "type": "object"
                    },
                    "cinder": {
                        "description": "Represents a cinder volume resource in Openstack. A Cinder volume must exist before mounting to a container. The volume must also be in the same region as the kubelet. Cinder volumes support ownership management and SELinux relabeling.",
                        "properties": {
                            "fsType": {
                                "description": "Filesystem type to mount. Must be a filesystem type supported by the host operating system. Examples: \"ext4\", \"xfs\", \"ntfs\". Implicitly inferred to be \"ext4\" if unspecified. More info: https://releases.k8s.io/HEAD/examples/mysql-cinder-pd/README.md",
                                "type": "string"
                            },
                            "readOnly": {
                                "description": "Optional: Defaults to false (read/write). ReadOnly here will force the ReadOnly setting in VolumeMounts. More info: https://releases.k8s.io/HEAD/examples/mysql-cinder-pd/README.md",
                                "type": "boolean"
                            },
                            "secretRef": {
                                "description": "LocalObjectReference contains enough information to let you locate the referenced object inside the same namespace.",
                                "properties": {
                                    "name": {
                                        "description": "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
                                        "type": "string"
                                    }
                                },
                                "type": "object"
                            },
                            "volumeID": {
                                "description": "volume id used to identify the volume in cinder More info: https://releases.k8s.io/HEAD/examples/mysql-cinder-pd/README.md",
                                "type": "string"
                            }
                        },
                        "required": [
                            "volumeID"
                        ],
                        "type": "object"
                    },
                    "configMap": {
                        "description": "Adapts a ConfigMap into a volume.\n\nThe contents of the target ConfigMap's Data field will be presented in a volume as files using the keys in the Data field as the file names, unless the items element is populated with specific mappings of keys to paths. ConfigMap volumes support ownership management and SELinux relabeling.",
                        "properties": {
                            "defaultMode": {
                                "description": "Optional: mode bits to use on created files by default. Must be a value between 0 and 0777. Defaults to 0644. Directories within the path are not affected by this setting. This might be in conflict with other options that affect the file mode, like fsGroup, and the result can be other mode bits set.",
                                "format": "int32",
                                "type": "integer"
                            },
                            "items": {
                                "description": "If unspecified, each key-value pair in the Data field of the referenced ConfigMap will be projected into the volume as a file whose name is the key and content is the value. If specified, the listed keys will be projected into the specified paths, and unlisted keys will not be present. If a key is specified which is not present in the ConfigMap, the volume setup will error unless it is marked optional. Paths must be relative and may not contain the '..' path or start with '..'.",
                                "items": {
                                    "description": "Maps a string key to a path within a volume.",
                                    "properties": {
                                        "key": {
                                            "description": "The key to project.",
                                            "type": "string"
                                        },
                                        "mode": {
                                            "description": "Optional: mode bits to use on this file, must be a value between 0 and 0777. If not specified, the volume defaultMode will be used. This might be in conflict with other options that affect the file mode, like fsGroup, and the result can be other mode bits set.",
                                            "format": "int32",
                                            "type": "integer"
                                        },
                                        "path": {
                                            "description": "The relative path of the file to map the key to. May not be an absolute path. May not contain the path element '..'. May not start with the string '..'.",
                                            "type": "string"
                                        }
                                    },
                                    "required": [
                                        "key",
                                        "path"
                                    ],
                                    "type": "object"
                                },
                                "type": "array"
                            },
                            "name": {
                                "description": "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
                                "type": "string"
                            },
                            "optional": {
                                "description": "Specify whether the ConfigMap or it's keys must be defined",
                                "type": "boolean"
                            }
                        },
                        "type": "object"
                    },
                    "csi": {
                        "description": "Represents a source location of a volume to mount, managed by an external CSI driver",
                        "properties": {
                            "driver": {
                                "description": "Driver is the name of the CSI driver that handles this volume. Consult with your admin for the correct name as registered in the cluster.",
                                "type": "string"
                            },
                            "fsType": {
                                "description": "Filesystem type to mount. Ex. \"ext4\", \"xfs\", \"ntfs\". If not provided, the empty value is passed to the associated CSI driver which will determine the default filesystem to apply.",
                                "type": "string"
                            },
                            "nodePublishSecretRef": {
                                "description": "LocalObjectReference contains enough information to let you locate the referenced object inside the same namespace.",
                                "properties": {
                                    "name": {
                                        "description": "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
                                        "type": "string"
                                    }
                                },
                                "type": "object"
                            },
                            "readOnly": {
                                "description": "Specifies a read-only configuration for the volume. Defaults to false (read/write).",
                                "type": "boolean"
                            },
                            "volumeAttributes": {
                                "additionalProperties": {
                                    "type": "string"
                                },
                                "description": "VolumeAttributes stores driver-specific properties that are passed to the CSI driver. Consult your driver's documentation for supported values.",
                                "type": "object"
                            }
                        },
                        "required": [
                            "driver"
                        ],
                        "type": "object"
                    },
                    "downwardAPI": {
                        "description": "DownwardAPIVolumeSource represents a volume containing downward API info. Downward API volumes support ownership management and SELinux relabeling.",
                        "properties": {
                            "defaultMode": {
                                "description": "Optional: mode bits to use on created files by default. Must be a value between 0 and 0777. Defaults to 0644. Directories within the path are not affected by this setting. This might be in conflict with other options that affect the file mode, like fsGroup, and the result can be other mode bits set.",
                                "format": "int32",
                                "type": "integer"
                            },
                            "items": {
                                "description": "Items is a list of downward API volume file",
                                "items": {
                                    "description": "DownwardAPIVolumeFile represents information to create the file containing the pod field",
                                    "properties": {
                                        "fieldRef": {
                                            "description": "ObjectFieldSelector selects an APIVersioned field of an object.",
                                            "properties": {
                                                "apiVersion": {
                                                    "description": "Version of the schema the FieldPath is written in terms of, defaults to \"v1\".",
                                                    "type": "string"
                                                },
                                                "fieldPath": {
                                                    "description": "Path of the field to select in the specified API version.",
                                                    "type": "string"
                                                }
                                            },
                                            "required": [
                                                "fieldPath"
                                            ],
                                            "type": "object"
                                        },
                                        "mode": {
                                            "description": "Optional: mode bits to use on this file, must be a value between 0 and 0777. If not specified, the volume defaultMode will be used. This might be in conflict with other options that affect the file mode, like fsGroup, and the result can be other mode bits set.",
                                            "format": "int32",
                                            "type": "integer"
                                        },
                                        "path": {
                                            "description": "Required: Path is  the relative path name of the file to be created. Must not be absolute or contain the '..' path. Must be utf-8 encoded. The first item of the relative path must not start with '..'",
                                            "type": "string"
                                        },
                                        "resourceFieldRef": {
                                            "description": "ResourceFieldSelector represents container resources (cpu, memory) and their output format",
                                            "properties": {
                                                "containerName": {
                                                    "description": "Container name: required for volumes, optional for env vars",
                                                    "type": "string"
                                                },
                                                "divisor": {
                                                    "description": "Quantity is a fixed-point representation of a number. It provides convenient marshaling/unmarshaling in JSON and YAML, in addition to String() and Int64() accessors.\n\nThe serialization format is:\n\n<quantity>        ::= <signedNumber><suffix>\n  (Note that <suffix> may be empty, from the \"\" case in <decimalSI>.)\n<digit>           ::= 0 | 1 | ... | 9 <digits>          ::= <digit> | <digit><digits> <number>          ::= <digits> | <digits>.<digits> | <digits>. | .<digits> <sign>            ::= \"+\" | \"-\" <signedNumber>    ::= <number> | <sign><number> <suffix>          ::= <binarySI> | <decimalExponent> | <decimalSI> <binarySI>        ::= Ki | Mi | Gi | Ti | Pi | Ei\n  (International System of units; See: http://physics.nist.gov/cuu/Units/binary.html)\n<decimalSI>       ::= m | \"\" | k | M | G | T | P | E\n  (Note that 1024 = 1Ki but 1000 = 1k; I didn't choose the capitalization.)\n<decimalExponent> ::= \"e\" <signedNumber> | \"E\" <signedNumber>\n\nNo matter which of the three exponent forms is used, no quantity may represent a number greater than 2^63-1 in magnitude, nor may it have more than 3 decimal places. Numbers larger or more precise will be capped or rounded up. (E.g.: 0.1m will rounded up to 1m.) This may be extended in the future if we require larger or smaller quantities.\n\nWhen a Quantity is parsed from a string, it will remember the type of suffix it had, and will use the same type again when it is serialized.\n\nBefore serializing, Quantity will be put in \"canonical form\". This means that Exponent/suffix will be adjusted up or down (with a corresponding increase or decrease in Mantissa) such that:\n  a. No precision is lost\n  b. No fractional digits will be emitted\n  c. The exponent (or suffix) is as large as possible.\nThe sign will be omitted unless the number is negative.\n\nExamples:\n  1.5 will be serialized as \"1500m\"\n  1.5Gi will be serialized as \"1536Mi\"\n\nNote that the quantity will NEVER be internally represented by a floating point number. That is the whole point of this exercise.\n\nNon-canonical values will still parse as long as they are well formed, but will be re-emitted in their canonical form. (So always use canonical form, or don't diff.)\n\nThis format is intended to make it difficult to use these numbers without writing some sort of special handling code in the hopes that that will cause implementors to also use a fixed point implementation.",
                                                    "type": "string"
                                                },
                                                "resource": {
                                                    "description": "Required: resource to select",
                                                    "type": "string"
                                                }
                                            },
                                            "required": [
                                                "resource"
                                            ],
                                            "type": "object"
                                        }
                                    },
                                    "required": [
                                        "path"
                                    ],
                                    "type": "object"
                                },
                                "type": "array"
                            }
                        },
                        "type": "object"
                    },
                    "emptyDir": {
                        "description": "Represents an empty directory for a pod. Empty directory volumes support ownership management and SELinux relabeling.",
                        "properties": {
                            "medium": {
                                "description": "What type of storage medium should back this directory. The default is \"\" which means to use the node's default medium. Must be an empty string (default) or Memory. More info: https://kubernetes.io/docs/concepts/storage/volumes#emptydir",
                                "type": "string"
                            },
                            "sizeLimit": {
                                "description": "Quantity is a fixed-point representation of a number. It provides convenient marshaling/unmarshaling in JSON and YAML, in addition to String() and Int64() accessors.\n\nThe serialization format is:\n\n<quantity>        ::= <signedNumber><suffix>\n  (Note that <suffix> may be empty, from the \"\" case in <decimalSI>.)\n<digit>           ::= 0 | 1 | ... | 9 <digits>          ::= <digit> | <digit><digits> <number>          ::= <digits> | <digits>.<digits> | <digits>. | .<digits> <sign>            ::= \"+\" | \"-\" <signedNumber>    ::= <number> | <sign><number> <suffix>          ::= <binarySI> | <decimalExponent> | <decimalSI> <binarySI>        ::= Ki | Mi | Gi | Ti | Pi | Ei\n  (International System of units; See: http://physics.nist.gov/cuu/Units/binary.html)\n<decimalSI>       ::= m | \"\" | k | M | G | T | P | E\n  (Note that 1024 = 1Ki but 1000 = 1k; I didn't choose the capitalization.)\n<decimalExponent> ::= \"e\" <signedNumber> | \"E\" <signedNumber>\n\nNo matter which of the three exponent forms is used, no quantity may represent a number greater than 2^63-1 in magnitude, nor may it have more than 3 decimal places. Numbers larger or more precise will be capped or rounded up. (E.g.: 0.1m will rounded up to 1m.) This may be extended in the future if we require larger or smaller quantities.\n\nWhen a Quantity is parsed from a string, it will remember the type of suffix it had, and will use the same type again when it is serialized.\n\nBefore serializing, Quantity will be put in \"canonical form\". This means that Exponent/suffix will be adjusted up or down (with a corresponding increase or decrease in Mantissa) such that:\n  a. No precision is lost\n  b. No fractional digits will be emitted\n  c. The exponent (or suffix) is as large as possible.\nThe sign will be omitted unless the number is negative.\n\nExamples:\n  1.5 will be serialized as \"1500m\"\n  1.5Gi will be serialized as \"1536Mi\"\n\nNote that the quantity will NEVER be internally represented by a floating point number. That is the whole point of this exercise.\n\nNon-canonical values will still parse as long as they are well formed, but will be re-emitted in their canonical form. (So always use canonical form, or don't diff.)\n\nThis format is intended to make it difficult to use these numbers without writing some sort of special handling code in the hopes that that will cause implementors to also use a fixed point implementation.",
                                "type": "string"
                            }
                        },
                        "type": "object"
                    },
                    "fc": {
                        "description": "Represents a Fibre Channel volume. Fibre Channel volumes can only be mounted as read/write once. Fibre Channel volumes support ownership management and SELinux relabeling.",
                        "properties": {
                            "fsType": {
                                "description": "Filesystem type to mount. Must be a filesystem type supported by the host operating system. Ex. \"ext4\", \"xfs\", \"ntfs\". Implicitly inferred to be \"ext4\" if unspecified.",
                                "type": "string"
                            },
                            "lun": {
                                "description": "Optional: FC target lun number",
                                "format": "int32",
                                "type": "integer"
                            },
                            "readOnly": {
                                "description": "Optional: Defaults to false (read/write). ReadOnly here will force the ReadOnly setting in VolumeMounts.",
                                "type": "boolean"
                            },
                            "targetWWNs": {
                                "description": "Optional: FC target worldwide names (WWNs)",
                                "items": {
                                    "type": "string"
                                },
                                "type": "array"
                            },
                            "wwids": {
                                "description": "Optional: FC volume world wide identifiers (wwids) Either wwids or combination of targetWWNs and lun must be set, but not both simultaneously.",
                                "items": {
                                    "type": "string"
                                },
                                "type": "array"
                            }
                        },
                        "type": "object"
                    },
                    "flexVolume": {
                        "description": "FlexVolume represents a generic volume resource that is provisioned/attached using an exec based plugin.",
                        "properties": {
                            "driver": {
                                "description": "Driver is the name of the driver to use for this volume.",
                                "type": "string"
                            },
                            "fsType": {
                                "description": "Filesystem type to mount. Must be a filesystem type supported by the host operating system. Ex. \"ext4\", \"xfs\", \"ntfs\". The default filesystem depends on FlexVolume script.",
                                "type": "string"
                            },
                            "options": {
                                "additionalProperties": {
                                    "type": "string"
                                },
                                "description": "Optional: Extra command options if any.",
                                "type": "object"
                            },
                            "readOnly": {
                                "description": "Optional: Defaults to false (read/write). ReadOnly here will force the ReadOnly setting in VolumeMounts.",
                                "type": "boolean"
                            },
                            "secretRef": {
                                "description": "LocalObjectReference contains enough information to let you locate the referenced object inside the same namespace.",
                                "properties": {
                                    "name": {
                                        "description": "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
                                        "type": "string"
                                    }
                                },
                                "type": "object"
                            }
                        },
                        "required": [
                            "driver"
                        ],
                        "type": "object"
                    },
                    "flocker": {
                        "description": "Represents a Flocker volume mounted by the Flocker agent. One and only one of datasetName and datasetUUID should be set. Flocker volumes do not support ownership management or SELinux relabeling.",
                        "properties": {
                            "datasetName": {
                                "description": "Name of the dataset stored as metadata -> name on the dataset for Flocker should be considered as deprecated",
                                "type": "string"
                            },
                            "datasetUUID": {
                                "description": "UUID of the dataset. This is unique identifier of a Flocker dataset",
                                "type": "string"
                            }
                        },
                        "type": "object"
                    },
                    "gcePersistentDisk": {
                        "description": "Represents a Persistent Disk resource in Google Compute Engine.\n\nA GCE PD must exist before mounting to a container. The disk must also be in the same GCE project and zone as the kubelet. A GCE PD can only be mounted as read/write once or read-only many times. GCE PDs support ownership management and SELinux relabeling.",
                        "properties": {
                            "fsType": {
                                "description": "Filesystem type of the volume that you want to mount. Tip: Ensure that the filesystem type is supported by the host operating system. Examples: \"ext4\", \"xfs\", \"ntfs\". Implicitly inferred to be \"ext4\" if unspecified. More info: https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk",
                                "type": "string"
                            },
                            "partition": {
                                "description": "The partition in the volume that you want to mount. If omitted, the default is to mount by volume name. Examples: For volume /dev/sda1, you specify the partition as \"1\". Similarly, the volume partition for /dev/sda is \"0\" (or you can leave the property empty). More info: https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk",
                                "format": "int32",
                                "type": "integer"
                            },
                            "pdName": {
                                "description": "Unique name of the PD resource in GCE. Used to identify the disk in GCE. More info: https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk",
                                "type": "string"
                            },
                            "readOnly": {
                                "description": "ReadOnly here will force the ReadOnly setting in VolumeMounts. Defaults to false. More info: https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk",
                                "type": "boolean"
                            }
                        },
                        "required": [
                            "pdName"
                        ],
                        "type": "object"
                    },
                    "gitRepo": {
                        "description": "Represents a volume that is populated with the contents of a git repository. Git repo volumes do not support ownership management. Git repo volumes support SELinux relabeling.\n\nDEPRECATED: GitRepo is deprecated. To provision a container with a git repo, mount an EmptyDir into an InitContainer that clones the repo using git, then mount the EmptyDir into the Pod's container.",
                        "properties": {
                            "directory": {
                                "description": "Target directory name. Must not contain or start with '..'.  If '.' is supplied, the volume directory will be the git repository.  Otherwise, if specified, the volume will contain the git repository in the subdirectory with the given name.",
                                "type": "string"
                            },
                            "repository": {
                                "description": "Repository URL",
                                "type": "string"
                            },
                            "revision": {
                                "description": "Commit hash for the specified revision.",
                                "type": "string"
                            }
                        },
                        "required": [
                            "repository"
                        ],
                        "type": "object"
                    },
                    "glusterfs": {
                        "description": "Represents a Glusterfs mount that lasts the lifetime of a pod. Glusterfs volumes do not support ownership management or SELinux relabeling.",
                        "properties": {
                            "endpoints": {
                                "description": "EndpointsName is the endpoint name that details Glusterfs topology. More info: https://releases.k8s.io/HEAD/examples/volumes/glusterfs/README.md#create-a-pod",
                                "type": "string"
                            },
                            "path": {
                                "description": "Path is the Glusterfs volume path. More info: https://releases.k8s.io/HEAD/examples/volumes/glusterfs/README.md#create-a-pod",
                                "type": "string"
                            },
                            "readOnly": {
                                "description": "ReadOnly here will force the Glusterfs volume to be mounted with read-only permissions. Defaults to false. More info: https://releases.k8s.io/HEAD/examples/volumes/glusterfs/README.md#create-a-pod",
                                "type": "boolean"
                            }
                        },
                        "required": [
                            "endpoints",
                            "path"
                        ],
                        "type": "object"
                    },
                    "hostPath": {
                        "description": "Represents a host path mapped into a pod. Host path volumes do not support ownership management or SELinux relabeling.",
                        "properties": {
                            "path": {
                                "description": "Path of the directory on the host. If the path is a symlink, it will follow the link to the real path. More info: https://kubernetes.io/docs/concepts/storage/volumes#hostpath",
                                "type": "string"
                            },
                            "type": {
                                "description": "Type for HostPath Volume Defaults to \"\" More info: https://kubernetes.io/docs/concepts/storage/volumes#hostpath",
                                "type": "string"
                            }
                        },
                        "required": [
                            "path"
                        ],
                        "type": "object"
                    },
                    "iscsi": {
                        "description": "Represents an ISCSI disk. ISCSI volumes can only be mounted as read/write once. ISCSI volumes support ownership management and SELinux relabeling.",
                        "properties": {
                            "chapAuthDiscovery": {
                                "description": "whether support iSCSI Discovery CHAP authentication",
                                "type": "boolean"
                            },
                            "chapAuthSession": {
                                "description": "whether support iSCSI Session CHAP authentication",
                                "type": "boolean"
                            },
                            "fsType": {
                                "description": "Filesystem type of the volume that you want to mount. Tip: Ensure that the filesystem type is supported by the host operating system. Examples: \"ext4\", \"xfs\", \"ntfs\". Implicitly inferred to be \"ext4\" if unspecified. More info: https://kubernetes.io/docs/concepts/storage/volumes#iscsi",
                                "type": "string"
                            },
                            "initiatorName": {
                                "description": "Custom iSCSI Initiator Name. If initiatorName is specified with iscsiInterface simultaneously, new iSCSI interface <target portal>:<volume name> will be created for the connection.",
                                "type": "string"
                            },
                            "iqn": {
                                "description": "Target iSCSI Qualified Name.",
                                "type": "string"
                            },
                            "iscsiInterface": {
                                "description": "iSCSI Interface Name that uses an iSCSI transport. Defaults to 'default' (tcp).",
                                "type": "string"
                            },
                            "lun": {
                                "description": "iSCSI Target Lun number.",
                                "format": "int32",
                                "type": "integer"
                            },
                            "portals": {
                                "description": "iSCSI Target Portal List. The portal is either an IP or ip_addr:port if the port is other than default (typically TCP ports 860 and 3260).",
                                "items": {
                                    "type": "string"
                                },
                                "type": "array"
                            },
                            "readOnly": {
                                "description": "ReadOnly here will force the ReadOnly setting in VolumeMounts. Defaults to false.",
                                "type": "boolean"
                            },
                            "secretRef": {
                                "description": "LocalObjectReference contains enough information to let you locate the referenced object inside the same namespace.",
                                "properties": {
                                    "name": {
                                        "description": "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
                                        "type": "string"
                                    }
                                },
                                "type": "object"
                            },
                            "targetPortal": {
                                "description": "iSCSI Target Portal. The Portal is either an IP or ip_addr:port if the port is other than default (typically TCP ports 860 and 3260).",
                                "type": "string"
                            }
                        },
                        "required": [
                            "targetPortal",
                            "iqn",
                            "lun"
                        ],
                        "type": "object"
                    },
                    "name": {
                        "description": "Volume's name. Must be a DNS_LABEL and unique within the pod. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
                        "type": "string"
                    },
                    "nfs": {
                        "description": "Represents an NFS mount that lasts the lifetime of a pod. NFS volumes do not support ownership management or SELinux relabeling.",
                        "properties": {
                            "path": {
                                "description": "Path that is exported by the NFS server. More info: https://kubernetes.io/docs/concepts/storage/volumes#nfs",
                                "type": "string"
                            },
                            "readOnly": {
                                "description": "ReadOnly here will force the NFS export to be mounted with read-only permissions. Defaults to false. More info: https://kubernetes.io/docs/concepts/storage/volumes#nfs",
                                "type": "boolean"
                            },
                            "server": {
                                "description": "Server is the hostname or IP address of the NFS server. More info: https://kubernetes.io/docs/concepts/storage/volumes#nfs",
                                "type": "string"
                            }
                        },
                        "required": [
                            "server",
                            "path"
                        ],
                        "type": "object"
                    },
                    "persistentVolumeClaim": {
                        "description": "PersistentVolumeClaimVolumeSource references the user's PVC in the same namespace. This volume finds the bound PV and mounts that volume for the pod. A PersistentVolumeClaimVolumeSource is, essentially, a wrapper around another type of volume that is owned by someone else (the system).",
                        "properties": {
                            "claimName": {
                                "description": "ClaimName is the name of a PersistentVolumeClaim in the same namespace as the pod using this volume. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims",
                                "type": "string"
                            },
                            "readOnly": {
                                "description": "Will force the ReadOnly setting in VolumeMounts. Default false.",
                                "type": "boolean"
                            }
                        },
                        "required": [
                            "claimName"
                        ],
                        "type": "object"
                    },
                    "photonPersistentDisk": {
                        "description": "Represents a Photon Controller persistent disk resource.",
                        "properties": {
                            "fsType": {
                                "description": "Filesystem type to mount. Must be a filesystem type supported by the host operating system. Ex. \"ext4\", \"xfs\", \"ntfs\". Implicitly inferred to be \"ext4\" if unspecified.",
                                "type": "string"
                            },
                            "pdID": {
                                "description": "ID that identifies Photon Controller persistent disk",
                                "type": "string"
                            }
                        },
                        "required": [
                            "pdID"
                        ],
                        "type": "object"
                    },
                    "portworxVolume": {
                        "description": "PortworxVolumeSource represents a Portworx volume resource.",
                        "properties": {
                            "fsType": {
                                "description": "FSType represents the filesystem type to mount Must be a filesystem type supported by the host operating system. Ex. \"ext4\", \"xfs\". Implicitly inferred to be \"ext4\" if unspecified.",
                                "type": "string"
                            },
                            "readOnly": {
                                "description": "Defaults to false (read/write). ReadOnly here will force the ReadOnly setting in VolumeMounts.",
                                "type": "boolean"
                            },
                            "volumeID": {
                                "description": "VolumeID uniquely identifies a Portworx volume",
                                "type": "string"
                            }
                        },
                        "required": [
                            "volumeID"
                        ],
                        "type": "object"
                    },
                    "projected": {
                        "description": "Represents a projected volume source",
                        "properties": {
                            "defaultMode": {
                                "description": "Mode bits to use on created files by default. Must be a value between 0 and 0777. Directories within the path are not affected by this setting. This might be in conflict with other options that affect the file mode, like fsGroup, and the result can be other mode bits set.",
                                "format": "int32",
                                "type": "integer"
                            },
                            "sources": {
                                "description": "list of volume projections",
                                "items": {
                                    "description": "Projection that may be projected along with other supported volume types",
                                    "properties": {
                                        "configMap": {
                                            "description": "Adapts a ConfigMap into a projected volume.\n\nThe contents of the target ConfigMap's Data field will be presented in a projected volume as files using the keys in the Data field as the file names, unless the items element is populated with specific mappings of keys to paths. Note that this is identical to a configmap volume source without the default mode.",
                                            "properties": {
                                                "items": {
                                                    "description": "If unspecified, each key-value pair in the Data field of the referenced ConfigMap will be projected into the volume as a file whose name is the key and content is the value. If specified, the listed keys will be projected into the specified paths, and unlisted keys will not be present. If a key is specified which is not present in the ConfigMap, the volume setup will error unless it is marked optional. Paths must be relative and may not contain the '..' path or start with '..'.",
                                                    "items": {
                                                        "description": "Maps a string key to a path within a volume.",
                                                        "properties": {
                                                            "key": {
                                                                "description": "The key to project.",
                                                                "type": "string"
                                                            },
                                                            "mode": {
                                                                "description": "Optional: mode bits to use on this file, must be a value between 0 and 0777. If not specified, the volume defaultMode will be used. This might be in conflict with other options that affect the file mode, like fsGroup, and the result can be other mode bits set.",
                                                                "format": "int32",
                                                                "type": "integer"
                                                            },
                                                            "path": {
                                                                "description": "The relative path of the file to map the key to. May not be an absolute path. May not contain the path element '..'. May not start with the string '..'.",
                                                                "type": "string"
                                                            }
                                                        },
                                                        "required": [
                                                            "key",
                                                            "path"
                                                        ],
                                                        "type": "object"
                                                    },
                                                    "type": "array"
                                                },
                                                "name": {
                                                    "description": "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
                                                    "type": "string"
                                                },
                                                "optional": {
                                                    "description": "Specify whether the ConfigMap or it's keys must be defined",
                                                    "type": "boolean"
                                                }
                                            },
                                            "type": "object"
                                        },
                                        "downwardAPI": {
                                            "description": "Represents downward API info for projecting into a projected volume. Note that this is identical to a downwardAPI volume source without the default mode.",
                                            "properties": {
                                                "items": {
                                                    "description": "Items is a list of DownwardAPIVolume file",
                                                    "items": {
                                                        "description": "DownwardAPIVolumeFile represents information to create the file containing the pod field",
                                                        "properties": {
                                                            "fieldRef": {
                                                                "description": "ObjectFieldSelector selects an APIVersioned field of an object.",
                                                                "properties": {
                                                                    "apiVersion": {
                                                                        "description": "Version of the schema the FieldPath is written in terms of, defaults to \"v1\".",
                                                                        "type": "string"
                                                                    },
                                                                    "fieldPath": {
                                                                        "description": "Path of the field to select in the specified API version.",
                                                                        "type": "string"
                                                                    }
                                                                },
                                                                "required": [
                                                                    "fieldPath"
                                                                ],
                                                                "type": "object"
                                                            },
                                                            "mode": {
                                                                "description": "Optional: mode bits to use on this file, must be a value between 0 and 0777. If not specified, the volume defaultMode will be used. This might be in conflict with other options that affect the file mode, like fsGroup, and the result can be other mode bits set.",
                                                                "format": "int32",
                                                                "type": "integer"
                                                            },
                                                            "path": {
                                                                "description": "Required: Path is  the relative path name of the file to be created. Must not be absolute or contain the '..' path. Must be utf-8 encoded. The first item of the relative path must not start with '..'",
                                                                "type": "string"
                                                            },
                                                            "resourceFieldRef": {
                                                                "description": "ResourceFieldSelector represents container resources (cpu, memory) and their output format",
                                                                "properties": {
                                                                    "containerName": {
                                                                        "description": "Container name: required for volumes, optional for env vars",
                                                                        "type": "string"
                                                                    },
                                                                    "divisor": {
                                                                        "description": "Quantity is a fixed-point representation of a number. It provides convenient marshaling/unmarshaling in JSON and YAML, in addition to String() and Int64() accessors.\n\nThe serialization format is:\n\n<quantity>        ::= <signedNumber><suffix>\n  (Note that <suffix> may be empty, from the \"\" case in <decimalSI>.)\n<digit>           ::= 0 | 1 | ... | 9 <digits>          ::= <digit> | <digit><digits> <number>          ::= <digits> | <digits>.<digits> | <digits>. | .<digits> <sign>            ::= \"+\" | \"-\" <signedNumber>    ::= <number> | <sign><number> <suffix>          ::= <binarySI> | <decimalExponent> | <decimalSI> <binarySI>        ::= Ki | Mi | Gi | Ti | Pi | Ei\n  (International System of units; See: http://physics.nist.gov/cuu/Units/binary.html)\n<decimalSI>       ::= m | \"\" | k | M | G | T | P | E\n  (Note that 1024 = 1Ki but 1000 = 1k; I didn't choose the capitalization.)\n<decimalExponent> ::= \"e\" <signedNumber> | \"E\" <signedNumber>\n\nNo matter which of the three exponent forms is used, no quantity may represent a number greater than 2^63-1 in magnitude, nor may it have more than 3 decimal places. Numbers larger or more precise will be capped or rounded up. (E.g.: 0.1m will rounded up to 1m.) This may be extended in the future if we require larger or smaller quantities.\n\nWhen a Quantity is parsed from a string, it will remember the type of suffix it had, and will use the same type again when it is serialized.\n\nBefore serializing, Quantity will be put in \"canonical form\". This means that Exponent/suffix will be adjusted up or down (with a corresponding increase or decrease in Mantissa) such that:\n  a. No precision is lost\n  b. No fractional digits will be emitted\n  c. The exponent (or suffix) is as large as possible.\nThe sign will be omitted unless the number is negative.\n\nExamples:\n  1.5 will be serialized as \"1500m\"\n  1.5Gi will be serialized as \"1536Mi\"\n\nNote that the quantity will NEVER be internally represented by a floating point number. That is the whole point of this exercise.\n\nNon-canonical values will still parse as long as they are well formed, but will be re-emitted in their canonical form. (So always use canonical form, or don't diff.)\n\nThis format is intended to make it difficult to use these numbers without writing some sort of special handling code in the hopes that that will cause implementors to also use a fixed point implementation.",
                                                                        "type": "string"
                                                                    },
                                                                    "resource": {
                                                                        "description": "Required: resource to select",
                                                                        "type": "string"
                                                                    }
                                                                },
                                                                "required": [
                                                                    "resource"
                                                                ],
                                                                "type": "object"
                                                            }
                                                        },
                                                        "required": [
                                                            "path"
                                                        ],
                                                        "type": "object"
                                                    },
                                                    "type": "array"
                                                }
                                            },
                                            "type": "object"
                                        },
                                        "secret": {
                                            "description": "Adapts a secret into a projected volume.\n\nThe contents of the target Secret's Data field will be presented in a projected volume as files using the keys in the Data field as the file names. Note that this is identical to a secret volume source without the default mode.",
                                            "properties": {
                                                "items": {
                                                    "description": "If unspecified, each key-value pair in the Data field of the referenced Secret will be projected into the volume as a file whose name is the key and content is the value. If specified, the listed keys will be projected into the specified paths, and unlisted keys will not be present. If a key is specified which is not present in the Secret, the volume setup will error unless it is marked optional. Paths must be relative and may not contain the '..' path or start with '..'.",
                                                    "items": {
                                                        "description": "Maps a string key to a path within a volume.",
                                                        "properties": {
                                                            "key": {
                                                                "description": "The key to project.",
                                                                "type": "string"
                                                            },
                                                            "mode": {
                                                                "description": "Optional: mode bits to use on this file, must be a value between 0 and 0777. If not specified, the volume defaultMode will be used. This might be in conflict with other options that affect the file mode, like fsGroup, and the result can be other mode bits set.",
                                                                "format": "int32",
                                                                "type": "integer"
                                                            },
                                                            "path": {
                                                                "description": "The relative path of the file to map the key to. May not be an absolute path. May not contain the path element '..'. May not start with the string '..'.",
                                                                "type": "string"
                                                            }
                                                        },
                                                        "required": [
                                                            "key",
                                                            "path"
                                                        ],
                                                        "type": "object"
                                                    },
                                                    "type": "array"
                                                },
                                                "name": {
                                                    "description": "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
                                                    "type": "string"
                                                },
                                                "optional": {
                                                    "description": "Specify whether the Secret or its key must be defined",
                                                    "type": "boolean"
                                                }
                                            },
                                            "type": "object"
                                        },
                                        "serviceAccountToken": {
                                            "description": "ServiceAccountTokenProjection represents a projected service account token volume. This projection can be used to insert a service account token into the pods runtime filesystem for use against APIs (Kubernetes API Server or otherwise).",
                                            "properties": {
                                                "audience": {
                                                    "description": "Audience is the intended audience of the token. A recipient of a token must identify itself with an identifier specified in the audience of the token, and otherwise should reject the token. The audience defaults to the identifier of the apiserver.",
                                                    "type": "string"
                                                },
                                                "expirationSeconds": {
                                                    "description": "ExpirationSeconds is the requested duration of validity of the service account token. As the token approaches expiration, the kubelet volume plugin will proactively rotate the service account token. The kubelet will start trying to rotate the token if the token is older than 80 percent of its time to live or if the token is older than 24 hours.Defaults to 1 hour and must be at least 10 minutes.",
                                                    "format": "int64",
                                                    "type": "integer"
                                                },
                                                "path": {
                                                    "description": "Path is the path relative to the mount point of the file to project the token into.",
                                                    "type": "string"
                                                }
                                            },
                                            "required": [
                                                "path"
                                            ],
                                            "type": "object"
                                        }
                                    },
                                    "type": "object"
                                },
                                "type": "array"
                            }
                        },
                        "required": [
                            "sources"
                        ],
                        "type": "object"
                    },
                    "quobyte": {
                        "description": "Represents a Quobyte mount that lasts the lifetime of a pod. Quobyte volumes do not support ownership management or SELinux relabeling.",
                        "properties": {
                            "group": {
                                "description": "Group to map volume access to Default is no group",
                                "type": "string"
                            },
                            "readOnly": {
                                "description": "ReadOnly here will force the Quobyte volume to be mounted with read-only permissions. Defaults to false.",
                                "type": "boolean"
                            },
                            "registry": {
                                "description": "Registry represents a single or multiple Quobyte Registry services specified as a string as host:port pair (multiple entries are separated with commas) which acts as the central registry for volumes",
                                "type": "string"
                            },
                            "tenant": {
                                "description": "Tenant owning the given Quobyte volume in the Backend Used with dynamically provisioned Quobyte volumes, value is set by the plugin",
                                "type": "string"
                            },
                            "user": {
                                "description": "User to map volume access to Defaults to serivceaccount user",
                                "type": "string"
                            },
                            "volume": {
                                "description": "Volume is a string that references an already created Quobyte volume by name.",
                                "type": "string"
                            }
                        },
                        "required": [
                            "registry",
                            "volume"
                        ],
                        "type": "object"
                    },
                    "rbd": {
                        "description": "Represents a Rados Block Device mount that lasts the lifetime of a pod. RBD volumes support ownership management and SELinux relabeling.",
                        "properties": {
                            "fsType": {
                                "description": "Filesystem type of the volume that you want to mount. Tip: Ensure that the filesystem type is supported by the host operating system. Examples: \"ext4\", \"xfs\", \"ntfs\". Implicitly inferred to be \"ext4\" if unspecified. More info: https://kubernetes.io/docs/concepts/storage/volumes#rbd",
                                "type": "string"
                            },
                            "image": {
                                "description": "The rados image name. More info: https://releases.k8s.io/HEAD/examples/volumes/rbd/README.md#how-to-use-it",
                                "type": "string"
                            },
                            "keyring": {
                                "description": "Keyring is the path to key ring for RBDUser. Default is /etc/ceph/keyring. More info: https://releases.k8s.io/HEAD/examples/volumes/rbd/README.md#how-to-use-it",
                                "type": "string"
                            },
                            "monitors": {
                                "description": "A collection of Ceph monitors. More info: https://releases.k8s.io/HEAD/examples/volumes/rbd/README.md#how-to-use-it",
                                "items": {
                                    "type": "string"
                                },
                                "type": "array"
                            },
                            "pool": {
                                "description": "The rados pool name. Default is rbd. More info: https://releases.k8s.io/HEAD/examples/volumes/rbd/README.md#how-to-use-it",
                                "type": "string"
                            },
                            "readOnly": {
                                "description": "ReadOnly here will force the ReadOnly setting in VolumeMounts. Defaults to false. More info: https://releases.k8s.io/HEAD/examples/volumes/rbd/README.md#how-to-use-it",
                                "type": "boolean"
                            },
                            "secretRef": {
                                "description": "LocalObjectReference contains enough information to let you locate the referenced object inside the same namespace.",
                                "properties": {
                                    "name": {
                                        "description": "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
                                        "type": "string"
                                    }
                                },
                                "type": "object"
                            },
                            "user": {
                                "description": "The rados user name. Default is admin. More info: https://releases.k8s.io/HEAD/examples/volumes/rbd/README.md#how-to-use-it",
                                "type": "string"
                            }
                        },
                        "required": [
                            "monitors",
                            "image"
                        ],
                        "type": "object"
                    },
                    "scaleIO": {
                        "description": "ScaleIOVolumeSource represents a persistent ScaleIO volume",
                        "properties": {
                            "fsType": {
                                "description": "Filesystem type to mount. Must be a filesystem type supported by the host operating system. Ex. \"ext4\", \"xfs\", \"ntfs\". Default is \"xfs\".",
                                "type": "string"
                            },
                            "gateway": {
                                "description": "The host address of the ScaleIO API Gateway.",
                                "type": "string"
                            },
                            "protectionDomain": {
                                "description": "The name of the ScaleIO Protection Domain for the configured storage.",
                                "type": "string"
                            },
                            "readOnly": {
                                "description": "Defaults to false (read/write). ReadOnly here will force the ReadOnly setting in VolumeMounts.",
                                "type": "boolean"
                            },
                            "secretRef": {
                                "description": "LocalObjectReference contains enough information to let you locate the referenced object inside the same namespace.",
                                "properties": {
                                    "name": {
                                        "description": "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
                                        "type": "string"
                                    }
                                },
                                "type": "object"
                            },
                            "sslEnabled": {
                                "description": "Flag to enable/disable SSL communication with Gateway, default false",
                                "type": "boolean"
                            },
                            "storageMode": {
                                "description": "Indicates whether the storage for a volume should be ThickProvisioned or ThinProvisioned. Default is ThinProvisioned.",
                                "type": "string"
                            },
                            "storagePool": {
                                "description": "The ScaleIO Storage Pool associated with the protection domain.",
                                "type": "string"
                            },
                            "system": {
                                "description": "The name of the storage system as configured in ScaleIO.",
                                "type": "string"
                            },
                            "volumeName": {
                                "description": "The name of a volume already created in the ScaleIO system that is associated with this volume source.",
                                "type": "string"
                            }
                        },
                        "required": [
                            "gateway",
                            "system",
                            "secretRef"
                        ],
                        "type": "object"
                    },
                    "secret": {
                        "description": "Adapts a Secret into a volume.\n\nThe contents of the target Secret's Data field will be presented in a volume as files using the keys in the Data field as the file names. Secret volumes support ownership management and SELinux relabeling.",
                        "properties": {
                            "defaultMode": {
                                "description": "Optional: mode bits to use on created files by default. Must be a value between 0 and 0777. Defaults to 0644. Directories within the path are not affected by this setting. This might be in conflict with other options that affect the file mode, like fsGroup, and the result can be other mode bits set.",
                                "format": "int32",
                                "type": "integer"
                            },
                            "items": {
                                "description": "If unspecified, each key-value pair in the Data field of the referenced Secret will be projected into the volume as a file whose name is the key and content is the value. If specified, the listed keys will be projected into the specified paths, and unlisted keys will not be present. If a key is specified which is not present in the Secret, the volume setup will error unless it is marked optional. Paths must be relative and may not contain the '..' path or start with '..'.",
                                "items": {
                                    "description": "Maps a string key to a path within a volume.",
                                    "properties": {
                                        "key": {
                                            "description": "The key to project.",
                                            "type": "string"
                                        },
                                        "mode": {
                                            "description": "Optional: mode bits to use on this file, must be a value between 0 and 0777. If not specified, the volume defaultMode will be used. This might be in conflict with other options that affect the file mode, like fsGroup, and the result can be other mode bits set.",
                                            "format": "int32",
                                            "type": "integer"
                                        },
                                        "path": {
                                            "description": "The relative path of the file to map the key to. May not be an absolute path. May not contain the path element '..'. May not start with the string '..'.",
                                            "type": "string"
                                        }
                                    },
                                    "required": [
                                        "key",
                                        "path"
                                    ],
                                    "type": "object"
                                },
                                "type": "array"
                            },
                            "optional": {
                                "description": "Specify whether the Secret or it's keys must be defined",
                                "type": "boolean"
                            },
                            "secretName": {
                                "description": "Name of the secret in the pod's namespace to use. More info: https://kubernetes.io/docs/concepts/storage/volumes#secret",
                                "type": "string"
                            }
                        },
                        "type": "object"
                    },
                    "storageos": {
                        "description": "Represents a StorageOS persistent volume resource.",
                        "properties": {
                            "fsType": {
                                "description": "Filesystem type to mount. Must be a filesystem type supported by the host operating system. Ex. \"ext4\", \"xfs\", \"ntfs\". Implicitly inferred to be \"ext4\" if unspecified.",
                                "type": "string"
                            },
                            "readOnly": {
                                "description": "Defaults to false (read/write). ReadOnly here will force the ReadOnly setting in VolumeMounts.",
                                "type": "boolean"
                            },
                            "secretRef": {
                                "description": "LocalObjectReference contains enough information to let you locate the referenced object inside the same namespace.",
                                "properties": {
                                    "name": {
                                        "description": "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
                                        "type": "string"
                                    }
                                },
                                "type": "object"
                            },
                            "volumeName": {
                                "description": "VolumeName is the human-readable name of the StorageOS volume.  Volume names are only unique within a namespace.",
                                "type": "string"
                            },
                            "volumeNamespace": {
                                "description": "VolumeNamespace specifies the scope of the volume within StorageOS.  If no namespace is specified then the Pod's namespace will be used.  This allows the Kubernetes name scoping to be mirrored within StorageOS for tighter integration. Set VolumeName to any name to override the default behaviour. Set to \"default\" if you are not using namespaces within StorageOS. Namespaces that do not pre-exist within StorageOS will be created.",
                                "type": "string"
                            }
                        },
                        "type": "object"
                    },
                    "vsphereVolume": {
                        "description": "Represents a vSphere volume resource.",
                        "properties": {
                            "fsType": {
                                "description": "Filesystem type to mount. Must be a filesystem type supported by the host operating system. Ex. \"ext4\", \"xfs\", \"ntfs\". Implicitly inferred to be \"ext4\" if unspecified.",
                                "type": "string"
                            },
                            "storagePolicyID": {
                                "description": "Storage Policy Based Management (SPBM) profile ID associated with the StoragePolicyName.",
                                "type": "string"
                            },
                            "storagePolicyName": {
                                "description": "Storage Policy Based Management (SPBM) profile name.",
                                "type": "string"
                            },
                            "volumePath": {
                                "description": "Path that identifies vSphere volume vmdk",
                                "type": "string"
                            }
                        },
                        "required": [
                            "volumePath"
                        ],
                        "type": "object"
                    }
                },
                "required": [
                    "name"
                ],
                "type": "object"
            },
            "type": "array",
            "x-kubernetes-patch-merge-key": "name",
            "x-kubernetes-patch-strategy": "merge,retainKeys"
        }
    },
    "required": [
        "containers"
    ],
    "type": "object"
}
{{- end }}
