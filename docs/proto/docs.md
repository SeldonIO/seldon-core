# Protocol Documentation
<a name="top"/>

## Table of Contents

- [generated.proto](#generated.proto)
    - [IntOrString](#k8s.io.apimachinery.pkg.util.intstr.IntOrString)
  
  
  
  

- [generated.proto](#generated.proto)
    - [RawExtension](#k8s.io.apimachinery.pkg.runtime.RawExtension)
    - [TypeMeta](#k8s.io.apimachinery.pkg.runtime.TypeMeta)
    - [Unknown](#k8s.io.apimachinery.pkg.runtime.Unknown)
  
  
  
  

- [generated.proto](#generated.proto)
  
  
  
  

- [generated.proto](#generated.proto)
    - [APIGroup](#k8s.io.apimachinery.pkg.apis.meta.v1.APIGroup)
    - [APIGroupList](#k8s.io.apimachinery.pkg.apis.meta.v1.APIGroupList)
    - [APIResource](#k8s.io.apimachinery.pkg.apis.meta.v1.APIResource)
    - [APIResourceList](#k8s.io.apimachinery.pkg.apis.meta.v1.APIResourceList)
    - [APIVersions](#k8s.io.apimachinery.pkg.apis.meta.v1.APIVersions)
    - [DeleteOptions](#k8s.io.apimachinery.pkg.apis.meta.v1.DeleteOptions)
    - [Duration](#k8s.io.apimachinery.pkg.apis.meta.v1.Duration)
    - [ExportOptions](#k8s.io.apimachinery.pkg.apis.meta.v1.ExportOptions)
    - [GetOptions](#k8s.io.apimachinery.pkg.apis.meta.v1.GetOptions)
    - [GroupKind](#k8s.io.apimachinery.pkg.apis.meta.v1.GroupKind)
    - [GroupResource](#k8s.io.apimachinery.pkg.apis.meta.v1.GroupResource)
    - [GroupVersion](#k8s.io.apimachinery.pkg.apis.meta.v1.GroupVersion)
    - [GroupVersionForDiscovery](#k8s.io.apimachinery.pkg.apis.meta.v1.GroupVersionForDiscovery)
    - [GroupVersionKind](#k8s.io.apimachinery.pkg.apis.meta.v1.GroupVersionKind)
    - [GroupVersionResource](#k8s.io.apimachinery.pkg.apis.meta.v1.GroupVersionResource)
    - [Initializer](#k8s.io.apimachinery.pkg.apis.meta.v1.Initializer)
    - [Initializers](#k8s.io.apimachinery.pkg.apis.meta.v1.Initializers)
    - [LabelSelector](#k8s.io.apimachinery.pkg.apis.meta.v1.LabelSelector)
    - [LabelSelector.MatchLabelsEntry](#k8s.io.apimachinery.pkg.apis.meta.v1.LabelSelector.MatchLabelsEntry)
    - [LabelSelectorRequirement](#k8s.io.apimachinery.pkg.apis.meta.v1.LabelSelectorRequirement)
    - [List](#k8s.io.apimachinery.pkg.apis.meta.v1.List)
    - [ListMeta](#k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta)
    - [ListOptions](#k8s.io.apimachinery.pkg.apis.meta.v1.ListOptions)
    - [MicroTime](#k8s.io.apimachinery.pkg.apis.meta.v1.MicroTime)
    - [ObjectMeta](#k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta)
    - [ObjectMeta.AnnotationsEntry](#k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta.AnnotationsEntry)
    - [ObjectMeta.LabelsEntry](#k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta.LabelsEntry)
    - [OwnerReference](#k8s.io.apimachinery.pkg.apis.meta.v1.OwnerReference)
    - [Preconditions](#k8s.io.apimachinery.pkg.apis.meta.v1.Preconditions)
    - [RootPaths](#k8s.io.apimachinery.pkg.apis.meta.v1.RootPaths)
    - [ServerAddressByClientCIDR](#k8s.io.apimachinery.pkg.apis.meta.v1.ServerAddressByClientCIDR)
    - [Status](#k8s.io.apimachinery.pkg.apis.meta.v1.Status)
    - [StatusCause](#k8s.io.apimachinery.pkg.apis.meta.v1.StatusCause)
    - [StatusDetails](#k8s.io.apimachinery.pkg.apis.meta.v1.StatusDetails)
    - [Time](#k8s.io.apimachinery.pkg.apis.meta.v1.Time)
    - [Timestamp](#k8s.io.apimachinery.pkg.apis.meta.v1.Timestamp)
    - [TypeMeta](#k8s.io.apimachinery.pkg.apis.meta.v1.TypeMeta)
    - [Verbs](#k8s.io.apimachinery.pkg.apis.meta.v1.Verbs)
    - [WatchEvent](#k8s.io.apimachinery.pkg.apis.meta.v1.WatchEvent)
  
  
  
  

- [generated.proto](#generated.proto)
    - [Quantity](#k8s.io.apimachinery.pkg.api.resource.Quantity)
  
  
  
  

- [v1.proto](#v1.proto)
    - [AWSElasticBlockStoreVolumeSource](#k8s.io.api.core.v1.AWSElasticBlockStoreVolumeSource)
    - [Affinity](#k8s.io.api.core.v1.Affinity)
    - [AttachedVolume](#k8s.io.api.core.v1.AttachedVolume)
    - [AvoidPods](#k8s.io.api.core.v1.AvoidPods)
    - [AzureDiskVolumeSource](#k8s.io.api.core.v1.AzureDiskVolumeSource)
    - [AzureFilePersistentVolumeSource](#k8s.io.api.core.v1.AzureFilePersistentVolumeSource)
    - [AzureFileVolumeSource](#k8s.io.api.core.v1.AzureFileVolumeSource)
    - [Binding](#k8s.io.api.core.v1.Binding)
    - [Capabilities](#k8s.io.api.core.v1.Capabilities)
    - [CephFSPersistentVolumeSource](#k8s.io.api.core.v1.CephFSPersistentVolumeSource)
    - [CephFSVolumeSource](#k8s.io.api.core.v1.CephFSVolumeSource)
    - [CinderVolumeSource](#k8s.io.api.core.v1.CinderVolumeSource)
    - [ClientIPConfig](#k8s.io.api.core.v1.ClientIPConfig)
    - [ComponentCondition](#k8s.io.api.core.v1.ComponentCondition)
    - [ComponentStatus](#k8s.io.api.core.v1.ComponentStatus)
    - [ComponentStatusList](#k8s.io.api.core.v1.ComponentStatusList)
    - [ConfigMap](#k8s.io.api.core.v1.ConfigMap)
    - [ConfigMap.DataEntry](#k8s.io.api.core.v1.ConfigMap.DataEntry)
    - [ConfigMapEnvSource](#k8s.io.api.core.v1.ConfigMapEnvSource)
    - [ConfigMapKeySelector](#k8s.io.api.core.v1.ConfigMapKeySelector)
    - [ConfigMapList](#k8s.io.api.core.v1.ConfigMapList)
    - [ConfigMapProjection](#k8s.io.api.core.v1.ConfigMapProjection)
    - [ConfigMapVolumeSource](#k8s.io.api.core.v1.ConfigMapVolumeSource)
    - [Container](#k8s.io.api.core.v1.Container)
    - [ContainerImage](#k8s.io.api.core.v1.ContainerImage)
    - [ContainerPort](#k8s.io.api.core.v1.ContainerPort)
    - [ContainerState](#k8s.io.api.core.v1.ContainerState)
    - [ContainerStateRunning](#k8s.io.api.core.v1.ContainerStateRunning)
    - [ContainerStateTerminated](#k8s.io.api.core.v1.ContainerStateTerminated)
    - [ContainerStateWaiting](#k8s.io.api.core.v1.ContainerStateWaiting)
    - [ContainerStatus](#k8s.io.api.core.v1.ContainerStatus)
    - [DaemonEndpoint](#k8s.io.api.core.v1.DaemonEndpoint)
    - [DeleteOptions](#k8s.io.api.core.v1.DeleteOptions)
    - [DownwardAPIProjection](#k8s.io.api.core.v1.DownwardAPIProjection)
    - [DownwardAPIVolumeFile](#k8s.io.api.core.v1.DownwardAPIVolumeFile)
    - [DownwardAPIVolumeSource](#k8s.io.api.core.v1.DownwardAPIVolumeSource)
    - [EmptyDirVolumeSource](#k8s.io.api.core.v1.EmptyDirVolumeSource)
    - [EndpointAddress](#k8s.io.api.core.v1.EndpointAddress)
    - [EndpointPort](#k8s.io.api.core.v1.EndpointPort)
    - [EndpointSubset](#k8s.io.api.core.v1.EndpointSubset)
    - [Endpoints](#k8s.io.api.core.v1.Endpoints)
    - [EndpointsList](#k8s.io.api.core.v1.EndpointsList)
    - [EnvFromSource](#k8s.io.api.core.v1.EnvFromSource)
    - [EnvVar](#k8s.io.api.core.v1.EnvVar)
    - [EnvVarSource](#k8s.io.api.core.v1.EnvVarSource)
    - [Event](#k8s.io.api.core.v1.Event)
    - [EventList](#k8s.io.api.core.v1.EventList)
    - [EventSource](#k8s.io.api.core.v1.EventSource)
    - [ExecAction](#k8s.io.api.core.v1.ExecAction)
    - [FCVolumeSource](#k8s.io.api.core.v1.FCVolumeSource)
    - [FlexVolumeSource](#k8s.io.api.core.v1.FlexVolumeSource)
    - [FlexVolumeSource.OptionsEntry](#k8s.io.api.core.v1.FlexVolumeSource.OptionsEntry)
    - [FlockerVolumeSource](#k8s.io.api.core.v1.FlockerVolumeSource)
    - [GCEPersistentDiskVolumeSource](#k8s.io.api.core.v1.GCEPersistentDiskVolumeSource)
    - [GitRepoVolumeSource](#k8s.io.api.core.v1.GitRepoVolumeSource)
    - [GlusterfsVolumeSource](#k8s.io.api.core.v1.GlusterfsVolumeSource)
    - [HTTPGetAction](#k8s.io.api.core.v1.HTTPGetAction)
    - [HTTPHeader](#k8s.io.api.core.v1.HTTPHeader)
    - [Handler](#k8s.io.api.core.v1.Handler)
    - [HostAlias](#k8s.io.api.core.v1.HostAlias)
    - [HostPathVolumeSource](#k8s.io.api.core.v1.HostPathVolumeSource)
    - [ISCSIVolumeSource](#k8s.io.api.core.v1.ISCSIVolumeSource)
    - [KeyToPath](#k8s.io.api.core.v1.KeyToPath)
    - [Lifecycle](#k8s.io.api.core.v1.Lifecycle)
    - [LimitRange](#k8s.io.api.core.v1.LimitRange)
    - [LimitRangeItem](#k8s.io.api.core.v1.LimitRangeItem)
    - [LimitRangeItem.DefaultEntry](#k8s.io.api.core.v1.LimitRangeItem.DefaultEntry)
    - [LimitRangeItem.DefaultRequestEntry](#k8s.io.api.core.v1.LimitRangeItem.DefaultRequestEntry)
    - [LimitRangeItem.MaxEntry](#k8s.io.api.core.v1.LimitRangeItem.MaxEntry)
    - [LimitRangeItem.MaxLimitRequestRatioEntry](#k8s.io.api.core.v1.LimitRangeItem.MaxLimitRequestRatioEntry)
    - [LimitRangeItem.MinEntry](#k8s.io.api.core.v1.LimitRangeItem.MinEntry)
    - [LimitRangeList](#k8s.io.api.core.v1.LimitRangeList)
    - [LimitRangeSpec](#k8s.io.api.core.v1.LimitRangeSpec)
    - [List](#k8s.io.api.core.v1.List)
    - [ListOptions](#k8s.io.api.core.v1.ListOptions)
    - [LoadBalancerIngress](#k8s.io.api.core.v1.LoadBalancerIngress)
    - [LoadBalancerStatus](#k8s.io.api.core.v1.LoadBalancerStatus)
    - [LocalObjectReference](#k8s.io.api.core.v1.LocalObjectReference)
    - [LocalVolumeSource](#k8s.io.api.core.v1.LocalVolumeSource)
    - [NFSVolumeSource](#k8s.io.api.core.v1.NFSVolumeSource)
    - [Namespace](#k8s.io.api.core.v1.Namespace)
    - [NamespaceList](#k8s.io.api.core.v1.NamespaceList)
    - [NamespaceSpec](#k8s.io.api.core.v1.NamespaceSpec)
    - [NamespaceStatus](#k8s.io.api.core.v1.NamespaceStatus)
    - [Node](#k8s.io.api.core.v1.Node)
    - [NodeAddress](#k8s.io.api.core.v1.NodeAddress)
    - [NodeAffinity](#k8s.io.api.core.v1.NodeAffinity)
    - [NodeCondition](#k8s.io.api.core.v1.NodeCondition)
    - [NodeConfigSource](#k8s.io.api.core.v1.NodeConfigSource)
    - [NodeDaemonEndpoints](#k8s.io.api.core.v1.NodeDaemonEndpoints)
    - [NodeList](#k8s.io.api.core.v1.NodeList)
    - [NodeProxyOptions](#k8s.io.api.core.v1.NodeProxyOptions)
    - [NodeResources](#k8s.io.api.core.v1.NodeResources)
    - [NodeResources.CapacityEntry](#k8s.io.api.core.v1.NodeResources.CapacityEntry)
    - [NodeSelector](#k8s.io.api.core.v1.NodeSelector)
    - [NodeSelectorRequirement](#k8s.io.api.core.v1.NodeSelectorRequirement)
    - [NodeSelectorTerm](#k8s.io.api.core.v1.NodeSelectorTerm)
    - [NodeSpec](#k8s.io.api.core.v1.NodeSpec)
    - [NodeStatus](#k8s.io.api.core.v1.NodeStatus)
    - [NodeStatus.AllocatableEntry](#k8s.io.api.core.v1.NodeStatus.AllocatableEntry)
    - [NodeStatus.CapacityEntry](#k8s.io.api.core.v1.NodeStatus.CapacityEntry)
    - [NodeSystemInfo](#k8s.io.api.core.v1.NodeSystemInfo)
    - [ObjectFieldSelector](#k8s.io.api.core.v1.ObjectFieldSelector)
    - [ObjectMeta](#k8s.io.api.core.v1.ObjectMeta)
    - [ObjectMeta.AnnotationsEntry](#k8s.io.api.core.v1.ObjectMeta.AnnotationsEntry)
    - [ObjectMeta.LabelsEntry](#k8s.io.api.core.v1.ObjectMeta.LabelsEntry)
    - [ObjectReference](#k8s.io.api.core.v1.ObjectReference)
    - [PersistentVolume](#k8s.io.api.core.v1.PersistentVolume)
    - [PersistentVolumeClaim](#k8s.io.api.core.v1.PersistentVolumeClaim)
    - [PersistentVolumeClaimCondition](#k8s.io.api.core.v1.PersistentVolumeClaimCondition)
    - [PersistentVolumeClaimList](#k8s.io.api.core.v1.PersistentVolumeClaimList)
    - [PersistentVolumeClaimSpec](#k8s.io.api.core.v1.PersistentVolumeClaimSpec)
    - [PersistentVolumeClaimStatus](#k8s.io.api.core.v1.PersistentVolumeClaimStatus)
    - [PersistentVolumeClaimStatus.CapacityEntry](#k8s.io.api.core.v1.PersistentVolumeClaimStatus.CapacityEntry)
    - [PersistentVolumeClaimVolumeSource](#k8s.io.api.core.v1.PersistentVolumeClaimVolumeSource)
    - [PersistentVolumeList](#k8s.io.api.core.v1.PersistentVolumeList)
    - [PersistentVolumeSource](#k8s.io.api.core.v1.PersistentVolumeSource)
    - [PersistentVolumeSpec](#k8s.io.api.core.v1.PersistentVolumeSpec)
    - [PersistentVolumeSpec.CapacityEntry](#k8s.io.api.core.v1.PersistentVolumeSpec.CapacityEntry)
    - [PersistentVolumeStatus](#k8s.io.api.core.v1.PersistentVolumeStatus)
    - [PhotonPersistentDiskVolumeSource](#k8s.io.api.core.v1.PhotonPersistentDiskVolumeSource)
    - [Pod](#k8s.io.api.core.v1.Pod)
    - [PodAffinity](#k8s.io.api.core.v1.PodAffinity)
    - [PodAffinityTerm](#k8s.io.api.core.v1.PodAffinityTerm)
    - [PodAntiAffinity](#k8s.io.api.core.v1.PodAntiAffinity)
    - [PodAttachOptions](#k8s.io.api.core.v1.PodAttachOptions)
    - [PodCondition](#k8s.io.api.core.v1.PodCondition)
    - [PodExecOptions](#k8s.io.api.core.v1.PodExecOptions)
    - [PodList](#k8s.io.api.core.v1.PodList)
    - [PodLogOptions](#k8s.io.api.core.v1.PodLogOptions)
    - [PodPortForwardOptions](#k8s.io.api.core.v1.PodPortForwardOptions)
    - [PodProxyOptions](#k8s.io.api.core.v1.PodProxyOptions)
    - [PodSecurityContext](#k8s.io.api.core.v1.PodSecurityContext)
    - [PodSignature](#k8s.io.api.core.v1.PodSignature)
    - [PodSpec](#k8s.io.api.core.v1.PodSpec)
    - [PodSpec.NodeSelectorEntry](#k8s.io.api.core.v1.PodSpec.NodeSelectorEntry)
    - [PodStatus](#k8s.io.api.core.v1.PodStatus)
    - [PodStatusResult](#k8s.io.api.core.v1.PodStatusResult)
    - [PodTemplate](#k8s.io.api.core.v1.PodTemplate)
    - [PodTemplateList](#k8s.io.api.core.v1.PodTemplateList)
    - [PodTemplateSpec](#k8s.io.api.core.v1.PodTemplateSpec)
    - [PortworxVolumeSource](#k8s.io.api.core.v1.PortworxVolumeSource)
    - [Preconditions](#k8s.io.api.core.v1.Preconditions)
    - [PreferAvoidPodsEntry](#k8s.io.api.core.v1.PreferAvoidPodsEntry)
    - [PreferredSchedulingTerm](#k8s.io.api.core.v1.PreferredSchedulingTerm)
    - [Probe](#k8s.io.api.core.v1.Probe)
    - [ProjectedVolumeSource](#k8s.io.api.core.v1.ProjectedVolumeSource)
    - [QuobyteVolumeSource](#k8s.io.api.core.v1.QuobyteVolumeSource)
    - [RBDPersistentVolumeSource](#k8s.io.api.core.v1.RBDPersistentVolumeSource)
    - [RBDVolumeSource](#k8s.io.api.core.v1.RBDVolumeSource)
    - [RangeAllocation](#k8s.io.api.core.v1.RangeAllocation)
    - [ReplicationController](#k8s.io.api.core.v1.ReplicationController)
    - [ReplicationControllerCondition](#k8s.io.api.core.v1.ReplicationControllerCondition)
    - [ReplicationControllerList](#k8s.io.api.core.v1.ReplicationControllerList)
    - [ReplicationControllerSpec](#k8s.io.api.core.v1.ReplicationControllerSpec)
    - [ReplicationControllerSpec.SelectorEntry](#k8s.io.api.core.v1.ReplicationControllerSpec.SelectorEntry)
    - [ReplicationControllerStatus](#k8s.io.api.core.v1.ReplicationControllerStatus)
    - [ResourceFieldSelector](#k8s.io.api.core.v1.ResourceFieldSelector)
    - [ResourceQuota](#k8s.io.api.core.v1.ResourceQuota)
    - [ResourceQuotaList](#k8s.io.api.core.v1.ResourceQuotaList)
    - [ResourceQuotaSpec](#k8s.io.api.core.v1.ResourceQuotaSpec)
    - [ResourceQuotaSpec.HardEntry](#k8s.io.api.core.v1.ResourceQuotaSpec.HardEntry)
    - [ResourceQuotaStatus](#k8s.io.api.core.v1.ResourceQuotaStatus)
    - [ResourceQuotaStatus.HardEntry](#k8s.io.api.core.v1.ResourceQuotaStatus.HardEntry)
    - [ResourceQuotaStatus.UsedEntry](#k8s.io.api.core.v1.ResourceQuotaStatus.UsedEntry)
    - [ResourceRequirements](#k8s.io.api.core.v1.ResourceRequirements)
    - [ResourceRequirements.LimitsEntry](#k8s.io.api.core.v1.ResourceRequirements.LimitsEntry)
    - [ResourceRequirements.RequestsEntry](#k8s.io.api.core.v1.ResourceRequirements.RequestsEntry)
    - [SELinuxOptions](#k8s.io.api.core.v1.SELinuxOptions)
    - [ScaleIOPersistentVolumeSource](#k8s.io.api.core.v1.ScaleIOPersistentVolumeSource)
    - [ScaleIOVolumeSource](#k8s.io.api.core.v1.ScaleIOVolumeSource)
    - [Secret](#k8s.io.api.core.v1.Secret)
    - [Secret.DataEntry](#k8s.io.api.core.v1.Secret.DataEntry)
    - [Secret.StringDataEntry](#k8s.io.api.core.v1.Secret.StringDataEntry)
    - [SecretEnvSource](#k8s.io.api.core.v1.SecretEnvSource)
    - [SecretKeySelector](#k8s.io.api.core.v1.SecretKeySelector)
    - [SecretList](#k8s.io.api.core.v1.SecretList)
    - [SecretProjection](#k8s.io.api.core.v1.SecretProjection)
    - [SecretReference](#k8s.io.api.core.v1.SecretReference)
    - [SecretVolumeSource](#k8s.io.api.core.v1.SecretVolumeSource)
    - [SecurityContext](#k8s.io.api.core.v1.SecurityContext)
    - [SerializedReference](#k8s.io.api.core.v1.SerializedReference)
    - [Service](#k8s.io.api.core.v1.Service)
    - [ServiceAccount](#k8s.io.api.core.v1.ServiceAccount)
    - [ServiceAccountList](#k8s.io.api.core.v1.ServiceAccountList)
    - [ServiceList](#k8s.io.api.core.v1.ServiceList)
    - [ServicePort](#k8s.io.api.core.v1.ServicePort)
    - [ServiceProxyOptions](#k8s.io.api.core.v1.ServiceProxyOptions)
    - [ServiceSpec](#k8s.io.api.core.v1.ServiceSpec)
    - [ServiceSpec.SelectorEntry](#k8s.io.api.core.v1.ServiceSpec.SelectorEntry)
    - [ServiceStatus](#k8s.io.api.core.v1.ServiceStatus)
    - [SessionAffinityConfig](#k8s.io.api.core.v1.SessionAffinityConfig)
    - [StorageOSPersistentVolumeSource](#k8s.io.api.core.v1.StorageOSPersistentVolumeSource)
    - [StorageOSVolumeSource](#k8s.io.api.core.v1.StorageOSVolumeSource)
    - [Sysctl](#k8s.io.api.core.v1.Sysctl)
    - [TCPSocketAction](#k8s.io.api.core.v1.TCPSocketAction)
    - [Taint](#k8s.io.api.core.v1.Taint)
    - [Toleration](#k8s.io.api.core.v1.Toleration)
    - [Volume](#k8s.io.api.core.v1.Volume)
    - [VolumeMount](#k8s.io.api.core.v1.VolumeMount)
    - [VolumeProjection](#k8s.io.api.core.v1.VolumeProjection)
    - [VolumeSource](#k8s.io.api.core.v1.VolumeSource)
    - [VsphereVirtualDiskVolumeSource](#k8s.io.api.core.v1.VsphereVirtualDiskVolumeSource)
    - [WeightedPodAffinityTerm](#k8s.io.api.core.v1.WeightedPodAffinityTerm)
  
  
  
  

- [seldon_deployment.proto](#seldon_deployment.proto)
    - [DeploymentSpec](#seldon.protos.DeploymentSpec)
    - [DeploymentSpec.AnnotationsEntry](#seldon.protos.DeploymentSpec.AnnotationsEntry)
    - [DeploymentStatus](#seldon.protos.DeploymentStatus)
    - [Endpoint](#seldon.protos.Endpoint)
    - [Parameter](#seldon.protos.Parameter)
    - [PredictiveUnit](#seldon.protos.PredictiveUnit)
    - [PredictorSpec](#seldon.protos.PredictorSpec)
    - [PredictorSpec.AnnotationsEntry](#seldon.protos.PredictorSpec.AnnotationsEntry)
    - [PredictorStatus](#seldon.protos.PredictorStatus)
    - [SeldonDeployment](#seldon.protos.SeldonDeployment)
  
    - [Endpoint.EndpointType](#seldon.protos.Endpoint.EndpointType)
    - [Parameter.ParameterType](#seldon.protos.Parameter.ParameterType)
    - [PredictiveUnit.PredictiveUnitSubtype](#seldon.protos.PredictiveUnit.PredictiveUnitSubtype)
    - [PredictiveUnit.PredictiveUnitType](#seldon.protos.PredictiveUnit.PredictiveUnitType)
  
  
  

- [Scalar Value Types](#scalar-value-types)



<a name="generated.proto"/>
<p align="right"><a href="#top">Top</a></p>

## generated.proto



<a name="k8s.io.apimachinery.pkg.util.intstr.IntOrString"/>

### IntOrString
IntOrString is a type that can hold an int32 or a string.  When used in
JSON or YAML marshalling and unmarshalling, it produces or consumes the
inner type.  This allows you to have, for example, a JSON field that can
accept a name or number.
TODO: Rename to Int32OrString

&#43;protobuf=true
&#43;protobuf.options.(gogoproto.goproto_stringer)=false
&#43;k8s:openapi-gen=true


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [int64](#int64) | optional |  |
| intVal | [int32](#int32) | optional |  |
| strVal | [string](#string) | optional |  |





 

 

 

 



<a name="generated.proto"/>
<p align="right"><a href="#top">Top</a></p>

## generated.proto



<a name="k8s.io.apimachinery.pkg.runtime.RawExtension"/>

### RawExtension
RawExtension is used to hold extensions in external versions.

To use this, make a field which has RawExtension as its type in your external, versioned
struct, and Object in your internal struct. You also need to register your
various plugin types.

Internal package:
type MyAPIObject struct {
	runtime.TypeMeta `json:&#34;,inline&#34;`
	MyPlugin runtime.Object `json:&#34;myPlugin&#34;`
}
type PluginA struct {
	AOption string `json:&#34;aOption&#34;`
}

External package:
type MyAPIObject struct {
	runtime.TypeMeta `json:&#34;,inline&#34;`
	MyPlugin runtime.RawExtension `json:&#34;myPlugin&#34;`
}
type PluginA struct {
	AOption string `json:&#34;aOption&#34;`
}

On the wire, the JSON will look something like this:
{
	&#34;kind&#34;:&#34;MyAPIObject&#34;,
	&#34;apiVersion&#34;:&#34;v1&#34;,
	&#34;myPlugin&#34;: {
		&#34;kind&#34;:&#34;PluginA&#34;,
		&#34;aOption&#34;:&#34;foo&#34;,
	},
}

So what happens? Decode first uses json or yaml to unmarshal the serialized data into
your external MyAPIObject. That causes the raw JSON to be stored, but not unpacked.
The next step is to copy (using pkg/conversion) into the internal struct. The runtime
package&#39;s DefaultScheme has conversion functions installed which will unpack the
JSON stored in RawExtension, turning it into the correct object type, and storing it
in the Object. (TODO: In the case where the object is of an unknown type, a
runtime.Unknown object will be created and stored.)

&#43;k8s:deepcopy-gen=true
&#43;protobuf=true
&#43;k8s:openapi-gen=true


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| raw | [bytes](#bytes) | optional | Raw is the underlying serialization of this object. TODO: Determine how to detect ContentType and ContentEncoding of &#39;Raw&#39; data. |






<a name="k8s.io.apimachinery.pkg.runtime.TypeMeta"/>

### TypeMeta
TypeMeta is shared by all top level objects. The proper way to use it is to inline it in your type,
like this:
type MyAwesomeAPIObject struct {
runtime.TypeMeta    `json:&#34;,inline&#34;`
... // other fields
}
func (obj *MyAwesomeAPIObject) SetGroupVersionKind(gvk *metav1.GroupVersionKind) { metav1.UpdateTypeMeta(obj,gvk) }; GroupVersionKind() *GroupVersionKind

TypeMeta is provided here for convenience. You may use it directly from this package or define
your own with the same fields.

&#43;k8s:deepcopy-gen=false
&#43;protobuf=true
&#43;k8s:openapi-gen=true


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| apiVersion | [string](#string) | optional | &#43;optional |
| kind | [string](#string) | optional | &#43;optional |






<a name="k8s.io.apimachinery.pkg.runtime.Unknown"/>

### Unknown
Unknown allows api objects with unknown types to be passed-through. This can be used
to deal with the API objects from a plug-in. Unknown objects still have functioning
TypeMeta features-- kind, version, etc.
TODO: Make this object have easy access to field based accessors and settors for
metadata and field mutatation.

&#43;k8s:deepcopy-gen=true
&#43;k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
&#43;protobuf=true
&#43;k8s:openapi-gen=true


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| typeMeta | [TypeMeta](#k8s.io.apimachinery.pkg.runtime.TypeMeta) | optional |  |
| raw | [bytes](#bytes) | optional | Raw will hold the complete serialized object which couldn&#39;t be matched with a registered type. Most likely, nothing should be done with this except for passing it through the system. |
| contentEncoding | [string](#string) | optional | ContentEncoding is encoding used to encode &#39;Raw&#39; data. Unspecified means no encoding. |
| contentType | [string](#string) | optional | ContentType is serialization method used to serialize &#39;Raw&#39;. Unspecified means ContentTypeJSON. |





 

 

 

 



<a name="generated.proto"/>
<p align="right"><a href="#top">Top</a></p>

## generated.proto


 

 

 

 



<a name="generated.proto"/>
<p align="right"><a href="#top">Top</a></p>

## generated.proto



<a name="k8s.io.apimachinery.pkg.apis.meta.v1.APIGroup"/>

### APIGroup
APIGroup contains the name, the supported versions, and the preferred version
of a group.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | optional | name is the name of the group. |
| versions | [GroupVersionForDiscovery](#k8s.io.apimachinery.pkg.apis.meta.v1.GroupVersionForDiscovery) | repeated | versions are the versions supported in this group. |
| preferredVersion | [GroupVersionForDiscovery](#k8s.io.apimachinery.pkg.apis.meta.v1.GroupVersionForDiscovery) | optional | preferredVersion is the version preferred by the API server, which probably is the storage version. &#43;optional |
| serverAddressByClientCIDRs | [ServerAddressByClientCIDR](#k8s.io.apimachinery.pkg.apis.meta.v1.ServerAddressByClientCIDR) | repeated | a map of client CIDR to server address that is serving this group. This is to help clients reach servers in the most network-efficient way possible. Clients can use the appropriate server address as per the CIDR that they match. In case of multiple matches, clients should use the longest matching CIDR. The server returns only those CIDRs that it thinks that the client can match. For example: the master will return an internal IP CIDR only, if the client reaches the server using an internal IP. Server looks at X-Forwarded-For header or X-Real-Ip header or request.RemoteAddr (in that order) to get the client IP. |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.APIGroupList"/>

### APIGroupList
APIGroupList is a list of APIGroup, to allow clients to discover the API at
apis.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| groups | [APIGroup](#k8s.io.apimachinery.pkg.apis.meta.v1.APIGroup) | repeated | groups is a list of APIGroup. |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.APIResource"/>

### APIResource
APIResource specifies the name of a resource and whether it is namespaced.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | optional | name is the plural name of the resource. |
| singularName | [string](#string) | optional | singularName is the singular name of the resource. This allows clients to handle plural and singular opaquely. The singularName is more correct for reporting status on a single item and both singular and plural are allowed from the kubectl CLI interface. |
| namespaced | [bool](#bool) | optional | namespaced indicates if a resource is namespaced or not. |
| group | [string](#string) | optional | group is the preferred group of the resource. Empty implies the group of the containing resource list. For subresources, this may have a different value, for example: Scale&#34;. |
| version | [string](#string) | optional | version is the preferred version of the resource. Empty implies the version of the containing resource list For subresources, this may have a different value, for example: v1 (while inside a v1beta1 version of the core resource&#39;s group)&#34;. |
| kind | [string](#string) | optional | kind is the kind for the resource (e.g. &#39;Foo&#39; is the kind for a resource &#39;foo&#39;) |
| verbs | [Verbs](#k8s.io.apimachinery.pkg.apis.meta.v1.Verbs) | optional | verbs is a list of supported kube verbs (this includes get, list, watch, create, update, patch, delete, deletecollection, and proxy) |
| shortNames | [string](#string) | repeated | shortNames is a list of suggested short names of the resource. |
| categories | [string](#string) | repeated | categories is a list of the grouped resources this resource belongs to (e.g. &#39;all&#39;) |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.APIResourceList"/>

### APIResourceList
APIResourceList is a list of APIResource, it is used to expose the name of the
resources supported in a specific group and version, and if the resource
is namespaced.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| groupVersion | [string](#string) | optional | groupVersion is the group and version this APIResourceList is for. |
| resources | [APIResource](#k8s.io.apimachinery.pkg.apis.meta.v1.APIResource) | repeated | resources contains the name of the resources and if they are namespaced. |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.APIVersions"/>

### APIVersions
APIVersions lists the versions that are available, to allow clients to
discover the API at /api, which is the root path of the legacy v1 API.

&#43;protobuf.options.(gogoproto.goproto_stringer)=false
&#43;k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| versions | [string](#string) | repeated | versions are the api versions that are available. |
| serverAddressByClientCIDRs | [ServerAddressByClientCIDR](#k8s.io.apimachinery.pkg.apis.meta.v1.ServerAddressByClientCIDR) | repeated | a map of client CIDR to server address that is serving this group. This is to help clients reach servers in the most network-efficient way possible. Clients can use the appropriate server address as per the CIDR that they match. In case of multiple matches, clients should use the longest matching CIDR. The server returns only those CIDRs that it thinks that the client can match. For example: the master will return an internal IP CIDR only, if the client reaches the server using an internal IP. Server looks at X-Forwarded-For header or X-Real-Ip header or request.RemoteAddr (in that order) to get the client IP. |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.DeleteOptions"/>

### DeleteOptions
DeleteOptions may be provided when deleting an API object.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| gracePeriodSeconds | [int64](#int64) | optional | The duration in seconds before the object should be deleted. Value must be non-negative integer. The value zero indicates delete immediately. If this value is nil, the default grace period for the specified type will be used. Defaults to a per object value if not specified. zero means delete immediately. &#43;optional |
| preconditions | [Preconditions](#k8s.io.apimachinery.pkg.apis.meta.v1.Preconditions) | optional | Must be fulfilled before a deletion is carried out. If not possible, a 409 Conflict status will be returned. &#43;optional |
| orphanDependents | [bool](#bool) | optional | Deprecated: please use the PropagationPolicy, this field will be deprecated in 1.7. Should the dependent objects be orphaned. If true/false, the &#34;orphan&#34; finalizer will be added to/removed from the object&#39;s finalizers list. Either this field or PropagationPolicy may be set, but not both. &#43;optional |
| propagationPolicy | [string](#string) | optional | Whether and how garbage collection will be performed. Either this field or OrphanDependents may be set, but not both. The default policy is decided by the existing finalizer set in the metadata.finalizers and the resource-specific default policy. &#43;optional |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.Duration"/>

### Duration
Duration is a wrapper around time.Duration which supports correct
marshaling to YAML and JSON. In particular, it marshals into strings, which
can be used as map keys in json.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| duration | [int64](#int64) | optional |  |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.ExportOptions"/>

### ExportOptions
ExportOptions is the query options to the standard REST get call.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| export | [bool](#bool) | optional | Should this value be exported. Export strips fields that a user can not specify. |
| exact | [bool](#bool) | optional | Should the export be exact. Exact export maintains cluster-specific fields like &#39;Namespace&#39;. |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.GetOptions"/>

### GetOptions
GetOptions is the standard query options to the standard REST get call.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resourceVersion | [string](#string) | optional | When specified: - if unset, then the result is returned from remote storage based on quorum-read flag; - if it&#39;s 0, then we simply return what we currently have in cache, no guarantee; - if set to non zero, then the result is at least as fresh as given rv. |
| includeUninitialized | [bool](#bool) | optional | If true, partially initialized resources are included in the response. &#43;optional |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.GroupKind"/>

### GroupKind
GroupKind specifies a Group and a Kind, but does not force a version.  This is useful for identifying
concepts during lookup stages without having partially valid types

&#43;protobuf.options.(gogoproto.goproto_stringer)=false


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group | [string](#string) | optional |  |
| kind | [string](#string) | optional |  |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.GroupResource"/>

### GroupResource
GroupResource specifies a Group and a Resource, but does not force a version.  This is useful for identifying
concepts during lookup stages without having partially valid types

&#43;protobuf.options.(gogoproto.goproto_stringer)=false


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group | [string](#string) | optional |  |
| resource | [string](#string) | optional |  |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.GroupVersion"/>

### GroupVersion
GroupVersion contains the &#34;group&#34; and the &#34;version&#34;, which uniquely identifies the API.

&#43;protobuf.options.(gogoproto.goproto_stringer)=false


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group | [string](#string) | optional |  |
| version | [string](#string) | optional |  |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.GroupVersionForDiscovery"/>

### GroupVersionForDiscovery
GroupVersion contains the &#34;group/version&#34; and &#34;version&#34; string of a version.
It is made a struct to keep extensibility.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| groupVersion | [string](#string) | optional | groupVersion specifies the API group and version in the form &#34;group/version&#34; |
| version | [string](#string) | optional | version specifies the version in the form of &#34;version&#34;. This is to save the clients the trouble of splitting the GroupVersion. |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.GroupVersionKind"/>

### GroupVersionKind
GroupVersionKind unambiguously identifies a kind.  It doesn&#39;t anonymously include GroupVersion
to avoid automatic coersion.  It doesn&#39;t use a GroupVersion to avoid custom marshalling

&#43;protobuf.options.(gogoproto.goproto_stringer)=false


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group | [string](#string) | optional |  |
| version | [string](#string) | optional |  |
| kind | [string](#string) | optional |  |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.GroupVersionResource"/>

### GroupVersionResource
GroupVersionResource unambiguously identifies a resource.  It doesn&#39;t anonymously include GroupVersion
to avoid automatic coersion.  It doesn&#39;t use a GroupVersion to avoid custom marshalling

&#43;protobuf.options.(gogoproto.goproto_stringer)=false


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group | [string](#string) | optional |  |
| version | [string](#string) | optional |  |
| resource | [string](#string) | optional |  |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.Initializer"/>

### Initializer
Initializer is information about an initializer that has not yet completed.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | optional | name of the process that is responsible for initializing this object. |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.Initializers"/>

### Initializers
Initializers tracks the progress of initialization.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pending | [Initializer](#k8s.io.apimachinery.pkg.apis.meta.v1.Initializer) | repeated | Pending is a list of initializers that must execute in order before this object is visible. When the last pending initializer is removed, and no failing result is set, the initializers struct will be set to nil and the object is considered as initialized and visible to all clients. &#43;patchMergeKey=name &#43;patchStrategy=merge |
| result | [Status](#k8s.io.apimachinery.pkg.apis.meta.v1.Status) | optional | If result is set with the Failure field, the object will be persisted to storage and then deleted, ensuring that other clients can observe the deletion. |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.LabelSelector"/>

### LabelSelector
A label selector is a label query over a set of resources. The result of matchLabels and
matchExpressions are ANDed. An empty label selector matches all objects. A null
label selector matches no objects.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| matchLabels | [LabelSelector.MatchLabelsEntry](#k8s.io.apimachinery.pkg.apis.meta.v1.LabelSelector.MatchLabelsEntry) | repeated | matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is &#34;key&#34;, the operator is &#34;In&#34;, and the values array contains only &#34;value&#34;. The requirements are ANDed. &#43;optional |
| matchExpressions | [LabelSelectorRequirement](#k8s.io.apimachinery.pkg.apis.meta.v1.LabelSelectorRequirement) | repeated | matchExpressions is a list of label selector requirements. The requirements are ANDed. &#43;optional |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.LabelSelector.MatchLabelsEntry"/>

### LabelSelector.MatchLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [string](#string) | optional |  |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.LabelSelectorRequirement"/>

### LabelSelectorRequirement
A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional | key is the label key that the selector applies to. &#43;patchMergeKey=key &#43;patchStrategy=merge |
| operator | [string](#string) | optional | operator represents a key&#39;s relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist. |
| values | [string](#string) | repeated | values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch. &#43;optional |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.List"/>

### List
List holds a list of objects, which may not be known by the server.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [ListMeta](#k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta) | optional | Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds &#43;optional |
| items | [.k8s.io.apimachinery.pkg.runtime.RawExtension](#k8s.io.apimachinery.pkg.apis.meta.v1..k8s.io.apimachinery.pkg.runtime.RawExtension) | repeated | List of objects |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta"/>

### ListMeta
ListMeta describes metadata that synthetic resources must have, including lists and
various status objects. A resource may have only one of {ObjectMeta, ListMeta}.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| selfLink | [string](#string) | optional | selfLink is a URL representing this object. Populated by the system. Read-only. &#43;optional |
| resourceVersion | [string](#string) | optional | String that identifies the server&#39;s internal version of this object that can be used by clients to determine when objects have changed. Value must be treated as opaque by clients and passed unmodified back to the server. Populated by the system. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency &#43;optional |
| continue | [string](#string) | optional | continue may be set if the user set a limit on the number of items returned, and indicates that the server has more data available. The value is opaque and may be used to issue another request to the endpoint that served this list to retrieve the next set of available objects. Continuing a list may not be possible if the server configuration has changed or more than a few minutes have passed. The resourceVersion field returned when using this continue value will be identical to the value in the first response. |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.ListOptions"/>

### ListOptions
ListOptions is the query options to a standard REST list call.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| labelSelector | [string](#string) | optional | A selector to restrict the list of returned objects by their labels. Defaults to everything. &#43;optional |
| fieldSelector | [string](#string) | optional | A selector to restrict the list of returned objects by their fields. Defaults to everything. &#43;optional |
| includeUninitialized | [bool](#bool) | optional | If true, partially initialized resources are included in the response. &#43;optional |
| watch | [bool](#bool) | optional | Watch for changes to the described resources and return them as a stream of add, update, and remove notifications. Specify resourceVersion. &#43;optional |
| resourceVersion | [string](#string) | optional | When specified with a watch call, shows changes that occur after that particular version of a resource. Defaults to changes from the beginning of history. When specified for list: - if unset, then the result is returned from remote storage based on quorum-read flag; - if it&#39;s 0, then we simply return what we currently have in cache, no guarantee; - if set to non zero, then the result is at least as fresh as given rv. &#43;optional |
| timeoutSeconds | [int64](#int64) | optional | Timeout for the list/watch call. &#43;optional |
| limit | [int64](#int64) | optional | limit is a maximum number of responses to return for a list call. If more items exist, the server will set the `continue` field on the list metadata to a value that can be used with the same initial query to retrieve the next set of results. Setting a limit may return fewer than the requested amount of items (up to zero items) in the event all requested objects are filtered out and clients should only use the presence of the continue field to determine whether more results are available. Servers may choose not to support the limit argument and will return all of the available results. If limit is specified and the continue field is empty, clients may assume that no more results are available. This field is not supported if watch is true. The server guarantees that the objects returned when using continue will be identical to issuing a single list call without a limit - that is, no objects created, modified, or deleted after the first request is issued will be included in any subsequent continued requests. This is sometimes referred to as a consistent snapshot, and ensures that a client that is using limit to receive smaller chunks of a very large result can ensure they see all possible objects. If objects are updated during a chunked list the version of the object that was present at the time the first list result was calculated is returned. |
| continue | [string](#string) | optional | The continue option should be set when retrieving more results from the server. Since this value is server defined, clients may only use the continue value from a previous query result with identical query parameters (except for the value of continue) and the server may reject a continue value it does not recognize. If the specified continue value is no longer valid whether due to expiration (generally five to fifteen minutes) or a configuration change on the server the server will respond with a 410 ResourceExpired error indicating the client must restart their list without the continue field. This field is not supported when watch is true. Clients may start a watch from the last resourceVersion value returned by the server and not miss any modifications. |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.MicroTime"/>

### MicroTime
MicroTime is version of Time with microsecond level precision.

&#43;protobuf.options.marshal=false
&#43;protobuf.as=Timestamp
&#43;protobuf.options.(gogoproto.goproto_stringer)=false


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| seconds | [int64](#int64) | optional | Represents seconds of UTC time since Unix epoch 1970-01-01T00:00:00Z. Must be from 0001-01-01T00:00:00Z to 9999-12-31T23:59:59Z inclusive. |
| nanos | [int32](#int32) | optional | Non-negative fractions of a second at nanosecond resolution. Negative second values with fractions must still have non-negative nanos values that count forward in time. Must be from 0 to 999,999,999 inclusive. This field may be limited in precision depending on context. |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta"/>

### ObjectMeta
ObjectMeta is metadata that all persisted resources must have, which includes all objects
users must create.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | optional | Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names &#43;optional |
| generateName | [string](#string) | optional | GenerateName is an optional prefix, used by the server, to generate a unique name ONLY IF the Name field has not been provided. If this field is used, the name returned to the client will be different than the name passed. This value will also be combined with a unique suffix. The provided value has the same validation rules as the Name field, and may be truncated by the length of the suffix required to make the value unique on the server. If this field is specified and the generated name exists, the server will NOT return a 409 - instead, it will either return 201 Created or 500 with Reason ServerTimeout indicating a unique name could not be found in the time allotted, and the client should retry (optionally after the time indicated in the Retry-After header). Applied only if Name is not specified. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#idempotency &#43;optional |
| namespace | [string](#string) | optional | Namespace defines the space within each name must be unique. An empty namespace is equivalent to the &#34;default&#34; namespace, but &#34;default&#34; is the canonical representation. Not all objects are required to be scoped to a namespace - the value of this field for those objects will be empty. Must be a DNS_LABEL. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/namespaces &#43;optional |
| selfLink | [string](#string) | optional | SelfLink is a URL representing this object. Populated by the system. Read-only. &#43;optional |
| uid | [string](#string) | optional | UID is the unique in time and space value for this object. It is typically generated by the server on successful creation of a resource and is not allowed to change on PUT operations. Populated by the system. Read-only. More info: http://kubernetes.io/docs/user-guide/identifiers#uids &#43;optional |
| resourceVersion | [string](#string) | optional | An opaque value that represents the internal version of this object that can be used by clients to determine when objects have changed. May be used for optimistic concurrency, change detection, and the watch operation on a resource or set of resources. Clients must treat these values as opaque and passed unmodified back to the server. They may only be valid for a particular resource or set of resources. Populated by the system. Read-only. Value must be treated as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency &#43;optional |
| generation | [int64](#int64) | optional | A sequence number representing a specific generation of the desired state. Populated by the system. Read-only. &#43;optional |
| creationTimestamp | [Time](#k8s.io.apimachinery.pkg.apis.meta.v1.Time) | optional | CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC. Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata &#43;optional |
| deletionTimestamp | [Time](#k8s.io.apimachinery.pkg.apis.meta.v1.Time) | optional | DeletionTimestamp is RFC 3339 date and time at which this resource will be deleted. This field is set by the server when a graceful deletion is requested by the user, and is not directly settable by a client. The resource is expected to be deleted (no longer visible from resource lists, and not reachable by name) after the time in this field. Once set, this value may not be unset or be set further into the future, although it may be shortened or the resource may be deleted prior to this time. For example, a user may request that a pod is deleted in 30 seconds. The Kubelet will react by sending a graceful termination signal to the containers in the pod. After that 30 seconds, the Kubelet will send a hard termination signal (SIGKILL) to the container and after cleanup, remove the pod from the API. In the presence of network partitions, this object may still exist after this timestamp, until an administrator or automated process can determine the resource is fully terminated. If not set, graceful deletion of the object has not been requested. Populated by the system when a graceful deletion is requested. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata &#43;optional |
| deletionGracePeriodSeconds | [int64](#int64) | optional | Number of seconds allowed for this object to gracefully terminate before it will be removed from the system. Only set when deletionTimestamp is also set. May only be shortened. Read-only. &#43;optional |
| labels | [ObjectMeta.LabelsEntry](#k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta.LabelsEntry) | repeated | Map of string keys and values that can be used to organize and categorize (scope and select) objects. May match selectors of replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels &#43;optional |
| annotations | [ObjectMeta.AnnotationsEntry](#k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta.AnnotationsEntry) | repeated | Annotations is an unstructured key value map stored with a resource that may be set by external tools to store and retrieve arbitrary metadata. They are not queryable and should be preserved when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations &#43;optional |
| ownerReferences | [OwnerReference](#k8s.io.apimachinery.pkg.apis.meta.v1.OwnerReference) | repeated | List of objects depended by this object. If ALL objects in the list have been deleted, this object will be garbage collected. If this object is managed by a controller, then an entry in this list will point to this controller, with the controller field set to true. There cannot be more than one managing controller. &#43;optional &#43;patchMergeKey=uid &#43;patchStrategy=merge |
| initializers | [Initializers](#k8s.io.apimachinery.pkg.apis.meta.v1.Initializers) | optional | An initializer is a controller which enforces some system invariant at object creation time. This field is a list of initializers that have not yet acted on this object. If nil or empty, this object has been completely initialized. Otherwise, the object is considered uninitialized and is hidden (in list/watch and get calls) from clients that haven&#39;t explicitly asked to observe uninitialized objects. When an object is created, the system will populate this list with the current set of initializers. Only privileged users may set or modify this list. Once it is empty, it may not be modified further by any user. |
| finalizers | [string](#string) | repeated | Must be empty before the object is deleted from the registry. Each entry is an identifier for the responsible component that will remove the entry from the list. If the deletionTimestamp of the object is non-nil, entries in this list can only be removed. &#43;optional &#43;patchStrategy=merge |
| clusterName | [string](#string) | optional | The name of the cluster which the object belongs to. This is used to distinguish resources with same name and namespace in different clusters. This field is not set anywhere right now and apiserver is going to ignore it if set in create or update request. &#43;optional |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta.AnnotationsEntry"/>

### ObjectMeta.AnnotationsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [string](#string) | optional |  |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta.LabelsEntry"/>

### ObjectMeta.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [string](#string) | optional |  |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.OwnerReference"/>

### OwnerReference
OwnerReference contains enough information to let you identify an owning
object. Currently, an owning object must be in the same namespace, so there
is no namespace field.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| apiVersion | [string](#string) | optional | API version of the referent. |
| kind | [string](#string) | optional | Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds |
| name | [string](#string) | optional | Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names |
| uid | [string](#string) | optional | UID of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#uids |
| controller | [bool](#bool) | optional | If true, this reference points to the managing controller. &#43;optional |
| blockOwnerDeletion | [bool](#bool) | optional | If true, AND if the owner has the &#34;foregroundDeletion&#34; finalizer, then the owner cannot be deleted from the key-value store until this reference is removed. Defaults to false. To set this field, a user needs &#34;delete&#34; permission of the owner, otherwise 422 (Unprocessable Entity) will be returned. &#43;optional |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.Preconditions"/>

### Preconditions
Preconditions must be fulfilled before an operation (update, delete, etc.) is carried out.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| uid | [string](#string) | optional | Specifies the target UID. &#43;optional |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.RootPaths"/>

### RootPaths
RootPaths lists the paths available at root.
For example: &#34;/healthz&#34;, &#34;/apis&#34;.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| paths | [string](#string) | repeated | paths are the paths available at root. |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.ServerAddressByClientCIDR"/>

### ServerAddressByClientCIDR
ServerAddressByClientCIDR helps the client to determine the server address that they should use, depending on the clientCIDR that they match.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| clientCIDR | [string](#string) | optional | The CIDR with which clients can match their IP to figure out the server address that they should use. |
| serverAddress | [string](#string) | optional | Address of this server, suitable for a client that matches the above CIDR. This can be a hostname, hostname:port, IP or IP:port. |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.Status"/>

### Status
Status is a return value for calls that don&#39;t return other objects.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [ListMeta](#k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta) | optional | Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds &#43;optional |
| status | [string](#string) | optional | Status of the operation. One of: &#34;Success&#34; or &#34;Failure&#34;. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status &#43;optional |
| message | [string](#string) | optional | A human-readable description of the status of this operation. &#43;optional |
| reason | [string](#string) | optional | A machine-readable description of why this operation is in the &#34;Failure&#34; status. If this value is empty there is no information available. A Reason clarifies an HTTP status code but does not override it. &#43;optional |
| details | [StatusDetails](#k8s.io.apimachinery.pkg.apis.meta.v1.StatusDetails) | optional | Extended data associated with the reason. Each reason may define its own extended details. This field is optional and the data returned is not guaranteed to conform to any schema except that defined by the reason type. &#43;optional |
| code | [int32](#int32) | optional | Suggested HTTP return code for this status, 0 if not set. &#43;optional |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.StatusCause"/>

### StatusCause
StatusCause provides more information about an api.Status failure, including
cases when multiple errors are encountered.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| reason | [string](#string) | optional | A machine-readable description of the cause of the error. If this value is empty there is no information available. &#43;optional |
| message | [string](#string) | optional | A human-readable description of the cause of the error. This field may be presented as-is to a reader. &#43;optional |
| field | [string](#string) | optional | The field of the resource that has caused this error, as named by its JSON serialization. May include dot and postfix notation for nested attributes. Arrays are zero-indexed. Fields may appear more than once in an array of causes due to fields having multiple errors. Optional. Examples: &#34;name&#34; - the field &#34;name&#34; on the current resource &#34;items[0].name&#34; - the field &#34;name&#34; on the first array entry in &#34;items&#34; &#43;optional |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.StatusDetails"/>

### StatusDetails
StatusDetails is a set of additional properties that MAY be set by the
server to provide additional information about a response. The Reason
field of a Status object defines what attributes will be set. Clients
must ignore fields that do not match the defined type of each attribute,
and should assume that any attribute may be empty, invalid, or under
defined.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | optional | The name attribute of the resource associated with the status StatusReason (when there is a single name which can be described). &#43;optional |
| group | [string](#string) | optional | The group attribute of the resource associated with the status StatusReason. &#43;optional |
| kind | [string](#string) | optional | The kind attribute of the resource associated with the status StatusReason. On some operations may differ from the requested resource Kind. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds &#43;optional |
| uid | [string](#string) | optional | UID of the resource. (when there is a single resource which can be described). More info: http://kubernetes.io/docs/user-guide/identifiers#uids &#43;optional |
| causes | [StatusCause](#k8s.io.apimachinery.pkg.apis.meta.v1.StatusCause) | repeated | The Causes array includes more details associated with the StatusReason failure. Not all StatusReasons may provide detailed causes. &#43;optional |
| retryAfterSeconds | [int32](#int32) | optional | If specified, the time in seconds before the operation should be retried. Some errors may indicate the client must take an alternate action - for those errors this field may indicate how long to wait before taking the alternate action. &#43;optional |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.Time"/>

### Time
Time is a wrapper around time.Time which supports correct
marshaling to YAML and JSON.  Wrappers are provided for many
of the factory methods that the time package offers.

&#43;protobuf.options.marshal=false
&#43;protobuf.as=Timestamp
&#43;protobuf.options.(gogoproto.goproto_stringer)=false


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| seconds | [int64](#int64) | optional | Represents seconds of UTC time since Unix epoch 1970-01-01T00:00:00Z. Must be from 0001-01-01T00:00:00Z to 9999-12-31T23:59:59Z inclusive. |
| nanos | [int32](#int32) | optional | Non-negative fractions of a second at nanosecond resolution. Negative second values with fractions must still have non-negative nanos values that count forward in time. Must be from 0 to 999,999,999 inclusive. This field may be limited in precision depending on context. |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.Timestamp"/>

### Timestamp
Timestamp is a struct that is equivalent to Time, but intended for
protobuf marshalling/unmarshalling. It is generated into a serialization
that matches Time. Do not use in Go structs.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| seconds | [int64](#int64) | optional | Represents seconds of UTC time since Unix epoch 1970-01-01T00:00:00Z. Must be from 0001-01-01T00:00:00Z to 9999-12-31T23:59:59Z inclusive. |
| nanos | [int32](#int32) | optional | Non-negative fractions of a second at nanosecond resolution. Negative second values with fractions must still have non-negative nanos values that count forward in time. Must be from 0 to 999,999,999 inclusive. This field may be limited in precision depending on context. |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.TypeMeta"/>

### TypeMeta
TypeMeta describes an individual object in an API response or request
with strings representing the type of the object and its API schema version.
Structures that are versioned or persisted should inline TypeMeta.

&#43;k8s:deepcopy-gen=false


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kind | [string](#string) | optional | Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds &#43;optional |
| apiVersion | [string](#string) | optional | APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources &#43;optional |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.Verbs"/>

### Verbs
Verbs masks the value so protobuf can generate

&#43;protobuf.nullable=true
&#43;protobuf.options.(gogoproto.goproto_stringer)=false


items, if empty, will result in an empty slice


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| items | [string](#string) | repeated |  |






<a name="k8s.io.apimachinery.pkg.apis.meta.v1.WatchEvent"/>

### WatchEvent
Event represents a single event to a watched resource.

&#43;protobuf=true
&#43;k8s:deepcopy-gen=true
&#43;k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [string](#string) | optional |  |
| object | [.k8s.io.apimachinery.pkg.runtime.RawExtension](#k8s.io.apimachinery.pkg.apis.meta.v1..k8s.io.apimachinery.pkg.runtime.RawExtension) | optional | Object is: If Type is Added or Modified: the new state of the object. If Type is Deleted: the state of the object immediately before deletion. If Type is Error: *Status is recommended; other types may make sense depending on context. |





 

 

 

 



<a name="generated.proto"/>
<p align="right"><a href="#top">Top</a></p>

## generated.proto



<a name="k8s.io.apimachinery.pkg.api.resource.Quantity"/>

### Quantity
Quantity is a fixed-point representation of a number.
It provides convenient marshaling/unmarshaling in JSON and YAML,
in addition to String() and Int64() accessors.

The serialization format is:

&lt;quantity&gt;        ::= &lt;signedNumber&gt;&lt;suffix&gt;
(Note that &lt;suffix&gt; may be empty, from the &#34;&#34; case in &lt;decimalSI&gt;.)
&lt;digit&gt;           ::= 0 | 1 | ... | 9
&lt;digits&gt;          ::= &lt;digit&gt; | &lt;digit&gt;&lt;digits&gt;
&lt;number&gt;          ::= &lt;digits&gt; | &lt;digits&gt;.&lt;digits&gt; | &lt;digits&gt;. | .&lt;digits&gt;
&lt;sign&gt;            ::= &#34;&#43;&#34; | &#34;-&#34;
&lt;signedNumber&gt;    ::= &lt;number&gt; | &lt;sign&gt;&lt;number&gt;
&lt;suffix&gt;          ::= &lt;binarySI&gt; | &lt;decimalExponent&gt; | &lt;decimalSI&gt;
&lt;binarySI&gt;        ::= Ki | Mi | Gi | Ti | Pi | Ei
(International System of units; See: http://physics.nist.gov/cuu/Units/binary.html)
&lt;decimalSI&gt;       ::= m | &#34;&#34; | k | M | G | T | P | E
(Note that 1024 = 1Ki but 1000 = 1k; I didn&#39;t choose the capitalization.)
&lt;decimalExponent&gt; ::= &#34;e&#34; &lt;signedNumber&gt; | &#34;E&#34; &lt;signedNumber&gt;

No matter which of the three exponent forms is used, no quantity may represent
a number greater than 2^63-1 in magnitude, nor may it have more than 3 decimal
places. Numbers larger or more precise will be capped or rounded up.
(E.g.: 0.1m will rounded up to 1m.)
This may be extended in the future if we require larger or smaller quantities.

When a Quantity is parsed from a string, it will remember the type of suffix
it had, and will use the same type again when it is serialized.

Before serializing, Quantity will be put in &#34;canonical form&#34;.
This means that Exponent/suffix will be adjusted up or down (with a
corresponding increase or decrease in Mantissa) such that:
a. No precision is lost
b. No fractional digits will be emitted
c. The exponent (or suffix) is as large as possible.
The sign will be omitted unless the number is negative.

Examples:
1.5 will be serialized as &#34;1500m&#34;
1.5Gi will be serialized as &#34;1536Mi&#34;

NOTE: We reserve the right to amend this canonical format, perhaps to
allow 1.5 to be canonical.
TODO: Remove above disclaimer after all bikeshedding about format is over,
or after March 2015.

Note that the quantity will NEVER be internally represented by a
floating point number. That is the whole point of this exercise.

Non-canonical values will still parse as long as they are well formed,
but will be re-emitted in their canonical form. (So always use canonical
form, or don&#39;t diff.)

This format is intended to make it difficult to use these numbers without
writing some sort of special handling code in the hopes that that will
cause implementors to also use a fixed point implementation.

&#43;protobuf=true
&#43;protobuf.embed=string
&#43;protobuf.options.marshal=false
&#43;protobuf.options.(gogoproto.goproto_stringer)=false
&#43;k8s:deepcopy-gen=true
&#43;k8s:openapi-gen=true


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| string | [string](#string) | optional |  |





 

 

 

 



<a name="v1.proto"/>
<p align="right"><a href="#top">Top</a></p>

## v1.proto



<a name="k8s.io.api.core.v1.AWSElasticBlockStoreVolumeSource"/>

### AWSElasticBlockStoreVolumeSource
Represents a Persistent Disk resource in AWS.

An AWS EBS disk must exist before mounting to a container. The disk
must also be in the same AWS zone as the kubelet. An AWS EBS disk
can only be mounted as read/write once. AWS EBS volumes support
ownership management and SELinux relabeling.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| volumeID | [string](#string) | optional | Unique ID of the persistent disk resource in AWS (Amazon EBS volume). More info: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore |
| fsType | [string](#string) | optional | Filesystem type of the volume that you want to mount. Tip: Ensure that the filesystem type is supported by the host operating system. Examples: &#34;ext4&#34;, &#34;xfs&#34;, &#34;ntfs&#34;. Implicitly inferred to be &#34;ext4&#34; if unspecified. More info: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore TODO: how do we prevent errors in the filesystem from compromising the machine &#43;optional |
| partition | [int32](#int32) | optional | The partition in the volume that you want to mount. If omitted, the default is to mount by volume name. Examples: For volume /dev/sda1, you specify the partition as &#34;1&#34;. Similarly, the volume partition for /dev/sda is &#34;0&#34; (or you can leave the property empty). &#43;optional |
| readOnly | [bool](#bool) | optional | Specify &#34;true&#34; to force and set the ReadOnly property in VolumeMounts to &#34;true&#34;. If omitted, the default is &#34;false&#34;. More info: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore &#43;optional |






<a name="k8s.io.api.core.v1.Affinity"/>

### Affinity
Affinity is a group of affinity scheduling rules.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| nodeAffinity | [NodeAffinity](#k8s.io.api.core.v1.NodeAffinity) | optional | Describes node affinity scheduling rules for the pod. &#43;optional |
| podAffinity | [PodAffinity](#k8s.io.api.core.v1.PodAffinity) | optional | Describes pod affinity scheduling rules (e.g. co-locate this pod in the same node, zone, etc. as some other pod(s)). &#43;optional |
| podAntiAffinity | [PodAntiAffinity](#k8s.io.api.core.v1.PodAntiAffinity) | optional | Describes pod anti-affinity scheduling rules (e.g. avoid putting this pod in the same node, zone, etc. as some other pod(s)). &#43;optional |






<a name="k8s.io.api.core.v1.AttachedVolume"/>

### AttachedVolume
AttachedVolume describes a volume attached to a node


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | optional | Name of the attached volume |
| devicePath | [string](#string) | optional | DevicePath represents the device path where the volume should be available |






<a name="k8s.io.api.core.v1.AvoidPods"/>

### AvoidPods
AvoidPods describes pods that should avoid this node. This is the value for a
Node annotation with key scheduler.alpha.kubernetes.io/preferAvoidPods and
will eventually become a field of NodeStatus.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| preferAvoidPods | [PreferAvoidPodsEntry](#k8s.io.api.core.v1.PreferAvoidPodsEntry) | repeated | Bounded-sized list of signatures of pods that should avoid this node, sorted in timestamp order from oldest to newest. Size of the slice is unspecified. &#43;optional |






<a name="k8s.io.api.core.v1.AzureDiskVolumeSource"/>

### AzureDiskVolumeSource
AzureDisk represents an Azure Data Disk mount on the host and bind mount to the pod.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| diskName | [string](#string) | optional | The Name of the data disk in the blob storage |
| diskURI | [string](#string) | optional | The URI the data disk in the blob storage |
| cachingMode | [string](#string) | optional | Host Caching mode: None, Read Only, Read Write. &#43;optional |
| fsType | [string](#string) | optional | Filesystem type to mount. Must be a filesystem type supported by the host operating system. Ex. &#34;ext4&#34;, &#34;xfs&#34;, &#34;ntfs&#34;. Implicitly inferred to be &#34;ext4&#34; if unspecified. &#43;optional |
| readOnly | [bool](#bool) | optional | Defaults to false (read/write). ReadOnly here will force the ReadOnly setting in VolumeMounts. &#43;optional |
| kind | [string](#string) | optional | Expected values Shared: multiple blob disks per storage account Dedicated: single blob disk per storage account Managed: azure managed data disk (only in managed availability set). defaults to shared |






<a name="k8s.io.api.core.v1.AzureFilePersistentVolumeSource"/>

### AzureFilePersistentVolumeSource
AzureFile represents an Azure File Service mount on the host and bind mount to the pod.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| secretName | [string](#string) | optional | the name of secret that contains Azure Storage Account Name and Key |
| shareName | [string](#string) | optional | Share Name |
| readOnly | [bool](#bool) | optional | Defaults to false (read/write). ReadOnly here will force the ReadOnly setting in VolumeMounts. &#43;optional |
| secretNamespace | [string](#string) | optional | the namespace of the secret that contains Azure Storage Account Name and Key default is the same as the Pod &#43;optional |






<a name="k8s.io.api.core.v1.AzureFileVolumeSource"/>

### AzureFileVolumeSource
AzureFile represents an Azure File Service mount on the host and bind mount to the pod.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| secretName | [string](#string) | optional | the name of secret that contains Azure Storage Account Name and Key |
| shareName | [string](#string) | optional | Share Name |
| readOnly | [bool](#bool) | optional | Defaults to false (read/write). ReadOnly here will force the ReadOnly setting in VolumeMounts. &#43;optional |






<a name="k8s.io.api.core.v1.Binding"/>

### Binding
Binding ties one object to another; for example, a pod is bound to a node by a scheduler.
Deprecated in 1.7, please use the bindings subresource of pods instead.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta) | optional | Standard object&#39;s metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata &#43;optional |
| target | [ObjectReference](#k8s.io.api.core.v1.ObjectReference) | optional | The target object that you want to bind to the standard object. |






<a name="k8s.io.api.core.v1.Capabilities"/>

### Capabilities
Adds and removes POSIX capabilities from running containers.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| add | [string](#string) | repeated | Added capabilities &#43;optional |
| drop | [string](#string) | repeated | Removed capabilities &#43;optional |






<a name="k8s.io.api.core.v1.CephFSPersistentVolumeSource"/>

### CephFSPersistentVolumeSource
Represents a Ceph Filesystem mount that lasts the lifetime of a pod
Cephfs volumes do not support ownership management or SELinux relabeling.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| monitors | [string](#string) | repeated | Required: Monitors is a collection of Ceph monitors More info: https://releases.k8s.io/HEAD/examples/volumes/cephfs/README.md#how-to-use-it |
| path | [string](#string) | optional | Optional: Used as the mounted root, rather than the full Ceph tree, default is &#43;optional |
| user | [string](#string) | optional | Optional: User is the rados user name, default is admin More info: https://releases.k8s.io/HEAD/examples/volumes/cephfs/README.md#how-to-use-it &#43;optional |
| secretFile | [string](#string) | optional | Optional: SecretFile is the path to key ring for User, default is /etc/ceph/user.secret More info: https://releases.k8s.io/HEAD/examples/volumes/cephfs/README.md#how-to-use-it &#43;optional |
| secretRef | [SecretReference](#k8s.io.api.core.v1.SecretReference) | optional | Optional: SecretRef is reference to the authentication secret for User, default is empty. More info: https://releases.k8s.io/HEAD/examples/volumes/cephfs/README.md#how-to-use-it &#43;optional |
| readOnly | [bool](#bool) | optional | Optional: Defaults to false (read/write). ReadOnly here will force the ReadOnly setting in VolumeMounts. More info: https://releases.k8s.io/HEAD/examples/volumes/cephfs/README.md#how-to-use-it &#43;optional |






<a name="k8s.io.api.core.v1.CephFSVolumeSource"/>

### CephFSVolumeSource
Represents a Ceph Filesystem mount that lasts the lifetime of a pod
Cephfs volumes do not support ownership management or SELinux relabeling.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| monitors | [string](#string) | repeated | Required: Monitors is a collection of Ceph monitors More info: https://releases.k8s.io/HEAD/examples/volumes/cephfs/README.md#how-to-use-it |
| path | [string](#string) | optional | Optional: Used as the mounted root, rather than the full Ceph tree, default is &#43;optional |
| user | [string](#string) | optional | Optional: User is the rados user name, default is admin More info: https://releases.k8s.io/HEAD/examples/volumes/cephfs/README.md#how-to-use-it &#43;optional |
| secretFile | [string](#string) | optional | Optional: SecretFile is the path to key ring for User, default is /etc/ceph/user.secret More info: https://releases.k8s.io/HEAD/examples/volumes/cephfs/README.md#how-to-use-it &#43;optional |
| secretRef | [LocalObjectReference](#k8s.io.api.core.v1.LocalObjectReference) | optional | Optional: SecretRef is reference to the authentication secret for User, default is empty. More info: https://releases.k8s.io/HEAD/examples/volumes/cephfs/README.md#how-to-use-it &#43;optional |
| readOnly | [bool](#bool) | optional | Optional: Defaults to false (read/write). ReadOnly here will force the ReadOnly setting in VolumeMounts. More info: https://releases.k8s.io/HEAD/examples/volumes/cephfs/README.md#how-to-use-it &#43;optional |






<a name="k8s.io.api.core.v1.CinderVolumeSource"/>

### CinderVolumeSource
Represents a cinder volume resource in Openstack.
A Cinder volume must exist before mounting to a container.
The volume must also be in the same region as the kubelet.
Cinder volumes support ownership management and SELinux relabeling.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| volumeID | [string](#string) | optional | volume id used to identify the volume in cinder More info: https://releases.k8s.io/HEAD/examples/mysql-cinder-pd/README.md |
| fsType | [string](#string) | optional | Filesystem type to mount. Must be a filesystem type supported by the host operating system. Examples: &#34;ext4&#34;, &#34;xfs&#34;, &#34;ntfs&#34;. Implicitly inferred to be &#34;ext4&#34; if unspecified. More info: https://releases.k8s.io/HEAD/examples/mysql-cinder-pd/README.md &#43;optional |
| readOnly | [bool](#bool) | optional | Optional: Defaults to false (read/write). ReadOnly here will force the ReadOnly setting in VolumeMounts. More info: https://releases.k8s.io/HEAD/examples/mysql-cinder-pd/README.md &#43;optional |






<a name="k8s.io.api.core.v1.ClientIPConfig"/>

### ClientIPConfig
ClientIPConfig represents the configurations of Client IP based session affinity.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| timeoutSeconds | [int32](#int32) | optional | timeoutSeconds specifies the seconds of ClientIP type session sticky time. The value must be &gt;0 &amp;&amp; &lt;=86400(for 1 day) if ServiceAffinity == &#34;ClientIP&#34;. Default value is 10800(for 3 hours). &#43;optional |






<a name="k8s.io.api.core.v1.ComponentCondition"/>

### ComponentCondition
Information about the condition of a component.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [string](#string) | optional | Type of condition for a component. Valid value: &#34;Healthy&#34; |
| status | [string](#string) | optional | Status of the condition for a component. Valid values for &#34;Healthy&#34;: &#34;True&#34;, &#34;False&#34;, or &#34;Unknown&#34;. |
| message | [string](#string) | optional | Message about the condition for a component. For example, information about a health check. &#43;optional |
| error | [string](#string) | optional | Condition error code for a component. For example, a health check error code. &#43;optional |






<a name="k8s.io.api.core.v1.ComponentStatus"/>

### ComponentStatus
ComponentStatus (and ComponentStatusList) holds the cluster validation info.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta) | optional | Standard object&#39;s metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata &#43;optional |
| conditions | [ComponentCondition](#k8s.io.api.core.v1.ComponentCondition) | repeated | List of component conditions observed &#43;optional &#43;patchMergeKey=type &#43;patchStrategy=merge |






<a name="k8s.io.api.core.v1.ComponentStatusList"/>

### ComponentStatusList
Status of all the conditions for the component as a list of ComponentStatus objects.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta) | optional | Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds &#43;optional |
| items | [ComponentStatus](#k8s.io.api.core.v1.ComponentStatus) | repeated | List of ComponentStatus objects. |






<a name="k8s.io.api.core.v1.ConfigMap"/>

### ConfigMap
ConfigMap holds configuration data for pods to consume.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta) | optional | Standard object&#39;s metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata &#43;optional |
| data | [ConfigMap.DataEntry](#k8s.io.api.core.v1.ConfigMap.DataEntry) | repeated | Data contains the configuration data. Each key must consist of alphanumeric characters, &#39;-&#39;, &#39;_&#39; or &#39;.&#39;. &#43;optional |






<a name="k8s.io.api.core.v1.ConfigMap.DataEntry"/>

### ConfigMap.DataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [string](#string) | optional |  |






<a name="k8s.io.api.core.v1.ConfigMapEnvSource"/>

### ConfigMapEnvSource
ConfigMapEnvSource selects a ConfigMap to populate the environment
variables with.

The contents of the target ConfigMap&#39;s Data field will represent the
key-value pairs as environment variables.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| localObjectReference | [LocalObjectReference](#k8s.io.api.core.v1.LocalObjectReference) | optional | The ConfigMap to select from. |
| optional | [bool](#bool) | optional | Specify whether the ConfigMap must be defined &#43;optional |






<a name="k8s.io.api.core.v1.ConfigMapKeySelector"/>

### ConfigMapKeySelector
Selects a key from a ConfigMap.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| localObjectReference | [LocalObjectReference](#k8s.io.api.core.v1.LocalObjectReference) | optional | The ConfigMap to select from. |
| key | [string](#string) | optional | The key to select. |
| optional | [bool](#bool) | optional | Specify whether the ConfigMap or it&#39;s key must be defined &#43;optional |






<a name="k8s.io.api.core.v1.ConfigMapList"/>

### ConfigMapList
ConfigMapList is a resource containing a list of ConfigMap objects.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta) | optional | More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata &#43;optional |
| items | [ConfigMap](#k8s.io.api.core.v1.ConfigMap) | repeated | Items is the list of ConfigMaps. |






<a name="k8s.io.api.core.v1.ConfigMapProjection"/>

### ConfigMapProjection
Adapts a ConfigMap into a projected volume.

The contents of the target ConfigMap&#39;s Data field will be presented in a
projected volume as files using the keys in the Data field as the file names,
unless the items element is populated with specific mappings of keys to paths.
Note that this is identical to a configmap volume source without the default
mode.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| localObjectReference | [LocalObjectReference](#k8s.io.api.core.v1.LocalObjectReference) | optional |  |
| items | [KeyToPath](#k8s.io.api.core.v1.KeyToPath) | repeated | If unspecified, each key-value pair in the Data field of the referenced ConfigMap will be projected into the volume as a file whose name is the key and content is the value. If specified, the listed keys will be projected into the specified paths, and unlisted keys will not be present. If a key is specified which is not present in the ConfigMap, the volume setup will error unless it is marked optional. Paths must be relative and may not contain the &#39;..&#39; path or start with &#39;..&#39;. &#43;optional |
| optional | [bool](#bool) | optional | Specify whether the ConfigMap or it&#39;s keys must be defined &#43;optional |






<a name="k8s.io.api.core.v1.ConfigMapVolumeSource"/>

### ConfigMapVolumeSource
Adapts a ConfigMap into a volume.

The contents of the target ConfigMap&#39;s Data field will be presented in a
volume as files using the keys in the Data field as the file names, unless
the items element is populated with specific mappings of keys to paths.
ConfigMap volumes support ownership management and SELinux relabeling.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| localObjectReference | [LocalObjectReference](#k8s.io.api.core.v1.LocalObjectReference) | optional |  |
| items | [KeyToPath](#k8s.io.api.core.v1.KeyToPath) | repeated | If unspecified, each key-value pair in the Data field of the referenced ConfigMap will be projected into the volume as a file whose name is the key and content is the value. If specified, the listed keys will be projected into the specified paths, and unlisted keys will not be present. If a key is specified which is not present in the ConfigMap, the volume setup will error unless it is marked optional. Paths must be relative and may not contain the &#39;..&#39; path or start with &#39;..&#39;. &#43;optional |
| defaultMode | [int32](#int32) | optional | Optional: mode bits to use on created files by default. Must be a value between 0 and 0777. Defaults to 0644. Directories within the path are not affected by this setting. This might be in conflict with other options that affect the file mode, like fsGroup, and the result can be other mode bits set. &#43;optional |
| optional | [bool](#bool) | optional | Specify whether the ConfigMap or it&#39;s keys must be defined &#43;optional |






<a name="k8s.io.api.core.v1.Container"/>

### Container
A single application container that you want to run within a pod.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | optional | Name of the container specified as a DNS_LABEL. Each container in a pod must have a unique name (DNS_LABEL). Cannot be updated. |
| image | [string](#string) | optional | Docker image name. More info: https://kubernetes.io/docs/concepts/containers/images This field is optional to allow higher level config management to default or override container images in workload controllers like Deployments and StatefulSets. &#43;optional |
| command | [string](#string) | repeated | Entrypoint array. Not executed within a shell. The docker image&#39;s ENTRYPOINT is used if this is not provided. Variable references $(VAR_NAME) are expanded using the container&#39;s environment. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Cannot be updated. More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell &#43;optional |
| args | [string](#string) | repeated | Arguments to the entrypoint. The docker image&#39;s CMD is used if this is not provided. Variable references $(VAR_NAME) are expanded using the container&#39;s environment. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Cannot be updated. More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell &#43;optional |
| workingDir | [string](#string) | optional | Container&#39;s working directory. If not specified, the container runtime&#39;s default will be used, which might be configured in the container image. Cannot be updated. &#43;optional |
| ports | [ContainerPort](#k8s.io.api.core.v1.ContainerPort) | repeated | List of ports to expose from the container. Exposing a port here gives the system additional information about the network connections a container uses, but is primarily informational. Not specifying a port here DOES NOT prevent that port from being exposed. Any port which is listening on the default &#34;0.0.0.0&#34; address inside a container will be accessible from the network. Cannot be updated. &#43;optional &#43;patchMergeKey=containerPort &#43;patchStrategy=merge |
| envFrom | [EnvFromSource](#k8s.io.api.core.v1.EnvFromSource) | repeated | List of sources to populate environment variables in the container. The keys defined within a source must be a C_IDENTIFIER. All invalid keys will be reported as an event when the container is starting. When a key exists in multiple sources, the value associated with the last source will take precedence. Values defined by an Env with a duplicate key will take precedence. Cannot be updated. &#43;optional |
| env | [EnvVar](#k8s.io.api.core.v1.EnvVar) | repeated | List of environment variables to set in the container. Cannot be updated. &#43;optional &#43;patchMergeKey=name &#43;patchStrategy=merge |
| resources | [ResourceRequirements](#k8s.io.api.core.v1.ResourceRequirements) | optional | Compute Resources required by this container. Cannot be updated. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources &#43;optional |
| volumeMounts | [VolumeMount](#k8s.io.api.core.v1.VolumeMount) | repeated | Pod volumes to mount into the container&#39;s filesystem. Cannot be updated. &#43;optional &#43;patchMergeKey=mountPath &#43;patchStrategy=merge |
| livenessProbe | [Probe](#k8s.io.api.core.v1.Probe) | optional | Periodic probe of container liveness. Container will be restarted if the probe fails. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes &#43;optional |
| readinessProbe | [Probe](#k8s.io.api.core.v1.Probe) | optional | Periodic probe of container service readiness. Container will be removed from service endpoints if the probe fails. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes &#43;optional |
| lifecycle | [Lifecycle](#k8s.io.api.core.v1.Lifecycle) | optional | Actions that the management system should take in response to container lifecycle events. Cannot be updated. &#43;optional |
| terminationMessagePath | [string](#string) | optional | Optional: Path at which the file to which the container&#39;s termination message will be written is mounted into the container&#39;s filesystem. Message written is intended to be brief final status, such as an assertion failure message. Will be truncated by the node if greater than 4096 bytes. The total message length across all containers will be limited to 12kb. Defaults to /dev/termination-log. Cannot be updated. &#43;optional |
| terminationMessagePolicy | [string](#string) | optional | Indicate how the termination message should be populated. File will use the contents of terminationMessagePath to populate the container status message on both success and failure. FallbackToLogsOnError will use the last chunk of container log output if the termination message file is empty and the container exited with an error. The log output is limited to 2048 bytes or 80 lines, whichever is smaller. Defaults to File. Cannot be updated. &#43;optional |
| imagePullPolicy | [string](#string) | optional | Image pull policy. One of Always, Never, IfNotPresent. Defaults to Always if :latest tag is specified, or IfNotPresent otherwise. Cannot be updated. More info: https://kubernetes.io/docs/concepts/containers/images#updating-images &#43;optional |
| securityContext | [SecurityContext](#k8s.io.api.core.v1.SecurityContext) | optional | Security options the pod should run with. More info: https://kubernetes.io/docs/concepts/policy/security-context More info: https://kubernetes.io/docs/tasks/configure-pod-container/security-context &#43;optional |
| stdin | [bool](#bool) | optional | Whether this container should allocate a buffer for stdin in the container runtime. If this is not set, reads from stdin in the container will always result in EOF. Default is false. &#43;optional |
| stdinOnce | [bool](#bool) | optional | Whether the container runtime should close the stdin channel after it has been opened by a single attach. When stdin is true the stdin stream will remain open across multiple attach sessions. If stdinOnce is set to true, stdin is opened on container start, is empty until the first client attaches to stdin, and then remains open and accepts data until the client disconnects, at which time stdin is closed and remains closed until the container is restarted. If this flag is false, a container processes that reads from stdin will never receive an EOF. Default is false &#43;optional |
| tty | [bool](#bool) | optional | Whether this container should allocate a TTY for itself, also requires &#39;stdin&#39; to be true. Default is false. &#43;optional |






<a name="k8s.io.api.core.v1.ContainerImage"/>

### ContainerImage
Describe a container image


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| names | [string](#string) | repeated | Names by which this image is known. e.g. [&#34;gcr.io/google_containers/hyperkube:v1.0.7&#34;, &#34;dockerhub.io/google_containers/hyperkube:v1.0.7&#34;] |
| sizeBytes | [int64](#int64) | optional | The size of the image in bytes. &#43;optional |






<a name="k8s.io.api.core.v1.ContainerPort"/>

### ContainerPort
ContainerPort represents a network port in a single container.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | optional | If specified, this must be an IANA_SVC_NAME and unique within the pod. Each named port in a pod must have a unique name. Name for the port that can be referred to by services. &#43;optional |
| hostPort | [int32](#int32) | optional | Number of port to expose on the host. If specified, this must be a valid port number, 0 &lt; x &lt; 65536. If HostNetwork is specified, this must match ContainerPort. Most containers do not need this. &#43;optional |
| containerPort | [int32](#int32) | optional | Number of port to expose on the pod&#39;s IP address. This must be a valid port number, 0 &lt; x &lt; 65536. |
| protocol | [string](#string) | optional | Protocol for port. Must be UDP or TCP. Defaults to &#34;TCP&#34;. &#43;optional |
| hostIP | [string](#string) | optional | What host IP to bind the external port to. &#43;optional |






<a name="k8s.io.api.core.v1.ContainerState"/>

### ContainerState
ContainerState holds a possible state of container.
Only one of its members may be specified.
If none of them is specified, the default one is ContainerStateWaiting.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| waiting | [ContainerStateWaiting](#k8s.io.api.core.v1.ContainerStateWaiting) | optional | Details about a waiting container &#43;optional |
| running | [ContainerStateRunning](#k8s.io.api.core.v1.ContainerStateRunning) | optional | Details about a running container &#43;optional |
| terminated | [ContainerStateTerminated](#k8s.io.api.core.v1.ContainerStateTerminated) | optional | Details about a terminated container &#43;optional |






<a name="k8s.io.api.core.v1.ContainerStateRunning"/>

### ContainerStateRunning
ContainerStateRunning is a running state of a container.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| startedAt | [.k8s.io.apimachinery.pkg.apis.meta.v1.Time](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.Time) | optional | Time at which the container was last (re-)started &#43;optional |






<a name="k8s.io.api.core.v1.ContainerStateTerminated"/>

### ContainerStateTerminated
ContainerStateTerminated is a terminated state of a container.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| exitCode | [int32](#int32) | optional | Exit status from the last termination of the container |
| signal | [int32](#int32) | optional | Signal from the last termination of the container &#43;optional |
| reason | [string](#string) | optional | (brief) reason from the last termination of the container &#43;optional |
| message | [string](#string) | optional | Message regarding the last termination of the container &#43;optional |
| startedAt | [.k8s.io.apimachinery.pkg.apis.meta.v1.Time](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.Time) | optional | Time at which previous execution of the container started &#43;optional |
| finishedAt | [.k8s.io.apimachinery.pkg.apis.meta.v1.Time](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.Time) | optional | Time at which the container last terminated &#43;optional |
| containerID | [string](#string) | optional | Container&#39;s ID in the format &#39;docker://&lt;container_id&gt;&#39; &#43;optional |






<a name="k8s.io.api.core.v1.ContainerStateWaiting"/>

### ContainerStateWaiting
ContainerStateWaiting is a waiting state of a container.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| reason | [string](#string) | optional | (brief) reason the container is not yet running. &#43;optional |
| message | [string](#string) | optional | Message regarding why the container is not yet running. &#43;optional |






<a name="k8s.io.api.core.v1.ContainerStatus"/>

### ContainerStatus
ContainerStatus contains details for the current status of this container.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | optional | This must be a DNS_LABEL. Each container in a pod must have a unique name. Cannot be updated. |
| state | [ContainerState](#k8s.io.api.core.v1.ContainerState) | optional | Details about the container&#39;s current condition. &#43;optional |
| lastState | [ContainerState](#k8s.io.api.core.v1.ContainerState) | optional | Details about the container&#39;s last termination condition. &#43;optional |
| ready | [bool](#bool) | optional | Specifies whether the container has passed its readiness probe. |
| restartCount | [int32](#int32) | optional | The number of times the container has been restarted, currently based on the number of dead containers that have not yet been removed. Note that this is calculated from dead containers. But those containers are subject to garbage collection. This value will get capped at 5 by GC. |
| image | [string](#string) | optional | The image the container is running. More info: https://kubernetes.io/docs/concepts/containers/images TODO(dchen1107): Which image the container is running with? |
| imageID | [string](#string) | optional | ImageID of the container&#39;s image. |
| containerID | [string](#string) | optional | Container&#39;s ID in the format &#39;docker://&lt;container_id&gt;&#39;. &#43;optional |






<a name="k8s.io.api.core.v1.DaemonEndpoint"/>

### DaemonEndpoint
DaemonEndpoint contains information about a single Daemon endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Port | [int32](#int32) | optional | Port number of the given endpoint. |






<a name="k8s.io.api.core.v1.DeleteOptions"/>

### DeleteOptions
DeleteOptions may be provided when deleting an API object
DEPRECATED: This type has been moved to meta/v1 and will be removed soon.
&#43;k8s:openapi-gen=false


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| gracePeriodSeconds | [int64](#int64) | optional | The duration in seconds before the object should be deleted. Value must be non-negative integer. The value zero indicates delete immediately. If this value is nil, the default grace period for the specified type will be used. Defaults to a per object value if not specified. zero means delete immediately. &#43;optional |
| preconditions | [Preconditions](#k8s.io.api.core.v1.Preconditions) | optional | Must be fulfilled before a deletion is carried out. If not possible, a 409 Conflict status will be returned. &#43;optional |
| orphanDependents | [bool](#bool) | optional | Deprecated: please use the PropagationPolicy, this field will be deprecated in 1.7. Should the dependent objects be orphaned. If true/false, the &#34;orphan&#34; finalizer will be added to/removed from the object&#39;s finalizers list. Either this field or PropagationPolicy may be set, but not both. &#43;optional |
| propagationPolicy | [string](#string) | optional | Whether and how garbage collection will be performed. Either this field or OrphanDependents may be set, but not both. The default policy is decided by the existing finalizer set in the metadata.finalizers and the resource-specific default policy. &#43;optional |






<a name="k8s.io.api.core.v1.DownwardAPIProjection"/>

### DownwardAPIProjection
Represents downward API info for projecting into a projected volume.
Note that this is identical to a downwardAPI volume source without the default
mode.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| items | [DownwardAPIVolumeFile](#k8s.io.api.core.v1.DownwardAPIVolumeFile) | repeated | Items is a list of DownwardAPIVolume file &#43;optional |






<a name="k8s.io.api.core.v1.DownwardAPIVolumeFile"/>

### DownwardAPIVolumeFile
DownwardAPIVolumeFile represents information to create the file containing the pod field


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [string](#string) | optional | Required: Path is the relative path name of the file to be created. Must not be absolute or contain the &#39;..&#39; path. Must be utf-8 encoded. The first item of the relative path must not start with &#39;..&#39; |
| fieldRef | [ObjectFieldSelector](#k8s.io.api.core.v1.ObjectFieldSelector) | optional | Required: Selects a field of the pod: only annotations, labels, name and namespace are supported. &#43;optional |
| resourceFieldRef | [ResourceFieldSelector](#k8s.io.api.core.v1.ResourceFieldSelector) | optional | Selects a resource of the container: only resources limits and requests (limits.cpu, limits.memory, requests.cpu and requests.memory) are currently supported. &#43;optional |
| mode | [int32](#int32) | optional | Optional: mode bits to use on this file, must be a value between 0 and 0777. If not specified, the volume defaultMode will be used. This might be in conflict with other options that affect the file mode, like fsGroup, and the result can be other mode bits set. &#43;optional |






<a name="k8s.io.api.core.v1.DownwardAPIVolumeSource"/>

### DownwardAPIVolumeSource
DownwardAPIVolumeSource represents a volume containing downward API info.
Downward API volumes support ownership management and SELinux relabeling.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| items | [DownwardAPIVolumeFile](#k8s.io.api.core.v1.DownwardAPIVolumeFile) | repeated | Items is a list of downward API volume file &#43;optional |
| defaultMode | [int32](#int32) | optional | Optional: mode bits to use on created files by default. Must be a value between 0 and 0777. Defaults to 0644. Directories within the path are not affected by this setting. This might be in conflict with other options that affect the file mode, like fsGroup, and the result can be other mode bits set. &#43;optional |






<a name="k8s.io.api.core.v1.EmptyDirVolumeSource"/>

### EmptyDirVolumeSource
Represents an empty directory for a pod.
Empty directory volumes support ownership management and SELinux relabeling.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| medium | [string](#string) | optional | What type of storage medium should back this directory. The default is &#34;&#34; which means to use the node&#39;s default medium. Must be an empty string (default) or Memory. More info: https://kubernetes.io/docs/concepts/storage/volumes#emptydir &#43;optional |
| sizeLimit | [.k8s.io.apimachinery.pkg.api.resource.Quantity](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.api.resource.Quantity) | optional | Total amount of local storage required for this EmptyDir volume. The size limit is also applicable for memory medium. The maximum usage on memory medium EmptyDir would be the minimum value between the SizeLimit specified here and the sum of memory limits of all containers in a pod. The default is nil which means that the limit is undefined. More info: http://kubernetes.io/docs/user-guide/volumes#emptydir &#43;optional |






<a name="k8s.io.api.core.v1.EndpointAddress"/>

### EndpointAddress
EndpointAddress is a tuple that describes single IP address.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ip | [string](#string) | optional | The IP of this endpoint. May not be loopback (127.0.0.0/8), link-local (169.254.0.0/16), or link-local multicast ((224.0.0.0/24). IPv6 is also accepted but not fully supported on all platforms. Also, certain kubernetes components, like kube-proxy, are not IPv6 ready. TODO: This should allow hostname or IP, See #4447. |
| hostname | [string](#string) | optional | The Hostname of this endpoint &#43;optional |
| nodeName | [string](#string) | optional | Optional: Node hosting this endpoint. This can be used to determine endpoints local to a node. &#43;optional |
| targetRef | [ObjectReference](#k8s.io.api.core.v1.ObjectReference) | optional | Reference to object providing the endpoint. &#43;optional |






<a name="k8s.io.api.core.v1.EndpointPort"/>

### EndpointPort
EndpointPort is a tuple that describes a single port.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | optional | The name of this port (corresponds to ServicePort.Name). Must be a DNS_LABEL. Optional only if one port is defined. &#43;optional |
| port | [int32](#int32) | optional | The port number of the endpoint. |
| protocol | [string](#string) | optional | The IP protocol for this port. Must be UDP or TCP. Default is TCP. &#43;optional |






<a name="k8s.io.api.core.v1.EndpointSubset"/>

### EndpointSubset
EndpointSubset is a group of addresses with a common set of ports. The
expanded set of endpoints is the Cartesian product of Addresses x Ports.
For example, given:
{
Addresses: [{&#34;ip&#34;: &#34;10.10.1.1&#34;}, {&#34;ip&#34;: &#34;10.10.2.2&#34;}],
Ports:     [{&#34;name&#34;: &#34;a&#34;, &#34;port&#34;: 8675}, {&#34;name&#34;: &#34;b&#34;, &#34;port&#34;: 309}]
}
The resulting set of endpoints can be viewed as:
a: [ 10.10.1.1:8675, 10.10.2.2:8675 ],
b: [ 10.10.1.1:309, 10.10.2.2:309 ]


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| addresses | [EndpointAddress](#k8s.io.api.core.v1.EndpointAddress) | repeated | IP addresses which offer the related ports that are marked as ready. These endpoints should be considered safe for load balancers and clients to utilize. &#43;optional |
| notReadyAddresses | [EndpointAddress](#k8s.io.api.core.v1.EndpointAddress) | repeated | IP addresses which offer the related ports but are not currently marked as ready because they have not yet finished starting, have recently failed a readiness check, or have recently failed a liveness check. &#43;optional |
| ports | [EndpointPort](#k8s.io.api.core.v1.EndpointPort) | repeated | Port numbers available on the related IP addresses. &#43;optional |






<a name="k8s.io.api.core.v1.Endpoints"/>

### Endpoints
Endpoints is a collection of endpoints that implement the actual service. Example:
Name: &#34;mysvc&#34;,
Subsets: [
{
Addresses: [{&#34;ip&#34;: &#34;10.10.1.1&#34;}, {&#34;ip&#34;: &#34;10.10.2.2&#34;}],
Ports: [{&#34;name&#34;: &#34;a&#34;, &#34;port&#34;: 8675}, {&#34;name&#34;: &#34;b&#34;, &#34;port&#34;: 309}]
},
{
Addresses: [{&#34;ip&#34;: &#34;10.10.3.3&#34;}],
Ports: [{&#34;name&#34;: &#34;a&#34;, &#34;port&#34;: 93}, {&#34;name&#34;: &#34;b&#34;, &#34;port&#34;: 76}]
},
]


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta) | optional | Standard object&#39;s metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata &#43;optional |
| subsets | [EndpointSubset](#k8s.io.api.core.v1.EndpointSubset) | repeated | The set of all endpoints is the union of all subsets. Addresses are placed into subsets according to the IPs they share. A single address with multiple ports, some of which are ready and some of which are not (because they come from different containers) will result in the address being displayed in different subsets for the different ports. No address will appear in both Addresses and NotReadyAddresses in the same subset. Sets of addresses and ports that comprise a service. |






<a name="k8s.io.api.core.v1.EndpointsList"/>

### EndpointsList
EndpointsList is a list of endpoints.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta) | optional | Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds &#43;optional |
| items | [Endpoints](#k8s.io.api.core.v1.Endpoints) | repeated | List of endpoints. |






<a name="k8s.io.api.core.v1.EnvFromSource"/>

### EnvFromSource
EnvFromSource represents the source of a set of ConfigMaps


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| prefix | [string](#string) | optional | An optional identifer to prepend to each key in the ConfigMap. Must be a C_IDENTIFIER. &#43;optional |
| configMapRef | [ConfigMapEnvSource](#k8s.io.api.core.v1.ConfigMapEnvSource) | optional | The ConfigMap to select from &#43;optional |
| secretRef | [SecretEnvSource](#k8s.io.api.core.v1.SecretEnvSource) | optional | The Secret to select from &#43;optional |






<a name="k8s.io.api.core.v1.EnvVar"/>

### EnvVar
EnvVar represents an environment variable present in a Container.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | optional | Name of the environment variable. Must be a C_IDENTIFIER. |
| value | [string](#string) | optional | Variable references $(VAR_NAME) are expanded using the previous defined environment variables in the container and any service environment variables. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Defaults to &#34;&#34;. &#43;optional |
| valueFrom | [EnvVarSource](#k8s.io.api.core.v1.EnvVarSource) | optional | Source for the environment variable&#39;s value. Cannot be used if value is not empty. &#43;optional |






<a name="k8s.io.api.core.v1.EnvVarSource"/>

### EnvVarSource
EnvVarSource represents a source for the value of an EnvVar.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fieldRef | [ObjectFieldSelector](#k8s.io.api.core.v1.ObjectFieldSelector) | optional | Selects a field of the pod: supports metadata.name, metadata.namespace, metadata.labels, metadata.annotations, spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP. &#43;optional |
| resourceFieldRef | [ResourceFieldSelector](#k8s.io.api.core.v1.ResourceFieldSelector) | optional | Selects a resource of the container: only resources limits and requests (limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported. &#43;optional |
| configMapKeyRef | [ConfigMapKeySelector](#k8s.io.api.core.v1.ConfigMapKeySelector) | optional | Selects a key of a ConfigMap. &#43;optional |
| secretKeyRef | [SecretKeySelector](#k8s.io.api.core.v1.SecretKeySelector) | optional | Selects a key of a secret in the pod&#39;s namespace &#43;optional |






<a name="k8s.io.api.core.v1.Event"/>

### Event
Event is a report of an event somewhere in the cluster.
TODO: Decide whether to store these separately or with the object they apply to.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta) | optional | Standard object&#39;s metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata |
| involvedObject | [ObjectReference](#k8s.io.api.core.v1.ObjectReference) | optional | The object that this event is about. |
| reason | [string](#string) | optional | This should be a short, machine understandable string that gives the reason for the transition into the object&#39;s current status. TODO: provide exact specification for format. &#43;optional |
| message | [string](#string) | optional | A human-readable description of the status of this operation. TODO: decide on maximum length. &#43;optional |
| source | [EventSource](#k8s.io.api.core.v1.EventSource) | optional | The component reporting this event. Should be a short machine understandable string. &#43;optional |
| firstTimestamp | [.k8s.io.apimachinery.pkg.apis.meta.v1.Time](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.Time) | optional | The time at which the event was first recorded. (Time of server receipt is in TypeMeta.) &#43;optional |
| lastTimestamp | [.k8s.io.apimachinery.pkg.apis.meta.v1.Time](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.Time) | optional | The time at which the most recent occurrence of this event was recorded. &#43;optional |
| count | [int32](#int32) | optional | The number of times this event has occurred. &#43;optional |
| type | [string](#string) | optional | Type of this event (Normal, Warning), new types could be added in the future &#43;optional |






<a name="k8s.io.api.core.v1.EventList"/>

### EventList
EventList is a list of events.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta) | optional | Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds &#43;optional |
| items | [Event](#k8s.io.api.core.v1.Event) | repeated | List of events |






<a name="k8s.io.api.core.v1.EventSource"/>

### EventSource
EventSource contains information for an event.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| component | [string](#string) | optional | Component from which the event is generated. &#43;optional |
| host | [string](#string) | optional | Node name on which the event is generated. &#43;optional |






<a name="k8s.io.api.core.v1.ExecAction"/>

### ExecAction
ExecAction describes a &#34;run in container&#34; action.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| command | [string](#string) | repeated | Command is the command line to execute inside the container, the working directory for the command is root (&#39;/&#39;) in the container&#39;s filesystem. The command is simply exec&#39;d, it is not run inside a shell, so traditional shell instructions (&#39;|&#39;, etc) won&#39;t work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy. &#43;optional |






<a name="k8s.io.api.core.v1.FCVolumeSource"/>

### FCVolumeSource
Represents a Fibre Channel volume.
Fibre Channel volumes can only be mounted as read/write once.
Fibre Channel volumes support ownership management and SELinux relabeling.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| targetWWNs | [string](#string) | repeated | Optional: FC target worldwide names (WWNs) &#43;optional |
| lun | [int32](#int32) | optional | Optional: FC target lun number &#43;optional |
| fsType | [string](#string) | optional | Filesystem type to mount. Must be a filesystem type supported by the host operating system. Ex. &#34;ext4&#34;, &#34;xfs&#34;, &#34;ntfs&#34;. Implicitly inferred to be &#34;ext4&#34; if unspecified. TODO: how do we prevent errors in the filesystem from compromising the machine &#43;optional |
| readOnly | [bool](#bool) | optional | Optional: Defaults to false (read/write). ReadOnly here will force the ReadOnly setting in VolumeMounts. &#43;optional |
| wwids | [string](#string) | repeated | Optional: FC volume world wide identifiers (wwids) Either wwids or combination of targetWWNs and lun must be set, but not both simultaneously. &#43;optional |






<a name="k8s.io.api.core.v1.FlexVolumeSource"/>

### FlexVolumeSource
FlexVolume represents a generic volume resource that is
provisioned/attached using an exec based plugin. This is an alpha feature and may change in future.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| driver | [string](#string) | optional | Driver is the name of the driver to use for this volume. |
| fsType | [string](#string) | optional | Filesystem type to mount. Must be a filesystem type supported by the host operating system. Ex. &#34;ext4&#34;, &#34;xfs&#34;, &#34;ntfs&#34;. The default filesystem depends on FlexVolume script. &#43;optional |
| secretRef | [LocalObjectReference](#k8s.io.api.core.v1.LocalObjectReference) | optional | Optional: SecretRef is reference to the secret object containing sensitive information to pass to the plugin scripts. This may be empty if no secret object is specified. If the secret object contains more than one secret, all secrets are passed to the plugin scripts. &#43;optional |
| readOnly | [bool](#bool) | optional | Optional: Defaults to false (read/write). ReadOnly here will force the ReadOnly setting in VolumeMounts. &#43;optional |
| options | [FlexVolumeSource.OptionsEntry](#k8s.io.api.core.v1.FlexVolumeSource.OptionsEntry) | repeated | Optional: Extra command options if any. &#43;optional |






<a name="k8s.io.api.core.v1.FlexVolumeSource.OptionsEntry"/>

### FlexVolumeSource.OptionsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [string](#string) | optional |  |






<a name="k8s.io.api.core.v1.FlockerVolumeSource"/>

### FlockerVolumeSource
Represents a Flocker volume mounted by the Flocker agent.
One and only one of datasetName and datasetUUID should be set.
Flocker volumes do not support ownership management or SELinux relabeling.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| datasetName | [string](#string) | optional | Name of the dataset stored as metadata -&gt; name on the dataset for Flocker should be considered as deprecated &#43;optional |
| datasetUUID | [string](#string) | optional | UUID of the dataset. This is unique identifier of a Flocker dataset &#43;optional |






<a name="k8s.io.api.core.v1.GCEPersistentDiskVolumeSource"/>

### GCEPersistentDiskVolumeSource
Represents a Persistent Disk resource in Google Compute Engine.

A GCE PD must exist before mounting to a container. The disk must
also be in the same GCE project and zone as the kubelet. A GCE PD
can only be mounted as read/write once or read-only many times. GCE
PDs support ownership management and SELinux relabeling.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pdName | [string](#string) | optional | Unique name of the PD resource in GCE. Used to identify the disk in GCE. More info: https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk |
| fsType | [string](#string) | optional | Filesystem type of the volume that you want to mount. Tip: Ensure that the filesystem type is supported by the host operating system. Examples: &#34;ext4&#34;, &#34;xfs&#34;, &#34;ntfs&#34;. Implicitly inferred to be &#34;ext4&#34; if unspecified. More info: https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk TODO: how do we prevent errors in the filesystem from compromising the machine &#43;optional |
| partition | [int32](#int32) | optional | The partition in the volume that you want to mount. If omitted, the default is to mount by volume name. Examples: For volume /dev/sda1, you specify the partition as &#34;1&#34;. Similarly, the volume partition for /dev/sda is &#34;0&#34; (or you can leave the property empty). More info: https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk &#43;optional |
| readOnly | [bool](#bool) | optional | ReadOnly here will force the ReadOnly setting in VolumeMounts. Defaults to false. More info: https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk &#43;optional |






<a name="k8s.io.api.core.v1.GitRepoVolumeSource"/>

### GitRepoVolumeSource
Represents a volume that is populated with the contents of a git repository.
Git repo volumes do not support ownership management.
Git repo volumes support SELinux relabeling.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| repository | [string](#string) | optional | Repository URL |
| revision | [string](#string) | optional | Commit hash for the specified revision. &#43;optional |
| directory | [string](#string) | optional | Target directory name. Must not contain or start with &#39;..&#39;. If &#39;.&#39; is supplied, the volume directory will be the git repository. Otherwise, if specified, the volume will contain the git repository in the subdirectory with the given name. &#43;optional |






<a name="k8s.io.api.core.v1.GlusterfsVolumeSource"/>

### GlusterfsVolumeSource
Represents a Glusterfs mount that lasts the lifetime of a pod.
Glusterfs volumes do not support ownership management or SELinux relabeling.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| endpoints | [string](#string) | optional | EndpointsName is the endpoint name that details Glusterfs topology. More info: https://releases.k8s.io/HEAD/examples/volumes/glusterfs/README.md#create-a-pod |
| path | [string](#string) | optional | Path is the Glusterfs volume path. More info: https://releases.k8s.io/HEAD/examples/volumes/glusterfs/README.md#create-a-pod |
| readOnly | [bool](#bool) | optional | ReadOnly here will force the Glusterfs volume to be mounted with read-only permissions. Defaults to false. More info: https://releases.k8s.io/HEAD/examples/volumes/glusterfs/README.md#create-a-pod &#43;optional |






<a name="k8s.io.api.core.v1.HTTPGetAction"/>

### HTTPGetAction
HTTPGetAction describes an action based on HTTP Get requests.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [string](#string) | optional | Path to access on the HTTP server. &#43;optional |
| port | [.k8s.io.apimachinery.pkg.util.intstr.IntOrString](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.util.intstr.IntOrString) | optional | Name or number of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME. |
| host | [string](#string) | optional | Host name to connect to, defaults to the pod IP. You probably want to set &#34;Host&#34; in httpHeaders instead. &#43;optional |
| scheme | [string](#string) | optional | Scheme to use for connecting to the host. Defaults to HTTP. &#43;optional |
| httpHeaders | [HTTPHeader](#k8s.io.api.core.v1.HTTPHeader) | repeated | Custom headers to set in the request. HTTP allows repeated headers. &#43;optional |






<a name="k8s.io.api.core.v1.HTTPHeader"/>

### HTTPHeader
HTTPHeader describes a custom header to be used in HTTP probes


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | optional | The header field name |
| value | [string](#string) | optional | The header field value |






<a name="k8s.io.api.core.v1.Handler"/>

### Handler
Handler defines a specific action that should be taken
TODO: pass structured data to these actions, and document that data here.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| exec | [ExecAction](#k8s.io.api.core.v1.ExecAction) | optional | One and only one of the following should be specified. Exec specifies the action to take. &#43;optional |
| httpGet | [HTTPGetAction](#k8s.io.api.core.v1.HTTPGetAction) | optional | HTTPGet specifies the http request to perform. &#43;optional |
| tcpSocket | [TCPSocketAction](#k8s.io.api.core.v1.TCPSocketAction) | optional | TCPSocket specifies an action involving a TCP port. TCP hooks not yet supported TODO: implement a realistic TCP lifecycle hook &#43;optional |






<a name="k8s.io.api.core.v1.HostAlias"/>

### HostAlias
HostAlias holds the mapping between IP and hostnames that will be injected as an entry in the
pod&#39;s hosts file.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ip | [string](#string) | optional | IP address of the host file entry. |
| hostnames | [string](#string) | repeated | Hostnames for the above IP address. |






<a name="k8s.io.api.core.v1.HostPathVolumeSource"/>

### HostPathVolumeSource
Represents a host path mapped into a pod.
Host path volumes do not support ownership management or SELinux relabeling.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [string](#string) | optional | Path of the directory on the host. If the path is a symlink, it will follow the link to the real path. More info: https://kubernetes.io/docs/concepts/storage/volumes#hostpath |
| type | [string](#string) | optional | Type for HostPath Volume Defaults to &#34;&#34; More info: https://kubernetes.io/docs/concepts/storage/volumes#hostpath &#43;optional |






<a name="k8s.io.api.core.v1.ISCSIVolumeSource"/>

### ISCSIVolumeSource
Represents an ISCSI disk.
ISCSI volumes can only be mounted as read/write once.
ISCSI volumes support ownership management and SELinux relabeling.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| targetPortal | [string](#string) | optional | iSCSI target portal. The portal is either an IP or ip_addr:port if the port is other than default (typically TCP ports 860 and 3260). |
| iqn | [string](#string) | optional | Target iSCSI Qualified Name. |
| lun | [int32](#int32) | optional | iSCSI target lun number. |
| iscsiInterface | [string](#string) | optional | Optional: Defaults to &#39;default&#39; (tcp). iSCSI interface name that uses an iSCSI transport. &#43;optional |
| fsType | [string](#string) | optional | Filesystem type of the volume that you want to mount. Tip: Ensure that the filesystem type is supported by the host operating system. Examples: &#34;ext4&#34;, &#34;xfs&#34;, &#34;ntfs&#34;. Implicitly inferred to be &#34;ext4&#34; if unspecified. More info: https://kubernetes.io/docs/concepts/storage/volumes#iscsi TODO: how do we prevent errors in the filesystem from compromising the machine &#43;optional |
| readOnly | [bool](#bool) | optional | ReadOnly here will force the ReadOnly setting in VolumeMounts. Defaults to false. &#43;optional |
| portals | [string](#string) | repeated | iSCSI target portal List. The portal is either an IP or ip_addr:port if the port is other than default (typically TCP ports 860 and 3260). &#43;optional |
| chapAuthDiscovery | [bool](#bool) | optional | whether support iSCSI Discovery CHAP authentication &#43;optional |
| chapAuthSession | [bool](#bool) | optional | whether support iSCSI Session CHAP authentication &#43;optional |
| secretRef | [LocalObjectReference](#k8s.io.api.core.v1.LocalObjectReference) | optional | CHAP secret for iSCSI target and initiator authentication &#43;optional |
| initiatorName | [string](#string) | optional | Custom iSCSI initiator name. If initiatorName is specified with iscsiInterface simultaneously, new iSCSI interface &lt;target portal&gt;:&lt;volume name&gt; will be created for the connection. &#43;optional |






<a name="k8s.io.api.core.v1.KeyToPath"/>

### KeyToPath
Maps a string key to a path within a volume.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional | The key to project. |
| path | [string](#string) | optional | The relative path of the file to map the key to. May not be an absolute path. May not contain the path element &#39;..&#39;. May not start with the string &#39;..&#39;. |
| mode | [int32](#int32) | optional | Optional: mode bits to use on this file, must be a value between 0 and 0777. If not specified, the volume defaultMode will be used. This might be in conflict with other options that affect the file mode, like fsGroup, and the result can be other mode bits set. &#43;optional |






<a name="k8s.io.api.core.v1.Lifecycle"/>

### Lifecycle
Lifecycle describes actions that the management system should take in response to container lifecycle
events. For the PostStart and PreStop lifecycle handlers, management of the container blocks
until the action is complete, unless the container process fails, in which case the handler is aborted.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| postStart | [Handler](#k8s.io.api.core.v1.Handler) | optional | PostStart is called immediately after a container is created. If the handler fails, the container is terminated and restarted according to its restart policy. Other management of the container blocks until the hook completes. More info: https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/#container-hooks &#43;optional |
| preStop | [Handler](#k8s.io.api.core.v1.Handler) | optional | PreStop is called immediately before a container is terminated. The container is terminated after the handler completes. The reason for termination is passed to the handler. Regardless of the outcome of the handler, the container is eventually terminated. Other management of the container blocks until the hook completes. More info: https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/#container-hooks &#43;optional |






<a name="k8s.io.api.core.v1.LimitRange"/>

### LimitRange
LimitRange sets resource usage limits for each kind of resource in a Namespace.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta) | optional | Standard object&#39;s metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata &#43;optional |
| spec | [LimitRangeSpec](#k8s.io.api.core.v1.LimitRangeSpec) | optional | Spec defines the limits enforced. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status &#43;optional |






<a name="k8s.io.api.core.v1.LimitRangeItem"/>

### LimitRangeItem
LimitRangeItem defines a min/max usage limit for any resource that matches on kind.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [string](#string) | optional | Type of resource that this limit applies to. &#43;optional |
| max | [LimitRangeItem.MaxEntry](#k8s.io.api.core.v1.LimitRangeItem.MaxEntry) | repeated | Max usage constraints on this kind by resource name. &#43;optional |
| min | [LimitRangeItem.MinEntry](#k8s.io.api.core.v1.LimitRangeItem.MinEntry) | repeated | Min usage constraints on this kind by resource name. &#43;optional |
| default | [LimitRangeItem.DefaultEntry](#k8s.io.api.core.v1.LimitRangeItem.DefaultEntry) | repeated | Default resource requirement limit value by resource name if resource limit is omitted. &#43;optional |
| defaultRequest | [LimitRangeItem.DefaultRequestEntry](#k8s.io.api.core.v1.LimitRangeItem.DefaultRequestEntry) | repeated | DefaultRequest is the default resource requirement request value by resource name if resource request is omitted. &#43;optional |
| maxLimitRequestRatio | [LimitRangeItem.MaxLimitRequestRatioEntry](#k8s.io.api.core.v1.LimitRangeItem.MaxLimitRequestRatioEntry) | repeated | MaxLimitRequestRatio if specified, the named resource must have a request and limit that are both non-zero where limit divided by request is less than or equal to the enumerated value; this represents the max burst for the named resource. &#43;optional |






<a name="k8s.io.api.core.v1.LimitRangeItem.DefaultEntry"/>

### LimitRangeItem.DefaultEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [.k8s.io.apimachinery.pkg.api.resource.Quantity](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.api.resource.Quantity) | optional |  |






<a name="k8s.io.api.core.v1.LimitRangeItem.DefaultRequestEntry"/>

### LimitRangeItem.DefaultRequestEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [.k8s.io.apimachinery.pkg.api.resource.Quantity](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.api.resource.Quantity) | optional |  |






<a name="k8s.io.api.core.v1.LimitRangeItem.MaxEntry"/>

### LimitRangeItem.MaxEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [.k8s.io.apimachinery.pkg.api.resource.Quantity](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.api.resource.Quantity) | optional |  |






<a name="k8s.io.api.core.v1.LimitRangeItem.MaxLimitRequestRatioEntry"/>

### LimitRangeItem.MaxLimitRequestRatioEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [.k8s.io.apimachinery.pkg.api.resource.Quantity](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.api.resource.Quantity) | optional |  |






<a name="k8s.io.api.core.v1.LimitRangeItem.MinEntry"/>

### LimitRangeItem.MinEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [.k8s.io.apimachinery.pkg.api.resource.Quantity](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.api.resource.Quantity) | optional |  |






<a name="k8s.io.api.core.v1.LimitRangeList"/>

### LimitRangeList
LimitRangeList is a list of LimitRange items.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta) | optional | Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds &#43;optional |
| items | [LimitRange](#k8s.io.api.core.v1.LimitRange) | repeated | Items is a list of LimitRange objects. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container |






<a name="k8s.io.api.core.v1.LimitRangeSpec"/>

### LimitRangeSpec
LimitRangeSpec defines a min/max usage limit for resources that match on kind.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| limits | [LimitRangeItem](#k8s.io.api.core.v1.LimitRangeItem) | repeated | Limits is the list of LimitRangeItem objects that are enforced. |






<a name="k8s.io.api.core.v1.List"/>

### List
List holds a list of objects, which may not be known by the server.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta) | optional | Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds &#43;optional |
| items | [.k8s.io.apimachinery.pkg.runtime.RawExtension](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.runtime.RawExtension) | repeated | List of objects |






<a name="k8s.io.api.core.v1.ListOptions"/>

### ListOptions
ListOptions is the query options to a standard REST list call.
DEPRECATED: This type has been moved to meta/v1 and will be removed soon.
&#43;k8s:openapi-gen=false


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| labelSelector | [string](#string) | optional | A selector to restrict the list of returned objects by their labels. Defaults to everything. &#43;optional |
| fieldSelector | [string](#string) | optional | A selector to restrict the list of returned objects by their fields. Defaults to everything. &#43;optional |
| includeUninitialized | [bool](#bool) | optional | If true, partially initialized resources are included in the response. &#43;optional |
| watch | [bool](#bool) | optional | Watch for changes to the described resources and return them as a stream of add, update, and remove notifications. Specify resourceVersion. &#43;optional |
| resourceVersion | [string](#string) | optional | When specified with a watch call, shows changes that occur after that particular version of a resource. Defaults to changes from the beginning of history. When specified for list: - if unset, then the result is returned from remote storage based on quorum-read flag; - if it&#39;s 0, then we simply return what we currently have in cache, no guarantee; - if set to non zero, then the result is at least as fresh as given rv. &#43;optional |
| timeoutSeconds | [int64](#int64) | optional | Timeout for the list/watch call. &#43;optional |






<a name="k8s.io.api.core.v1.LoadBalancerIngress"/>

### LoadBalancerIngress
LoadBalancerIngress represents the status of a load-balancer ingress point:
traffic intended for the service should be sent to an ingress point.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ip | [string](#string) | optional | IP is set for load-balancer ingress points that are IP based (typically GCE or OpenStack load-balancers) &#43;optional |
| hostname | [string](#string) | optional | Hostname is set for load-balancer ingress points that are DNS based (typically AWS load-balancers) &#43;optional |






<a name="k8s.io.api.core.v1.LoadBalancerStatus"/>

### LoadBalancerStatus
LoadBalancerStatus represents the status of a load-balancer.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ingress | [LoadBalancerIngress](#k8s.io.api.core.v1.LoadBalancerIngress) | repeated | Ingress is a list containing ingress points for the load-balancer. Traffic intended for the service should be sent to these ingress points. &#43;optional |






<a name="k8s.io.api.core.v1.LocalObjectReference"/>

### LocalObjectReference
LocalObjectReference contains enough information to let you locate the
referenced object inside the same namespace.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | optional | Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid? &#43;optional |






<a name="k8s.io.api.core.v1.LocalVolumeSource"/>

### LocalVolumeSource
Local represents directly-attached storage with node affinity


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [string](#string) | optional | The full path to the volume on the node For alpha, this path must be a directory Once block as a source is supported, then this path can point to a block device |






<a name="k8s.io.api.core.v1.NFSVolumeSource"/>

### NFSVolumeSource
Represents an NFS mount that lasts the lifetime of a pod.
NFS volumes do not support ownership management or SELinux relabeling.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| server | [string](#string) | optional | Server is the hostname or IP address of the NFS server. More info: https://kubernetes.io/docs/concepts/storage/volumes#nfs |
| path | [string](#string) | optional | Path that is exported by the NFS server. More info: https://kubernetes.io/docs/concepts/storage/volumes#nfs |
| readOnly | [bool](#bool) | optional | ReadOnly here will force the NFS export to be mounted with read-only permissions. Defaults to false. More info: https://kubernetes.io/docs/concepts/storage/volumes#nfs &#43;optional |






<a name="k8s.io.api.core.v1.Namespace"/>

### Namespace
Namespace provides a scope for Names.
Use of multiple namespaces is optional.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta) | optional | Standard object&#39;s metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata &#43;optional |
| spec | [NamespaceSpec](#k8s.io.api.core.v1.NamespaceSpec) | optional | Spec defines the behavior of the Namespace. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status &#43;optional |
| status | [NamespaceStatus](#k8s.io.api.core.v1.NamespaceStatus) | optional | Status describes the current status of a Namespace. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status &#43;optional |






<a name="k8s.io.api.core.v1.NamespaceList"/>

### NamespaceList
NamespaceList is a list of Namespaces.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta) | optional | Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds &#43;optional |
| items | [Namespace](#k8s.io.api.core.v1.Namespace) | repeated | Items is the list of Namespace objects in the list. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces |






<a name="k8s.io.api.core.v1.NamespaceSpec"/>

### NamespaceSpec
NamespaceSpec describes the attributes on a Namespace.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| finalizers | [string](#string) | repeated | Finalizers is an opaque list of values that must be empty to permanently remove object from storage. More info: https://kubernetes.io/docs/tasks/administer-cluster/namespaces &#43;optional |






<a name="k8s.io.api.core.v1.NamespaceStatus"/>

### NamespaceStatus
NamespaceStatus is information about the current status of a Namespace.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| phase | [string](#string) | optional | Phase is the current lifecycle phase of the namespace. More info: https://kubernetes.io/docs/tasks/administer-cluster/namespaces &#43;optional |






<a name="k8s.io.api.core.v1.Node"/>

### Node
Node is a worker node in Kubernetes.
Each node will have a unique identifier in the cache (i.e. in etcd).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta) | optional | Standard object&#39;s metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata &#43;optional |
| spec | [NodeSpec](#k8s.io.api.core.v1.NodeSpec) | optional | Spec defines the behavior of a node. https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status &#43;optional |
| status | [NodeStatus](#k8s.io.api.core.v1.NodeStatus) | optional | Most recently observed status of the node. Populated by the system. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status &#43;optional |






<a name="k8s.io.api.core.v1.NodeAddress"/>

### NodeAddress
NodeAddress contains information for the node&#39;s address.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [string](#string) | optional | Node address type, one of Hostname, ExternalIP or InternalIP. |
| address | [string](#string) | optional | The node address. |






<a name="k8s.io.api.core.v1.NodeAffinity"/>

### NodeAffinity
Node affinity is a group of node affinity scheduling rules.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| requiredDuringSchedulingIgnoredDuringExecution | [NodeSelector](#k8s.io.api.core.v1.NodeSelector) | optional | If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to an update), the system may or may not try to eventually evict the pod from its node. &#43;optional |
| preferredDuringSchedulingIgnoredDuringExecution | [PreferredSchedulingTerm](#k8s.io.api.core.v1.PreferredSchedulingTerm) | repeated | The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding &#34;weight&#34; to the sum if the node matches the corresponding matchExpressions; the node(s) with the highest sum are the most preferred. &#43;optional |






<a name="k8s.io.api.core.v1.NodeCondition"/>

### NodeCondition
NodeCondition contains condition information for a node.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [string](#string) | optional | Type of node condition. |
| status | [string](#string) | optional | Status of the condition, one of True, False, Unknown. |
| lastHeartbeatTime | [.k8s.io.apimachinery.pkg.apis.meta.v1.Time](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.Time) | optional | Last time we got an update on a given condition. &#43;optional |
| lastTransitionTime | [.k8s.io.apimachinery.pkg.apis.meta.v1.Time](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.Time) | optional | Last time the condition transit from one status to another. &#43;optional |
| reason | [string](#string) | optional | (brief) reason for the condition&#39;s last transition. &#43;optional |
| message | [string](#string) | optional | Human readable message indicating details about last transition. &#43;optional |






<a name="k8s.io.api.core.v1.NodeConfigSource"/>

### NodeConfigSource
NodeConfigSource specifies a source of node configuration. Exactly one subfield (excluding metadata) must be non-nil.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| configMapRef | [ObjectReference](#k8s.io.api.core.v1.ObjectReference) | optional |  |






<a name="k8s.io.api.core.v1.NodeDaemonEndpoints"/>

### NodeDaemonEndpoints
NodeDaemonEndpoints lists ports opened by daemons running on the Node.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeletEndpoint | [DaemonEndpoint](#k8s.io.api.core.v1.DaemonEndpoint) | optional | Endpoint on which Kubelet is listening. &#43;optional |






<a name="k8s.io.api.core.v1.NodeList"/>

### NodeList
NodeList is the whole list of all Nodes which have been registered with master.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta) | optional | Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds &#43;optional |
| items | [Node](#k8s.io.api.core.v1.Node) | repeated | List of nodes |






<a name="k8s.io.api.core.v1.NodeProxyOptions"/>

### NodeProxyOptions
NodeProxyOptions is the query options to a Node&#39;s proxy call.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [string](#string) | optional | Path is the URL path to use for the current proxy request to node. &#43;optional |






<a name="k8s.io.api.core.v1.NodeResources"/>

### NodeResources
NodeResources is an object for conveying resource information about a node.
see http://releases.k8s.io/HEAD/docs/design/resources.md for more details.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capacity | [NodeResources.CapacityEntry](#k8s.io.api.core.v1.NodeResources.CapacityEntry) | repeated | Capacity represents the available resources of a node |






<a name="k8s.io.api.core.v1.NodeResources.CapacityEntry"/>

### NodeResources.CapacityEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [.k8s.io.apimachinery.pkg.api.resource.Quantity](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.api.resource.Quantity) | optional |  |






<a name="k8s.io.api.core.v1.NodeSelector"/>

### NodeSelector
A node selector represents the union of the results of one or more label queries
over a set of nodes; that is, it represents the OR of the selectors represented
by the node selector terms.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| nodeSelectorTerms | [NodeSelectorTerm](#k8s.io.api.core.v1.NodeSelectorTerm) | repeated | Required. A list of node selector terms. The terms are ORed. |






<a name="k8s.io.api.core.v1.NodeSelectorRequirement"/>

### NodeSelectorRequirement
A node selector requirement is a selector that contains values, a key, and an operator
that relates the key and values.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional | The label key that the selector applies to. |
| operator | [string](#string) | optional | Represents a key&#39;s relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt. |
| values | [string](#string) | repeated | An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch. &#43;optional |






<a name="k8s.io.api.core.v1.NodeSelectorTerm"/>

### NodeSelectorTerm
A null or empty node selector term matches no objects.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| matchExpressions | [NodeSelectorRequirement](#k8s.io.api.core.v1.NodeSelectorRequirement) | repeated | Required. A list of node selector requirements. The requirements are ANDed. |






<a name="k8s.io.api.core.v1.NodeSpec"/>

### NodeSpec
NodeSpec describes the attributes that a node is created with.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| podCIDR | [string](#string) | optional | PodCIDR represents the pod IP range assigned to the node. &#43;optional |
| externalID | [string](#string) | optional | External ID of the node assigned by some machine database (e.g. a cloud provider). Deprecated. &#43;optional |
| providerID | [string](#string) | optional | ID of the node assigned by the cloud provider in the format: &lt;ProviderName&gt;://&lt;ProviderSpecificNodeID&gt; &#43;optional |
| unschedulable | [bool](#bool) | optional | Unschedulable controls node schedulability of new pods. By default, node is schedulable. More info: https://kubernetes.io/docs/concepts/nodes/node/#manual-node-administration &#43;optional |
| taints | [Taint](#k8s.io.api.core.v1.Taint) | repeated | If specified, the node&#39;s taints. &#43;optional |
| configSource | [NodeConfigSource](#k8s.io.api.core.v1.NodeConfigSource) | optional | If specified, the source to get node configuration from The DynamicKubeletConfig feature gate must be enabled for the Kubelet to use this field &#43;optional |






<a name="k8s.io.api.core.v1.NodeStatus"/>

### NodeStatus
NodeStatus is information about the current status of a node.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capacity | [NodeStatus.CapacityEntry](#k8s.io.api.core.v1.NodeStatus.CapacityEntry) | repeated | Capacity represents the total resources of a node. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#capacity &#43;optional |
| allocatable | [NodeStatus.AllocatableEntry](#k8s.io.api.core.v1.NodeStatus.AllocatableEntry) | repeated | Allocatable represents the resources of a node that are available for scheduling. Defaults to Capacity. &#43;optional |
| phase | [string](#string) | optional | NodePhase is the recently observed lifecycle phase of the node. More info: https://kubernetes.io/docs/concepts/nodes/node/#phase The field is never populated, and now is deprecated. &#43;optional |
| conditions | [NodeCondition](#k8s.io.api.core.v1.NodeCondition) | repeated | Conditions is an array of current observed node conditions. More info: https://kubernetes.io/docs/concepts/nodes/node/#condition &#43;optional &#43;patchMergeKey=type &#43;patchStrategy=merge |
| addresses | [NodeAddress](#k8s.io.api.core.v1.NodeAddress) | repeated | List of addresses reachable to the node. Queried from cloud provider, if available. More info: https://kubernetes.io/docs/concepts/nodes/node/#addresses &#43;optional &#43;patchMergeKey=type &#43;patchStrategy=merge |
| daemonEndpoints | [NodeDaemonEndpoints](#k8s.io.api.core.v1.NodeDaemonEndpoints) | optional | Endpoints of daemons running on the Node. &#43;optional |
| nodeInfo | [NodeSystemInfo](#k8s.io.api.core.v1.NodeSystemInfo) | optional | Set of ids/uuids to uniquely identify the node. More info: https://kubernetes.io/docs/concepts/nodes/node/#info &#43;optional |
| images | [ContainerImage](#k8s.io.api.core.v1.ContainerImage) | repeated | List of container images on this node &#43;optional |
| volumesInUse | [string](#string) | repeated | List of attachable volumes in use (mounted) by the node. &#43;optional |
| volumesAttached | [AttachedVolume](#k8s.io.api.core.v1.AttachedVolume) | repeated | List of volumes that are attached to the node. &#43;optional |






<a name="k8s.io.api.core.v1.NodeStatus.AllocatableEntry"/>

### NodeStatus.AllocatableEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [.k8s.io.apimachinery.pkg.api.resource.Quantity](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.api.resource.Quantity) | optional |  |






<a name="k8s.io.api.core.v1.NodeStatus.CapacityEntry"/>

### NodeStatus.CapacityEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [.k8s.io.apimachinery.pkg.api.resource.Quantity](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.api.resource.Quantity) | optional |  |






<a name="k8s.io.api.core.v1.NodeSystemInfo"/>

### NodeSystemInfo
NodeSystemInfo is a set of ids/uuids to uniquely identify the node.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| machineID | [string](#string) | optional | MachineID reported by the node. For unique machine identification in the cluster this field is preferred. Learn more from man(5) machine-id: http://man7.org/linux/man-pages/man5/machine-id.5.html |
| systemUUID | [string](#string) | optional | SystemUUID reported by the node. For unique machine identification MachineID is preferred. This field is specific to Red Hat hosts https://access.redhat.com/documentation/en-US/Red_Hat_Subscription_Management/1/html/RHSM/getting-system-uuid.html |
| bootID | [string](#string) | optional | Boot ID reported by the node. |
| kernelVersion | [string](#string) | optional | Kernel Version reported by the node from &#39;uname -r&#39; (e.g. 3.16.0-0.bpo.4-amd64). |
| osImage | [string](#string) | optional | OS Image reported by the node from /etc/os-release (e.g. Debian GNU/Linux 7 (wheezy)). |
| containerRuntimeVersion | [string](#string) | optional | ContainerRuntime Version reported by the node through runtime remote API (e.g. docker://1.5.0). |
| kubeletVersion | [string](#string) | optional | Kubelet Version reported by the node. |
| kubeProxyVersion | [string](#string) | optional | KubeProxy Version reported by the node. |
| operatingSystem | [string](#string) | optional | The Operating System reported by the node |
| architecture | [string](#string) | optional | The Architecture reported by the node |






<a name="k8s.io.api.core.v1.ObjectFieldSelector"/>

### ObjectFieldSelector
ObjectFieldSelector selects an APIVersioned field of an object.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| apiVersion | [string](#string) | optional | Version of the schema the FieldPath is written in terms of, defaults to &#34;v1&#34;. &#43;optional |
| fieldPath | [string](#string) | optional | Path of the field to select in the specified API version. |






<a name="k8s.io.api.core.v1.ObjectMeta"/>

### ObjectMeta
ObjectMeta is metadata that all persisted resources must have, which includes all objects
users must create.
DEPRECATED: Use k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta instead - this type will be removed soon.
&#43;k8s:openapi-gen=false


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | optional | Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names &#43;optional |
| generateName | [string](#string) | optional | GenerateName is an optional prefix, used by the server, to generate a unique name ONLY IF the Name field has not been provided. If this field is used, the name returned to the client will be different than the name passed. This value will also be combined with a unique suffix. The provided value has the same validation rules as the Name field, and may be truncated by the length of the suffix required to make the value unique on the server. If this field is specified and the generated name exists, the server will NOT return a 409 - instead, it will either return 201 Created or 500 with Reason ServerTimeout indicating a unique name could not be found in the time allotted, and the client should retry (optionally after the time indicated in the Retry-After header). Applied only if Name is not specified. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#idempotency &#43;optional |
| namespace | [string](#string) | optional | Namespace defines the space within each name must be unique. An empty namespace is equivalent to the &#34;default&#34; namespace, but &#34;default&#34; is the canonical representation. Not all objects are required to be scoped to a namespace - the value of this field for those objects will be empty. Must be a DNS_LABEL. Cannot be updated. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces &#43;optional |
| selfLink | [string](#string) | optional | SelfLink is a URL representing this object. Populated by the system. Read-only. &#43;optional |
| uid | [string](#string) | optional | UID is the unique in time and space value for this object. It is typically generated by the server on successful creation of a resource and is not allowed to change on PUT operations. Populated by the system. Read-only. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#uids &#43;optional |
| resourceVersion | [string](#string) | optional | An opaque value that represents the internal version of this object that can be used by clients to determine when objects have changed. May be used for optimistic concurrency, change detection, and the watch operation on a resource or set of resources. Clients must treat these values as opaque and passed unmodified back to the server. They may only be valid for a particular resource or set of resources. Populated by the system. Read-only. Value must be treated as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency &#43;optional |
| generation | [int64](#int64) | optional | A sequence number representing a specific generation of the desired state. Populated by the system. Read-only. &#43;optional |
| creationTimestamp | [.k8s.io.apimachinery.pkg.apis.meta.v1.Time](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.Time) | optional | CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC. Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata &#43;optional |
| deletionTimestamp | [.k8s.io.apimachinery.pkg.apis.meta.v1.Time](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.Time) | optional | DeletionTimestamp is RFC 3339 date and time at which this resource will be deleted. This field is set by the server when a graceful deletion is requested by the user, and is not directly settable by a client. The resource is expected to be deleted (no longer visible from resource lists, and not reachable by name) after the time in this field. Once set, this value may not be unset or be set further into the future, although it may be shortened or the resource may be deleted prior to this time. For example, a user may request that a pod is deleted in 30 seconds. The Kubelet will react by sending a graceful termination signal to the containers in the pod. After that 30 seconds, the Kubelet will send a hard termination signal (SIGKILL) to the container and after cleanup, remove the pod from the API. In the presence of network partitions, this object may still exist after this timestamp, until an administrator or automated process can determine the resource is fully terminated. If not set, graceful deletion of the object has not been requested. Populated by the system when a graceful deletion is requested. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata &#43;optional |
| deletionGracePeriodSeconds | [int64](#int64) | optional | Number of seconds allowed for this object to gracefully terminate before it will be removed from the system. Only set when deletionTimestamp is also set. May only be shortened. Read-only. &#43;optional |
| labels | [ObjectMeta.LabelsEntry](#k8s.io.api.core.v1.ObjectMeta.LabelsEntry) | repeated | Map of string keys and values that can be used to organize and categorize (scope and select) objects. May match selectors of replication controllers and services. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels &#43;optional |
| annotations | [ObjectMeta.AnnotationsEntry](#k8s.io.api.core.v1.ObjectMeta.AnnotationsEntry) | repeated | Annotations is an unstructured key value map stored with a resource that may be set by external tools to store and retrieve arbitrary metadata. They are not queryable and should be preserved when modifying objects. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations &#43;optional |
| ownerReferences | [.k8s.io.apimachinery.pkg.apis.meta.v1.OwnerReference](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.OwnerReference) | repeated | List of objects depended by this object. If ALL objects in the list have been deleted, this object will be garbage collected. If this object is managed by a controller, then an entry in this list will point to this controller, with the controller field set to true. There cannot be more than one managing controller. &#43;optional &#43;patchMergeKey=uid &#43;patchStrategy=merge |
| initializers | [.k8s.io.apimachinery.pkg.apis.meta.v1.Initializers](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.Initializers) | optional | An initializer is a controller which enforces some system invariant at object creation time. This field is a list of initializers that have not yet acted on this object. If nil or empty, this object has been completely initialized. Otherwise, the object is considered uninitialized and is hidden (in list/watch and get calls) from clients that haven&#39;t explicitly asked to observe uninitialized objects. When an object is created, the system will populate this list with the current set of initializers. Only privileged users may set or modify this list. Once it is empty, it may not be modified further by any user. |
| finalizers | [string](#string) | repeated | Must be empty before the object is deleted from the registry. Each entry is an identifier for the responsible component that will remove the entry from the list. If the deletionTimestamp of the object is non-nil, entries in this list can only be removed. &#43;optional &#43;patchStrategy=merge |
| clusterName | [string](#string) | optional | The name of the cluster which the object belongs to. This is used to distinguish resources with same name and namespace in different clusters. This field is not set anywhere right now and apiserver is going to ignore it if set in create or update request. &#43;optional |






<a name="k8s.io.api.core.v1.ObjectMeta.AnnotationsEntry"/>

### ObjectMeta.AnnotationsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [string](#string) | optional |  |






<a name="k8s.io.api.core.v1.ObjectMeta.LabelsEntry"/>

### ObjectMeta.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [string](#string) | optional |  |






<a name="k8s.io.api.core.v1.ObjectReference"/>

### ObjectReference
ObjectReference contains enough information to let you inspect or modify the referred object.
&#43;k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kind | [string](#string) | optional | Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds &#43;optional |
| namespace | [string](#string) | optional | Namespace of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces &#43;optional |
| name | [string](#string) | optional | Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names &#43;optional |
| uid | [string](#string) | optional | UID of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#uids &#43;optional |
| apiVersion | [string](#string) | optional | API version of the referent. &#43;optional |
| resourceVersion | [string](#string) | optional | Specific resourceVersion to which this reference is made, if any. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency &#43;optional |
| fieldPath | [string](#string) | optional | If referring to a piece of an object instead of an entire object, this string should contain a valid JSON/Go field access statement, such as desiredState.manifest.containers[2]. For example, if the object reference is to a container within a pod, this would take on a value like: &#34;spec.containers{name}&#34; (where &#34;name&#34; refers to the name of the container that triggered the event) or if no container name is specified &#34;spec.containers[2]&#34; (container with index 2 in this pod). This syntax is chosen only to have some well-defined way of referencing a part of an object. TODO: this design is not final and this field is subject to change in the future. &#43;optional |






<a name="k8s.io.api.core.v1.PersistentVolume"/>

### PersistentVolume
PersistentVolume (PV) is a storage resource provisioned by an administrator.
It is analogous to a node.
More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta) | optional | Standard object&#39;s metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata &#43;optional |
| spec | [PersistentVolumeSpec](#k8s.io.api.core.v1.PersistentVolumeSpec) | optional | Spec defines a specification of a persistent volume owned by the cluster. Provisioned by an administrator. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistent-volumes &#43;optional |
| status | [PersistentVolumeStatus](#k8s.io.api.core.v1.PersistentVolumeStatus) | optional | Status represents the current information/status for the persistent volume. Populated by the system. Read-only. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistent-volumes &#43;optional |






<a name="k8s.io.api.core.v1.PersistentVolumeClaim"/>

### PersistentVolumeClaim
PersistentVolumeClaim is a user&#39;s request for and claim to a persistent volume


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta) | optional | Standard object&#39;s metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata &#43;optional |
| spec | [PersistentVolumeClaimSpec](#k8s.io.api.core.v1.PersistentVolumeClaimSpec) | optional | Spec defines the desired characteristics of a volume requested by a pod author. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims &#43;optional |
| status | [PersistentVolumeClaimStatus](#k8s.io.api.core.v1.PersistentVolumeClaimStatus) | optional | Status represents the current information/status of a persistent volume claim. Read-only. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims &#43;optional |






<a name="k8s.io.api.core.v1.PersistentVolumeClaimCondition"/>

### PersistentVolumeClaimCondition
PersistentVolumeClaimCondition contails details about state of pvc


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [string](#string) | optional |  |
| status | [string](#string) | optional |  |
| lastProbeTime | [.k8s.io.apimachinery.pkg.apis.meta.v1.Time](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.Time) | optional | Last time we probed the condition. &#43;optional |
| lastTransitionTime | [.k8s.io.apimachinery.pkg.apis.meta.v1.Time](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.Time) | optional | Last time the condition transitioned from one status to another. &#43;optional |
| reason | [string](#string) | optional | Unique, this should be a short, machine understandable string that gives the reason for condition&#39;s last transition. If it reports &#34;ResizeStarted&#34; that means the underlying persistent volume is being resized. &#43;optional |
| message | [string](#string) | optional | Human-readable message indicating details about last transition. &#43;optional |






<a name="k8s.io.api.core.v1.PersistentVolumeClaimList"/>

### PersistentVolumeClaimList
PersistentVolumeClaimList is a list of PersistentVolumeClaim items.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta) | optional | Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds &#43;optional |
| items | [PersistentVolumeClaim](#k8s.io.api.core.v1.PersistentVolumeClaim) | repeated | A list of persistent volume claims. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims |






<a name="k8s.io.api.core.v1.PersistentVolumeClaimSpec"/>

### PersistentVolumeClaimSpec
PersistentVolumeClaimSpec describes the common attributes of storage devices
and allows a Source for provider-specific attributes


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| accessModes | [string](#string) | repeated | AccessModes contains the desired access modes the volume should have. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1 &#43;optional |
| selector | [.k8s.io.apimachinery.pkg.apis.meta.v1.LabelSelector](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.LabelSelector) | optional | A label query over volumes to consider for binding. &#43;optional |
| resources | [ResourceRequirements](#k8s.io.api.core.v1.ResourceRequirements) | optional | Resources represents the minimum resources the volume should have. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources &#43;optional |
| volumeName | [string](#string) | optional | VolumeName is the binding reference to the PersistentVolume backing this claim. &#43;optional |
| storageClassName | [string](#string) | optional | Name of the StorageClass required by the claim. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1 &#43;optional |






<a name="k8s.io.api.core.v1.PersistentVolumeClaimStatus"/>

### PersistentVolumeClaimStatus
PersistentVolumeClaimStatus is the current status of a persistent volume claim.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| phase | [string](#string) | optional | Phase represents the current phase of PersistentVolumeClaim. &#43;optional |
| accessModes | [string](#string) | repeated | AccessModes contains the actual access modes the volume backing the PVC has. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1 &#43;optional |
| capacity | [PersistentVolumeClaimStatus.CapacityEntry](#k8s.io.api.core.v1.PersistentVolumeClaimStatus.CapacityEntry) | repeated | Represents the actual resources of the underlying volume. &#43;optional |
| conditions | [PersistentVolumeClaimCondition](#k8s.io.api.core.v1.PersistentVolumeClaimCondition) | repeated | Current Condition of persistent volume claim. If underlying persistent volume is being resized then the Condition will be set to &#39;ResizeStarted&#39;. &#43;optional &#43;patchMergeKey=type &#43;patchStrategy=merge |






<a name="k8s.io.api.core.v1.PersistentVolumeClaimStatus.CapacityEntry"/>

### PersistentVolumeClaimStatus.CapacityEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [.k8s.io.apimachinery.pkg.api.resource.Quantity](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.api.resource.Quantity) | optional |  |






<a name="k8s.io.api.core.v1.PersistentVolumeClaimVolumeSource"/>

### PersistentVolumeClaimVolumeSource
PersistentVolumeClaimVolumeSource references the user&#39;s PVC in the same namespace.
This volume finds the bound PV and mounts that volume for the pod. A
PersistentVolumeClaimVolumeSource is, essentially, a wrapper around another
type of volume that is owned by someone else (the system).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| claimName | [string](#string) | optional | ClaimName is the name of a PersistentVolumeClaim in the same namespace as the pod using this volume. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims |
| readOnly | [bool](#bool) | optional | Will force the ReadOnly setting in VolumeMounts. Default false. &#43;optional |






<a name="k8s.io.api.core.v1.PersistentVolumeList"/>

### PersistentVolumeList
PersistentVolumeList is a list of PersistentVolume items.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta) | optional | Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds &#43;optional |
| items | [PersistentVolume](#k8s.io.api.core.v1.PersistentVolume) | repeated | List of persistent volumes. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes |






<a name="k8s.io.api.core.v1.PersistentVolumeSource"/>

### PersistentVolumeSource
PersistentVolumeSource is similar to VolumeSource but meant for the
administrator who creates PVs. Exactly one of its members must be set.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| gcePersistentDisk | [GCEPersistentDiskVolumeSource](#k8s.io.api.core.v1.GCEPersistentDiskVolumeSource) | optional | GCEPersistentDisk represents a GCE Disk resource that is attached to a kubelet&#39;s host machine and then exposed to the pod. Provisioned by an admin. More info: https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk &#43;optional |
| awsElasticBlockStore | [AWSElasticBlockStoreVolumeSource](#k8s.io.api.core.v1.AWSElasticBlockStoreVolumeSource) | optional | AWSElasticBlockStore represents an AWS Disk resource that is attached to a kubelet&#39;s host machine and then exposed to the pod. More info: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore &#43;optional |
| hostPath | [HostPathVolumeSource](#k8s.io.api.core.v1.HostPathVolumeSource) | optional | HostPath represents a directory on the host. Provisioned by a developer or tester. This is useful for single-node development and testing only! On-host storage is not supported in any way and WILL NOT WORK in a multi-node cluster. More info: https://kubernetes.io/docs/concepts/storage/volumes#hostpath &#43;optional |
| glusterfs | [GlusterfsVolumeSource](#k8s.io.api.core.v1.GlusterfsVolumeSource) | optional | Glusterfs represents a Glusterfs volume that is attached to a host and exposed to the pod. Provisioned by an admin. More info: https://releases.k8s.io/HEAD/examples/volumes/glusterfs/README.md &#43;optional |
| nfs | [NFSVolumeSource](#k8s.io.api.core.v1.NFSVolumeSource) | optional | NFS represents an NFS mount on the host. Provisioned by an admin. More info: https://kubernetes.io/docs/concepts/storage/volumes#nfs &#43;optional |
| rbd | [RBDPersistentVolumeSource](#k8s.io.api.core.v1.RBDPersistentVolumeSource) | optional | RBD represents a Rados Block Device mount on the host that shares a pod&#39;s lifetime. More info: https://releases.k8s.io/HEAD/examples/volumes/rbd/README.md &#43;optional |
| iscsi | [ISCSIVolumeSource](#k8s.io.api.core.v1.ISCSIVolumeSource) | optional | ISCSI represents an ISCSI Disk resource that is attached to a kubelet&#39;s host machine and then exposed to the pod. Provisioned by an admin. &#43;optional |
| cinder | [CinderVolumeSource](#k8s.io.api.core.v1.CinderVolumeSource) | optional | Cinder represents a cinder volume attached and mounted on kubelets host machine More info: https://releases.k8s.io/HEAD/examples/mysql-cinder-pd/README.md &#43;optional |
| cephfs | [CephFSPersistentVolumeSource](#k8s.io.api.core.v1.CephFSPersistentVolumeSource) | optional | CephFS represents a Ceph FS mount on the host that shares a pod&#39;s lifetime &#43;optional |
| fc | [FCVolumeSource](#k8s.io.api.core.v1.FCVolumeSource) | optional | FC represents a Fibre Channel resource that is attached to a kubelet&#39;s host machine and then exposed to the pod. &#43;optional |
| flocker | [FlockerVolumeSource](#k8s.io.api.core.v1.FlockerVolumeSource) | optional | Flocker represents a Flocker volume attached to a kubelet&#39;s host machine and exposed to the pod for its usage. This depends on the Flocker control service being running &#43;optional |
| flexVolume | [FlexVolumeSource](#k8s.io.api.core.v1.FlexVolumeSource) | optional | FlexVolume represents a generic volume resource that is provisioned/attached using an exec based plugin. This is an alpha feature and may change in future. &#43;optional |
| azureFile | [AzureFilePersistentVolumeSource](#k8s.io.api.core.v1.AzureFilePersistentVolumeSource) | optional | AzureFile represents an Azure File Service mount on the host and bind mount to the pod. &#43;optional |
| vsphereVolume | [VsphereVirtualDiskVolumeSource](#k8s.io.api.core.v1.VsphereVirtualDiskVolumeSource) | optional | VsphereVolume represents a vSphere volume attached and mounted on kubelets host machine &#43;optional |
| quobyte | [QuobyteVolumeSource](#k8s.io.api.core.v1.QuobyteVolumeSource) | optional | Quobyte represents a Quobyte mount on the host that shares a pod&#39;s lifetime &#43;optional |
| azureDisk | [AzureDiskVolumeSource](#k8s.io.api.core.v1.AzureDiskVolumeSource) | optional | AzureDisk represents an Azure Data Disk mount on the host and bind mount to the pod. &#43;optional |
| photonPersistentDisk | [PhotonPersistentDiskVolumeSource](#k8s.io.api.core.v1.PhotonPersistentDiskVolumeSource) | optional | PhotonPersistentDisk represents a PhotonController persistent disk attached and mounted on kubelets host machine |
| portworxVolume | [PortworxVolumeSource](#k8s.io.api.core.v1.PortworxVolumeSource) | optional | PortworxVolume represents a portworx volume attached and mounted on kubelets host machine &#43;optional |
| scaleIO | [ScaleIOPersistentVolumeSource](#k8s.io.api.core.v1.ScaleIOPersistentVolumeSource) | optional | ScaleIO represents a ScaleIO persistent volume attached and mounted on Kubernetes nodes. &#43;optional |
| local | [LocalVolumeSource](#k8s.io.api.core.v1.LocalVolumeSource) | optional | Local represents directly-attached storage with node affinity &#43;optional |
| storageos | [StorageOSPersistentVolumeSource](#k8s.io.api.core.v1.StorageOSPersistentVolumeSource) | optional | StorageOS represents a StorageOS volume that is attached to the kubelet&#39;s host machine and mounted into the pod More info: https://releases.k8s.io/HEAD/examples/volumes/storageos/README.md &#43;optional |






<a name="k8s.io.api.core.v1.PersistentVolumeSpec"/>

### PersistentVolumeSpec
PersistentVolumeSpec is the specification of a persistent volume.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capacity | [PersistentVolumeSpec.CapacityEntry](#k8s.io.api.core.v1.PersistentVolumeSpec.CapacityEntry) | repeated | A description of the persistent volume&#39;s resources and capacity. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#capacity &#43;optional |
| persistentVolumeSource | [PersistentVolumeSource](#k8s.io.api.core.v1.PersistentVolumeSource) | optional | The actual volume backing the persistent volume. |
| accessModes | [string](#string) | repeated | AccessModes contains all ways the volume can be mounted. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes &#43;optional |
| claimRef | [ObjectReference](#k8s.io.api.core.v1.ObjectReference) | optional | ClaimRef is part of a bi-directional binding between PersistentVolume and PersistentVolumeClaim. Expected to be non-nil when bound. claim.VolumeName is the authoritative bind between PV and PVC. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#binding &#43;optional |
| persistentVolumeReclaimPolicy | [string](#string) | optional | What happens to a persistent volume when released from its claim. Valid options are Retain (default) and Recycle. Recycling must be supported by the volume plugin underlying this persistent volume. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#reclaiming &#43;optional |
| storageClassName | [string](#string) | optional | Name of StorageClass to which this persistent volume belongs. Empty value means that this volume does not belong to any StorageClass. &#43;optional |
| mountOptions | [string](#string) | repeated | A list of mount options, e.g. [&#34;ro&#34;, &#34;soft&#34;]. Not validated - mount will simply fail if one is invalid. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes/#mount-options &#43;optional |






<a name="k8s.io.api.core.v1.PersistentVolumeSpec.CapacityEntry"/>

### PersistentVolumeSpec.CapacityEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [.k8s.io.apimachinery.pkg.api.resource.Quantity](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.api.resource.Quantity) | optional |  |






<a name="k8s.io.api.core.v1.PersistentVolumeStatus"/>

### PersistentVolumeStatus
PersistentVolumeStatus is the current status of a persistent volume.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| phase | [string](#string) | optional | Phase indicates if a volume is available, bound to a claim, or released by a claim. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#phase &#43;optional |
| message | [string](#string) | optional | A human-readable message indicating details about why the volume is in this state. &#43;optional |
| reason | [string](#string) | optional | Reason is a brief CamelCase string that describes any failure and is meant for machine parsing and tidy display in the CLI. &#43;optional |






<a name="k8s.io.api.core.v1.PhotonPersistentDiskVolumeSource"/>

### PhotonPersistentDiskVolumeSource
Represents a Photon Controller persistent disk resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pdID | [string](#string) | optional | ID that identifies Photon Controller persistent disk |
| fsType | [string](#string) | optional | Filesystem type to mount. Must be a filesystem type supported by the host operating system. Ex. &#34;ext4&#34;, &#34;xfs&#34;, &#34;ntfs&#34;. Implicitly inferred to be &#34;ext4&#34; if unspecified. |






<a name="k8s.io.api.core.v1.Pod"/>

### Pod
Pod is a collection of containers that can run on a host. This resource is created
by clients and scheduled onto hosts.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta) | optional | Standard object&#39;s metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata &#43;optional |
| spec | [PodSpec](#k8s.io.api.core.v1.PodSpec) | optional | Specification of the desired behavior of the pod. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status &#43;optional |
| status | [PodStatus](#k8s.io.api.core.v1.PodStatus) | optional | Most recently observed status of the pod. This data may not be up to date. Populated by the system. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status &#43;optional |






<a name="k8s.io.api.core.v1.PodAffinity"/>

### PodAffinity
Pod affinity is a group of inter pod affinity scheduling rules.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| requiredDuringSchedulingIgnoredDuringExecution | [PodAffinityTerm](#k8s.io.api.core.v1.PodAffinityTerm) | repeated | If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied. &#43;optional |
| preferredDuringSchedulingIgnoredDuringExecution | [WeightedPodAffinityTerm](#k8s.io.api.core.v1.WeightedPodAffinityTerm) | repeated | The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding &#34;weight&#34; to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred. &#43;optional |






<a name="k8s.io.api.core.v1.PodAffinityTerm"/>

### PodAffinityTerm
Defines a set of pods (namely those matching the labelSelector
relative to the given namespace(s)) that this pod should be
co-located (affinity) or not co-located (anti-affinity) with,
where co-located is defined as running on a node whose value of
the label with key &lt;topologyKey&gt; matches that of any node on which
a pod of the set of pods is running


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| labelSelector | [.k8s.io.apimachinery.pkg.apis.meta.v1.LabelSelector](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.LabelSelector) | optional | A label query over a set of resources, in this case pods. &#43;optional |
| namespaces | [string](#string) | repeated | namespaces specifies which namespaces the labelSelector applies to (matches against); null or empty list means &#34;this pod&#39;s namespace&#34; |
| topologyKey | [string](#string) | optional | This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. For PreferredDuringScheduling pod anti-affinity, empty topologyKey is interpreted as &#34;all topologies&#34; (&#34;all topologies&#34; here means all the topologyKeys indicated by scheduler command-line argument --failure-domains); for affinity and for RequiredDuringScheduling pod anti-affinity, empty topologyKey is not allowed. &#43;optional |






<a name="k8s.io.api.core.v1.PodAntiAffinity"/>

### PodAntiAffinity
Pod anti affinity is a group of inter pod anti affinity scheduling rules.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| requiredDuringSchedulingIgnoredDuringExecution | [PodAffinityTerm](#k8s.io.api.core.v1.PodAffinityTerm) | repeated | If the anti-affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the anti-affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied. &#43;optional |
| preferredDuringSchedulingIgnoredDuringExecution | [WeightedPodAffinityTerm](#k8s.io.api.core.v1.WeightedPodAffinityTerm) | repeated | The scheduler will prefer to schedule pods to nodes that satisfy the anti-affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling anti-affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding &#34;weight&#34; to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred. &#43;optional |






<a name="k8s.io.api.core.v1.PodAttachOptions"/>

### PodAttachOptions
PodAttachOptions is the query options to a Pod&#39;s remote attach call.
---
TODO: merge w/ PodExecOptions below for stdin, stdout, etc
and also when we cut V2, we should export a &#34;StreamOptions&#34; or somesuch that contains Stdin, Stdout, Stder and TTY


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| stdin | [bool](#bool) | optional | Stdin if true, redirects the standard input stream of the pod for this call. Defaults to false. &#43;optional |
| stdout | [bool](#bool) | optional | Stdout if true indicates that stdout is to be redirected for the attach call. Defaults to true. &#43;optional |
| stderr | [bool](#bool) | optional | Stderr if true indicates that stderr is to be redirected for the attach call. Defaults to true. &#43;optional |
| tty | [bool](#bool) | optional | TTY if true indicates that a tty will be allocated for the attach call. This is passed through the container runtime so the tty is allocated on the worker node by the container runtime. Defaults to false. &#43;optional |
| container | [string](#string) | optional | The container in which to execute the command. Defaults to only container if there is only one container in the pod. &#43;optional |






<a name="k8s.io.api.core.v1.PodCondition"/>

### PodCondition
PodCondition contains details for the current condition of this pod.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [string](#string) | optional | Type is the type of the condition. Currently only Ready. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#pod-conditions |
| status | [string](#string) | optional | Status is the status of the condition. Can be True, False, Unknown. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#pod-conditions |
| lastProbeTime | [.k8s.io.apimachinery.pkg.apis.meta.v1.Time](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.Time) | optional | Last time we probed the condition. &#43;optional |
| lastTransitionTime | [.k8s.io.apimachinery.pkg.apis.meta.v1.Time](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.Time) | optional | Last time the condition transitioned from one status to another. &#43;optional |
| reason | [string](#string) | optional | Unique, one-word, CamelCase reason for the condition&#39;s last transition. &#43;optional |
| message | [string](#string) | optional | Human-readable message indicating details about last transition. &#43;optional |






<a name="k8s.io.api.core.v1.PodExecOptions"/>

### PodExecOptions
PodExecOptions is the query options to a Pod&#39;s remote exec call.
---
TODO: This is largely identical to PodAttachOptions above, make sure they stay in sync and see about merging
and also when we cut V2, we should export a &#34;StreamOptions&#34; or somesuch that contains Stdin, Stdout, Stder and TTY


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| stdin | [bool](#bool) | optional | Redirect the standard input stream of the pod for this call. Defaults to false. &#43;optional |
| stdout | [bool](#bool) | optional | Redirect the standard output stream of the pod for this call. Defaults to true. &#43;optional |
| stderr | [bool](#bool) | optional | Redirect the standard error stream of the pod for this call. Defaults to true. &#43;optional |
| tty | [bool](#bool) | optional | TTY if true indicates that a tty will be allocated for the exec call. Defaults to false. &#43;optional |
| container | [string](#string) | optional | Container in which to execute the command. Defaults to only container if there is only one container in the pod. &#43;optional |
| command | [string](#string) | repeated | Command is the remote command to execute. argv array. Not executed within a shell. |






<a name="k8s.io.api.core.v1.PodList"/>

### PodList
PodList is a list of Pods.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta) | optional | Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds &#43;optional |
| items | [Pod](#k8s.io.api.core.v1.Pod) | repeated | List of pods. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md |






<a name="k8s.io.api.core.v1.PodLogOptions"/>

### PodLogOptions
PodLogOptions is the query options for a Pod&#39;s logs REST call.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| container | [string](#string) | optional | The container for which to stream logs. Defaults to only container if there is one container in the pod. &#43;optional |
| follow | [bool](#bool) | optional | Follow the log stream of the pod. Defaults to false. &#43;optional |
| previous | [bool](#bool) | optional | Return previous terminated container logs. Defaults to false. &#43;optional |
| sinceSeconds | [int64](#int64) | optional | A relative time in seconds before the current time from which to show logs. If this value precedes the time a pod was started, only logs since the pod start will be returned. If this value is in the future, no logs will be returned. Only one of sinceSeconds or sinceTime may be specified. &#43;optional |
| sinceTime | [.k8s.io.apimachinery.pkg.apis.meta.v1.Time](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.Time) | optional | An RFC3339 timestamp from which to show logs. If this value precedes the time a pod was started, only logs since the pod start will be returned. If this value is in the future, no logs will be returned. Only one of sinceSeconds or sinceTime may be specified. &#43;optional |
| timestamps | [bool](#bool) | optional | If true, add an RFC3339 or RFC3339Nano timestamp at the beginning of every line of log output. Defaults to false. &#43;optional |
| tailLines | [int64](#int64) | optional | If set, the number of lines from the end of the logs to show. If not specified, logs are shown from the creation of the container or sinceSeconds or sinceTime &#43;optional |
| limitBytes | [int64](#int64) | optional | If set, the number of bytes to read from the server before terminating the log output. This may not display a complete final line of logging, and may return slightly more or slightly less than the specified limit. &#43;optional |






<a name="k8s.io.api.core.v1.PodPortForwardOptions"/>

### PodPortForwardOptions
PodPortForwardOptions is the query options to a Pod&#39;s port forward call
when using WebSockets.
The `port` query parameter must specify the port or
ports (comma separated) to forward over.
Port forwarding over SPDY does not use these options. It requires the port
to be passed in the `port` header as part of request.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ports | [int32](#int32) | repeated | List of ports to forward Required when using WebSockets &#43;optional |






<a name="k8s.io.api.core.v1.PodProxyOptions"/>

### PodProxyOptions
PodProxyOptions is the query options to a Pod&#39;s proxy call.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [string](#string) | optional | Path is the URL path to use for the current proxy request to pod. &#43;optional |






<a name="k8s.io.api.core.v1.PodSecurityContext"/>

### PodSecurityContext
PodSecurityContext holds pod-level security attributes and common container settings.
Some fields are also present in container.securityContext.  Field values of
container.securityContext take precedence over field values of PodSecurityContext.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| seLinuxOptions | [SELinuxOptions](#k8s.io.api.core.v1.SELinuxOptions) | optional | The SELinux context to be applied to all containers. If unspecified, the container runtime will allocate a random SELinux context for each container. May also be set in SecurityContext. If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence for that container. &#43;optional |
| runAsUser | [int64](#int64) | optional | The UID to run the entrypoint of the container process. Defaults to user specified in image metadata if unspecified. May also be set in SecurityContext. If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence for that container. &#43;optional |
| runAsNonRoot | [bool](#bool) | optional | Indicates that the container must run as a non-root user. If true, the Kubelet will validate the image at runtime to ensure that it does not run as UID 0 (root) and fail to start the container if it does. If unset or false, no such validation will be performed. May also be set in SecurityContext. If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence. &#43;optional |
| supplementalGroups | [int64](#int64) | repeated | A list of groups applied to the first process run in each container, in addition to the container&#39;s primary GID. If unspecified, no groups will be added to any container. &#43;optional |
| fsGroup | [int64](#int64) | optional | A special supplemental group that applies to all containers in a pod. Some volume types allow the Kubelet to change the ownership of that volume to be owned by the pod: 1. The owning GID will be the FSGroup 2. The setgid bit is set (new files created in the volume will be owned by FSGroup) 3. The permission bits are OR&#39;d with rw-rw---- If unset, the Kubelet will not modify the ownership and permissions of any volume. &#43;optional |






<a name="k8s.io.api.core.v1.PodSignature"/>

### PodSignature
Describes the class of pods that should avoid this node.
Exactly one field should be set.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| podController | [.k8s.io.apimachinery.pkg.apis.meta.v1.OwnerReference](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.OwnerReference) | optional | Reference to controller whose pods should avoid this node. &#43;optional |






<a name="k8s.io.api.core.v1.PodSpec"/>

### PodSpec
PodSpec is a description of a pod.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| volumes | [Volume](#k8s.io.api.core.v1.Volume) | repeated | List of volumes that can be mounted by containers belonging to the pod. More info: https://kubernetes.io/docs/concepts/storage/volumes &#43;optional &#43;patchMergeKey=name &#43;patchStrategy=merge,retainKeys |
| initContainers | [Container](#k8s.io.api.core.v1.Container) | repeated | List of initialization containers belonging to the pod. Init containers are executed in order prior to containers being started. If any init container fails, the pod is considered to have failed and is handled according to its restartPolicy. The name for an init container or normal container must be unique among all containers. Init containers may not have Lifecycle actions, Readiness probes, or Liveness probes. The resourceRequirements of an init container are taken into account during scheduling by finding the highest request/limit for each resource type, and then using the max of of that value or the sum of the normal containers. Limits are applied to init containers in a similar fashion. Init containers cannot currently be added or removed. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers &#43;patchMergeKey=name &#43;patchStrategy=merge |
| containers | [Container](#k8s.io.api.core.v1.Container) | repeated | List of containers belonging to the pod. Containers cannot currently be added or removed. There must be at least one container in a Pod. Cannot be updated. &#43;patchMergeKey=name &#43;patchStrategy=merge |
| restartPolicy | [string](#string) | optional | Restart policy for all containers within the pod. One of Always, OnFailure, Never. Default to Always. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#restart-policy &#43;optional |
| terminationGracePeriodSeconds | [int64](#int64) | optional | Optional duration in seconds the pod needs to terminate gracefully. May be decreased in delete request. Value must be non-negative integer. The value zero indicates delete immediately. If this value is nil, the default grace period will be used instead. The grace period is the duration in seconds after the processes running in the pod are sent a termination signal and the time when the processes are forcibly halted with a kill signal. Set this value longer than the expected cleanup time for your process. Defaults to 30 seconds. &#43;optional |
| activeDeadlineSeconds | [int64](#int64) | optional | Optional duration in seconds the pod may be active on the node relative to StartTime before the system will actively try to mark it failed and kill associated containers. Value must be a positive integer. &#43;optional |
| dnsPolicy | [string](#string) | optional | Set DNS policy for containers within the pod. One of &#39;ClusterFirstWithHostNet&#39;, &#39;ClusterFirst&#39; or &#39;Default&#39;. Defaults to &#34;ClusterFirst&#34;. To have DNS options set along with hostNetwork, you have to specify DNS policy explicitly to &#39;ClusterFirstWithHostNet&#39;. &#43;optional |
| nodeSelector | [PodSpec.NodeSelectorEntry](#k8s.io.api.core.v1.PodSpec.NodeSelectorEntry) | repeated | NodeSelector is a selector which must be true for the pod to fit on a node. Selector which must match a node&#39;s labels for the pod to be scheduled on that node. More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node &#43;optional |
| serviceAccountName | [string](#string) | optional | ServiceAccountName is the name of the ServiceAccount to use to run this pod. More info: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account &#43;optional |
| serviceAccount | [string](#string) | optional | DeprecatedServiceAccount is a depreciated alias for ServiceAccountName. Deprecated: Use serviceAccountName instead. &#43;k8s:conversion-gen=false &#43;optional |
| automountServiceAccountToken | [bool](#bool) | optional | AutomountServiceAccountToken indicates whether a service account token should be automatically mounted. &#43;optional |
| nodeName | [string](#string) | optional | NodeName is a request to schedule this pod onto a specific node. If it is non-empty, the scheduler simply schedules this pod onto that node, assuming that it fits resource requirements. &#43;optional |
| hostNetwork | [bool](#bool) | optional | Host networking requested for this pod. Use the host&#39;s network namespace. If this option is set, the ports that will be used must be specified. Default to false. &#43;k8s:conversion-gen=false &#43;optional |
| hostPID | [bool](#bool) | optional | Use the host&#39;s pid namespace. Optional: Default to false. &#43;k8s:conversion-gen=false &#43;optional |
| hostIPC | [bool](#bool) | optional | Use the host&#39;s ipc namespace. Optional: Default to false. &#43;k8s:conversion-gen=false &#43;optional |
| securityContext | [PodSecurityContext](#k8s.io.api.core.v1.PodSecurityContext) | optional | SecurityContext holds pod-level security attributes and common container settings. Optional: Defaults to empty. See type description for default values of each field. &#43;optional |
| imagePullSecrets | [LocalObjectReference](#k8s.io.api.core.v1.LocalObjectReference) | repeated | ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec. If specified, these secrets will be passed to individual puller implementations for them to use. For example, in the case of docker, only DockerConfig type secrets are honored. More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod &#43;optional &#43;patchMergeKey=name &#43;patchStrategy=merge |
| hostname | [string](#string) | optional | Specifies the hostname of the Pod If not specified, the pod&#39;s hostname will be set to a system-defined value. &#43;optional |
| subdomain | [string](#string) | optional | If specified, the fully qualified Pod hostname will be &#34;&lt;hostname&gt;.&lt;subdomain&gt;.&lt;pod namespace&gt;.svc.&lt;cluster domain&gt;&#34;. If not specified, the pod will not have a domainname at all. &#43;optional |
| affinity | [Affinity](#k8s.io.api.core.v1.Affinity) | optional | If specified, the pod&#39;s scheduling constraints &#43;optional |
| schedulerName | [string](#string) | optional | If specified, the pod will be dispatched by specified scheduler. If not specified, the pod will be dispatched by default scheduler. &#43;optional |
| tolerations | [Toleration](#k8s.io.api.core.v1.Toleration) | repeated | If specified, the pod&#39;s tolerations. &#43;optional |
| hostAliases | [HostAlias](#k8s.io.api.core.v1.HostAlias) | repeated | HostAliases is an optional list of hosts and IPs that will be injected into the pod&#39;s hosts file if specified. This is only valid for non-hostNetwork pods. &#43;optional &#43;patchMergeKey=ip &#43;patchStrategy=merge |
| priorityClassName | [string](#string) | optional | If specified, indicates the pod&#39;s priority. &#34;SYSTEM&#34; is a special keyword which indicates the highest priority. Any other name must be defined by creating a PriorityClass object with that name. If not specified, the pod priority will be default or zero if there is no default. &#43;optional |
| priority | [int32](#int32) | optional | The priority value. Various system components use this field to find the priority of the pod. When Priority Admission Controller is enabled, it prevents users from setting this field. The admission controller populates this field from PriorityClassName. The higher the value, the higher the priority. &#43;optional |






<a name="k8s.io.api.core.v1.PodSpec.NodeSelectorEntry"/>

### PodSpec.NodeSelectorEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [string](#string) | optional |  |






<a name="k8s.io.api.core.v1.PodStatus"/>

### PodStatus
PodStatus represents information about the status of a pod. Status may trail the actual
state of a system.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| phase | [string](#string) | optional | Current condition of the pod. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#pod-phase &#43;optional |
| conditions | [PodCondition](#k8s.io.api.core.v1.PodCondition) | repeated | Current service state of pod. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#pod-conditions &#43;optional &#43;patchMergeKey=type &#43;patchStrategy=merge |
| message | [string](#string) | optional | A human readable message indicating details about why the pod is in this condition. &#43;optional |
| reason | [string](#string) | optional | A brief CamelCase message indicating details about why the pod is in this state. e.g. &#39;Evicted&#39; &#43;optional |
| hostIP | [string](#string) | optional | IP address of the host to which the pod is assigned. Empty if not yet scheduled. &#43;optional |
| podIP | [string](#string) | optional | IP address allocated to the pod. Routable at least within the cluster. Empty if not yet allocated. &#43;optional |
| startTime | [.k8s.io.apimachinery.pkg.apis.meta.v1.Time](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.Time) | optional | RFC 3339 date and time at which the object was acknowledged by the Kubelet. This is before the Kubelet pulled the container image(s) for the pod. &#43;optional |
| initContainerStatuses | [ContainerStatus](#k8s.io.api.core.v1.ContainerStatus) | repeated | The list has one entry per init container in the manifest. The most recent successful init container will have ready = true, the most recently started container will have startTime set. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#pod-and-container-status |
| containerStatuses | [ContainerStatus](#k8s.io.api.core.v1.ContainerStatus) | repeated | The list has one entry per container in the manifest. Each entry is currently the output of `docker inspect`. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#pod-and-container-status &#43;optional |
| qosClass | [string](#string) | optional | The Quality of Service (QOS) classification assigned to the pod based on resource requirements See PodQOSClass type for available QOS classes More info: https://github.com/kubernetes/kubernetes/blob/master/docs/design/resource-qos.md &#43;optional |






<a name="k8s.io.api.core.v1.PodStatusResult"/>

### PodStatusResult
PodStatusResult is a wrapper for PodStatus returned by kubelet that can be encode/decoded


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta) | optional | Standard object&#39;s metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata &#43;optional |
| status | [PodStatus](#k8s.io.api.core.v1.PodStatus) | optional | Most recently observed status of the pod. This data may not be up to date. Populated by the system. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status &#43;optional |






<a name="k8s.io.api.core.v1.PodTemplate"/>

### PodTemplate
PodTemplate describes a template for creating copies of a predefined pod.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta) | optional | Standard object&#39;s metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata &#43;optional |
| template | [PodTemplateSpec](#k8s.io.api.core.v1.PodTemplateSpec) | optional | Template defines the pods that will be created from this pod template. https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status &#43;optional |






<a name="k8s.io.api.core.v1.PodTemplateList"/>

### PodTemplateList
PodTemplateList is a list of PodTemplates.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta) | optional | Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds &#43;optional |
| items | [PodTemplate](#k8s.io.api.core.v1.PodTemplate) | repeated | List of pod templates |






<a name="k8s.io.api.core.v1.PodTemplateSpec"/>

### PodTemplateSpec
PodTemplateSpec describes the data a pod should have when created from a template


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta) | optional | Standard object&#39;s metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata &#43;optional |
| spec | [PodSpec](#k8s.io.api.core.v1.PodSpec) | optional | Specification of the desired behavior of the pod. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status &#43;optional |






<a name="k8s.io.api.core.v1.PortworxVolumeSource"/>

### PortworxVolumeSource
PortworxVolumeSource represents a Portworx volume resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| volumeID | [string](#string) | optional | VolumeID uniquely identifies a Portworx volume |
| fsType | [string](#string) | optional | FSType represents the filesystem type to mount Must be a filesystem type supported by the host operating system. Ex. &#34;ext4&#34;, &#34;xfs&#34;. Implicitly inferred to be &#34;ext4&#34; if unspecified. |
| readOnly | [bool](#bool) | optional | Defaults to false (read/write). ReadOnly here will force the ReadOnly setting in VolumeMounts. &#43;optional |






<a name="k8s.io.api.core.v1.Preconditions"/>

### Preconditions
Preconditions must be fulfilled before an operation (update, delete, etc.) is carried out.
&#43;k8s:openapi-gen=false


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| uid | [string](#string) | optional | Specifies the target UID. &#43;optional |






<a name="k8s.io.api.core.v1.PreferAvoidPodsEntry"/>

### PreferAvoidPodsEntry
Describes a class of pods that should avoid this node.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| podSignature | [PodSignature](#k8s.io.api.core.v1.PodSignature) | optional | The class of pods. |
| evictionTime | [.k8s.io.apimachinery.pkg.apis.meta.v1.Time](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.Time) | optional | Time at which this entry was added to the list. &#43;optional |
| reason | [string](#string) | optional | (brief) reason why this entry was added to the list. &#43;optional |
| message | [string](#string) | optional | Human readable message indicating why this entry was added to the list. &#43;optional |






<a name="k8s.io.api.core.v1.PreferredSchedulingTerm"/>

### PreferredSchedulingTerm
An empty preferred scheduling term matches all objects with implicit weight 0
(i.e. it&#39;s a no-op). A null preferred scheduling term matches no objects (i.e. is also a no-op).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| weight | [int32](#int32) | optional | Weight associated with matching the corresponding nodeSelectorTerm, in the range 1-100. |
| preference | [NodeSelectorTerm](#k8s.io.api.core.v1.NodeSelectorTerm) | optional | A node selector term, associated with the corresponding weight. |






<a name="k8s.io.api.core.v1.Probe"/>

### Probe
Probe describes a health check to be performed against a container to determine whether it is
alive or ready to receive traffic.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| handler | [Handler](#k8s.io.api.core.v1.Handler) | optional | The action taken to determine the health of a container |
| initialDelaySeconds | [int32](#int32) | optional | Number of seconds after the container has started before liveness probes are initiated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes &#43;optional |
| timeoutSeconds | [int32](#int32) | optional | Number of seconds after which the probe times out. Defaults to 1 second. Minimum value is 1. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes &#43;optional |
| periodSeconds | [int32](#int32) | optional | How often (in seconds) to perform the probe. Default to 10 seconds. Minimum value is 1. &#43;optional |
| successThreshold | [int32](#int32) | optional | Minimum consecutive successes for the probe to be considered successful after having failed. Defaults to 1. Must be 1 for liveness. Minimum value is 1. &#43;optional |
| failureThreshold | [int32](#int32) | optional | Minimum consecutive failures for the probe to be considered failed after having succeeded. Defaults to 3. Minimum value is 1. &#43;optional |






<a name="k8s.io.api.core.v1.ProjectedVolumeSource"/>

### ProjectedVolumeSource
Represents a projected volume source


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sources | [VolumeProjection](#k8s.io.api.core.v1.VolumeProjection) | repeated | list of volume projections |
| defaultMode | [int32](#int32) | optional | Mode bits to use on created files by default. Must be a value between 0 and 0777. Directories within the path are not affected by this setting. This might be in conflict with other options that affect the file mode, like fsGroup, and the result can be other mode bits set. &#43;optional |






<a name="k8s.io.api.core.v1.QuobyteVolumeSource"/>

### QuobyteVolumeSource
Represents a Quobyte mount that lasts the lifetime of a pod.
Quobyte volumes do not support ownership management or SELinux relabeling.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| registry | [string](#string) | optional | Registry represents a single or multiple Quobyte Registry services specified as a string as host:port pair (multiple entries are separated with commas) which acts as the central registry for volumes |
| volume | [string](#string) | optional | Volume is a string that references an already created Quobyte volume by name. |
| readOnly | [bool](#bool) | optional | ReadOnly here will force the Quobyte volume to be mounted with read-only permissions. Defaults to false. &#43;optional |
| user | [string](#string) | optional | User to map volume access to Defaults to serivceaccount user &#43;optional |
| group | [string](#string) | optional | Group to map volume access to Default is no group &#43;optional |






<a name="k8s.io.api.core.v1.RBDPersistentVolumeSource"/>

### RBDPersistentVolumeSource
Represents a Rados Block Device mount that lasts the lifetime of a pod.
RBD volumes support ownership management and SELinux relabeling.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| monitors | [string](#string) | repeated | A collection of Ceph monitors. More info: https://releases.k8s.io/HEAD/examples/volumes/rbd/README.md#how-to-use-it |
| image | [string](#string) | optional | The rados image name. More info: https://releases.k8s.io/HEAD/examples/volumes/rbd/README.md#how-to-use-it |
| fsType | [string](#string) | optional | Filesystem type of the volume that you want to mount. Tip: Ensure that the filesystem type is supported by the host operating system. Examples: &#34;ext4&#34;, &#34;xfs&#34;, &#34;ntfs&#34;. Implicitly inferred to be &#34;ext4&#34; if unspecified. More info: https://kubernetes.io/docs/concepts/storage/volumes#rbd TODO: how do we prevent errors in the filesystem from compromising the machine &#43;optional |
| pool | [string](#string) | optional | The rados pool name. Default is rbd. More info: https://releases.k8s.io/HEAD/examples/volumes/rbd/README.md#how-to-use-it &#43;optional |
| user | [string](#string) | optional | The rados user name. Default is admin. More info: https://releases.k8s.io/HEAD/examples/volumes/rbd/README.md#how-to-use-it &#43;optional |
| keyring | [string](#string) | optional | Keyring is the path to key ring for RBDUser. Default is /etc/ceph/keyring. More info: https://releases.k8s.io/HEAD/examples/volumes/rbd/README.md#how-to-use-it &#43;optional |
| secretRef | [SecretReference](#k8s.io.api.core.v1.SecretReference) | optional | SecretRef is name of the authentication secret for RBDUser. If provided overrides keyring. Default is nil. More info: https://releases.k8s.io/HEAD/examples/volumes/rbd/README.md#how-to-use-it &#43;optional |
| readOnly | [bool](#bool) | optional | ReadOnly here will force the ReadOnly setting in VolumeMounts. Defaults to false. More info: https://releases.k8s.io/HEAD/examples/volumes/rbd/README.md#how-to-use-it &#43;optional |






<a name="k8s.io.api.core.v1.RBDVolumeSource"/>

### RBDVolumeSource
Represents a Rados Block Device mount that lasts the lifetime of a pod.
RBD volumes support ownership management and SELinux relabeling.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| monitors | [string](#string) | repeated | A collection of Ceph monitors. More info: https://releases.k8s.io/HEAD/examples/volumes/rbd/README.md#how-to-use-it |
| image | [string](#string) | optional | The rados image name. More info: https://releases.k8s.io/HEAD/examples/volumes/rbd/README.md#how-to-use-it |
| fsType | [string](#string) | optional | Filesystem type of the volume that you want to mount. Tip: Ensure that the filesystem type is supported by the host operating system. Examples: &#34;ext4&#34;, &#34;xfs&#34;, &#34;ntfs&#34;. Implicitly inferred to be &#34;ext4&#34; if unspecified. More info: https://kubernetes.io/docs/concepts/storage/volumes#rbd TODO: how do we prevent errors in the filesystem from compromising the machine &#43;optional |
| pool | [string](#string) | optional | The rados pool name. Default is rbd. More info: https://releases.k8s.io/HEAD/examples/volumes/rbd/README.md#how-to-use-it &#43;optional |
| user | [string](#string) | optional | The rados user name. Default is admin. More info: https://releases.k8s.io/HEAD/examples/volumes/rbd/README.md#how-to-use-it &#43;optional |
| keyring | [string](#string) | optional | Keyring is the path to key ring for RBDUser. Default is /etc/ceph/keyring. More info: https://releases.k8s.io/HEAD/examples/volumes/rbd/README.md#how-to-use-it &#43;optional |
| secretRef | [LocalObjectReference](#k8s.io.api.core.v1.LocalObjectReference) | optional | SecretRef is name of the authentication secret for RBDUser. If provided overrides keyring. Default is nil. More info: https://releases.k8s.io/HEAD/examples/volumes/rbd/README.md#how-to-use-it &#43;optional |
| readOnly | [bool](#bool) | optional | ReadOnly here will force the ReadOnly setting in VolumeMounts. Defaults to false. More info: https://releases.k8s.io/HEAD/examples/volumes/rbd/README.md#how-to-use-it &#43;optional |






<a name="k8s.io.api.core.v1.RangeAllocation"/>

### RangeAllocation
RangeAllocation is not a public type.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta) | optional | Standard object&#39;s metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata &#43;optional |
| range | [string](#string) | optional | Range is string that identifies the range represented by &#39;data&#39;. |
| data | [bytes](#bytes) | optional | Data is a bit array containing all allocated addresses in the previous segment. |






<a name="k8s.io.api.core.v1.ReplicationController"/>

### ReplicationController
ReplicationController represents the configuration of a replication controller.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta) | optional | If the Labels of a ReplicationController are empty, they are defaulted to be the same as the Pod(s) that the replication controller manages. Standard object&#39;s metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata &#43;optional |
| spec | [ReplicationControllerSpec](#k8s.io.api.core.v1.ReplicationControllerSpec) | optional | Spec defines the specification of the desired behavior of the replication controller. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status &#43;optional |
| status | [ReplicationControllerStatus](#k8s.io.api.core.v1.ReplicationControllerStatus) | optional | Status is the most recently observed status of the replication controller. This data may be out of date by some window of time. Populated by the system. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status &#43;optional |






<a name="k8s.io.api.core.v1.ReplicationControllerCondition"/>

### ReplicationControllerCondition
ReplicationControllerCondition describes the state of a replication controller at a certain point.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [string](#string) | optional | Type of replication controller condition. |
| status | [string](#string) | optional | Status of the condition, one of True, False, Unknown. |
| lastTransitionTime | [.k8s.io.apimachinery.pkg.apis.meta.v1.Time](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.Time) | optional | The last time the condition transitioned from one status to another. &#43;optional |
| reason | [string](#string) | optional | The reason for the condition&#39;s last transition. &#43;optional |
| message | [string](#string) | optional | A human readable message indicating details about the transition. &#43;optional |






<a name="k8s.io.api.core.v1.ReplicationControllerList"/>

### ReplicationControllerList
ReplicationControllerList is a collection of replication controllers.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta) | optional | Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds &#43;optional |
| items | [ReplicationController](#k8s.io.api.core.v1.ReplicationController) | repeated | List of replication controllers. More info: https://kubernetes.io/docs/concepts/workloads/controllers/replicationcontroller |






<a name="k8s.io.api.core.v1.ReplicationControllerSpec"/>

### ReplicationControllerSpec
ReplicationControllerSpec is the specification of a replication controller.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| replicas | [int32](#int32) | optional | Replicas is the number of desired replicas. This is a pointer to distinguish between explicit zero and unspecified. Defaults to 1. More info: https://kubernetes.io/docs/concepts/workloads/controllers/replicationcontroller#what-is-a-replicationcontroller &#43;optional |
| minReadySeconds | [int32](#int32) | optional | Minimum number of seconds for which a newly created pod should be ready without any of its container crashing, for it to be considered available. Defaults to 0 (pod will be considered available as soon as it is ready) &#43;optional |
| selector | [ReplicationControllerSpec.SelectorEntry](#k8s.io.api.core.v1.ReplicationControllerSpec.SelectorEntry) | repeated | Selector is a label query over pods that should match the Replicas count. If Selector is empty, it is defaulted to the labels present on the Pod template. Label keys and values that must match in order to be controlled by this replication controller, if empty defaulted to labels on Pod template. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors &#43;optional |
| template | [PodTemplateSpec](#k8s.io.api.core.v1.PodTemplateSpec) | optional | Template is the object that describes the pod that will be created if insufficient replicas are detected. This takes precedence over a TemplateRef. More info: https://kubernetes.io/docs/concepts/workloads/controllers/replicationcontroller#pod-template &#43;optional |






<a name="k8s.io.api.core.v1.ReplicationControllerSpec.SelectorEntry"/>

### ReplicationControllerSpec.SelectorEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [string](#string) | optional |  |






<a name="k8s.io.api.core.v1.ReplicationControllerStatus"/>

### ReplicationControllerStatus
ReplicationControllerStatus represents the current status of a replication
controller.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| replicas | [int32](#int32) | optional | Replicas is the most recently oberved number of replicas. More info: https://kubernetes.io/docs/concepts/workloads/controllers/replicationcontroller#what-is-a-replicationcontroller |
| fullyLabeledReplicas | [int32](#int32) | optional | The number of pods that have labels matching the labels of the pod template of the replication controller. &#43;optional |
| readyReplicas | [int32](#int32) | optional | The number of ready replicas for this replication controller. &#43;optional |
| availableReplicas | [int32](#int32) | optional | The number of available replicas (ready for at least minReadySeconds) for this replication controller. &#43;optional |
| observedGeneration | [int64](#int64) | optional | ObservedGeneration reflects the generation of the most recently observed replication controller. &#43;optional |
| conditions | [ReplicationControllerCondition](#k8s.io.api.core.v1.ReplicationControllerCondition) | repeated | Represents the latest available observations of a replication controller&#39;s current state. &#43;optional &#43;patchMergeKey=type &#43;patchStrategy=merge |






<a name="k8s.io.api.core.v1.ResourceFieldSelector"/>

### ResourceFieldSelector
ResourceFieldSelector represents container resources (cpu, memory) and their output format


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| containerName | [string](#string) | optional | Container name: required for volumes, optional for env vars &#43;optional |
| resource | [string](#string) | optional | Required: resource to select |
| divisor | [.k8s.io.apimachinery.pkg.api.resource.Quantity](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.api.resource.Quantity) | optional | Specifies the output format of the exposed resources, defaults to &#34;1&#34; &#43;optional |






<a name="k8s.io.api.core.v1.ResourceQuota"/>

### ResourceQuota
ResourceQuota sets aggregate quota restrictions enforced per namespace


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta) | optional | Standard object&#39;s metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata &#43;optional |
| spec | [ResourceQuotaSpec](#k8s.io.api.core.v1.ResourceQuotaSpec) | optional | Spec defines the desired quota. https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status &#43;optional |
| status | [ResourceQuotaStatus](#k8s.io.api.core.v1.ResourceQuotaStatus) | optional | Status defines the actual enforced quota and its current usage. https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status &#43;optional |






<a name="k8s.io.api.core.v1.ResourceQuotaList"/>

### ResourceQuotaList
ResourceQuotaList is a list of ResourceQuota items.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta) | optional | Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds &#43;optional |
| items | [ResourceQuota](#k8s.io.api.core.v1.ResourceQuota) | repeated | Items is a list of ResourceQuota objects. More info: https://kubernetes.io/docs/concepts/policy/resource-quotas |






<a name="k8s.io.api.core.v1.ResourceQuotaSpec"/>

### ResourceQuotaSpec
ResourceQuotaSpec defines the desired hard limits to enforce for Quota.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hard | [ResourceQuotaSpec.HardEntry](#k8s.io.api.core.v1.ResourceQuotaSpec.HardEntry) | repeated | Hard is the set of desired hard limits for each named resource. More info: https://kubernetes.io/docs/concepts/policy/resource-quotas &#43;optional |
| scopes | [string](#string) | repeated | A collection of filters that must match each object tracked by a quota. If not specified, the quota matches all objects. &#43;optional |






<a name="k8s.io.api.core.v1.ResourceQuotaSpec.HardEntry"/>

### ResourceQuotaSpec.HardEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [.k8s.io.apimachinery.pkg.api.resource.Quantity](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.api.resource.Quantity) | optional |  |






<a name="k8s.io.api.core.v1.ResourceQuotaStatus"/>

### ResourceQuotaStatus
ResourceQuotaStatus defines the enforced hard limits and observed use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hard | [ResourceQuotaStatus.HardEntry](#k8s.io.api.core.v1.ResourceQuotaStatus.HardEntry) | repeated | Hard is the set of enforced hard limits for each named resource. More info: https://kubernetes.io/docs/concepts/policy/resource-quotas &#43;optional |
| used | [ResourceQuotaStatus.UsedEntry](#k8s.io.api.core.v1.ResourceQuotaStatus.UsedEntry) | repeated | Used is the current observed total usage of the resource in the namespace. &#43;optional |






<a name="k8s.io.api.core.v1.ResourceQuotaStatus.HardEntry"/>

### ResourceQuotaStatus.HardEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [.k8s.io.apimachinery.pkg.api.resource.Quantity](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.api.resource.Quantity) | optional |  |






<a name="k8s.io.api.core.v1.ResourceQuotaStatus.UsedEntry"/>

### ResourceQuotaStatus.UsedEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [.k8s.io.apimachinery.pkg.api.resource.Quantity](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.api.resource.Quantity) | optional |  |






<a name="k8s.io.api.core.v1.ResourceRequirements"/>

### ResourceRequirements
ResourceRequirements describes the compute resource requirements.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| limits | [ResourceRequirements.LimitsEntry](#k8s.io.api.core.v1.ResourceRequirements.LimitsEntry) | repeated | Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container &#43;optional |
| requests | [ResourceRequirements.RequestsEntry](#k8s.io.api.core.v1.ResourceRequirements.RequestsEntry) | repeated | Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container &#43;optional |






<a name="k8s.io.api.core.v1.ResourceRequirements.LimitsEntry"/>

### ResourceRequirements.LimitsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [.k8s.io.apimachinery.pkg.api.resource.Quantity](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.api.resource.Quantity) | optional |  |






<a name="k8s.io.api.core.v1.ResourceRequirements.RequestsEntry"/>

### ResourceRequirements.RequestsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [.k8s.io.apimachinery.pkg.api.resource.Quantity](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.api.resource.Quantity) | optional |  |






<a name="k8s.io.api.core.v1.SELinuxOptions"/>

### SELinuxOptions
SELinuxOptions are the labels to be applied to the container


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user | [string](#string) | optional | User is a SELinux user label that applies to the container. &#43;optional |
| role | [string](#string) | optional | Role is a SELinux role label that applies to the container. &#43;optional |
| type | [string](#string) | optional | Type is a SELinux type label that applies to the container. &#43;optional |
| level | [string](#string) | optional | Level is SELinux level label that applies to the container. &#43;optional |






<a name="k8s.io.api.core.v1.ScaleIOPersistentVolumeSource"/>

### ScaleIOPersistentVolumeSource
ScaleIOPersistentVolumeSource represents a persistent ScaleIO volume


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| gateway | [string](#string) | optional | The host address of the ScaleIO API Gateway. |
| system | [string](#string) | optional | The name of the storage system as configured in ScaleIO. |
| secretRef | [SecretReference](#k8s.io.api.core.v1.SecretReference) | optional | SecretRef references to the secret for ScaleIO user and other sensitive information. If this is not provided, Login operation will fail. |
| sslEnabled | [bool](#bool) | optional | Flag to enable/disable SSL communication with Gateway, default false &#43;optional |
| protectionDomain | [string](#string) | optional | The name of the ScaleIO Protection Domain for the configured storage. &#43;optional |
| storagePool | [string](#string) | optional | The ScaleIO Storage Pool associated with the protection domain. &#43;optional |
| storageMode | [string](#string) | optional | Indicates whether the storage for a volume should be ThickProvisioned or ThinProvisioned. &#43;optional |
| volumeName | [string](#string) | optional | The name of a volume already created in the ScaleIO system that is associated with this volume source. |
| fsType | [string](#string) | optional | Filesystem type to mount. Must be a filesystem type supported by the host operating system. Ex. &#34;ext4&#34;, &#34;xfs&#34;, &#34;ntfs&#34;. Implicitly inferred to be &#34;ext4&#34; if unspecified. &#43;optional |
| readOnly | [bool](#bool) | optional | Defaults to false (read/write). ReadOnly here will force the ReadOnly setting in VolumeMounts. &#43;optional |






<a name="k8s.io.api.core.v1.ScaleIOVolumeSource"/>

### ScaleIOVolumeSource
ScaleIOVolumeSource represents a persistent ScaleIO volume


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| gateway | [string](#string) | optional | The host address of the ScaleIO API Gateway. |
| system | [string](#string) | optional | The name of the storage system as configured in ScaleIO. |
| secretRef | [LocalObjectReference](#k8s.io.api.core.v1.LocalObjectReference) | optional | SecretRef references to the secret for ScaleIO user and other sensitive information. If this is not provided, Login operation will fail. |
| sslEnabled | [bool](#bool) | optional | Flag to enable/disable SSL communication with Gateway, default false &#43;optional |
| protectionDomain | [string](#string) | optional | The name of the ScaleIO Protection Domain for the configured storage. &#43;optional |
| storagePool | [string](#string) | optional | The ScaleIO Storage Pool associated with the protection domain. &#43;optional |
| storageMode | [string](#string) | optional | Indicates whether the storage for a volume should be ThickProvisioned or ThinProvisioned. &#43;optional |
| volumeName | [string](#string) | optional | The name of a volume already created in the ScaleIO system that is associated with this volume source. |
| fsType | [string](#string) | optional | Filesystem type to mount. Must be a filesystem type supported by the host operating system. Ex. &#34;ext4&#34;, &#34;xfs&#34;, &#34;ntfs&#34;. Implicitly inferred to be &#34;ext4&#34; if unspecified. &#43;optional |
| readOnly | [bool](#bool) | optional | Defaults to false (read/write). ReadOnly here will force the ReadOnly setting in VolumeMounts. &#43;optional |






<a name="k8s.io.api.core.v1.Secret"/>

### Secret
Secret holds secret data of a certain type. The total bytes of the values in
the Data field must be less than MaxSecretSize bytes.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta) | optional | Standard object&#39;s metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata &#43;optional |
| data | [Secret.DataEntry](#k8s.io.api.core.v1.Secret.DataEntry) | repeated | Data contains the secret data. Each key must consist of alphanumeric characters, &#39;-&#39;, &#39;_&#39; or &#39;.&#39;. The serialized form of the secret data is a base64 encoded string, representing the arbitrary (possibly non-string) data value here. Described in https://tools.ietf.org/html/rfc4648#section-4 &#43;optional |
| stringData | [Secret.StringDataEntry](#k8s.io.api.core.v1.Secret.StringDataEntry) | repeated | stringData allows specifying non-binary secret data in string form. It is provided as a write-only convenience method. All keys and values are merged into the data field on write, overwriting any existing values. It is never output when reading from the API. &#43;k8s:conversion-gen=false &#43;optional |
| type | [string](#string) | optional | Used to facilitate programmatic handling of secret data. &#43;optional |






<a name="k8s.io.api.core.v1.Secret.DataEntry"/>

### Secret.DataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [bytes](#bytes) | optional |  |






<a name="k8s.io.api.core.v1.Secret.StringDataEntry"/>

### Secret.StringDataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [string](#string) | optional |  |






<a name="k8s.io.api.core.v1.SecretEnvSource"/>

### SecretEnvSource
SecretEnvSource selects a Secret to populate the environment
variables with.

The contents of the target Secret&#39;s Data field will represent the
key-value pairs as environment variables.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| localObjectReference | [LocalObjectReference](#k8s.io.api.core.v1.LocalObjectReference) | optional | The Secret to select from. |
| optional | [bool](#bool) | optional | Specify whether the Secret must be defined &#43;optional |






<a name="k8s.io.api.core.v1.SecretKeySelector"/>

### SecretKeySelector
SecretKeySelector selects a key of a Secret.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| localObjectReference | [LocalObjectReference](#k8s.io.api.core.v1.LocalObjectReference) | optional | The name of the secret in the pod&#39;s namespace to select from. |
| key | [string](#string) | optional | The key of the secret to select from. Must be a valid secret key. |
| optional | [bool](#bool) | optional | Specify whether the Secret or it&#39;s key must be defined &#43;optional |






<a name="k8s.io.api.core.v1.SecretList"/>

### SecretList
SecretList is a list of Secret.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta) | optional | Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds &#43;optional |
| items | [Secret](#k8s.io.api.core.v1.Secret) | repeated | Items is a list of secret objects. More info: https://kubernetes.io/docs/concepts/configuration/secret |






<a name="k8s.io.api.core.v1.SecretProjection"/>

### SecretProjection
Adapts a secret into a projected volume.

The contents of the target Secret&#39;s Data field will be presented in a
projected volume as files using the keys in the Data field as the file names.
Note that this is identical to a secret volume source without the default
mode.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| localObjectReference | [LocalObjectReference](#k8s.io.api.core.v1.LocalObjectReference) | optional |  |
| items | [KeyToPath](#k8s.io.api.core.v1.KeyToPath) | repeated | If unspecified, each key-value pair in the Data field of the referenced Secret will be projected into the volume as a file whose name is the key and content is the value. If specified, the listed keys will be projected into the specified paths, and unlisted keys will not be present. If a key is specified which is not present in the Secret, the volume setup will error unless it is marked optional. Paths must be relative and may not contain the &#39;..&#39; path or start with &#39;..&#39;. &#43;optional |
| optional | [bool](#bool) | optional | Specify whether the Secret or its key must be defined &#43;optional |






<a name="k8s.io.api.core.v1.SecretReference"/>

### SecretReference
SecretReference represents a Secret Reference. It has enough information to retrieve secret
in any namespace


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | optional | Name is unique within a namespace to reference a secret resource. &#43;optional |
| namespace | [string](#string) | optional | Namespace defines the space within which the secret name must be unique. &#43;optional |






<a name="k8s.io.api.core.v1.SecretVolumeSource"/>

### SecretVolumeSource
Adapts a Secret into a volume.

The contents of the target Secret&#39;s Data field will be presented in a volume
as files using the keys in the Data field as the file names.
Secret volumes support ownership management and SELinux relabeling.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| secretName | [string](#string) | optional | Name of the secret in the pod&#39;s namespace to use. More info: https://kubernetes.io/docs/concepts/storage/volumes#secret &#43;optional |
| items | [KeyToPath](#k8s.io.api.core.v1.KeyToPath) | repeated | If unspecified, each key-value pair in the Data field of the referenced Secret will be projected into the volume as a file whose name is the key and content is the value. If specified, the listed keys will be projected into the specified paths, and unlisted keys will not be present. If a key is specified which is not present in the Secret, the volume setup will error unless it is marked optional. Paths must be relative and may not contain the &#39;..&#39; path or start with &#39;..&#39;. &#43;optional |
| defaultMode | [int32](#int32) | optional | Optional: mode bits to use on created files by default. Must be a value between 0 and 0777. Defaults to 0644. Directories within the path are not affected by this setting. This might be in conflict with other options that affect the file mode, like fsGroup, and the result can be other mode bits set. &#43;optional |
| optional | [bool](#bool) | optional | Specify whether the Secret or it&#39;s keys must be defined &#43;optional |






<a name="k8s.io.api.core.v1.SecurityContext"/>

### SecurityContext
SecurityContext holds security configuration that will be applied to a container.
Some fields are present in both SecurityContext and PodSecurityContext.  When both
are set, the values in SecurityContext take precedence.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capabilities | [Capabilities](#k8s.io.api.core.v1.Capabilities) | optional | The capabilities to add/drop when running containers. Defaults to the default set of capabilities granted by the container runtime. &#43;optional |
| privileged | [bool](#bool) | optional | Run container in privileged mode. Processes in privileged containers are essentially equivalent to root on the host. Defaults to false. &#43;optional |
| seLinuxOptions | [SELinuxOptions](#k8s.io.api.core.v1.SELinuxOptions) | optional | The SELinux context to be applied to the container. If unspecified, the container runtime will allocate a random SELinux context for each container. May also be set in PodSecurityContext. If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence. &#43;optional |
| runAsUser | [int64](#int64) | optional | The UID to run the entrypoint of the container process. Defaults to user specified in image metadata if unspecified. May also be set in PodSecurityContext. If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence. &#43;optional |
| runAsNonRoot | [bool](#bool) | optional | Indicates that the container must run as a non-root user. If true, the Kubelet will validate the image at runtime to ensure that it does not run as UID 0 (root) and fail to start the container if it does. If unset or false, no such validation will be performed. May also be set in PodSecurityContext. If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence. &#43;optional |
| readOnlyRootFilesystem | [bool](#bool) | optional | Whether this container has a read-only root filesystem. Default is false. &#43;optional |
| allowPrivilegeEscalation | [bool](#bool) | optional | AllowPrivilegeEscalation controls whether a process can gain more privileges than its parent process. This bool directly controls if the no_new_privs flag will be set on the container process. AllowPrivilegeEscalation is true always when the container is: 1) run as Privileged 2) has CAP_SYS_ADMIN &#43;optional |






<a name="k8s.io.api.core.v1.SerializedReference"/>

### SerializedReference
SerializedReference is a reference to serialized object.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| reference | [ObjectReference](#k8s.io.api.core.v1.ObjectReference) | optional | The reference to an object in the system. &#43;optional |






<a name="k8s.io.api.core.v1.Service"/>

### Service
Service is a named abstraction of software service (for example, mysql) consisting of local port
(for example 3306) that the proxy listens on, and the selector that determines which pods
will answer requests sent through the proxy.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta) | optional | Standard object&#39;s metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata &#43;optional |
| spec | [ServiceSpec](#k8s.io.api.core.v1.ServiceSpec) | optional | Spec defines the behavior of a service. https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status &#43;optional |
| status | [ServiceStatus](#k8s.io.api.core.v1.ServiceStatus) | optional | Most recently observed status of the service. Populated by the system. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status &#43;optional |






<a name="k8s.io.api.core.v1.ServiceAccount"/>

### ServiceAccount
ServiceAccount binds together:
a name, understood by users, and perhaps by peripheral systems, for an identity
a principal that can be authenticated and authorized
a set of secrets


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta) | optional | Standard object&#39;s metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata &#43;optional |
| secrets | [ObjectReference](#k8s.io.api.core.v1.ObjectReference) | repeated | Secrets is the list of secrets allowed to be used by pods running using this ServiceAccount. More info: https://kubernetes.io/docs/concepts/configuration/secret &#43;optional &#43;patchMergeKey=name &#43;patchStrategy=merge |
| imagePullSecrets | [LocalObjectReference](#k8s.io.api.core.v1.LocalObjectReference) | repeated | ImagePullSecrets is a list of references to secrets in the same namespace to use for pulling any images in pods that reference this ServiceAccount. ImagePullSecrets are distinct from Secrets because Secrets can be mounted in the pod, but ImagePullSecrets are only accessed by the kubelet. More info: https://kubernetes.io/docs/concepts/containers/images/#specifying-imagepullsecrets-on-a-pod &#43;optional |
| automountServiceAccountToken | [bool](#bool) | optional | AutomountServiceAccountToken indicates whether pods running as this service account should have an API token automatically mounted. Can be overridden at the pod level. &#43;optional |






<a name="k8s.io.api.core.v1.ServiceAccountList"/>

### ServiceAccountList
ServiceAccountList is a list of ServiceAccount objects


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta) | optional | Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds &#43;optional |
| items | [ServiceAccount](#k8s.io.api.core.v1.ServiceAccount) | repeated | List of ServiceAccounts. More info: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account |






<a name="k8s.io.api.core.v1.ServiceList"/>

### ServiceList
ServiceList holds a list of services.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.ListMeta) | optional | Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds &#43;optional |
| items | [Service](#k8s.io.api.core.v1.Service) | repeated | List of services |






<a name="k8s.io.api.core.v1.ServicePort"/>

### ServicePort
ServicePort contains information on service&#39;s port.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | optional | The name of this port within the service. This must be a DNS_LABEL. All ports within a ServiceSpec must have unique names. This maps to the &#39;Name&#39; field in EndpointPort objects. Optional if only one ServicePort is defined on this service. &#43;optional |
| protocol | [string](#string) | optional | The IP protocol for this port. Supports &#34;TCP&#34; and &#34;UDP&#34;. Default is TCP. &#43;optional |
| port | [int32](#int32) | optional | The port that will be exposed by this service. |
| targetPort | [.k8s.io.apimachinery.pkg.util.intstr.IntOrString](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.util.intstr.IntOrString) | optional | Number or name of the port to access on the pods targeted by the service. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME. If this is a string, it will be looked up as a named port in the target Pod&#39;s container ports. If this is not specified, the value of the &#39;port&#39; field is used (an identity map). This field is ignored for services with clusterIP=None, and should be omitted or set equal to the &#39;port&#39; field. More info: https://kubernetes.io/docs/concepts/services-networking/service/#defining-a-service &#43;optional |
| nodePort | [int32](#int32) | optional | The port on each node on which this service is exposed when type=NodePort or LoadBalancer. Usually assigned by the system. If specified, it will be allocated to the service if unused or else creation of the service will fail. Default is to auto-allocate a port if the ServiceType of this Service requires one. More info: https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport &#43;optional |






<a name="k8s.io.api.core.v1.ServiceProxyOptions"/>

### ServiceProxyOptions
ServiceProxyOptions is the query options to a Service&#39;s proxy call.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [string](#string) | optional | Path is the part of URLs that include service endpoints, suffixes, and parameters to use for the current proxy request to service. For example, the whole request URL is http://localhost/api/v1/namespaces/kube-system/services/elasticsearch-logging/_search?q=user:kimchy. Path is _search?q=user:kimchy. &#43;optional |






<a name="k8s.io.api.core.v1.ServiceSpec"/>

### ServiceSpec
ServiceSpec describes the attributes that a user creates on a service.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ports | [ServicePort](#k8s.io.api.core.v1.ServicePort) | repeated | The list of ports that are exposed by this service. More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies &#43;patchMergeKey=port &#43;patchStrategy=merge |
| selector | [ServiceSpec.SelectorEntry](#k8s.io.api.core.v1.ServiceSpec.SelectorEntry) | repeated | Route service traffic to pods with label keys and values matching this selector. If empty or not present, the service is assumed to have an external process managing its endpoints, which Kubernetes will not modify. Only applies to types ClusterIP, NodePort, and LoadBalancer. Ignored if type is ExternalName. More info: https://kubernetes.io/docs/concepts/services-networking/service &#43;optional |
| clusterIP | [string](#string) | optional | clusterIP is the IP address of the service and is usually assigned randomly by the master. If an address is specified manually and is not in use by others, it will be allocated to the service; otherwise, creation of the service will fail. This field can not be changed through updates. Valid values are &#34;None&#34;, empty string (&#34;&#34;), or a valid IP address. &#34;None&#34; can be specified for headless services when proxying is not required. Only applies to types ClusterIP, NodePort, and LoadBalancer. Ignored if type is ExternalName. More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies &#43;optional |
| type | [string](#string) | optional | type determines how the Service is exposed. Defaults to ClusterIP. Valid options are ExternalName, ClusterIP, NodePort, and LoadBalancer. &#34;ExternalName&#34; maps to the specified externalName. &#34;ClusterIP&#34; allocates a cluster-internal IP address for load-balancing to endpoints. Endpoints are determined by the selector or if that is not specified, by manual construction of an Endpoints object. If clusterIP is &#34;None&#34;, no virtual IP is allocated and the endpoints are published as a set of endpoints rather than a stable IP. &#34;NodePort&#34; builds on ClusterIP and allocates a port on every node which routes to the clusterIP. &#34;LoadBalancer&#34; builds on NodePort and creates an external load-balancer (if supported in the current cloud) which routes to the clusterIP. More info: https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services---service-types &#43;optional |
| externalIPs | [string](#string) | repeated | externalIPs is a list of IP addresses for which nodes in the cluster will also accept traffic for this service. These IPs are not managed by Kubernetes. The user is responsible for ensuring that traffic arrives at a node with this IP. A common example is external load-balancers that are not part of the Kubernetes system. &#43;optional |
| sessionAffinity | [string](#string) | optional | Supports &#34;ClientIP&#34; and &#34;None&#34;. Used to maintain session affinity. Enable client IP based session affinity. Must be ClientIP or None. Defaults to None. More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies &#43;optional |
| loadBalancerIP | [string](#string) | optional | Only applies to Service Type: LoadBalancer LoadBalancer will get created with the IP specified in this field. This feature depends on whether the underlying cloud-provider supports specifying the loadBalancerIP when a load balancer is created. This field will be ignored if the cloud-provider does not support the feature. &#43;optional |
| loadBalancerSourceRanges | [string](#string) | repeated | If specified and supported by the platform, this will restrict traffic through the cloud-provider load-balancer will be restricted to the specified client IPs. This field will be ignored if the cloud-provider does not support the feature.&#34; More info: https://kubernetes.io/docs/tasks/access-application-cluster/configure-cloud-provider-firewall &#43;optional |
| externalName | [string](#string) | optional | externalName is the external reference that kubedns or equivalent will return as a CNAME record for this service. No proxying will be involved. Must be a valid DNS name and requires Type to be ExternalName. &#43;optional |
| externalTrafficPolicy | [string](#string) | optional | externalTrafficPolicy denotes if this Service desires to route external traffic to node-local or cluster-wide endpoints. &#34;Local&#34; preserves the client source IP and avoids a second hop for LoadBalancer and Nodeport type services, but risks potentially imbalanced traffic spreading. &#34;Cluster&#34; obscures the client source IP and may cause a second hop to another node, but should have good overall load-spreading. &#43;optional |
| healthCheckNodePort | [int32](#int32) | optional | healthCheckNodePort specifies the healthcheck nodePort for the service. If not specified, HealthCheckNodePort is created by the service api backend with the allocated nodePort. Will use user-specified nodePort value if specified by the client. Only effects when Type is set to LoadBalancer and ExternalTrafficPolicy is set to Local. &#43;optional |
| publishNotReadyAddresses | [bool](#bool) | optional | publishNotReadyAddresses, when set to true, indicates that DNS implementations must publish the notReadyAddresses of subsets for the Endpoints associated with the Service. The default value is false. The primary use case for setting this field is to use a StatefulSet&#39;s Headless Service to propagate SRV records for its Pods without respect to their readiness for purpose of peer discovery. This field will replace the service.alpha.kubernetes.io/tolerate-unready-endpoints when that annotation is deprecated and all clients have been converted to use this field. &#43;optional |
| sessionAffinityConfig | [SessionAffinityConfig](#k8s.io.api.core.v1.SessionAffinityConfig) | optional | sessionAffinityConfig contains the configurations of session affinity. &#43;optional |






<a name="k8s.io.api.core.v1.ServiceSpec.SelectorEntry"/>

### ServiceSpec.SelectorEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [string](#string) | optional |  |






<a name="k8s.io.api.core.v1.ServiceStatus"/>

### ServiceStatus
ServiceStatus represents the current status of a service.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| loadBalancer | [LoadBalancerStatus](#k8s.io.api.core.v1.LoadBalancerStatus) | optional | LoadBalancer contains the current status of the load-balancer, if one is present. &#43;optional |






<a name="k8s.io.api.core.v1.SessionAffinityConfig"/>

### SessionAffinityConfig
SessionAffinityConfig represents the configurations of session affinity.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| clientIP | [ClientIPConfig](#k8s.io.api.core.v1.ClientIPConfig) | optional | clientIP contains the configurations of Client IP based session affinity. &#43;optional |






<a name="k8s.io.api.core.v1.StorageOSPersistentVolumeSource"/>

### StorageOSPersistentVolumeSource
Represents a StorageOS persistent volume resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| volumeName | [string](#string) | optional | VolumeName is the human-readable name of the StorageOS volume. Volume names are only unique within a namespace. |
| volumeNamespace | [string](#string) | optional | VolumeNamespace specifies the scope of the volume within StorageOS. If no namespace is specified then the Pod&#39;s namespace will be used. This allows the Kubernetes name scoping to be mirrored within StorageOS for tighter integration. Set VolumeName to any name to override the default behaviour. Set to &#34;default&#34; if you are not using namespaces within StorageOS. Namespaces that do not pre-exist within StorageOS will be created. &#43;optional |
| fsType | [string](#string) | optional | Filesystem type to mount. Must be a filesystem type supported by the host operating system. Ex. &#34;ext4&#34;, &#34;xfs&#34;, &#34;ntfs&#34;. Implicitly inferred to be &#34;ext4&#34; if unspecified. &#43;optional |
| readOnly | [bool](#bool) | optional | Defaults to false (read/write). ReadOnly here will force the ReadOnly setting in VolumeMounts. &#43;optional |
| secretRef | [ObjectReference](#k8s.io.api.core.v1.ObjectReference) | optional | SecretRef specifies the secret to use for obtaining the StorageOS API credentials. If not specified, default values will be attempted. &#43;optional |






<a name="k8s.io.api.core.v1.StorageOSVolumeSource"/>

### StorageOSVolumeSource
Represents a StorageOS persistent volume resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| volumeName | [string](#string) | optional | VolumeName is the human-readable name of the StorageOS volume. Volume names are only unique within a namespace. |
| volumeNamespace | [string](#string) | optional | VolumeNamespace specifies the scope of the volume within StorageOS. If no namespace is specified then the Pod&#39;s namespace will be used. This allows the Kubernetes name scoping to be mirrored within StorageOS for tighter integration. Set VolumeName to any name to override the default behaviour. Set to &#34;default&#34; if you are not using namespaces within StorageOS. Namespaces that do not pre-exist within StorageOS will be created. &#43;optional |
| fsType | [string](#string) | optional | Filesystem type to mount. Must be a filesystem type supported by the host operating system. Ex. &#34;ext4&#34;, &#34;xfs&#34;, &#34;ntfs&#34;. Implicitly inferred to be &#34;ext4&#34; if unspecified. &#43;optional |
| readOnly | [bool](#bool) | optional | Defaults to false (read/write). ReadOnly here will force the ReadOnly setting in VolumeMounts. &#43;optional |
| secretRef | [LocalObjectReference](#k8s.io.api.core.v1.LocalObjectReference) | optional | SecretRef specifies the secret to use for obtaining the StorageOS API credentials. If not specified, default values will be attempted. &#43;optional |






<a name="k8s.io.api.core.v1.Sysctl"/>

### Sysctl
Sysctl defines a kernel parameter to be set


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | optional | Name of a property to set |
| value | [string](#string) | optional | Value of a property to set |






<a name="k8s.io.api.core.v1.TCPSocketAction"/>

### TCPSocketAction
TCPSocketAction describes an action based on opening a socket


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| port | [.k8s.io.apimachinery.pkg.util.intstr.IntOrString](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.util.intstr.IntOrString) | optional | Number or name of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME. |
| host | [string](#string) | optional | Optional: Host name to connect to, defaults to the pod IP. &#43;optional |






<a name="k8s.io.api.core.v1.Taint"/>

### Taint
The node this Taint is attached to has the &#34;effect&#34; on
any pod that does not tolerate the Taint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional | Required. The taint key to be applied to a node. |
| value | [string](#string) | optional | Required. The taint value corresponding to the taint key. &#43;optional |
| effect | [string](#string) | optional | Required. The effect of the taint on pods that do not tolerate the taint. Valid effects are NoSchedule, PreferNoSchedule and NoExecute. |
| timeAdded | [.k8s.io.apimachinery.pkg.apis.meta.v1.Time](#k8s.io.api.core.v1..k8s.io.apimachinery.pkg.apis.meta.v1.Time) | optional | TimeAdded represents the time at which the taint was added. It is only written for NoExecute taints. &#43;optional |






<a name="k8s.io.api.core.v1.Toleration"/>

### Toleration
The pod this Toleration is attached to tolerates any taint that matches
the triple &lt;key,value,effect&gt; using the matching operator &lt;operator&gt;.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional | Key is the taint key that the toleration applies to. Empty means match all taint keys. If the key is empty, operator must be Exists; this combination means to match all values and all keys. &#43;optional |
| operator | [string](#string) | optional | Operator represents a key&#39;s relationship to the value. Valid operators are Exists and Equal. Defaults to Equal. Exists is equivalent to wildcard for value, so that a pod can tolerate all taints of a particular category. &#43;optional |
| value | [string](#string) | optional | Value is the taint value the toleration matches to. If the operator is Exists, the value should be empty, otherwise just a regular string. &#43;optional |
| effect | [string](#string) | optional | Effect indicates the taint effect to match. Empty means match all taint effects. When specified, allowed values are NoSchedule, PreferNoSchedule and NoExecute. &#43;optional |
| tolerationSeconds | [int64](#int64) | optional | TolerationSeconds represents the period of time the toleration (which must be of effect NoExecute, otherwise this field is ignored) tolerates the taint. By default, it is not set, which means tolerate the taint forever (do not evict). Zero and negative values will be treated as 0 (evict immediately) by the system. &#43;optional |






<a name="k8s.io.api.core.v1.Volume"/>

### Volume
Volume represents a named volume in a pod that may be accessed by any container in the pod.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | optional | Volume&#39;s name. Must be a DNS_LABEL and unique within the pod. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names |
| volumeSource | [VolumeSource](#k8s.io.api.core.v1.VolumeSource) | optional | VolumeSource represents the location and type of the mounted volume. If not specified, the Volume is implied to be an EmptyDir. This implied behavior is deprecated and will be removed in a future version. |






<a name="k8s.io.api.core.v1.VolumeMount"/>

### VolumeMount
VolumeMount describes a mounting of a Volume within a container.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | optional | This must match the Name of a Volume. |
| readOnly | [bool](#bool) | optional | Mounted read-only if true, read-write otherwise (false or unspecified). Defaults to false. &#43;optional |
| mountPath | [string](#string) | optional | Path within the container at which the volume should be mounted. Must not contain &#39;:&#39;. |
| subPath | [string](#string) | optional | Path within the volume from which the container&#39;s volume should be mounted. Defaults to &#34;&#34; (volume&#39;s root). &#43;optional |
| mountPropagation | [string](#string) | optional | mountPropagation determines how mounts are propagated from the host to container and the other way around. When not set, MountPropagationHostToContainer is used. This field is alpha in 1.8 and can be reworked or removed in a future release. &#43;optional |






<a name="k8s.io.api.core.v1.VolumeProjection"/>

### VolumeProjection
Projection that may be projected along with other supported volume types


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| secret | [SecretProjection](#k8s.io.api.core.v1.SecretProjection) | optional | information about the secret data to project |
| downwardAPI | [DownwardAPIProjection](#k8s.io.api.core.v1.DownwardAPIProjection) | optional | information about the downwardAPI data to project |
| configMap | [ConfigMapProjection](#k8s.io.api.core.v1.ConfigMapProjection) | optional | information about the configMap data to project |






<a name="k8s.io.api.core.v1.VolumeSource"/>

### VolumeSource
Represents the source of a volume to mount.
Only one of its members may be specified.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hostPath | [HostPathVolumeSource](#k8s.io.api.core.v1.HostPathVolumeSource) | optional | HostPath represents a pre-existing file or directory on the host machine that is directly exposed to the container. This is generally used for system agents or other privileged things that are allowed to see the host machine. Most containers will NOT need this. More info: https://kubernetes.io/docs/concepts/storage/volumes#hostpath --- TODO(jonesdl) We need to restrict who can use host directory mounts and who can/can not mount host directories as read/write. &#43;optional |
| emptyDir | [EmptyDirVolumeSource](#k8s.io.api.core.v1.EmptyDirVolumeSource) | optional | EmptyDir represents a temporary directory that shares a pod&#39;s lifetime. More info: https://kubernetes.io/docs/concepts/storage/volumes#emptydir &#43;optional |
| gcePersistentDisk | [GCEPersistentDiskVolumeSource](#k8s.io.api.core.v1.GCEPersistentDiskVolumeSource) | optional | GCEPersistentDisk represents a GCE Disk resource that is attached to a kubelet&#39;s host machine and then exposed to the pod. More info: https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk &#43;optional |
| awsElasticBlockStore | [AWSElasticBlockStoreVolumeSource](#k8s.io.api.core.v1.AWSElasticBlockStoreVolumeSource) | optional | AWSElasticBlockStore represents an AWS Disk resource that is attached to a kubelet&#39;s host machine and then exposed to the pod. More info: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore &#43;optional |
| gitRepo | [GitRepoVolumeSource](#k8s.io.api.core.v1.GitRepoVolumeSource) | optional | GitRepo represents a git repository at a particular revision. &#43;optional |
| secret | [SecretVolumeSource](#k8s.io.api.core.v1.SecretVolumeSource) | optional | Secret represents a secret that should populate this volume. More info: https://kubernetes.io/docs/concepts/storage/volumes#secret &#43;optional |
| nfs | [NFSVolumeSource](#k8s.io.api.core.v1.NFSVolumeSource) | optional | NFS represents an NFS mount on the host that shares a pod&#39;s lifetime More info: https://kubernetes.io/docs/concepts/storage/volumes#nfs &#43;optional |
| iscsi | [ISCSIVolumeSource](#k8s.io.api.core.v1.ISCSIVolumeSource) | optional | ISCSI represents an ISCSI Disk resource that is attached to a kubelet&#39;s host machine and then exposed to the pod. More info: https://releases.k8s.io/HEAD/examples/volumes/iscsi/README.md &#43;optional |
| glusterfs | [GlusterfsVolumeSource](#k8s.io.api.core.v1.GlusterfsVolumeSource) | optional | Glusterfs represents a Glusterfs mount on the host that shares a pod&#39;s lifetime. More info: https://releases.k8s.io/HEAD/examples/volumes/glusterfs/README.md &#43;optional |
| persistentVolumeClaim | [PersistentVolumeClaimVolumeSource](#k8s.io.api.core.v1.PersistentVolumeClaimVolumeSource) | optional | PersistentVolumeClaimVolumeSource represents a reference to a PersistentVolumeClaim in the same namespace. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims &#43;optional |
| rbd | [RBDVolumeSource](#k8s.io.api.core.v1.RBDVolumeSource) | optional | RBD represents a Rados Block Device mount on the host that shares a pod&#39;s lifetime. More info: https://releases.k8s.io/HEAD/examples/volumes/rbd/README.md &#43;optional |
| flexVolume | [FlexVolumeSource](#k8s.io.api.core.v1.FlexVolumeSource) | optional | FlexVolume represents a generic volume resource that is provisioned/attached using an exec based plugin. This is an alpha feature and may change in future. &#43;optional |
| cinder | [CinderVolumeSource](#k8s.io.api.core.v1.CinderVolumeSource) | optional | Cinder represents a cinder volume attached and mounted on kubelets host machine More info: https://releases.k8s.io/HEAD/examples/mysql-cinder-pd/README.md &#43;optional |
| cephfs | [CephFSVolumeSource](#k8s.io.api.core.v1.CephFSVolumeSource) | optional | CephFS represents a Ceph FS mount on the host that shares a pod&#39;s lifetime &#43;optional |
| flocker | [FlockerVolumeSource](#k8s.io.api.core.v1.FlockerVolumeSource) | optional | Flocker represents a Flocker volume attached to a kubelet&#39;s host machine. This depends on the Flocker control service being running &#43;optional |
| downwardAPI | [DownwardAPIVolumeSource](#k8s.io.api.core.v1.DownwardAPIVolumeSource) | optional | DownwardAPI represents downward API about the pod that should populate this volume &#43;optional |
| fc | [FCVolumeSource](#k8s.io.api.core.v1.FCVolumeSource) | optional | FC represents a Fibre Channel resource that is attached to a kubelet&#39;s host machine and then exposed to the pod. &#43;optional |
| azureFile | [AzureFileVolumeSource](#k8s.io.api.core.v1.AzureFileVolumeSource) | optional | AzureFile represents an Azure File Service mount on the host and bind mount to the pod. &#43;optional |
| configMap | [ConfigMapVolumeSource](#k8s.io.api.core.v1.ConfigMapVolumeSource) | optional | ConfigMap represents a configMap that should populate this volume &#43;optional |
| vsphereVolume | [VsphereVirtualDiskVolumeSource](#k8s.io.api.core.v1.VsphereVirtualDiskVolumeSource) | optional | VsphereVolume represents a vSphere volume attached and mounted on kubelets host machine &#43;optional |
| quobyte | [QuobyteVolumeSource](#k8s.io.api.core.v1.QuobyteVolumeSource) | optional | Quobyte represents a Quobyte mount on the host that shares a pod&#39;s lifetime &#43;optional |
| azureDisk | [AzureDiskVolumeSource](#k8s.io.api.core.v1.AzureDiskVolumeSource) | optional | AzureDisk represents an Azure Data Disk mount on the host and bind mount to the pod. &#43;optional |
| photonPersistentDisk | [PhotonPersistentDiskVolumeSource](#k8s.io.api.core.v1.PhotonPersistentDiskVolumeSource) | optional | PhotonPersistentDisk represents a PhotonController persistent disk attached and mounted on kubelets host machine |
| projected | [ProjectedVolumeSource](#k8s.io.api.core.v1.ProjectedVolumeSource) | optional | Items for all in one resources secrets, configmaps, and downward API |
| portworxVolume | [PortworxVolumeSource](#k8s.io.api.core.v1.PortworxVolumeSource) | optional | PortworxVolume represents a portworx volume attached and mounted on kubelets host machine &#43;optional |
| scaleIO | [ScaleIOVolumeSource](#k8s.io.api.core.v1.ScaleIOVolumeSource) | optional | ScaleIO represents a ScaleIO persistent volume attached and mounted on Kubernetes nodes. &#43;optional |
| storageos | [StorageOSVolumeSource](#k8s.io.api.core.v1.StorageOSVolumeSource) | optional | StorageOS represents a StorageOS volume attached and mounted on Kubernetes nodes. &#43;optional |






<a name="k8s.io.api.core.v1.VsphereVirtualDiskVolumeSource"/>

### VsphereVirtualDiskVolumeSource
Represents a vSphere volume resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| volumePath | [string](#string) | optional | Path that identifies vSphere volume vmdk |
| fsType | [string](#string) | optional | Filesystem type to mount. Must be a filesystem type supported by the host operating system. Ex. &#34;ext4&#34;, &#34;xfs&#34;, &#34;ntfs&#34;. Implicitly inferred to be &#34;ext4&#34; if unspecified. &#43;optional |
| storagePolicyName | [string](#string) | optional | Storage Policy Based Management (SPBM) profile name. &#43;optional |
| storagePolicyID | [string](#string) | optional | Storage Policy Based Management (SPBM) profile ID associated with the StoragePolicyName. &#43;optional |






<a name="k8s.io.api.core.v1.WeightedPodAffinityTerm"/>

### WeightedPodAffinityTerm
The weights of all of the matched WeightedPodAffinityTerm fields are added per-node to find the most preferred node(s)


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| weight | [int32](#int32) | optional | weight associated with matching the corresponding podAffinityTerm, in the range 1-100. |
| podAffinityTerm | [PodAffinityTerm](#k8s.io.api.core.v1.PodAffinityTerm) | optional | Required. A pod affinity term, associated with the corresponding weight. |





 

 

 

 



<a name="seldon_deployment.proto"/>
<p align="right"><a href="#top">Top</a></p>

## seldon_deployment.proto
[START declaration]


<a name="seldon.protos.DeploymentSpec"/>

### DeploymentSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | optional |  |
| predictors | [PredictorSpec](#seldon.protos.PredictorSpec) | repeated |  |
| endpoint | [Endpoint](#seldon.protos.Endpoint) | optional |  |
| oauth_key | [string](#string) | optional |  |
| oauth_secret | [string](#string) | optional |  |
| annotations | [DeploymentSpec.AnnotationsEntry](#seldon.protos.DeploymentSpec.AnnotationsEntry) | repeated |  |






<a name="seldon.protos.DeploymentSpec.AnnotationsEntry"/>

### DeploymentSpec.AnnotationsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [string](#string) | optional |  |






<a name="seldon.protos.DeploymentStatus"/>

### DeploymentStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| state | [string](#string) | optional |  |
| description | [string](#string) | optional |  |
| predictorStatus | [PredictorStatus](#seldon.protos.PredictorStatus) | repeated |  |






<a name="seldon.protos.Endpoint"/>

### Endpoint



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| service_host | [string](#string) | optional |  |
| service_port | [int32](#int32) | optional |  |
| type | [Endpoint.EndpointType](#seldon.protos.Endpoint.EndpointType) | optional |  |






<a name="seldon.protos.Parameter"/>

### Parameter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | required |  |
| value | [string](#string) | required |  |
| type | [Parameter.ParameterType](#seldon.protos.Parameter.ParameterType) | required |  |






<a name="seldon.protos.PredictiveUnit"/>

### PredictiveUnit



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | required | must match container name of component if subtype microservice |
| children | [PredictiveUnit](#seldon.protos.PredictiveUnit) | repeated |  |
| type | [PredictiveUnit.PredictiveUnitType](#seldon.protos.PredictiveUnit.PredictiveUnitType) | optional |  |
| subtype | [PredictiveUnit.PredictiveUnitSubtype](#seldon.protos.PredictiveUnit.PredictiveUnitSubtype) | optional |  |
| endpoint | [Endpoint](#seldon.protos.Endpoint) | optional |  |
| parameters | [Parameter](#seldon.protos.Parameter) | repeated |  |






<a name="seldon.protos.PredictorSpec"/>

### PredictorSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | required |  |
| graph | [PredictiveUnit](#seldon.protos.PredictiveUnit) | required |  |
| componentSpec | [.k8s.io.api.core.v1.PodTemplateSpec](#seldon.protos..k8s.io.api.core.v1.PodTemplateSpec) | required |  |
| replicas | [int32](#int32) | optional |  |
| annotations | [PredictorSpec.AnnotationsEntry](#seldon.protos.PredictorSpec.AnnotationsEntry) | repeated |  |






<a name="seldon.protos.PredictorSpec.AnnotationsEntry"/>

### PredictorSpec.AnnotationsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) | optional |  |
| value | [string](#string) | optional |  |






<a name="seldon.protos.PredictorStatus"/>

### PredictorStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | required |  |
| status | [string](#string) | optional |  |
| description | [string](#string) | optional |  |
| replicas | [int32](#int32) | optional |  |
| replicasAvailable | [int32](#int32) | optional |  |






<a name="seldon.protos.SeldonDeployment"/>

### SeldonDeployment



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| apiVersion | [string](#string) | required |  |
| kind | [string](#string) | required |  |
| metadata | [.k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta](#seldon.protos..k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta) | optional |  |
| spec | [DeploymentSpec](#seldon.protos.DeploymentSpec) | required |  |
| status | [DeploymentStatus](#seldon.protos.DeploymentStatus) | optional |  |





 


<a name="seldon.protos.Endpoint.EndpointType"/>

### Endpoint.EndpointType


| Name | Number | Description |
| ---- | ------ | ----------- |
| REST | 0 |  |
| GRPC | 1 |  |



<a name="seldon.protos.Parameter.ParameterType"/>

### Parameter.ParameterType


| Name | Number | Description |
| ---- | ------ | ----------- |
| INT | 0 |  |
| FLOAT | 1 |  |
| DOUBLE | 2 |  |
| STRING | 3 |  |
| BOOL | 4 |  |



<a name="seldon.protos.PredictiveUnit.PredictiveUnitSubtype"/>

### PredictiveUnit.PredictiveUnitSubtype


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN_SUBTYPE | 0 |  |
| MICROSERVICE | 1 |  |
| SIMPLE_MODEL | 2 |  |
| SIMPLE_ROUTER | 3 |  |
| RANDOM_ABTEST | 4 |  |
| AVERAGE_COMBINER | 5 |  |



<a name="seldon.protos.PredictiveUnit.PredictiveUnitType"/>

### PredictiveUnit.PredictiveUnitType


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN_TYPE | 0 |  |
| ROUTER | 1 |  |
| COMBINER | 2 |  |
| MODEL | 3 |  |
| TRANSFORMER | 4 |  |


 

 

 



## Scalar Value Types

| .proto Type | Notes | C++ Type | Java Type | Python Type |
| ----------- | ----- | -------- | --------- | ----------- |
| <a name="double" /> double |  | double | double | float |
| <a name="float" /> float |  | float | float | float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long |
| <a name="bool" /> bool |  | bool | boolean | boolean |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str |

