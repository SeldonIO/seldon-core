# Ksonnet Configuration

|Parameter|Description|Default|
|---------|-----------|-------|
| namespace | Namespace to use for the components. It is automatically inherited from the environment if not set. | default or from env |
| withRbac | Whether to include RBAC setup | true |
| withApife | Whether to include builtin API OAuth gateway server for ingress | true |
| withAmbassador | Whether to include Ambassador reverse proxy | false |
| apifeImage | Default image for API Front End | ```<latest seldon release>``` |
| apifeServiceType | API Front End Service Type | NodePort |
| operatorImage | Seldon cluster manager image version | ```<latest seldon release>``` |
| operatorSpringOpts | Operator spring opts | empty |
| operatorJavaOpts | Operator | java opts | empty |
| engineImage | Seldon engine image version | ```<latest seldon release>``` |
| grpcMaxMessageSize | Max gRPC message size | 4MB |


