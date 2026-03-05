# API Reference

Packages:

- [redroid.io/v1alpha1](#redroidiov1alpha1)

# redroid.io/v1alpha1

Resource Types:

- [RedroidInstance](#redroidinstance)

- [RedroidTask](#redroidtask)




## RedroidInstance
<sup><sup>[↩ Parent](#redroidiov1alpha1 )</sup></sup>






RedroidInstance is the Schema for the redroidinstances API.
It represents a single persistent Android container instance backed by overlayfs storage.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>redroid.io/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>RedroidInstance</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespec">spec</a></b></td>
        <td>object</td>
        <td>
          RedroidInstanceSpec defines the desired state of a RedroidInstance.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancestatus">status</a></b></td>
        <td>object</td>
        <td>
          RedroidInstanceStatus defines the observed state of RedroidInstance.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec
<sup><sup>[↩ Parent](#redroidinstance)</sup></sup>



RedroidInstanceSpec defines the desired state of a RedroidInstance.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>diffDataPVC</b></td>
        <td>string</td>
        <td>
          DiffDataPVC is the PVC name for the per-instance /data-diff volume (writable upper layer).<br/>
          <br/>
            <i>Default</i>: redroid-data-diff-pvc<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>index</b></td>
        <td>integer</td>
        <td>
          Index is the overlayfs partition index shared with RedroidTasks.
/data-base is shared across all instances; /data-diff/<Index> is this instance's private layer.<br/>
          <br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>sharedDataPVC</b></td>
        <td>string</td>
        <td>
          SharedDataPVC is the PVC name for the shared /data-base volume (read-only lower layer).<br/>
          <br/>
            <i>Default</i>: redroid-data-base-pvc<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>adbPort</b></td>
        <td>integer</td>
        <td>
          ADBPort is the ADB TCP port exposed by the container. Defaults to 5555.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 5555<br/>
            <i>Minimum</i>: 1<br/>
            <i>Maximum</i>: 65535<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespecaffinity">affinity</a></b></td>
        <td>object</td>
        <td>
          Affinity provides advanced scheduling constraints.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>baseMode</b></td>
        <td>boolean</td>
        <td>
          BaseMode makes this instance write directly to SharedDataPVC as /data,
bypassing the overlayfs per-instance layer. Use this to initialise the shared
base image (install common APKs, first-boot setup, etc.).

In base mode:
  - SharedDataPVC is mounted at /data (read-write).
  - DiffDataPVC is NOT mounted.
  - androidboot.use_redroid_overlayfs is set to 0.

Typical workflow:
  1. Create a RedroidInstance with baseMode: true, suspend: false.
  2. ADB-connect and perform initial setup (app installs, configs).
  3. Set suspend: true (or delete) once satisfied.
  4. Normal instances (baseMode: false) sharing the same SharedDataPVC
     will see the initialised state as their read-only lower layer.

WARNING: running a base-mode instance concurrently with normal instances
that share the same SharedDataPVC may corrupt their overlayfs lower layer.
Use spec.suspend or status.suspended on all normal instances first.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>extraArgs</b></td>
        <td>[]string</td>
        <td>
          ExtraArgs are additional androidboot.* arguments passed to the redroid container.
Supports $(VAR_NAME) substitution from ExtraEnv.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespecextraenvindex">extraEnv</a></b></td>
        <td>[]object</td>
        <td>
          ExtraEnv defines additional environment variables for the redroid container.
Supports valueFrom.secretKeyRef and valueFrom.configMapKeyRef for sensitive params.
Can be referenced in ExtraArgs via $(VAR_NAME) syntax.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>gpuMode</b></td>
        <td>enum</td>
        <td>
          GPUMode sets androidboot.redroid_gpu_mode. Defaults to "host".<br/>
          <br/>
            <i>Enum</i>: host, guest, auto, none<br/>
            <i>Default</i>: host<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>gpuNode</b></td>
        <td>string</td>
        <td>
          GPUNode sets androidboot.redroid_gpu_node (DRM device path).
When empty the Redroid runtime auto-detects the GPU node.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>image</b></td>
        <td>string</td>
        <td>
          Image is the redroid container image. Defaults to redroid/redroid:16.0.0-latest.<br/>
          <br/>
            <i>Default</i>: redroid/redroid:16.0.0-latest<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>imagePullPolicy</b></td>
        <td>string</td>
        <td>
          ImagePullPolicy for the redroid container image. Defaults to IfNotPresent.<br/>
          <br/>
            <i>Default</i>: IfNotPresent<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespecimagepullsecretsindex">imagePullSecrets</a></b></td>
        <td>[]object</td>
        <td>
          ImagePullSecrets references Secret resources for pulling private images.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespecnetwork">network</a></b></td>
        <td>object</td>
        <td>
          Network configures DNS and proxy settings for the Android instance.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>nodeSelector</b></td>
        <td>map[string]string</td>
        <td>
          NodeSelector constrains which node the Pod is scheduled on.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespecresources">resources</a></b></td>
        <td>object</td>
        <td>
          Resources optionally sets CPU/memory limits on the redroid container.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespecscreen">screen</a></b></td>
        <td>object</td>
        <td>
          Screen configures the virtual display (width, height, dpi, fps).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespecservice">service</a></b></td>
        <td>object</td>
        <td>
          Service customises the Kubernetes Service the operator creates to expose the ADB port.
The Service is always created; this field controls its type, annotations, and node port.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>suspend</b></td>
        <td>boolean</td>
        <td>
          Suspend controls whether the Android instance Pod should be running.
Set to false (default) to start the instance; true to stop it without deleting the resource.
This follows the same semantics as CronJob.spec.suspend.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespectolerationsindex">tolerations</a></b></td>
        <td>[]object</td>
        <td>
          Tolerations allows the Pod to be scheduled on nodes with matching taints.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity
<sup><sup>[↩ Parent](#redroidinstancespec)</sup></sup>



Affinity provides advanced scheduling constraints.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidinstancespecaffinitynodeaffinity">nodeAffinity</a></b></td>
        <td>object</td>
        <td>
          Describes node affinity scheduling rules for the pod.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespecaffinitypodaffinity">podAffinity</a></b></td>
        <td>object</td>
        <td>
          Describes pod affinity scheduling rules (e.g. co-locate this pod in the same node, zone, etc. as some other pod(s)).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespecaffinitypodantiaffinity">podAntiAffinity</a></b></td>
        <td>object</td>
        <td>
          Describes pod anti-affinity scheduling rules (e.g. avoid putting this pod in the same node, zone, etc. as some other pod(s)).<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.nodeAffinity
<sup><sup>[↩ Parent](#redroidinstancespecaffinity)</sup></sup>



Describes node affinity scheduling rules for the pod.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidinstancespecaffinitynodeaffinitypreferredduringschedulingignoredduringexecutionindex">preferredDuringSchedulingIgnoredDuringExecution</a></b></td>
        <td>[]object</td>
        <td>
          The scheduler will prefer to schedule pods to nodes that satisfy
the affinity expressions specified by this field, but it may choose
a node that violates one or more of the expressions. The node that is
most preferred is the one with the greatest sum of weights, i.e.
for each node that meets all of the scheduling requirements (resource
request, requiredDuringScheduling affinity expressions, etc.),
compute a sum by iterating through the elements of this field and adding
"weight" to the sum if the node matches the corresponding matchExpressions; the
node(s) with the highest sum are the most preferred.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespecaffinitynodeaffinityrequiredduringschedulingignoredduringexecution">requiredDuringSchedulingIgnoredDuringExecution</a></b></td>
        <td>object</td>
        <td>
          If the affinity requirements specified by this field are not met at
scheduling time, the pod will not be scheduled onto the node.
If the affinity requirements specified by this field cease to be met
at some point during pod execution (e.g. due to an update), the system
may or may not try to eventually evict the pod from its node.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[index]
<sup><sup>[↩ Parent](#redroidinstancespecaffinitynodeaffinity)</sup></sup>



An empty preferred scheduling term matches all objects with implicit weight 0
(i.e. it's a no-op). A null preferred scheduling term matches no objects (i.e. is also a no-op).

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidinstancespecaffinitynodeaffinitypreferredduringschedulingignoredduringexecutionindexpreference">preference</a></b></td>
        <td>object</td>
        <td>
          A node selector term, associated with the corresponding weight.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>weight</b></td>
        <td>integer</td>
        <td>
          Weight associated with matching the corresponding nodeSelectorTerm, in the range 1-100.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[index].preference
<sup><sup>[↩ Parent](#redroidinstancespecaffinitynodeaffinitypreferredduringschedulingignoredduringexecutionindex)</sup></sup>



A node selector term, associated with the corresponding weight.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidinstancespecaffinitynodeaffinitypreferredduringschedulingignoredduringexecutionindexpreferencematchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          A list of node selector requirements by node's labels.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespecaffinitynodeaffinitypreferredduringschedulingignoredduringexecutionindexpreferencematchfieldsindex">matchFields</a></b></td>
        <td>[]object</td>
        <td>
          A list of node selector requirements by node's fields.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[index].preference.matchExpressions[index]
<sup><sup>[↩ Parent](#redroidinstancespecaffinitynodeaffinitypreferredduringschedulingignoredduringexecutionindexpreference)</sup></sup>



A node selector requirement is a selector that contains values, a key, and an operator
that relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          Represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          An array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. If the operator is Gt or Lt, the values
array must have a single element, which will be interpreted as an integer.
This array is replaced during a strategic merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[index].preference.matchFields[index]
<sup><sup>[↩ Parent](#redroidinstancespecaffinitynodeaffinitypreferredduringschedulingignoredduringexecutionindexpreference)</sup></sup>



A node selector requirement is a selector that contains values, a key, and an operator
that relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          Represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          An array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. If the operator is Gt or Lt, the values
array must have a single element, which will be interpreted as an integer.
This array is replaced during a strategic merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution
<sup><sup>[↩ Parent](#redroidinstancespecaffinitynodeaffinity)</sup></sup>



If the affinity requirements specified by this field are not met at
scheduling time, the pod will not be scheduled onto the node.
If the affinity requirements specified by this field cease to be met
at some point during pod execution (e.g. due to an update), the system
may or may not try to eventually evict the pod from its node.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidinstancespecaffinitynodeaffinityrequiredduringschedulingignoredduringexecutionnodeselectortermsindex">nodeSelectorTerms</a></b></td>
        <td>[]object</td>
        <td>
          Required. A list of node selector terms. The terms are ORed.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[index]
<sup><sup>[↩ Parent](#redroidinstancespecaffinitynodeaffinityrequiredduringschedulingignoredduringexecution)</sup></sup>



A null or empty node selector term matches no objects. The requirements of
them are ANDed.
The TopologySelectorTerm type implements a subset of the NodeSelectorTerm.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidinstancespecaffinitynodeaffinityrequiredduringschedulingignoredduringexecutionnodeselectortermsindexmatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          A list of node selector requirements by node's labels.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespecaffinitynodeaffinityrequiredduringschedulingignoredduringexecutionnodeselectortermsindexmatchfieldsindex">matchFields</a></b></td>
        <td>[]object</td>
        <td>
          A list of node selector requirements by node's fields.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[index].matchExpressions[index]
<sup><sup>[↩ Parent](#redroidinstancespecaffinitynodeaffinityrequiredduringschedulingignoredduringexecutionnodeselectortermsindex)</sup></sup>



A node selector requirement is a selector that contains values, a key, and an operator
that relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          Represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          An array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. If the operator is Gt or Lt, the values
array must have a single element, which will be interpreted as an integer.
This array is replaced during a strategic merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[index].matchFields[index]
<sup><sup>[↩ Parent](#redroidinstancespecaffinitynodeaffinityrequiredduringschedulingignoredduringexecutionnodeselectortermsindex)</sup></sup>



A node selector requirement is a selector that contains values, a key, and an operator
that relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          Represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          An array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. If the operator is Gt or Lt, the values
array must have a single element, which will be interpreted as an integer.
This array is replaced during a strategic merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.podAffinity
<sup><sup>[↩ Parent](#redroidinstancespecaffinity)</sup></sup>



Describes pod affinity scheduling rules (e.g. co-locate this pod in the same node, zone, etc. as some other pod(s)).

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidinstancespecaffinitypodaffinitypreferredduringschedulingignoredduringexecutionindex">preferredDuringSchedulingIgnoredDuringExecution</a></b></td>
        <td>[]object</td>
        <td>
          The scheduler will prefer to schedule pods to nodes that satisfy
the affinity expressions specified by this field, but it may choose
a node that violates one or more of the expressions. The node that is
most preferred is the one with the greatest sum of weights, i.e.
for each node that meets all of the scheduling requirements (resource
request, requiredDuringScheduling affinity expressions, etc.),
compute a sum by iterating through the elements of this field and adding
"weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the
node(s) with the highest sum are the most preferred.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespecaffinitypodaffinityrequiredduringschedulingignoredduringexecutionindex">requiredDuringSchedulingIgnoredDuringExecution</a></b></td>
        <td>[]object</td>
        <td>
          If the affinity requirements specified by this field are not met at
scheduling time, the pod will not be scheduled onto the node.
If the affinity requirements specified by this field cease to be met
at some point during pod execution (e.g. due to a pod label update), the
system may or may not try to eventually evict the pod from its node.
When there are multiple elements, the lists of nodes corresponding to each
podAffinityTerm are intersected, i.e. all terms must be satisfied.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[index]
<sup><sup>[↩ Parent](#redroidinstancespecaffinitypodaffinity)</sup></sup>



The weights of all of the matched WeightedPodAffinityTerm fields are added per-node to find the most preferred node(s)

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidinstancespecaffinitypodaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinityterm">podAffinityTerm</a></b></td>
        <td>object</td>
        <td>
          Required. A pod affinity term, associated with the corresponding weight.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>weight</b></td>
        <td>integer</td>
        <td>
          weight associated with matching the corresponding podAffinityTerm,
in the range 1-100.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[index].podAffinityTerm
<sup><sup>[↩ Parent](#redroidinstancespecaffinitypodaffinitypreferredduringschedulingignoredduringexecutionindex)</sup></sup>



Required. A pod affinity term, associated with the corresponding weight.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>topologyKey</b></td>
        <td>string</td>
        <td>
          This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching
the labelSelector in the specified namespaces, where co-located is defined as running on a node
whose value of the label with key topologyKey matches that of any node on which any of the
selected pods is running.
Empty topologyKey is not allowed.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespecaffinitypodaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinitytermlabelselector">labelSelector</a></b></td>
        <td>object</td>
        <td>
          A label query over a set of resources, in this case pods.
If it's null, this PodAffinityTerm matches with no Pods.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabelKeys</b></td>
        <td>[]string</td>
        <td>
          MatchLabelKeys is a set of pod label keys to select which pods will
be taken into consideration. The keys are used to lookup values from the
incoming pod labels, those key-value labels are merged with `LabelSelector` as `key in (value)`
to select the group of existing pods which pods will be taken into consideration
for the incoming pod's pod (anti) affinity. Keys that don't exist in the incoming
pod labels will be ignored. The default value is empty.
The same key is forbidden to exist in both MatchLabelKeys and LabelSelector.
Also, MatchLabelKeys cannot be set when LabelSelector isn't set.
This is an alpha field and requires enabling MatchLabelKeysInPodAffinity feature gate.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>mismatchLabelKeys</b></td>
        <td>[]string</td>
        <td>
          MismatchLabelKeys is a set of pod label keys to select which pods will
be taken into consideration. The keys are used to lookup values from the
incoming pod labels, those key-value labels are merged with `LabelSelector` as `key notin (value)`
to select the group of existing pods which pods will be taken into consideration
for the incoming pod's pod (anti) affinity. Keys that don't exist in the incoming
pod labels will be ignored. The default value is empty.
The same key is forbidden to exist in both MismatchLabelKeys and LabelSelector.
Also, MismatchLabelKeys cannot be set when LabelSelector isn't set.
This is an alpha field and requires enabling MatchLabelKeysInPodAffinity feature gate.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespecaffinitypodaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinitytermnamespaceselector">namespaceSelector</a></b></td>
        <td>object</td>
        <td>
          A label query over the set of namespaces that the term applies to.
The term is applied to the union of the namespaces selected by this field
and the ones listed in the namespaces field.
null selector and null or empty namespaces list means "this pod's namespace".
An empty selector ({}) matches all namespaces.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespaces</b></td>
        <td>[]string</td>
        <td>
          namespaces specifies a static list of namespace names that the term applies to.
The term is applied to the union of the namespaces listed in this field
and the ones selected by namespaceSelector.
null or empty namespaces list and null namespaceSelector means "this pod's namespace".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[index].podAffinityTerm.labelSelector
<sup><sup>[↩ Parent](#redroidinstancespecaffinitypodaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinityterm)</sup></sup>



A label query over a set of resources, in this case pods.
If it's null, this PodAffinityTerm matches with no Pods.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidinstancespecaffinitypodaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinitytermlabelselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[index].podAffinityTerm.labelSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#redroidinstancespecaffinitypodaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinitytermlabelselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[index].podAffinityTerm.namespaceSelector
<sup><sup>[↩ Parent](#redroidinstancespecaffinitypodaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinityterm)</sup></sup>



A label query over the set of namespaces that the term applies to.
The term is applied to the union of the namespaces selected by this field
and the ones listed in the namespaces field.
null selector and null or empty namespaces list means "this pod's namespace".
An empty selector ({}) matches all namespaces.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidinstancespecaffinitypodaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinitytermnamespaceselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[index].podAffinityTerm.namespaceSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#redroidinstancespecaffinitypodaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinitytermnamespaceselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[index]
<sup><sup>[↩ Parent](#redroidinstancespecaffinitypodaffinity)</sup></sup>



Defines a set of pods (namely those matching the labelSelector
relative to the given namespace(s)) that this pod should be
co-located (affinity) or not co-located (anti-affinity) with,
where co-located is defined as running on a node whose value of
the label with key <topologyKey> matches that of any node on which
a pod of the set of pods is running

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>topologyKey</b></td>
        <td>string</td>
        <td>
          This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching
the labelSelector in the specified namespaces, where co-located is defined as running on a node
whose value of the label with key topologyKey matches that of any node on which any of the
selected pods is running.
Empty topologyKey is not allowed.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespecaffinitypodaffinityrequiredduringschedulingignoredduringexecutionindexlabelselector">labelSelector</a></b></td>
        <td>object</td>
        <td>
          A label query over a set of resources, in this case pods.
If it's null, this PodAffinityTerm matches with no Pods.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabelKeys</b></td>
        <td>[]string</td>
        <td>
          MatchLabelKeys is a set of pod label keys to select which pods will
be taken into consideration. The keys are used to lookup values from the
incoming pod labels, those key-value labels are merged with `LabelSelector` as `key in (value)`
to select the group of existing pods which pods will be taken into consideration
for the incoming pod's pod (anti) affinity. Keys that don't exist in the incoming
pod labels will be ignored. The default value is empty.
The same key is forbidden to exist in both MatchLabelKeys and LabelSelector.
Also, MatchLabelKeys cannot be set when LabelSelector isn't set.
This is an alpha field and requires enabling MatchLabelKeysInPodAffinity feature gate.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>mismatchLabelKeys</b></td>
        <td>[]string</td>
        <td>
          MismatchLabelKeys is a set of pod label keys to select which pods will
be taken into consideration. The keys are used to lookup values from the
incoming pod labels, those key-value labels are merged with `LabelSelector` as `key notin (value)`
to select the group of existing pods which pods will be taken into consideration
for the incoming pod's pod (anti) affinity. Keys that don't exist in the incoming
pod labels will be ignored. The default value is empty.
The same key is forbidden to exist in both MismatchLabelKeys and LabelSelector.
Also, MismatchLabelKeys cannot be set when LabelSelector isn't set.
This is an alpha field and requires enabling MatchLabelKeysInPodAffinity feature gate.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespecaffinitypodaffinityrequiredduringschedulingignoredduringexecutionindexnamespaceselector">namespaceSelector</a></b></td>
        <td>object</td>
        <td>
          A label query over the set of namespaces that the term applies to.
The term is applied to the union of the namespaces selected by this field
and the ones listed in the namespaces field.
null selector and null or empty namespaces list means "this pod's namespace".
An empty selector ({}) matches all namespaces.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespaces</b></td>
        <td>[]string</td>
        <td>
          namespaces specifies a static list of namespace names that the term applies to.
The term is applied to the union of the namespaces listed in this field
and the ones selected by namespaceSelector.
null or empty namespaces list and null namespaceSelector means "this pod's namespace".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[index].labelSelector
<sup><sup>[↩ Parent](#redroidinstancespecaffinitypodaffinityrequiredduringschedulingignoredduringexecutionindex)</sup></sup>



A label query over a set of resources, in this case pods.
If it's null, this PodAffinityTerm matches with no Pods.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidinstancespecaffinitypodaffinityrequiredduringschedulingignoredduringexecutionindexlabelselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[index].labelSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#redroidinstancespecaffinitypodaffinityrequiredduringschedulingignoredduringexecutionindexlabelselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[index].namespaceSelector
<sup><sup>[↩ Parent](#redroidinstancespecaffinitypodaffinityrequiredduringschedulingignoredduringexecutionindex)</sup></sup>



A label query over the set of namespaces that the term applies to.
The term is applied to the union of the namespaces selected by this field
and the ones listed in the namespaces field.
null selector and null or empty namespaces list means "this pod's namespace".
An empty selector ({}) matches all namespaces.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidinstancespecaffinitypodaffinityrequiredduringschedulingignoredduringexecutionindexnamespaceselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[index].namespaceSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#redroidinstancespecaffinitypodaffinityrequiredduringschedulingignoredduringexecutionindexnamespaceselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.podAntiAffinity
<sup><sup>[↩ Parent](#redroidinstancespecaffinity)</sup></sup>



Describes pod anti-affinity scheduling rules (e.g. avoid putting this pod in the same node, zone, etc. as some other pod(s)).

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidinstancespecaffinitypodantiaffinitypreferredduringschedulingignoredduringexecutionindex">preferredDuringSchedulingIgnoredDuringExecution</a></b></td>
        <td>[]object</td>
        <td>
          The scheduler will prefer to schedule pods to nodes that satisfy
the anti-affinity expressions specified by this field, but it may choose
a node that violates one or more of the expressions. The node that is
most preferred is the one with the greatest sum of weights, i.e.
for each node that meets all of the scheduling requirements (resource
request, requiredDuringScheduling anti-affinity expressions, etc.),
compute a sum by iterating through the elements of this field and adding
"weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the
node(s) with the highest sum are the most preferred.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespecaffinitypodantiaffinityrequiredduringschedulingignoredduringexecutionindex">requiredDuringSchedulingIgnoredDuringExecution</a></b></td>
        <td>[]object</td>
        <td>
          If the anti-affinity requirements specified by this field are not met at
scheduling time, the pod will not be scheduled onto the node.
If the anti-affinity requirements specified by this field cease to be met
at some point during pod execution (e.g. due to a pod label update), the
system may or may not try to eventually evict the pod from its node.
When there are multiple elements, the lists of nodes corresponding to each
podAffinityTerm are intersected, i.e. all terms must be satisfied.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[index]
<sup><sup>[↩ Parent](#redroidinstancespecaffinitypodantiaffinity)</sup></sup>



The weights of all of the matched WeightedPodAffinityTerm fields are added per-node to find the most preferred node(s)

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidinstancespecaffinitypodantiaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinityterm">podAffinityTerm</a></b></td>
        <td>object</td>
        <td>
          Required. A pod affinity term, associated with the corresponding weight.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>weight</b></td>
        <td>integer</td>
        <td>
          weight associated with matching the corresponding podAffinityTerm,
in the range 1-100.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[index].podAffinityTerm
<sup><sup>[↩ Parent](#redroidinstancespecaffinitypodantiaffinitypreferredduringschedulingignoredduringexecutionindex)</sup></sup>



Required. A pod affinity term, associated with the corresponding weight.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>topologyKey</b></td>
        <td>string</td>
        <td>
          This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching
the labelSelector in the specified namespaces, where co-located is defined as running on a node
whose value of the label with key topologyKey matches that of any node on which any of the
selected pods is running.
Empty topologyKey is not allowed.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespecaffinitypodantiaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinitytermlabelselector">labelSelector</a></b></td>
        <td>object</td>
        <td>
          A label query over a set of resources, in this case pods.
If it's null, this PodAffinityTerm matches with no Pods.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabelKeys</b></td>
        <td>[]string</td>
        <td>
          MatchLabelKeys is a set of pod label keys to select which pods will
be taken into consideration. The keys are used to lookup values from the
incoming pod labels, those key-value labels are merged with `LabelSelector` as `key in (value)`
to select the group of existing pods which pods will be taken into consideration
for the incoming pod's pod (anti) affinity. Keys that don't exist in the incoming
pod labels will be ignored. The default value is empty.
The same key is forbidden to exist in both MatchLabelKeys and LabelSelector.
Also, MatchLabelKeys cannot be set when LabelSelector isn't set.
This is an alpha field and requires enabling MatchLabelKeysInPodAffinity feature gate.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>mismatchLabelKeys</b></td>
        <td>[]string</td>
        <td>
          MismatchLabelKeys is a set of pod label keys to select which pods will
be taken into consideration. The keys are used to lookup values from the
incoming pod labels, those key-value labels are merged with `LabelSelector` as `key notin (value)`
to select the group of existing pods which pods will be taken into consideration
for the incoming pod's pod (anti) affinity. Keys that don't exist in the incoming
pod labels will be ignored. The default value is empty.
The same key is forbidden to exist in both MismatchLabelKeys and LabelSelector.
Also, MismatchLabelKeys cannot be set when LabelSelector isn't set.
This is an alpha field and requires enabling MatchLabelKeysInPodAffinity feature gate.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespecaffinitypodantiaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinitytermnamespaceselector">namespaceSelector</a></b></td>
        <td>object</td>
        <td>
          A label query over the set of namespaces that the term applies to.
The term is applied to the union of the namespaces selected by this field
and the ones listed in the namespaces field.
null selector and null or empty namespaces list means "this pod's namespace".
An empty selector ({}) matches all namespaces.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespaces</b></td>
        <td>[]string</td>
        <td>
          namespaces specifies a static list of namespace names that the term applies to.
The term is applied to the union of the namespaces listed in this field
and the ones selected by namespaceSelector.
null or empty namespaces list and null namespaceSelector means "this pod's namespace".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[index].podAffinityTerm.labelSelector
<sup><sup>[↩ Parent](#redroidinstancespecaffinitypodantiaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinityterm)</sup></sup>



A label query over a set of resources, in this case pods.
If it's null, this PodAffinityTerm matches with no Pods.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidinstancespecaffinitypodantiaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinitytermlabelselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[index].podAffinityTerm.labelSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#redroidinstancespecaffinitypodantiaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinitytermlabelselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[index].podAffinityTerm.namespaceSelector
<sup><sup>[↩ Parent](#redroidinstancespecaffinitypodantiaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinityterm)</sup></sup>



A label query over the set of namespaces that the term applies to.
The term is applied to the union of the namespaces selected by this field
and the ones listed in the namespaces field.
null selector and null or empty namespaces list means "this pod's namespace".
An empty selector ({}) matches all namespaces.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidinstancespecaffinitypodantiaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinitytermnamespaceselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[index].podAffinityTerm.namespaceSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#redroidinstancespecaffinitypodantiaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinitytermnamespaceselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[index]
<sup><sup>[↩ Parent](#redroidinstancespecaffinitypodantiaffinity)</sup></sup>



Defines a set of pods (namely those matching the labelSelector
relative to the given namespace(s)) that this pod should be
co-located (affinity) or not co-located (anti-affinity) with,
where co-located is defined as running on a node whose value of
the label with key <topologyKey> matches that of any node on which
a pod of the set of pods is running

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>topologyKey</b></td>
        <td>string</td>
        <td>
          This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching
the labelSelector in the specified namespaces, where co-located is defined as running on a node
whose value of the label with key topologyKey matches that of any node on which any of the
selected pods is running.
Empty topologyKey is not allowed.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespecaffinitypodantiaffinityrequiredduringschedulingignoredduringexecutionindexlabelselector">labelSelector</a></b></td>
        <td>object</td>
        <td>
          A label query over a set of resources, in this case pods.
If it's null, this PodAffinityTerm matches with no Pods.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabelKeys</b></td>
        <td>[]string</td>
        <td>
          MatchLabelKeys is a set of pod label keys to select which pods will
be taken into consideration. The keys are used to lookup values from the
incoming pod labels, those key-value labels are merged with `LabelSelector` as `key in (value)`
to select the group of existing pods which pods will be taken into consideration
for the incoming pod's pod (anti) affinity. Keys that don't exist in the incoming
pod labels will be ignored. The default value is empty.
The same key is forbidden to exist in both MatchLabelKeys and LabelSelector.
Also, MatchLabelKeys cannot be set when LabelSelector isn't set.
This is an alpha field and requires enabling MatchLabelKeysInPodAffinity feature gate.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>mismatchLabelKeys</b></td>
        <td>[]string</td>
        <td>
          MismatchLabelKeys is a set of pod label keys to select which pods will
be taken into consideration. The keys are used to lookup values from the
incoming pod labels, those key-value labels are merged with `LabelSelector` as `key notin (value)`
to select the group of existing pods which pods will be taken into consideration
for the incoming pod's pod (anti) affinity. Keys that don't exist in the incoming
pod labels will be ignored. The default value is empty.
The same key is forbidden to exist in both MismatchLabelKeys and LabelSelector.
Also, MismatchLabelKeys cannot be set when LabelSelector isn't set.
This is an alpha field and requires enabling MatchLabelKeysInPodAffinity feature gate.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespecaffinitypodantiaffinityrequiredduringschedulingignoredduringexecutionindexnamespaceselector">namespaceSelector</a></b></td>
        <td>object</td>
        <td>
          A label query over the set of namespaces that the term applies to.
The term is applied to the union of the namespaces selected by this field
and the ones listed in the namespaces field.
null selector and null or empty namespaces list means "this pod's namespace".
An empty selector ({}) matches all namespaces.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespaces</b></td>
        <td>[]string</td>
        <td>
          namespaces specifies a static list of namespace names that the term applies to.
The term is applied to the union of the namespaces listed in this field
and the ones selected by namespaceSelector.
null or empty namespaces list and null namespaceSelector means "this pod's namespace".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[index].labelSelector
<sup><sup>[↩ Parent](#redroidinstancespecaffinitypodantiaffinityrequiredduringschedulingignoredduringexecutionindex)</sup></sup>



A label query over a set of resources, in this case pods.
If it's null, this PodAffinityTerm matches with no Pods.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidinstancespecaffinitypodantiaffinityrequiredduringschedulingignoredduringexecutionindexlabelselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[index].labelSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#redroidinstancespecaffinitypodantiaffinityrequiredduringschedulingignoredduringexecutionindexlabelselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[index].namespaceSelector
<sup><sup>[↩ Parent](#redroidinstancespecaffinitypodantiaffinityrequiredduringschedulingignoredduringexecutionindex)</sup></sup>



A label query over the set of namespaces that the term applies to.
The term is applied to the union of the namespaces selected by this field
and the ones listed in the namespaces field.
null selector and null or empty namespaces list means "this pod's namespace".
An empty selector ({}) matches all namespaces.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidinstancespecaffinitypodantiaffinityrequiredduringschedulingignoredduringexecutionindexnamespaceselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[index].namespaceSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#redroidinstancespecaffinitypodantiaffinityrequiredduringschedulingignoredduringexecutionindexnamespaceselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.extraEnv[index]
<sup><sup>[↩ Parent](#redroidinstancespec)</sup></sup>



EnvVar represents an environment variable present in a Container.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the environment variable. Must be a C_IDENTIFIER.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Variable references $(VAR_NAME) are expanded
using the previously defined environment variables in the container and
any service environment variables. If a variable cannot be resolved,
the reference in the input string will be unchanged. Double $$ are reduced
to a single $, which allows for escaping the $(VAR_NAME) syntax: i.e.
"$$(VAR_NAME)" will produce the string literal "$(VAR_NAME)".
Escaped references will never be expanded, regardless of whether the variable
exists or not.
Defaults to "".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespecextraenvindexvaluefrom">valueFrom</a></b></td>
        <td>object</td>
        <td>
          Source for the environment variable's value. Cannot be used if value is not empty.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.extraEnv[index].valueFrom
<sup><sup>[↩ Parent](#redroidinstancespecextraenvindex)</sup></sup>



Source for the environment variable's value. Cannot be used if value is not empty.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidinstancespecextraenvindexvaluefromconfigmapkeyref">configMapKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a ConfigMap.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespecextraenvindexvaluefromfieldref">fieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a field of the pod: supports metadata.name, metadata.namespace, `metadata.labels['<KEY>']`, `metadata.annotations['<KEY>']`,
spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespecextraenvindexvaluefromresourcefieldref">resourceFieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a resource of the container: only resources limits and requests
(limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespecextraenvindexvaluefromsecretkeyref">secretKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a secret in the pod's namespace<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.extraEnv[index].valueFrom.configMapKeyRef
<sup><sup>[↩ Parent](#redroidinstancespecextraenvindexvaluefrom)</sup></sup>



Selects a key of a ConfigMap.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.extraEnv[index].valueFrom.fieldRef
<sup><sup>[↩ Parent](#redroidinstancespecextraenvindexvaluefrom)</sup></sup>



Selects a field of the pod: supports metadata.name, metadata.namespace, `metadata.labels['<KEY>']`, `metadata.annotations['<KEY>']`,
spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>fieldPath</b></td>
        <td>string</td>
        <td>
          Path of the field to select in the specified API version.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          Version of the schema the FieldPath is written in terms of, defaults to "v1".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.extraEnv[index].valueFrom.resourceFieldRef
<sup><sup>[↩ Parent](#redroidinstancespecextraenvindexvaluefrom)</sup></sup>



Selects a resource of the container: only resources limits and requests
(limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>resource</b></td>
        <td>string</td>
        <td>
          Required: resource to select<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>containerName</b></td>
        <td>string</td>
        <td>
          Container name: required for volumes, optional for env vars<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>divisor</b></td>
        <td>int or string</td>
        <td>
          Specifies the output format of the exposed resources, defaults to "1"<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.extraEnv[index].valueFrom.secretKeyRef
<sup><sup>[↩ Parent](#redroidinstancespecextraenvindexvaluefrom)</sup></sup>



Selects a key of a secret in the pod's namespace

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.imagePullSecrets[index]
<sup><sup>[↩ Parent](#redroidinstancespec)</sup></sup>



LocalObjectReference contains enough information to let you locate the
referenced object inside the same namespace.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.network
<sup><sup>[↩ Parent](#redroidinstancespec)</sup></sup>



Network configures DNS and proxy settings for the Android instance.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>dns</b></td>
        <td>[]string</td>
        <td>
          DNS lists DNS server addresses (up to N servers).
They map to androidboot.redroid_net_ndns and androidboot.redroid_net_dns<1..N>.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancespecnetworkproxy">proxy</a></b></td>
        <td>object</td>
        <td>
          Proxy configures HTTP/HTTPS proxy for outbound traffic.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.network.proxy
<sup><sup>[↩ Parent](#redroidinstancespecnetwork)</sup></sup>



Proxy configures HTTP/HTTPS proxy for outbound traffic.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>excludeList</b></td>
        <td>string</td>
        <td>
          ExcludeList is a comma-separated list of hosts bypassing the proxy.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          Host is the proxy server hostname or IP (used with Type=static).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>pac</b></td>
        <td>string</td>
        <td>
          PAC is the URL of a proxy auto-config file (used with Type=pac).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          Port is the proxy server port. Defaults to 3128.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 3128<br/>
            <i>Minimum</i>: 1<br/>
            <i>Maximum</i>: 65535<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type selects the proxy mode: static, pac, none, or unassigned.<br/>
          <br/>
            <i>Enum</i>: static, pac, none, unassigned<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.resources
<sup><sup>[↩ Parent](#redroidinstancespec)</sup></sup>



Resources optionally sets CPU/memory limits on the redroid container.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidinstancespecresourcesclaimsindex">claims</a></b></td>
        <td>[]object</td>
        <td>
          Claims lists the names of resources, defined in spec.resourceClaims,
that are used by this container.

This is an alpha field and requires enabling the
DynamicResourceAllocation feature gate.

This field is immutable. It can only be set for containers.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>limits</b></td>
        <td>map[string]int or string</td>
        <td>
          Limits describes the maximum amount of compute resources allowed.
More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>requests</b></td>
        <td>map[string]int or string</td>
        <td>
          Requests describes the minimum amount of compute resources required.
If Requests is omitted for a container, it defaults to Limits if that is explicitly specified,
otherwise to an implementation-defined value. Requests cannot exceed Limits.
More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.resources.claims[index]
<sup><sup>[↩ Parent](#redroidinstancespecresources)</sup></sup>



ResourceClaim references one entry in PodSpec.ResourceClaims.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name must match the name of one entry in pod.spec.resourceClaims of
the Pod where this field is used. It makes that resource available
inside a container.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.screen
<sup><sup>[↩ Parent](#redroidinstancespec)</sup></sup>



Screen configures the virtual display (width, height, dpi, fps).

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>dpi</b></td>
        <td>integer</td>
        <td>
          DPI is the screen density. Defaults to 320.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>fps</b></td>
        <td>integer</td>
        <td>
          FPS is the display frame rate. Defaults to 30 (Android 12+) or 15 (older).<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>height</b></td>
        <td>integer</td>
        <td>
          Height is the screen height in pixels. Defaults to 1280.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>width</b></td>
        <td>integer</td>
        <td>
          Width is the screen width in pixels. Defaults to 720.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.service
<sup><sup>[↩ Parent](#redroidinstancespec)</sup></sup>



Service customises the Kubernetes Service the operator creates to expose the ADB port.
The Service is always created; this field controls its type, annotations, and node port.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>annotations</b></td>
        <td>map[string]string</td>
        <td>
          Annotations are extra annotations merged onto the Service.
Useful for cloud-provider-specific behaviour, e.g. AWS NLB, GCP NEG.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>nodePort</b></td>
        <td>integer</td>
        <td>
          NodePort pins the node port when Type=NodePort.
If 0 or unset, Kubernetes auto-assigns a port from the node-port range.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 1<br/>
            <i>Maximum</i>: 65535<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type is the Service type. Defaults to ClusterIP.
Use NodePort or LoadBalancer to expose the ADB port outside the cluster.<br/>
          <br/>
            <i>Enum</i>: ClusterIP, NodePort, LoadBalancer<br/>
            <i>Default</i>: ClusterIP<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.spec.tolerations[index]
<sup><sup>[↩ Parent](#redroidinstancespec)</sup></sup>



The pod this Toleration is attached to tolerates any taint that matches
the triple <key,value,effect> using the matching operator <operator>.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>effect</b></td>
        <td>string</td>
        <td>
          Effect indicates the taint effect to match. Empty means match all taint effects.
When specified, allowed values are NoSchedule, PreferNoSchedule and NoExecute.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          Key is the taint key that the toleration applies to. Empty means match all taint keys.
If the key is empty, operator must be Exists; this combination means to match all values and all keys.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          Operator represents a key's relationship to the value.
Valid operators are Exists and Equal. Defaults to Equal.
Exists is equivalent to wildcard for value, so that a pod can
tolerate all taints of a particular category.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>tolerationSeconds</b></td>
        <td>integer</td>
        <td>
          TolerationSeconds represents the period of time the toleration (which must be
of effect NoExecute, otherwise this field is ignored) tolerates the taint. By default,
it is not set, which means tolerate the taint forever (do not evict). Zero and
negative values will be treated as 0 (evict immediately) by the system.<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Value is the taint value the toleration matches to.
If the operator is Exists, the value should be empty, otherwise just a regular string.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.status
<sup><sup>[↩ Parent](#redroidinstance)</sup></sup>



RedroidInstanceStatus defines the observed state of RedroidInstance.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>adbAddress</b></td>
        <td>string</td>
        <td>
          ADBAddress is the in-cluster address (host:port) to reach this instance's ADB.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancestatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          Conditions holds detailed status conditions (Ready, Scheduled).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          ObservedGeneration is the most recent generation observed by the controller.<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>phase</b></td>
        <td>enum</td>
        <td>
          Phase is the current lifecycle phase.<br/>
          <br/>
            <i>Enum</i>: Pending, Running, Stopped, Failed<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>podName</b></td>
        <td>string</td>
        <td>
          PodName is the name of the managed Pod, if it exists.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancestatussuspended">suspended</a></b></td>
        <td>object</td>
        <td>
          Suspended holds a suspend-override override set by controllers or manual operators.
The instance Pod is stopped while this field is non-nil, regardless of spec.suspend.
Unlike spec.suspend, this field is not reconciled by GitOps tools (e.g. Flux) so clearing
it does not cause config drift.

To manually suspend: kubectl patch redroidinstance <name> --subresource=status
  --type=merge -p '{"status":{"suspended":{"reason":"maintenance","actor":"manual"}}}'
To clear: kubectl patch redroidinstance <name> --subresource=status
  --type=merge -p '{"status":{"suspended":null}}'<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidinstancestatuswoken">woken</a></b></td>
        <td>object</td>
        <td>
          Woken holds a wake-override set by the RedroidTask controller when spec.wakeInstance
is true. The instance Pod runs while this field is non-nil, even if spec.suspend is true.
Unlike spec.suspend, this field is not reconciled by GitOps tools (e.g. Flux).

To manually wake: kubectl patch redroidinstance <name> --subresource=status
  --type=merge -p '{"status":{"woken":{"reason":"on-demand","actor":"manual"}}}'
To clear: kubectl patch redroidinstance <name> --subresource=status
  --type=merge -p '{"status":{"woken":null}}'<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.status.conditions[index]
<sup><sup>[↩ Parent](#redroidinstancestatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.status.suspended
<sup><sup>[↩ Parent](#redroidinstancestatus)</sup></sup>



Suspended holds a suspend-override override set by controllers or manual operators.
The instance Pod is stopped while this field is non-nil, regardless of spec.suspend.
Unlike spec.suspend, this field is not reconciled by GitOps tools (e.g. Flux) so clearing
it does not cause config drift.

To manually suspend: kubectl patch redroidinstance <name> --subresource=status
  --type=merge -p '{"status":{"suspended":{"reason":"maintenance","actor":"manual"}}}'
To clear: kubectl patch redroidinstance <name> --subresource=status
  --type=merge -p '{"status":{"suspended":null}}'

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>actor</b></td>
        <td>string</td>
        <td>
          Actor identifies who set the temporary suspend, e.g. "manual", "task/maa-task".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          Reason is a human-readable explanation for the temporary suspend.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>until</b></td>
        <td>string</td>
        <td>
          Until is an optional expiry timestamp. The controller automatically clears the
temporary suspend (and restarts the Pod) once this time has passed.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidInstance.status.woken
<sup><sup>[↩ Parent](#redroidinstancestatus)</sup></sup>



Woken holds a wake-override set by the RedroidTask controller when spec.wakeInstance
is true. The instance Pod runs while this field is non-nil, even if spec.suspend is true.
Unlike spec.suspend, this field is not reconciled by GitOps tools (e.g. Flux).

To manually wake: kubectl patch redroidinstance <name> --subresource=status
  --type=merge -p '{"status":{"woken":{"reason":"on-demand","actor":"manual"}}}'
To clear: kubectl patch redroidinstance <name> --subresource=status
  --type=merge -p '{"status":{"woken":null}}'

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>actor</b></td>
        <td>string</td>
        <td>
          Actor identifies who set the temporary wake, e.g. "task/maa-task".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          Reason is a human-readable explanation for the temporary wake.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>until</b></td>
        <td>string</td>
        <td>
          Until is an optional expiry timestamp. The controller automatically clears the
temporary wake (and lets spec.suspend take effect again) once this time has passed.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## RedroidTask
<sup><sup>[↩ Parent](#redroidiov1alpha1 )</sup></sup>






RedroidTask is the Schema for the redroidtasks API.
It describes a workload (one-shot or recurring via CronJob) that runs integration
tool containers against a set of RedroidInstance overlay partitions.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>redroid.io/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>RedroidTask</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspec">spec</a></b></td>
        <td>object</td>
        <td>
          RedroidTaskSpec defines the desired state of RedroidTask.<br/>
          <br/>
            <i>Validations</i>:<li>!(self.suspendInstance == true && self.wakeInstance == true): suspendInstance and wakeInstance are mutually exclusive</li>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskstatus">status</a></b></td>
        <td>object</td>
        <td>
          RedroidTaskStatus defines the observed state of RedroidTask.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec
<sup><sup>[↩ Parent](#redroidtask)</sup></sup>



RedroidTaskSpec defines the desired state of RedroidTask.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidtaskspecinstancesindex">instances</a></b></td>
        <td>[]object</td>
        <td>
          Instances lists the RedroidInstance resources that this task targets.
Each instance runs its own Job/CronJob execution, inheriting
the instance's image, overlayfs PVCs, and GPU settings.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecintegrationsindex">integrations</a></b></td>
        <td>[]object</td>
        <td>
          Integrations is the ordered list of tool containers executed per instance.
They run as regular containers alongside the redroid sidecar, sharing localhost.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>activeDeadlineSeconds</b></td>
        <td>integer</td>
        <td>
          ActiveDeadlineSeconds limits the duration of each Job in seconds.
Jobs exceeding this limit are terminated. 0 means no limit.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>backoffLimit</b></td>
        <td>integer</td>
        <td>
          BackoffLimit specifies the number of retries before marking the Job as failed.
Defaults to 0 (no retries).<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 0<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>failedJobsHistoryLimit</b></td>
        <td>integer</td>
        <td>
          FailedJobsHistoryLimit controls how many failed CronJob-spawned Jobs to retain. Defaults to 3.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 3<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecimagepullsecretsindex">imagePullSecrets</a></b></td>
        <td>[]object</td>
        <td>
          ImagePullSecrets applies to all integration containers in this task.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>parallelism</b></td>
        <td>integer</td>
        <td>
          Parallelism controls how many instance Pods run concurrently.
Defaults to the number of Instances (run all in parallel).<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>schedule</b></td>
        <td>string</td>
        <td>
          Schedule is a Cron expression for recurring execution (e.g. "0 4 * * *").
If empty, the task is one-shot and runs immediately upon creation.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>serviceAccountName</b></td>
        <td>string</td>
        <td>
          ServiceAccountName is the name of the ServiceAccount to use for all Pods
created by this task. Applies at the PodSpec level and is shared by every
integration container. If empty, the namespace default ServiceAccount is used.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>startingDeadlineSeconds</b></td>
        <td>integer</td>
        <td>
          StartingDeadlineSeconds is the optional deadline (seconds) for starting a CronJob
if it misses scheduled time. Ignored for one-shot tasks.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>successfulJobsHistoryLimit</b></td>
        <td>integer</td>
        <td>
          SuccessfulJobsHistoryLimit controls how many successful CronJob-spawned Jobs to retain. Defaults to 3.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 3<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>suspend</b></td>
        <td>boolean</td>
        <td>
          Suspend pauses CronJob execution without deleting it. Ignored for one-shot tasks.
Follows the same semantics as CronJob.spec.suspend.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>suspendInstance</b></td>
        <td>boolean</td>
        <td>
          SuspendInstance temporarily stops the referenced RedroidInstance Pod while the
one-shot Job runs, then automatically restores it on Job completion/failure.
This prevents overlayfs conflicts when both the instance Pod and the task Job
would otherwise write to the same /data-diff volume simultaneously.

Mechanism: the task controller sets status.suspended on each instance
before creating the Job, waits until the instance Pod is stopped (phase=Stopped),
then creates the Job. After the Job finishes the controller clears the field.

This field is ignored for CronJob-based (scheduled) tasks.
Mutually exclusive with WakeInstance.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>timezone</b></td>
        <td>string</td>
        <td>
          Timezone is the IANA timezone name for the CronJob schedule (e.g. "Asia/Shanghai").
Requires Kubernetes >= 1.27. Ignored for one-shot tasks.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>ttlSecondsAfterFinished</b></td>
        <td>integer</td>
        <td>
          TTLSecondsAfterFinished automatically removes completed one-shot Jobs after
this many seconds. Ignored for scheduled tasks.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindex">volumes</a></b></td>
        <td>[]object</td>
        <td>
          Volumes defines additional volumes to attach to the Job Pod.
Use this together with integration VolumeMounts to mount arbitrary volume
sources (Secrets, projected ConfigMaps, emptyDir, etc.) that are not covered
by the per-key Configs shorthand.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>wakeInstance</b></td>
        <td>boolean</td>
        <td>
          WakeInstance temporarily starts the referenced RedroidInstance Pod while the
one-shot Job runs, then restores the original suspended state on Job completion/failure.
Use this for on-demand instances that are normally stopped (spec.suspend: true).

Mechanism: the task controller sets status.woken on each instance before creating the
Job, waits until the instance Pod reaches phase=Running, then creates the Job.
After the Job finishes the controller clears status.woken, allowing spec.suspend to
take effect again.

This field is ignored for CronJob-based (scheduled) tasks.
Mutually exclusive with SuspendInstance.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index]
<sup><sup>[↩ Parent](#redroidtaskspec)</sup></sup>



InstanceRef selects a RedroidInstance by name to include in a task run.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the RedroidInstance name in the same namespace.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumemountsindex">volumeMounts</a></b></td>
        <td>[]object</td>
        <td>
          VolumeMounts adds extra volume mounts into every integration container
for this instance only. Use together with instance-level Volumes to
mount instance-specific ConfigMaps, Secrets, etc.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindex">volumes</a></b></td>
        <td>[]object</td>
        <td>
          Volumes adds additional volumes specific to this instance's Job Pod.
These are merged with task-level spec.volumes; an instance volume overrides
only user-defined task-level volumes with the same name. Reserved volumes
(data-base, data-diff, dev-dri) and controller-generated ConfigMap volumes
(cm-* prefix) are never overrideable — those retain precedence regardless.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumeMounts[index]
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindex)</sup></sup>



VolumeMount describes a mounting of a Volume within a container.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>mountPath</b></td>
        <td>string</td>
        <td>
          Path within the container at which the volume should be mounted.  Must
not contain ':'.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          This must match the Name of a Volume.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>mountPropagation</b></td>
        <td>string</td>
        <td>
          mountPropagation determines how mounts are propagated from the host
to container and the other way around.
When not set, MountPropagationNone is used.
This field is beta in 1.10.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          Mounted read-only if true, read-write otherwise (false or unspecified).
Defaults to false.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>subPath</b></td>
        <td>string</td>
        <td>
          Path within the volume from which the container's volume should be mounted.
Defaults to "" (volume's root).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>subPathExpr</b></td>
        <td>string</td>
        <td>
          Expanded path within the volume from which the container's volume should be mounted.
Behaves similarly to SubPath but environment variable references $(VAR_NAME) are expanded using the container's environment.
Defaults to "" (volume's root).
SubPathExpr and SubPath are mutually exclusive.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index]
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindex)</sup></sup>



Volume represents a named volume in a pod that may be accessed by any container in the pod.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          name of the volume.
Must be a DNS_LABEL and unique within the pod.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexawselasticblockstore">awsElasticBlockStore</a></b></td>
        <td>object</td>
        <td>
          awsElasticBlockStore represents an AWS Disk resource that is attached to a
kubelet's host machine and then exposed to the pod.
More info: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexazuredisk">azureDisk</a></b></td>
        <td>object</td>
        <td>
          azureDisk represents an Azure Data Disk mount on the host and bind mount to the pod.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexazurefile">azureFile</a></b></td>
        <td>object</td>
        <td>
          azureFile represents an Azure File Service mount on the host and bind mount to the pod.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexcephfs">cephfs</a></b></td>
        <td>object</td>
        <td>
          cephFS represents a Ceph FS mount on the host that shares a pod's lifetime<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexcinder">cinder</a></b></td>
        <td>object</td>
        <td>
          cinder represents a cinder volume attached and mounted on kubelets host machine.
More info: https://examples.k8s.io/mysql-cinder-pd/README.md<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexconfigmap">configMap</a></b></td>
        <td>object</td>
        <td>
          configMap represents a configMap that should populate this volume<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexcsi">csi</a></b></td>
        <td>object</td>
        <td>
          csi (Container Storage Interface) represents ephemeral storage that is handled by certain external CSI drivers (Beta feature).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexdownwardapi">downwardAPI</a></b></td>
        <td>object</td>
        <td>
          downwardAPI represents downward API about the pod that should populate this volume<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexemptydir">emptyDir</a></b></td>
        <td>object</td>
        <td>
          emptyDir represents a temporary directory that shares a pod's lifetime.
More info: https://kubernetes.io/docs/concepts/storage/volumes#emptydir<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexephemeral">ephemeral</a></b></td>
        <td>object</td>
        <td>
          ephemeral represents a volume that is handled by a cluster storage driver.
The volume's lifecycle is tied to the pod that defines it - it will be created before the pod starts,
and deleted when the pod is removed.

Use this if:
a) the volume is only needed while the pod runs,
b) features of normal volumes like restoring from snapshot or capacity
   tracking are needed,
c) the storage driver is specified through a storage class, and
d) the storage driver supports dynamic volume provisioning through
   a PersistentVolumeClaim (see EphemeralVolumeSource for more
   information on the connection between this volume type
   and PersistentVolumeClaim).

Use PersistentVolumeClaim or one of the vendor-specific
APIs for volumes that persist for longer than the lifecycle
of an individual pod.

Use CSI for light-weight local ephemeral volumes if the CSI driver is meant to
be used that way - see the documentation of the driver for
more information.

A pod can use both types of ephemeral volumes and
persistent volumes at the same time.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexfc">fc</a></b></td>
        <td>object</td>
        <td>
          fc represents a Fibre Channel resource that is attached to a kubelet's host machine and then exposed to the pod.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexflexvolume">flexVolume</a></b></td>
        <td>object</td>
        <td>
          flexVolume represents a generic volume resource that is
provisioned/attached using an exec based plugin.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexflocker">flocker</a></b></td>
        <td>object</td>
        <td>
          flocker represents a Flocker volume attached to a kubelet's host machine. This depends on the Flocker control service being running<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexgcepersistentdisk">gcePersistentDisk</a></b></td>
        <td>object</td>
        <td>
          gcePersistentDisk represents a GCE Disk resource that is attached to a
kubelet's host machine and then exposed to the pod.
More info: https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexgitrepo">gitRepo</a></b></td>
        <td>object</td>
        <td>
          gitRepo represents a git repository at a particular revision.
DEPRECATED: GitRepo is deprecated. To provision a container with a git repo, mount an
EmptyDir into an InitContainer that clones the repo using git, then mount the EmptyDir
into the Pod's container.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexglusterfs">glusterfs</a></b></td>
        <td>object</td>
        <td>
          glusterfs represents a Glusterfs mount on the host that shares a pod's lifetime.
More info: https://examples.k8s.io/volumes/glusterfs/README.md<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexhostpath">hostPath</a></b></td>
        <td>object</td>
        <td>
          hostPath represents a pre-existing file or directory on the host
machine that is directly exposed to the container. This is generally
used for system agents or other privileged things that are allowed
to see the host machine. Most containers will NOT need this.
More info: https://kubernetes.io/docs/concepts/storage/volumes#hostpath<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexiscsi">iscsi</a></b></td>
        <td>object</td>
        <td>
          iscsi represents an ISCSI Disk resource that is attached to a
kubelet's host machine and then exposed to the pod.
More info: https://examples.k8s.io/volumes/iscsi/README.md<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexnfs">nfs</a></b></td>
        <td>object</td>
        <td>
          nfs represents an NFS mount on the host that shares a pod's lifetime
More info: https://kubernetes.io/docs/concepts/storage/volumes#nfs<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexpersistentvolumeclaim">persistentVolumeClaim</a></b></td>
        <td>object</td>
        <td>
          persistentVolumeClaimVolumeSource represents a reference to a
PersistentVolumeClaim in the same namespace.
More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexphotonpersistentdisk">photonPersistentDisk</a></b></td>
        <td>object</td>
        <td>
          photonPersistentDisk represents a PhotonController persistent disk attached and mounted on kubelets host machine<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexportworxvolume">portworxVolume</a></b></td>
        <td>object</td>
        <td>
          portworxVolume represents a portworx volume attached and mounted on kubelets host machine<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexprojected">projected</a></b></td>
        <td>object</td>
        <td>
          projected items for all in one resources secrets, configmaps, and downward API<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexquobyte">quobyte</a></b></td>
        <td>object</td>
        <td>
          quobyte represents a Quobyte mount on the host that shares a pod's lifetime<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexrbd">rbd</a></b></td>
        <td>object</td>
        <td>
          rbd represents a Rados Block Device mount on the host that shares a pod's lifetime.
More info: https://examples.k8s.io/volumes/rbd/README.md<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexscaleio">scaleIO</a></b></td>
        <td>object</td>
        <td>
          scaleIO represents a ScaleIO persistent volume attached and mounted on Kubernetes nodes.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexsecret">secret</a></b></td>
        <td>object</td>
        <td>
          secret represents a secret that should populate this volume.
More info: https://kubernetes.io/docs/concepts/storage/volumes#secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexstorageos">storageos</a></b></td>
        <td>object</td>
        <td>
          storageOS represents a StorageOS volume attached and mounted on Kubernetes nodes.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexvspherevolume">vsphereVolume</a></b></td>
        <td>object</td>
        <td>
          vsphereVolume represents a vSphere volume attached and mounted on kubelets host machine<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].awsElasticBlockStore
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



awsElasticBlockStore represents an AWS Disk resource that is attached to a
kubelet's host machine and then exposed to the pod.
More info: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>volumeID</b></td>
        <td>string</td>
        <td>
          volumeID is unique ID of the persistent disk resource in AWS (Amazon EBS volume).
More info: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fsType is the filesystem type of the volume that you want to mount.
Tip: Ensure that the filesystem type is supported by the host operating system.
Examples: "ext4", "xfs", "ntfs". Implicitly inferred to be "ext4" if unspecified.
More info: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>partition</b></td>
        <td>integer</td>
        <td>
          partition is the partition in the volume that you want to mount.
If omitted, the default is to mount by volume name.
Examples: For volume /dev/sda1, you specify the partition as "1".
Similarly, the volume partition for /dev/sda is "0" (or you can leave the property empty).<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly value true will force the readOnly setting in VolumeMounts.
More info: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].azureDisk
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



azureDisk represents an Azure Data Disk mount on the host and bind mount to the pod.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>diskName</b></td>
        <td>string</td>
        <td>
          diskName is the Name of the data disk in the blob storage<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>diskURI</b></td>
        <td>string</td>
        <td>
          diskURI is the URI of data disk in the blob storage<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>cachingMode</b></td>
        <td>string</td>
        <td>
          cachingMode is the Host Caching mode: None, Read Only, Read Write.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fsType is Filesystem type to mount.
Must be a filesystem type supported by the host operating system.
Ex. "ext4", "xfs", "ntfs". Implicitly inferred to be "ext4" if unspecified.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          kind expected values are Shared: multiple blob disks per storage account  Dedicated: single blob disk per storage account  Managed: azure managed data disk (only in managed availability set). defaults to shared<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly Defaults to false (read/write). ReadOnly here will force
the ReadOnly setting in VolumeMounts.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].azureFile
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



azureFile represents an Azure File Service mount on the host and bind mount to the pod.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>secretName</b></td>
        <td>string</td>
        <td>
          secretName is the  name of secret that contains Azure Storage Account Name and Key<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>shareName</b></td>
        <td>string</td>
        <td>
          shareName is the azure share Name<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly defaults to false (read/write). ReadOnly here will force
the ReadOnly setting in VolumeMounts.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].cephfs
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



cephFS represents a Ceph FS mount on the host that shares a pod's lifetime

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>monitors</b></td>
        <td>[]string</td>
        <td>
          monitors is Required: Monitors is a collection of Ceph monitors
More info: https://examples.k8s.io/volumes/cephfs/README.md#how-to-use-it<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          path is Optional: Used as the mounted root, rather than the full Ceph tree, default is /<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly is Optional: Defaults to false (read/write). ReadOnly here will force
the ReadOnly setting in VolumeMounts.
More info: https://examples.k8s.io/volumes/cephfs/README.md#how-to-use-it<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretFile</b></td>
        <td>string</td>
        <td>
          secretFile is Optional: SecretFile is the path to key ring for User, default is /etc/ceph/user.secret
More info: https://examples.k8s.io/volumes/cephfs/README.md#how-to-use-it<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexcephfssecretref">secretRef</a></b></td>
        <td>object</td>
        <td>
          secretRef is Optional: SecretRef is reference to the authentication secret for User, default is empty.
More info: https://examples.k8s.io/volumes/cephfs/README.md#how-to-use-it<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>user</b></td>
        <td>string</td>
        <td>
          user is optional: User is the rados user name, default is admin
More info: https://examples.k8s.io/volumes/cephfs/README.md#how-to-use-it<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].cephfs.secretRef
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexcephfs)</sup></sup>



secretRef is Optional: SecretRef is reference to the authentication secret for User, default is empty.
More info: https://examples.k8s.io/volumes/cephfs/README.md#how-to-use-it

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].cinder
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



cinder represents a cinder volume attached and mounted on kubelets host machine.
More info: https://examples.k8s.io/mysql-cinder-pd/README.md

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>volumeID</b></td>
        <td>string</td>
        <td>
          volumeID used to identify the volume in cinder.
More info: https://examples.k8s.io/mysql-cinder-pd/README.md<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fsType is the filesystem type to mount.
Must be a filesystem type supported by the host operating system.
Examples: "ext4", "xfs", "ntfs". Implicitly inferred to be "ext4" if unspecified.
More info: https://examples.k8s.io/mysql-cinder-pd/README.md<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly defaults to false (read/write). ReadOnly here will force
the ReadOnly setting in VolumeMounts.
More info: https://examples.k8s.io/mysql-cinder-pd/README.md<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexcindersecretref">secretRef</a></b></td>
        <td>object</td>
        <td>
          secretRef is optional: points to a secret object containing parameters used to connect
to OpenStack.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].cinder.secretRef
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexcinder)</sup></sup>



secretRef is optional: points to a secret object containing parameters used to connect
to OpenStack.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].configMap
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



configMap represents a configMap that should populate this volume

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>defaultMode</b></td>
        <td>integer</td>
        <td>
          defaultMode is optional: mode bits used to set permissions on created files by default.
Must be an octal value between 0000 and 0777 or a decimal value between 0 and 511.
YAML accepts both octal and decimal values, JSON requires decimal values for mode bits.
Defaults to 0644.
Directories within the path are not affected by this setting.
This might be in conflict with other options that affect the file
mode, like fsGroup, and the result can be other mode bits set.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexconfigmapitemsindex">items</a></b></td>
        <td>[]object</td>
        <td>
          items if unspecified, each key-value pair in the Data field of the referenced
ConfigMap will be projected into the volume as a file whose name is the
key and content is the value. If specified, the listed keys will be
projected into the specified paths, and unlisted keys will not be
present. If a key is specified which is not present in the ConfigMap,
the volume setup will error unless it is marked optional. Paths must be
relative and may not contain the '..' path or start with '..'.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          optional specify whether the ConfigMap or its keys must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].configMap.items[index]
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexconfigmap)</sup></sup>



Maps a string key to a path within a volume.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the key to project.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          path is the relative path of the file to map the key to.
May not be an absolute path.
May not contain the path element '..'.
May not start with the string '..'.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>mode</b></td>
        <td>integer</td>
        <td>
          mode is Optional: mode bits used to set permissions on this file.
Must be an octal value between 0000 and 0777 or a decimal value between 0 and 511.
YAML accepts both octal and decimal values, JSON requires decimal values for mode bits.
If not specified, the volume defaultMode will be used.
This might be in conflict with other options that affect the file
mode, like fsGroup, and the result can be other mode bits set.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].csi
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



csi (Container Storage Interface) represents ephemeral storage that is handled by certain external CSI drivers (Beta feature).

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>driver</b></td>
        <td>string</td>
        <td>
          driver is the name of the CSI driver that handles this volume.
Consult with your admin for the correct name as registered in the cluster.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fsType to mount. Ex. "ext4", "xfs", "ntfs".
If not provided, the empty value is passed to the associated CSI driver
which will determine the default filesystem to apply.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexcsinodepublishsecretref">nodePublishSecretRef</a></b></td>
        <td>object</td>
        <td>
          nodePublishSecretRef is a reference to the secret object containing
sensitive information to pass to the CSI driver to complete the CSI
NodePublishVolume and NodeUnpublishVolume calls.
This field is optional, and  may be empty if no secret is required. If the
secret object contains more than one secret, all secret references are passed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly specifies a read-only configuration for the volume.
Defaults to false (read/write).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>volumeAttributes</b></td>
        <td>map[string]string</td>
        <td>
          volumeAttributes stores driver-specific properties that are passed to the CSI
driver. Consult your driver's documentation for supported values.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].csi.nodePublishSecretRef
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexcsi)</sup></sup>



nodePublishSecretRef is a reference to the secret object containing
sensitive information to pass to the CSI driver to complete the CSI
NodePublishVolume and NodeUnpublishVolume calls.
This field is optional, and  may be empty if no secret is required. If the
secret object contains more than one secret, all secret references are passed.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].downwardAPI
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



downwardAPI represents downward API about the pod that should populate this volume

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>defaultMode</b></td>
        <td>integer</td>
        <td>
          Optional: mode bits to use on created files by default. Must be a
Optional: mode bits used to set permissions on created files by default.
Must be an octal value between 0000 and 0777 or a decimal value between 0 and 511.
YAML accepts both octal and decimal values, JSON requires decimal values for mode bits.
Defaults to 0644.
Directories within the path are not affected by this setting.
This might be in conflict with other options that affect the file
mode, like fsGroup, and the result can be other mode bits set.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexdownwardapiitemsindex">items</a></b></td>
        <td>[]object</td>
        <td>
          Items is a list of downward API volume file<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].downwardAPI.items[index]
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexdownwardapi)</sup></sup>



DownwardAPIVolumeFile represents information to create the file containing the pod field

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          Required: Path is  the relative path name of the file to be created. Must not be absolute or contain the '..' path. Must be utf-8 encoded. The first item of the relative path must not start with '..'<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexdownwardapiitemsindexfieldref">fieldRef</a></b></td>
        <td>object</td>
        <td>
          Required: Selects a field of the pod: only annotations, labels, name and namespace are supported.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>mode</b></td>
        <td>integer</td>
        <td>
          Optional: mode bits used to set permissions on this file, must be an octal value
between 0000 and 0777 or a decimal value between 0 and 511.
YAML accepts both octal and decimal values, JSON requires decimal values for mode bits.
If not specified, the volume defaultMode will be used.
This might be in conflict with other options that affect the file
mode, like fsGroup, and the result can be other mode bits set.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexdownwardapiitemsindexresourcefieldref">resourceFieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a resource of the container: only resources limits and requests
(limits.cpu, limits.memory, requests.cpu and requests.memory) are currently supported.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].downwardAPI.items[index].fieldRef
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexdownwardapiitemsindex)</sup></sup>



Required: Selects a field of the pod: only annotations, labels, name and namespace are supported.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>fieldPath</b></td>
        <td>string</td>
        <td>
          Path of the field to select in the specified API version.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          Version of the schema the FieldPath is written in terms of, defaults to "v1".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].downwardAPI.items[index].resourceFieldRef
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexdownwardapiitemsindex)</sup></sup>



Selects a resource of the container: only resources limits and requests
(limits.cpu, limits.memory, requests.cpu and requests.memory) are currently supported.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>resource</b></td>
        <td>string</td>
        <td>
          Required: resource to select<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>containerName</b></td>
        <td>string</td>
        <td>
          Container name: required for volumes, optional for env vars<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>divisor</b></td>
        <td>int or string</td>
        <td>
          Specifies the output format of the exposed resources, defaults to "1"<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].emptyDir
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



emptyDir represents a temporary directory that shares a pod's lifetime.
More info: https://kubernetes.io/docs/concepts/storage/volumes#emptydir

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>medium</b></td>
        <td>string</td>
        <td>
          medium represents what type of storage medium should back this directory.
The default is "" which means to use the node's default medium.
Must be an empty string (default) or Memory.
More info: https://kubernetes.io/docs/concepts/storage/volumes#emptydir<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sizeLimit</b></td>
        <td>int or string</td>
        <td>
          sizeLimit is the total amount of local storage required for this EmptyDir volume.
The size limit is also applicable for memory medium.
The maximum usage on memory medium EmptyDir would be the minimum value between
the SizeLimit specified here and the sum of memory limits of all containers in a pod.
The default is nil which means that the limit is undefined.
More info: https://kubernetes.io/docs/concepts/storage/volumes#emptydir<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].ephemeral
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



ephemeral represents a volume that is handled by a cluster storage driver.
The volume's lifecycle is tied to the pod that defines it - it will be created before the pod starts,
and deleted when the pod is removed.

Use this if:
a) the volume is only needed while the pod runs,
b) features of normal volumes like restoring from snapshot or capacity
   tracking are needed,
c) the storage driver is specified through a storage class, and
d) the storage driver supports dynamic volume provisioning through
   a PersistentVolumeClaim (see EphemeralVolumeSource for more
   information on the connection between this volume type
   and PersistentVolumeClaim).

Use PersistentVolumeClaim or one of the vendor-specific
APIs for volumes that persist for longer than the lifecycle
of an individual pod.

Use CSI for light-weight local ephemeral volumes if the CSI driver is meant to
be used that way - see the documentation of the driver for
more information.

A pod can use both types of ephemeral volumes and
persistent volumes at the same time.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexephemeralvolumeclaimtemplate">volumeClaimTemplate</a></b></td>
        <td>object</td>
        <td>
          Will be used to create a stand-alone PVC to provision the volume.
The pod in which this EphemeralVolumeSource is embedded will be the
owner of the PVC, i.e. the PVC will be deleted together with the
pod.  The name of the PVC will be `<pod name>-<volume name>` where
`<volume name>` is the name from the `PodSpec.Volumes` array
entry. Pod validation will reject the pod if the concatenated name
is not valid for a PVC (for example, too long).

An existing PVC with that name that is not owned by the pod
will *not* be used for the pod to avoid using an unrelated
volume by mistake. Starting the pod is then blocked until
the unrelated PVC is removed. If such a pre-created PVC is
meant to be used by the pod, the PVC has to updated with an
owner reference to the pod once the pod exists. Normally
this should not be necessary, but it may be useful when
manually reconstructing a broken cluster.

This field is read-only and no changes will be made by Kubernetes
to the PVC after it has been created.

Required, must not be nil.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].ephemeral.volumeClaimTemplate
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexephemeral)</sup></sup>



Will be used to create a stand-alone PVC to provision the volume.
The pod in which this EphemeralVolumeSource is embedded will be the
owner of the PVC, i.e. the PVC will be deleted together with the
pod.  The name of the PVC will be `<pod name>-<volume name>` where
`<volume name>` is the name from the `PodSpec.Volumes` array
entry. Pod validation will reject the pod if the concatenated name
is not valid for a PVC (for example, too long).

An existing PVC with that name that is not owned by the pod
will *not* be used for the pod to avoid using an unrelated
volume by mistake. Starting the pod is then blocked until
the unrelated PVC is removed. If such a pre-created PVC is
meant to be used by the pod, the PVC has to updated with an
owner reference to the pod once the pod exists. Normally
this should not be necessary, but it may be useful when
manually reconstructing a broken cluster.

This field is read-only and no changes will be made by Kubernetes
to the PVC after it has been created.

Required, must not be nil.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexephemeralvolumeclaimtemplatespec">spec</a></b></td>
        <td>object</td>
        <td>
          The specification for the PersistentVolumeClaim. The entire content is
copied unchanged into the PVC that gets created from this
template. The same fields as in a PersistentVolumeClaim
are also valid here.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexephemeralvolumeclaimtemplatemetadata">metadata</a></b></td>
        <td>object</td>
        <td>
          May contain labels and annotations that will be copied into the PVC
when creating it. No other fields are allowed and will be rejected during
validation.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].ephemeral.volumeClaimTemplate.spec
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexephemeralvolumeclaimtemplate)</sup></sup>



The specification for the PersistentVolumeClaim. The entire content is
copied unchanged into the PVC that gets created from this
template. The same fields as in a PersistentVolumeClaim
are also valid here.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>accessModes</b></td>
        <td>[]string</td>
        <td>
          accessModes contains the desired access modes the volume should have.
More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexephemeralvolumeclaimtemplatespecdatasource">dataSource</a></b></td>
        <td>object</td>
        <td>
          dataSource field can be used to specify either:
* An existing VolumeSnapshot object (snapshot.storage.k8s.io/VolumeSnapshot)
* An existing PVC (PersistentVolumeClaim)
If the provisioner or an external controller can support the specified data source,
it will create a new volume based on the contents of the specified data source.
When the AnyVolumeDataSource feature gate is enabled, dataSource contents will be copied to dataSourceRef,
and dataSourceRef contents will be copied to dataSource when dataSourceRef.namespace is not specified.
If the namespace is specified, then dataSourceRef will not be copied to dataSource.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexephemeralvolumeclaimtemplatespecdatasourceref">dataSourceRef</a></b></td>
        <td>object</td>
        <td>
          dataSourceRef specifies the object from which to populate the volume with data, if a non-empty
volume is desired. This may be any object from a non-empty API group (non
core object) or a PersistentVolumeClaim object.
When this field is specified, volume binding will only succeed if the type of
the specified object matches some installed volume populator or dynamic
provisioner.
This field will replace the functionality of the dataSource field and as such
if both fields are non-empty, they must have the same value. For backwards
compatibility, when namespace isn't specified in dataSourceRef,
both fields (dataSource and dataSourceRef) will be set to the same
value automatically if one of them is empty and the other is non-empty.
When namespace is specified in dataSourceRef,
dataSource isn't set to the same value and must be empty.
There are three important differences between dataSource and dataSourceRef:
* While dataSource only allows two specific types of objects, dataSourceRef
  allows any non-core object, as well as PersistentVolumeClaim objects.
* While dataSource ignores disallowed values (dropping them), dataSourceRef
  preserves all values, and generates an error if a disallowed value is
  specified.
* While dataSource only allows local objects, dataSourceRef allows objects
  in any namespaces.
(Beta) Using this field requires the AnyVolumeDataSource feature gate to be enabled.
(Alpha) Using the namespace field of dataSourceRef requires the CrossNamespaceVolumeDataSource feature gate to be enabled.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexephemeralvolumeclaimtemplatespecresources">resources</a></b></td>
        <td>object</td>
        <td>
          resources represents the minimum resources the volume should have.
If RecoverVolumeExpansionFailure feature is enabled users are allowed to specify resource requirements
that are lower than previous value but must still be higher than capacity recorded in the
status field of the claim.
More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexephemeralvolumeclaimtemplatespecselector">selector</a></b></td>
        <td>object</td>
        <td>
          selector is a label query over volumes to consider for binding.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>storageClassName</b></td>
        <td>string</td>
        <td>
          storageClassName is the name of the StorageClass required by the claim.
More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>volumeAttributesClassName</b></td>
        <td>string</td>
        <td>
          volumeAttributesClassName may be used to set the VolumeAttributesClass used by this claim.
If specified, the CSI driver will create or update the volume with the attributes defined
in the corresponding VolumeAttributesClass. This has a different purpose than storageClassName,
it can be changed after the claim is created. An empty string value means that no VolumeAttributesClass
will be applied to the claim but it's not allowed to reset this field to empty string once it is set.
If unspecified and the PersistentVolumeClaim is unbound, the default VolumeAttributesClass
will be set by the persistentvolume controller if it exists.
If the resource referred to by volumeAttributesClass does not exist, this PersistentVolumeClaim will be
set to a Pending state, as reflected by the modifyVolumeStatus field, until such as a resource
exists.
More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#volumeattributesclass
(Alpha) Using this field requires the VolumeAttributesClass feature gate to be enabled.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>volumeMode</b></td>
        <td>string</td>
        <td>
          volumeMode defines what type of volume is required by the claim.
Value of Filesystem is implied when not included in claim spec.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>volumeName</b></td>
        <td>string</td>
        <td>
          volumeName is the binding reference to the PersistentVolume backing this claim.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].ephemeral.volumeClaimTemplate.spec.dataSource
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexephemeralvolumeclaimtemplatespec)</sup></sup>



dataSource field can be used to specify either:
* An existing VolumeSnapshot object (snapshot.storage.k8s.io/VolumeSnapshot)
* An existing PVC (PersistentVolumeClaim)
If the provisioner or an external controller can support the specified data source,
it will create a new volume based on the contents of the specified data source.
When the AnyVolumeDataSource feature gate is enabled, dataSource contents will be copied to dataSourceRef,
and dataSourceRef contents will be copied to dataSource when dataSourceRef.namespace is not specified.
If the namespace is specified, then dataSourceRef will not be copied to dataSource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind is the type of resource being referenced<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of resource being referenced<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiGroup</b></td>
        <td>string</td>
        <td>
          APIGroup is the group for the resource being referenced.
If APIGroup is not specified, the specified Kind must be in the core API group.
For any other third-party types, APIGroup is required.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].ephemeral.volumeClaimTemplate.spec.dataSourceRef
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexephemeralvolumeclaimtemplatespec)</sup></sup>



dataSourceRef specifies the object from which to populate the volume with data, if a non-empty
volume is desired. This may be any object from a non-empty API group (non
core object) or a PersistentVolumeClaim object.
When this field is specified, volume binding will only succeed if the type of
the specified object matches some installed volume populator or dynamic
provisioner.
This field will replace the functionality of the dataSource field and as such
if both fields are non-empty, they must have the same value. For backwards
compatibility, when namespace isn't specified in dataSourceRef,
both fields (dataSource and dataSourceRef) will be set to the same
value automatically if one of them is empty and the other is non-empty.
When namespace is specified in dataSourceRef,
dataSource isn't set to the same value and must be empty.
There are three important differences between dataSource and dataSourceRef:
* While dataSource only allows two specific types of objects, dataSourceRef
  allows any non-core object, as well as PersistentVolumeClaim objects.
* While dataSource ignores disallowed values (dropping them), dataSourceRef
  preserves all values, and generates an error if a disallowed value is
  specified.
* While dataSource only allows local objects, dataSourceRef allows objects
  in any namespaces.
(Beta) Using this field requires the AnyVolumeDataSource feature gate to be enabled.
(Alpha) Using the namespace field of dataSourceRef requires the CrossNamespaceVolumeDataSource feature gate to be enabled.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind is the type of resource being referenced<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of resource being referenced<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiGroup</b></td>
        <td>string</td>
        <td>
          APIGroup is the group for the resource being referenced.
If APIGroup is not specified, the specified Kind must be in the core API group.
For any other third-party types, APIGroup is required.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace is the namespace of resource being referenced
Note that when a namespace is specified, a gateway.networking.k8s.io/ReferenceGrant object is required in the referent namespace to allow that namespace's owner to accept the reference. See the ReferenceGrant documentation for details.
(Alpha) This field requires the CrossNamespaceVolumeDataSource feature gate to be enabled.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].ephemeral.volumeClaimTemplate.spec.resources
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexephemeralvolumeclaimtemplatespec)</sup></sup>



resources represents the minimum resources the volume should have.
If RecoverVolumeExpansionFailure feature is enabled users are allowed to specify resource requirements
that are lower than previous value but must still be higher than capacity recorded in the
status field of the claim.
More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>limits</b></td>
        <td>map[string]int or string</td>
        <td>
          Limits describes the maximum amount of compute resources allowed.
More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>requests</b></td>
        <td>map[string]int or string</td>
        <td>
          Requests describes the minimum amount of compute resources required.
If Requests is omitted for a container, it defaults to Limits if that is explicitly specified,
otherwise to an implementation-defined value. Requests cannot exceed Limits.
More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].ephemeral.volumeClaimTemplate.spec.selector
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexephemeralvolumeclaimtemplatespec)</sup></sup>



selector is a label query over volumes to consider for binding.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexephemeralvolumeclaimtemplatespecselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].ephemeral.volumeClaimTemplate.spec.selector.matchExpressions[index]
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexephemeralvolumeclaimtemplatespecselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].ephemeral.volumeClaimTemplate.metadata
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexephemeralvolumeclaimtemplate)</sup></sup>



May contain labels and annotations that will be copied into the PVC
when creating it. No other fields are allowed and will be rejected during
validation.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>annotations</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>finalizers</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>labels</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].fc
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



fc represents a Fibre Channel resource that is attached to a kubelet's host machine and then exposed to the pod.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fsType is the filesystem type to mount.
Must be a filesystem type supported by the host operating system.
Ex. "ext4", "xfs", "ntfs". Implicitly inferred to be "ext4" if unspecified.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>lun</b></td>
        <td>integer</td>
        <td>
          lun is Optional: FC target lun number<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly is Optional: Defaults to false (read/write). ReadOnly here will force
the ReadOnly setting in VolumeMounts.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>targetWWNs</b></td>
        <td>[]string</td>
        <td>
          targetWWNs is Optional: FC target worldwide names (WWNs)<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>wwids</b></td>
        <td>[]string</td>
        <td>
          wwids Optional: FC volume world wide identifiers (wwids)
Either wwids or combination of targetWWNs and lun must be set, but not both simultaneously.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].flexVolume
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



flexVolume represents a generic volume resource that is
provisioned/attached using an exec based plugin.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>driver</b></td>
        <td>string</td>
        <td>
          driver is the name of the driver to use for this volume.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fsType is the filesystem type to mount.
Must be a filesystem type supported by the host operating system.
Ex. "ext4", "xfs", "ntfs". The default filesystem depends on FlexVolume script.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>options</b></td>
        <td>map[string]string</td>
        <td>
          options is Optional: this field holds extra command options if any.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly is Optional: defaults to false (read/write). ReadOnly here will force
the ReadOnly setting in VolumeMounts.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexflexvolumesecretref">secretRef</a></b></td>
        <td>object</td>
        <td>
          secretRef is Optional: secretRef is reference to the secret object containing
sensitive information to pass to the plugin scripts. This may be
empty if no secret object is specified. If the secret object
contains more than one secret, all secrets are passed to the plugin
scripts.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].flexVolume.secretRef
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexflexvolume)</sup></sup>



secretRef is Optional: secretRef is reference to the secret object containing
sensitive information to pass to the plugin scripts. This may be
empty if no secret object is specified. If the secret object
contains more than one secret, all secrets are passed to the plugin
scripts.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].flocker
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



flocker represents a Flocker volume attached to a kubelet's host machine. This depends on the Flocker control service being running

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>datasetName</b></td>
        <td>string</td>
        <td>
          datasetName is Name of the dataset stored as metadata -> name on the dataset for Flocker
should be considered as deprecated<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>datasetUUID</b></td>
        <td>string</td>
        <td>
          datasetUUID is the UUID of the dataset. This is unique identifier of a Flocker dataset<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].gcePersistentDisk
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



gcePersistentDisk represents a GCE Disk resource that is attached to a
kubelet's host machine and then exposed to the pod.
More info: https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>pdName</b></td>
        <td>string</td>
        <td>
          pdName is unique name of the PD resource in GCE. Used to identify the disk in GCE.
More info: https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fsType is filesystem type of the volume that you want to mount.
Tip: Ensure that the filesystem type is supported by the host operating system.
Examples: "ext4", "xfs", "ntfs". Implicitly inferred to be "ext4" if unspecified.
More info: https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>partition</b></td>
        <td>integer</td>
        <td>
          partition is the partition in the volume that you want to mount.
If omitted, the default is to mount by volume name.
Examples: For volume /dev/sda1, you specify the partition as "1".
Similarly, the volume partition for /dev/sda is "0" (or you can leave the property empty).
More info: https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly here will force the ReadOnly setting in VolumeMounts.
Defaults to false.
More info: https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].gitRepo
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



gitRepo represents a git repository at a particular revision.
DEPRECATED: GitRepo is deprecated. To provision a container with a git repo, mount an
EmptyDir into an InitContainer that clones the repo using git, then mount the EmptyDir
into the Pod's container.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>repository</b></td>
        <td>string</td>
        <td>
          repository is the URL<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>directory</b></td>
        <td>string</td>
        <td>
          directory is the target directory name.
Must not contain or start with '..'.  If '.' is supplied, the volume directory will be the
git repository.  Otherwise, if specified, the volume will contain the git repository in
the subdirectory with the given name.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>revision</b></td>
        <td>string</td>
        <td>
          revision is the commit hash for the specified revision.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].glusterfs
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



glusterfs represents a Glusterfs mount on the host that shares a pod's lifetime.
More info: https://examples.k8s.io/volumes/glusterfs/README.md

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>endpoints</b></td>
        <td>string</td>
        <td>
          endpoints is the endpoint name that details Glusterfs topology.
More info: https://examples.k8s.io/volumes/glusterfs/README.md#create-a-pod<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          path is the Glusterfs volume path.
More info: https://examples.k8s.io/volumes/glusterfs/README.md#create-a-pod<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly here will force the Glusterfs volume to be mounted with read-only permissions.
Defaults to false.
More info: https://examples.k8s.io/volumes/glusterfs/README.md#create-a-pod<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].hostPath
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



hostPath represents a pre-existing file or directory on the host
machine that is directly exposed to the container. This is generally
used for system agents or other privileged things that are allowed
to see the host machine. Most containers will NOT need this.
More info: https://kubernetes.io/docs/concepts/storage/volumes#hostpath

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          path of the directory on the host.
If the path is a symlink, it will follow the link to the real path.
More info: https://kubernetes.io/docs/concepts/storage/volumes#hostpath<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type for HostPath Volume
Defaults to ""
More info: https://kubernetes.io/docs/concepts/storage/volumes#hostpath<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].iscsi
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



iscsi represents an ISCSI Disk resource that is attached to a
kubelet's host machine and then exposed to the pod.
More info: https://examples.k8s.io/volumes/iscsi/README.md

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>iqn</b></td>
        <td>string</td>
        <td>
          iqn is the target iSCSI Qualified Name.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>lun</b></td>
        <td>integer</td>
        <td>
          lun represents iSCSI Target Lun number.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>targetPortal</b></td>
        <td>string</td>
        <td>
          targetPortal is iSCSI Target Portal. The Portal is either an IP or ip_addr:port if the port
is other than default (typically TCP ports 860 and 3260).<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>chapAuthDiscovery</b></td>
        <td>boolean</td>
        <td>
          chapAuthDiscovery defines whether support iSCSI Discovery CHAP authentication<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>chapAuthSession</b></td>
        <td>boolean</td>
        <td>
          chapAuthSession defines whether support iSCSI Session CHAP authentication<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fsType is the filesystem type of the volume that you want to mount.
Tip: Ensure that the filesystem type is supported by the host operating system.
Examples: "ext4", "xfs", "ntfs". Implicitly inferred to be "ext4" if unspecified.
More info: https://kubernetes.io/docs/concepts/storage/volumes#iscsi<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>initiatorName</b></td>
        <td>string</td>
        <td>
          initiatorName is the custom iSCSI Initiator Name.
If initiatorName is specified with iscsiInterface simultaneously, new iSCSI interface
<target portal>:<volume name> will be created for the connection.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>iscsiInterface</b></td>
        <td>string</td>
        <td>
          iscsiInterface is the interface Name that uses an iSCSI transport.
Defaults to 'default' (tcp).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>portals</b></td>
        <td>[]string</td>
        <td>
          portals is the iSCSI Target Portal List. The portal is either an IP or ip_addr:port if the port
is other than default (typically TCP ports 860 and 3260).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly here will force the ReadOnly setting in VolumeMounts.
Defaults to false.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexiscsisecretref">secretRef</a></b></td>
        <td>object</td>
        <td>
          secretRef is the CHAP Secret for iSCSI target and initiator authentication<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].iscsi.secretRef
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexiscsi)</sup></sup>



secretRef is the CHAP Secret for iSCSI target and initiator authentication

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].nfs
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



nfs represents an NFS mount on the host that shares a pod's lifetime
More info: https://kubernetes.io/docs/concepts/storage/volumes#nfs

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          path that is exported by the NFS server.
More info: https://kubernetes.io/docs/concepts/storage/volumes#nfs<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>server</b></td>
        <td>string</td>
        <td>
          server is the hostname or IP address of the NFS server.
More info: https://kubernetes.io/docs/concepts/storage/volumes#nfs<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly here will force the NFS export to be mounted with read-only permissions.
Defaults to false.
More info: https://kubernetes.io/docs/concepts/storage/volumes#nfs<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].persistentVolumeClaim
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



persistentVolumeClaimVolumeSource represents a reference to a
PersistentVolumeClaim in the same namespace.
More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>claimName</b></td>
        <td>string</td>
        <td>
          claimName is the name of a PersistentVolumeClaim in the same namespace as the pod using this volume.
More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly Will force the ReadOnly setting in VolumeMounts.
Default false.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].photonPersistentDisk
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



photonPersistentDisk represents a PhotonController persistent disk attached and mounted on kubelets host machine

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>pdID</b></td>
        <td>string</td>
        <td>
          pdID is the ID that identifies Photon Controller persistent disk<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fsType is the filesystem type to mount.
Must be a filesystem type supported by the host operating system.
Ex. "ext4", "xfs", "ntfs". Implicitly inferred to be "ext4" if unspecified.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].portworxVolume
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



portworxVolume represents a portworx volume attached and mounted on kubelets host machine

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>volumeID</b></td>
        <td>string</td>
        <td>
          volumeID uniquely identifies a Portworx volume<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fSType represents the filesystem type to mount
Must be a filesystem type supported by the host operating system.
Ex. "ext4", "xfs". Implicitly inferred to be "ext4" if unspecified.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly defaults to false (read/write). ReadOnly here will force
the ReadOnly setting in VolumeMounts.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].projected
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



projected items for all in one resources secrets, configmaps, and downward API

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>defaultMode</b></td>
        <td>integer</td>
        <td>
          defaultMode are the mode bits used to set permissions on created files by default.
Must be an octal value between 0000 and 0777 or a decimal value between 0 and 511.
YAML accepts both octal and decimal values, JSON requires decimal values for mode bits.
Directories within the path are not affected by this setting.
This might be in conflict with other options that affect the file
mode, like fsGroup, and the result can be other mode bits set.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexprojectedsourcesindex">sources</a></b></td>
        <td>[]object</td>
        <td>
          sources is the list of volume projections<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].projected.sources[index]
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexprojected)</sup></sup>



Projection that may be projected along with other supported volume types

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexprojectedsourcesindexclustertrustbundle">clusterTrustBundle</a></b></td>
        <td>object</td>
        <td>
          ClusterTrustBundle allows a pod to access the `.spec.trustBundle` field
of ClusterTrustBundle objects in an auto-updating file.

Alpha, gated by the ClusterTrustBundleProjection feature gate.

ClusterTrustBundle objects can either be selected by name, or by the
combination of signer name and a label selector.

Kubelet performs aggressive normalization of the PEM contents written
into the pod filesystem.  Esoteric PEM features such as inter-block
comments and block headers are stripped.  Certificates are deduplicated.
The ordering of certificates within the file is arbitrary, and Kubelet
may change the order over time.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexprojectedsourcesindexconfigmap">configMap</a></b></td>
        <td>object</td>
        <td>
          configMap information about the configMap data to project<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexprojectedsourcesindexdownwardapi">downwardAPI</a></b></td>
        <td>object</td>
        <td>
          downwardAPI information about the downwardAPI data to project<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexprojectedsourcesindexsecret">secret</a></b></td>
        <td>object</td>
        <td>
          secret information about the secret data to project<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexprojectedsourcesindexserviceaccounttoken">serviceAccountToken</a></b></td>
        <td>object</td>
        <td>
          serviceAccountToken is information about the serviceAccountToken data to project<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].projected.sources[index].clusterTrustBundle
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexprojectedsourcesindex)</sup></sup>



ClusterTrustBundle allows a pod to access the `.spec.trustBundle` field
of ClusterTrustBundle objects in an auto-updating file.

Alpha, gated by the ClusterTrustBundleProjection feature gate.

ClusterTrustBundle objects can either be selected by name, or by the
combination of signer name and a label selector.

Kubelet performs aggressive normalization of the PEM contents written
into the pod filesystem.  Esoteric PEM features such as inter-block
comments and block headers are stripped.  Certificates are deduplicated.
The ordering of certificates within the file is arbitrary, and Kubelet
may change the order over time.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          Relative path from the volume root to write the bundle.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexprojectedsourcesindexclustertrustbundlelabelselector">labelSelector</a></b></td>
        <td>object</td>
        <td>
          Select all ClusterTrustBundles that match this label selector.  Only has
effect if signerName is set.  Mutually-exclusive with name.  If unset,
interpreted as "match nothing".  If set but empty, interpreted as "match
everything".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Select a single ClusterTrustBundle by object name.  Mutually-exclusive
with signerName and labelSelector.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          If true, don't block pod startup if the referenced ClusterTrustBundle(s)
aren't available.  If using name, then the named ClusterTrustBundle is
allowed not to exist.  If using signerName, then the combination of
signerName and labelSelector is allowed to match zero
ClusterTrustBundles.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>signerName</b></td>
        <td>string</td>
        <td>
          Select all ClusterTrustBundles that match this signer name.
Mutually-exclusive with name.  The contents of all selected
ClusterTrustBundles will be unified and deduplicated.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].projected.sources[index].clusterTrustBundle.labelSelector
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexprojectedsourcesindexclustertrustbundle)</sup></sup>



Select all ClusterTrustBundles that match this label selector.  Only has
effect if signerName is set.  Mutually-exclusive with name.  If unset,
interpreted as "match nothing".  If set but empty, interpreted as "match
everything".

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexprojectedsourcesindexclustertrustbundlelabelselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].projected.sources[index].clusterTrustBundle.labelSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexprojectedsourcesindexclustertrustbundlelabelselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].projected.sources[index].configMap
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexprojectedsourcesindex)</sup></sup>



configMap information about the configMap data to project

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexprojectedsourcesindexconfigmapitemsindex">items</a></b></td>
        <td>[]object</td>
        <td>
          items if unspecified, each key-value pair in the Data field of the referenced
ConfigMap will be projected into the volume as a file whose name is the
key and content is the value. If specified, the listed keys will be
projected into the specified paths, and unlisted keys will not be
present. If a key is specified which is not present in the ConfigMap,
the volume setup will error unless it is marked optional. Paths must be
relative and may not contain the '..' path or start with '..'.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          optional specify whether the ConfigMap or its keys must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].projected.sources[index].configMap.items[index]
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexprojectedsourcesindexconfigmap)</sup></sup>



Maps a string key to a path within a volume.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the key to project.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          path is the relative path of the file to map the key to.
May not be an absolute path.
May not contain the path element '..'.
May not start with the string '..'.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>mode</b></td>
        <td>integer</td>
        <td>
          mode is Optional: mode bits used to set permissions on this file.
Must be an octal value between 0000 and 0777 or a decimal value between 0 and 511.
YAML accepts both octal and decimal values, JSON requires decimal values for mode bits.
If not specified, the volume defaultMode will be used.
This might be in conflict with other options that affect the file
mode, like fsGroup, and the result can be other mode bits set.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].projected.sources[index].downwardAPI
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexprojectedsourcesindex)</sup></sup>



downwardAPI information about the downwardAPI data to project

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexprojectedsourcesindexdownwardapiitemsindex">items</a></b></td>
        <td>[]object</td>
        <td>
          Items is a list of DownwardAPIVolume file<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].projected.sources[index].downwardAPI.items[index]
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexprojectedsourcesindexdownwardapi)</sup></sup>



DownwardAPIVolumeFile represents information to create the file containing the pod field

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          Required: Path is  the relative path name of the file to be created. Must not be absolute or contain the '..' path. Must be utf-8 encoded. The first item of the relative path must not start with '..'<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexprojectedsourcesindexdownwardapiitemsindexfieldref">fieldRef</a></b></td>
        <td>object</td>
        <td>
          Required: Selects a field of the pod: only annotations, labels, name and namespace are supported.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>mode</b></td>
        <td>integer</td>
        <td>
          Optional: mode bits used to set permissions on this file, must be an octal value
between 0000 and 0777 or a decimal value between 0 and 511.
YAML accepts both octal and decimal values, JSON requires decimal values for mode bits.
If not specified, the volume defaultMode will be used.
This might be in conflict with other options that affect the file
mode, like fsGroup, and the result can be other mode bits set.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexprojectedsourcesindexdownwardapiitemsindexresourcefieldref">resourceFieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a resource of the container: only resources limits and requests
(limits.cpu, limits.memory, requests.cpu and requests.memory) are currently supported.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].projected.sources[index].downwardAPI.items[index].fieldRef
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexprojectedsourcesindexdownwardapiitemsindex)</sup></sup>



Required: Selects a field of the pod: only annotations, labels, name and namespace are supported.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>fieldPath</b></td>
        <td>string</td>
        <td>
          Path of the field to select in the specified API version.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          Version of the schema the FieldPath is written in terms of, defaults to "v1".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].projected.sources[index].downwardAPI.items[index].resourceFieldRef
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexprojectedsourcesindexdownwardapiitemsindex)</sup></sup>



Selects a resource of the container: only resources limits and requests
(limits.cpu, limits.memory, requests.cpu and requests.memory) are currently supported.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>resource</b></td>
        <td>string</td>
        <td>
          Required: resource to select<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>containerName</b></td>
        <td>string</td>
        <td>
          Container name: required for volumes, optional for env vars<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>divisor</b></td>
        <td>int or string</td>
        <td>
          Specifies the output format of the exposed resources, defaults to "1"<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].projected.sources[index].secret
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexprojectedsourcesindex)</sup></sup>



secret information about the secret data to project

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexprojectedsourcesindexsecretitemsindex">items</a></b></td>
        <td>[]object</td>
        <td>
          items if unspecified, each key-value pair in the Data field of the referenced
Secret will be projected into the volume as a file whose name is the
key and content is the value. If specified, the listed keys will be
projected into the specified paths, and unlisted keys will not be
present. If a key is specified which is not present in the Secret,
the volume setup will error unless it is marked optional. Paths must be
relative and may not contain the '..' path or start with '..'.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          optional field specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].projected.sources[index].secret.items[index]
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexprojectedsourcesindexsecret)</sup></sup>



Maps a string key to a path within a volume.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the key to project.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          path is the relative path of the file to map the key to.
May not be an absolute path.
May not contain the path element '..'.
May not start with the string '..'.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>mode</b></td>
        <td>integer</td>
        <td>
          mode is Optional: mode bits used to set permissions on this file.
Must be an octal value between 0000 and 0777 or a decimal value between 0 and 511.
YAML accepts both octal and decimal values, JSON requires decimal values for mode bits.
If not specified, the volume defaultMode will be used.
This might be in conflict with other options that affect the file
mode, like fsGroup, and the result can be other mode bits set.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].projected.sources[index].serviceAccountToken
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexprojectedsourcesindex)</sup></sup>



serviceAccountToken is information about the serviceAccountToken data to project

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          path is the path relative to the mount point of the file to project the
token into.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>audience</b></td>
        <td>string</td>
        <td>
          audience is the intended audience of the token. A recipient of a token
must identify itself with an identifier specified in the audience of the
token, and otherwise should reject the token. The audience defaults to the
identifier of the apiserver.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>expirationSeconds</b></td>
        <td>integer</td>
        <td>
          expirationSeconds is the requested duration of validity of the service
account token. As the token approaches expiration, the kubelet volume
plugin will proactively rotate the service account token. The kubelet will
start trying to rotate the token if the token is older than 80 percent of
its time to live or if the token is older than 24 hours.Defaults to 1 hour
and must be at least 10 minutes.<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].quobyte
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



quobyte represents a Quobyte mount on the host that shares a pod's lifetime

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>registry</b></td>
        <td>string</td>
        <td>
          registry represents a single or multiple Quobyte Registry services
specified as a string as host:port pair (multiple entries are separated with commas)
which acts as the central registry for volumes<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>volume</b></td>
        <td>string</td>
        <td>
          volume is a string that references an already created Quobyte volume by name.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>group</b></td>
        <td>string</td>
        <td>
          group to map volume access to
Default is no group<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly here will force the Quobyte volume to be mounted with read-only permissions.
Defaults to false.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>tenant</b></td>
        <td>string</td>
        <td>
          tenant owning the given Quobyte volume in the Backend
Used with dynamically provisioned Quobyte volumes, value is set by the plugin<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>user</b></td>
        <td>string</td>
        <td>
          user to map volume access to
Defaults to serivceaccount user<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].rbd
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



rbd represents a Rados Block Device mount on the host that shares a pod's lifetime.
More info: https://examples.k8s.io/volumes/rbd/README.md

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>image</b></td>
        <td>string</td>
        <td>
          image is the rados image name.
More info: https://examples.k8s.io/volumes/rbd/README.md#how-to-use-it<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>monitors</b></td>
        <td>[]string</td>
        <td>
          monitors is a collection of Ceph monitors.
More info: https://examples.k8s.io/volumes/rbd/README.md#how-to-use-it<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fsType is the filesystem type of the volume that you want to mount.
Tip: Ensure that the filesystem type is supported by the host operating system.
Examples: "ext4", "xfs", "ntfs". Implicitly inferred to be "ext4" if unspecified.
More info: https://kubernetes.io/docs/concepts/storage/volumes#rbd<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>keyring</b></td>
        <td>string</td>
        <td>
          keyring is the path to key ring for RBDUser.
Default is /etc/ceph/keyring.
More info: https://examples.k8s.io/volumes/rbd/README.md#how-to-use-it<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>pool</b></td>
        <td>string</td>
        <td>
          pool is the rados pool name.
Default is rbd.
More info: https://examples.k8s.io/volumes/rbd/README.md#how-to-use-it<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly here will force the ReadOnly setting in VolumeMounts.
Defaults to false.
More info: https://examples.k8s.io/volumes/rbd/README.md#how-to-use-it<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexrbdsecretref">secretRef</a></b></td>
        <td>object</td>
        <td>
          secretRef is name of the authentication secret for RBDUser. If provided
overrides keyring.
Default is nil.
More info: https://examples.k8s.io/volumes/rbd/README.md#how-to-use-it<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>user</b></td>
        <td>string</td>
        <td>
          user is the rados user name.
Default is admin.
More info: https://examples.k8s.io/volumes/rbd/README.md#how-to-use-it<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].rbd.secretRef
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexrbd)</sup></sup>



secretRef is name of the authentication secret for RBDUser. If provided
overrides keyring.
Default is nil.
More info: https://examples.k8s.io/volumes/rbd/README.md#how-to-use-it

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].scaleIO
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



scaleIO represents a ScaleIO persistent volume attached and mounted on Kubernetes nodes.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>gateway</b></td>
        <td>string</td>
        <td>
          gateway is the host address of the ScaleIO API Gateway.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexscaleiosecretref">secretRef</a></b></td>
        <td>object</td>
        <td>
          secretRef references to the secret for ScaleIO user and other
sensitive information. If this is not provided, Login operation will fail.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>system</b></td>
        <td>string</td>
        <td>
          system is the name of the storage system as configured in ScaleIO.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fsType is the filesystem type to mount.
Must be a filesystem type supported by the host operating system.
Ex. "ext4", "xfs", "ntfs".
Default is "xfs".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>protectionDomain</b></td>
        <td>string</td>
        <td>
          protectionDomain is the name of the ScaleIO Protection Domain for the configured storage.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly Defaults to false (read/write). ReadOnly here will force
the ReadOnly setting in VolumeMounts.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sslEnabled</b></td>
        <td>boolean</td>
        <td>
          sslEnabled Flag enable/disable SSL communication with Gateway, default false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>storageMode</b></td>
        <td>string</td>
        <td>
          storageMode indicates whether the storage for a volume should be ThickProvisioned or ThinProvisioned.
Default is ThinProvisioned.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>storagePool</b></td>
        <td>string</td>
        <td>
          storagePool is the ScaleIO Storage Pool associated with the protection domain.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>volumeName</b></td>
        <td>string</td>
        <td>
          volumeName is the name of a volume already created in the ScaleIO system
that is associated with this volume source.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].scaleIO.secretRef
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexscaleio)</sup></sup>



secretRef references to the secret for ScaleIO user and other
sensitive information. If this is not provided, Login operation will fail.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].secret
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



secret represents a secret that should populate this volume.
More info: https://kubernetes.io/docs/concepts/storage/volumes#secret

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>defaultMode</b></td>
        <td>integer</td>
        <td>
          defaultMode is Optional: mode bits used to set permissions on created files by default.
Must be an octal value between 0000 and 0777 or a decimal value between 0 and 511.
YAML accepts both octal and decimal values, JSON requires decimal values
for mode bits. Defaults to 0644.
Directories within the path are not affected by this setting.
This might be in conflict with other options that affect the file
mode, like fsGroup, and the result can be other mode bits set.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexsecretitemsindex">items</a></b></td>
        <td>[]object</td>
        <td>
          items If unspecified, each key-value pair in the Data field of the referenced
Secret will be projected into the volume as a file whose name is the
key and content is the value. If specified, the listed keys will be
projected into the specified paths, and unlisted keys will not be
present. If a key is specified which is not present in the Secret,
the volume setup will error unless it is marked optional. Paths must be
relative and may not contain the '..' path or start with '..'.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          optional field specify whether the Secret or its keys must be defined<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretName</b></td>
        <td>string</td>
        <td>
          secretName is the name of the secret in the pod's namespace to use.
More info: https://kubernetes.io/docs/concepts/storage/volumes#secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].secret.items[index]
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexsecret)</sup></sup>



Maps a string key to a path within a volume.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the key to project.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          path is the relative path of the file to map the key to.
May not be an absolute path.
May not contain the path element '..'.
May not start with the string '..'.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>mode</b></td>
        <td>integer</td>
        <td>
          mode is Optional: mode bits used to set permissions on this file.
Must be an octal value between 0000 and 0777 or a decimal value between 0 and 511.
YAML accepts both octal and decimal values, JSON requires decimal values for mode bits.
If not specified, the volume defaultMode will be used.
This might be in conflict with other options that affect the file
mode, like fsGroup, and the result can be other mode bits set.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].storageos
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



storageOS represents a StorageOS volume attached and mounted on Kubernetes nodes.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fsType is the filesystem type to mount.
Must be a filesystem type supported by the host operating system.
Ex. "ext4", "xfs", "ntfs". Implicitly inferred to be "ext4" if unspecified.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly defaults to false (read/write). ReadOnly here will force
the ReadOnly setting in VolumeMounts.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecinstancesindexvolumesindexstorageossecretref">secretRef</a></b></td>
        <td>object</td>
        <td>
          secretRef specifies the secret to use for obtaining the StorageOS API
credentials.  If not specified, default values will be attempted.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>volumeName</b></td>
        <td>string</td>
        <td>
          volumeName is the human-readable name of the StorageOS volume.  Volume
names are only unique within a namespace.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>volumeNamespace</b></td>
        <td>string</td>
        <td>
          volumeNamespace specifies the scope of the volume within StorageOS.  If no
namespace is specified then the Pod's namespace will be used.  This allows the
Kubernetes name scoping to be mirrored within StorageOS for tighter integration.
Set VolumeName to any name to override the default behaviour.
Set to "default" if you are not using namespaces within StorageOS.
Namespaces that do not pre-exist within StorageOS will be created.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].storageos.secretRef
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindexstorageos)</sup></sup>



secretRef specifies the secret to use for obtaining the StorageOS API
credentials.  If not specified, default values will be attempted.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.instances[index].volumes[index].vsphereVolume
<sup><sup>[↩ Parent](#redroidtaskspecinstancesindexvolumesindex)</sup></sup>



vsphereVolume represents a vSphere volume attached and mounted on kubelets host machine

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>volumePath</b></td>
        <td>string</td>
        <td>
          volumePath is the path that identifies vSphere volume vmdk<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fsType is filesystem type to mount.
Must be a filesystem type supported by the host operating system.
Ex. "ext4", "xfs", "ntfs". Implicitly inferred to be "ext4" if unspecified.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>storagePolicyID</b></td>
        <td>string</td>
        <td>
          storagePolicyID is the storage Policy Based Management (SPBM) profile ID associated with the StoragePolicyName.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>storagePolicyName</b></td>
        <td>string</td>
        <td>
          storagePolicyName is the storage Policy Based Management (SPBM) profile name.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.integrations[index]
<sup><sup>[↩ Parent](#redroidtaskspec)</sup></sup>



IntegrationSpec describes a tool container that runs against the Redroid ADB.
ADB address is injected as ADB_ADDRESS env var; instance index as INSTANCE_INDEX.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>image</b></td>
        <td>string</td>
        <td>
          Image is the container image for this tool.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is a unique identifier for this integration within the task.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>args</b></td>
        <td>[]string</td>
        <td>
          Args is passed to the container command.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>command</b></td>
        <td>[]string</td>
        <td>
          Command overrides the container entrypoint.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecintegrationsindexconfigsindex">configs</a></b></td>
        <td>[]object</td>
        <td>
          Configs mounts ConfigMap keys as files inside this container.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecintegrationsindexenvindex">env</a></b></td>
        <td>[]object</td>
        <td>
          Env provides additional environment variables, merged after ADB_ADDRESS/INSTANCE_INDEX.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>imagePullPolicy</b></td>
        <td>string</td>
        <td>
          ImagePullPolicy overrides the image pull policy. Defaults to Always.<br/>
          <br/>
            <i>Default</i>: Always<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecintegrationsindexresources">resources</a></b></td>
        <td>object</td>
        <td>
          Resources optionally sets CPU/memory limits on this container.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecintegrationsindexsecuritycontext">securityContext</a></b></td>
        <td>object</td>
        <td>
          SecurityContext sets per-container security options.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecintegrationsindexvolumemountsindex">volumeMounts</a></b></td>
        <td>[]object</td>
        <td>
          VolumeMounts adds extra volume mounts into the integration container.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>workingDir</b></td>
        <td>string</td>
        <td>
          WorkingDir sets the current working directory inside the container.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.integrations[index].configs[index]
<sup><sup>[↩ Parent](#redroidtaskspecintegrationsindex)</sup></sup>



ConfigFile mounts a single key from a ConfigMap as a file inside the integration container.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>configMapName</b></td>
        <td>string</td>
        <td>
          ConfigMapName is the name of the ConfigMap in the same namespace.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          Key is the ConfigMap key to mount.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>mountPath</b></td>
        <td>string</td>
        <td>
          MountPath is the absolute path inside the container where the file is placed.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### RedroidTask.spec.integrations[index].env[index]
<sup><sup>[↩ Parent](#redroidtaskspecintegrationsindex)</sup></sup>



EnvVar represents an environment variable present in a Container.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the environment variable. Must be a C_IDENTIFIER.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Variable references $(VAR_NAME) are expanded
using the previously defined environment variables in the container and
any service environment variables. If a variable cannot be resolved,
the reference in the input string will be unchanged. Double $$ are reduced
to a single $, which allows for escaping the $(VAR_NAME) syntax: i.e.
"$$(VAR_NAME)" will produce the string literal "$(VAR_NAME)".
Escaped references will never be expanded, regardless of whether the variable
exists or not.
Defaults to "".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecintegrationsindexenvindexvaluefrom">valueFrom</a></b></td>
        <td>object</td>
        <td>
          Source for the environment variable's value. Cannot be used if value is not empty.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.integrations[index].env[index].valueFrom
<sup><sup>[↩ Parent](#redroidtaskspecintegrationsindexenvindex)</sup></sup>



Source for the environment variable's value. Cannot be used if value is not empty.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidtaskspecintegrationsindexenvindexvaluefromconfigmapkeyref">configMapKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a ConfigMap.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecintegrationsindexenvindexvaluefromfieldref">fieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a field of the pod: supports metadata.name, metadata.namespace, `metadata.labels['<KEY>']`, `metadata.annotations['<KEY>']`,
spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecintegrationsindexenvindexvaluefromresourcefieldref">resourceFieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a resource of the container: only resources limits and requests
(limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecintegrationsindexenvindexvaluefromsecretkeyref">secretKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a secret in the pod's namespace<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.integrations[index].env[index].valueFrom.configMapKeyRef
<sup><sup>[↩ Parent](#redroidtaskspecintegrationsindexenvindexvaluefrom)</sup></sup>



Selects a key of a ConfigMap.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.integrations[index].env[index].valueFrom.fieldRef
<sup><sup>[↩ Parent](#redroidtaskspecintegrationsindexenvindexvaluefrom)</sup></sup>



Selects a field of the pod: supports metadata.name, metadata.namespace, `metadata.labels['<KEY>']`, `metadata.annotations['<KEY>']`,
spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>fieldPath</b></td>
        <td>string</td>
        <td>
          Path of the field to select in the specified API version.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          Version of the schema the FieldPath is written in terms of, defaults to "v1".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.integrations[index].env[index].valueFrom.resourceFieldRef
<sup><sup>[↩ Parent](#redroidtaskspecintegrationsindexenvindexvaluefrom)</sup></sup>



Selects a resource of the container: only resources limits and requests
(limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>resource</b></td>
        <td>string</td>
        <td>
          Required: resource to select<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>containerName</b></td>
        <td>string</td>
        <td>
          Container name: required for volumes, optional for env vars<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>divisor</b></td>
        <td>int or string</td>
        <td>
          Specifies the output format of the exposed resources, defaults to "1"<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.integrations[index].env[index].valueFrom.secretKeyRef
<sup><sup>[↩ Parent](#redroidtaskspecintegrationsindexenvindexvaluefrom)</sup></sup>



Selects a key of a secret in the pod's namespace

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.integrations[index].resources
<sup><sup>[↩ Parent](#redroidtaskspecintegrationsindex)</sup></sup>



Resources optionally sets CPU/memory limits on this container.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidtaskspecintegrationsindexresourcesclaimsindex">claims</a></b></td>
        <td>[]object</td>
        <td>
          Claims lists the names of resources, defined in spec.resourceClaims,
that are used by this container.

This is an alpha field and requires enabling the
DynamicResourceAllocation feature gate.

This field is immutable. It can only be set for containers.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>limits</b></td>
        <td>map[string]int or string</td>
        <td>
          Limits describes the maximum amount of compute resources allowed.
More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>requests</b></td>
        <td>map[string]int or string</td>
        <td>
          Requests describes the minimum amount of compute resources required.
If Requests is omitted for a container, it defaults to Limits if that is explicitly specified,
otherwise to an implementation-defined value. Requests cannot exceed Limits.
More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.integrations[index].resources.claims[index]
<sup><sup>[↩ Parent](#redroidtaskspecintegrationsindexresources)</sup></sup>



ResourceClaim references one entry in PodSpec.ResourceClaims.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name must match the name of one entry in pod.spec.resourceClaims of
the Pod where this field is used. It makes that resource available
inside a container.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### RedroidTask.spec.integrations[index].securityContext
<sup><sup>[↩ Parent](#redroidtaskspecintegrationsindex)</sup></sup>



SecurityContext sets per-container security options.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>allowPrivilegeEscalation</b></td>
        <td>boolean</td>
        <td>
          AllowPrivilegeEscalation controls whether a process can gain more
privileges than its parent process. This bool directly controls if
the no_new_privs flag will be set on the container process.
AllowPrivilegeEscalation is true always when the container is:
1) run as Privileged
2) has CAP_SYS_ADMIN
Note that this field cannot be set when spec.os.name is windows.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecintegrationsindexsecuritycontextcapabilities">capabilities</a></b></td>
        <td>object</td>
        <td>
          The capabilities to add/drop when running containers.
Defaults to the default set of capabilities granted by the container runtime.
Note that this field cannot be set when spec.os.name is windows.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>privileged</b></td>
        <td>boolean</td>
        <td>
          Run container in privileged mode.
Processes in privileged containers are essentially equivalent to root on the host.
Defaults to false.
Note that this field cannot be set when spec.os.name is windows.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>procMount</b></td>
        <td>string</td>
        <td>
          procMount denotes the type of proc mount to use for the containers.
The default is DefaultProcMount which uses the container runtime defaults for
readonly paths and masked paths.
This requires the ProcMountType feature flag to be enabled.
Note that this field cannot be set when spec.os.name is windows.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnlyRootFilesystem</b></td>
        <td>boolean</td>
        <td>
          Whether this container has a read-only root filesystem.
Default is false.
Note that this field cannot be set when spec.os.name is windows.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>runAsGroup</b></td>
        <td>integer</td>
        <td>
          The GID to run the entrypoint of the container process.
Uses runtime default if unset.
May also be set in PodSecurityContext.  If set in both SecurityContext and
PodSecurityContext, the value specified in SecurityContext takes precedence.
Note that this field cannot be set when spec.os.name is windows.<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>runAsNonRoot</b></td>
        <td>boolean</td>
        <td>
          Indicates that the container must run as a non-root user.
If true, the Kubelet will validate the image at runtime to ensure that it
does not run as UID 0 (root) and fail to start the container if it does.
If unset or false, no such validation will be performed.
May also be set in PodSecurityContext.  If set in both SecurityContext and
PodSecurityContext, the value specified in SecurityContext takes precedence.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>runAsUser</b></td>
        <td>integer</td>
        <td>
          The UID to run the entrypoint of the container process.
Defaults to user specified in image metadata if unspecified.
May also be set in PodSecurityContext.  If set in both SecurityContext and
PodSecurityContext, the value specified in SecurityContext takes precedence.
Note that this field cannot be set when spec.os.name is windows.<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecintegrationsindexsecuritycontextselinuxoptions">seLinuxOptions</a></b></td>
        <td>object</td>
        <td>
          The SELinux context to be applied to the container.
If unspecified, the container runtime will allocate a random SELinux context for each
container.  May also be set in PodSecurityContext.  If set in both SecurityContext and
PodSecurityContext, the value specified in SecurityContext takes precedence.
Note that this field cannot be set when spec.os.name is windows.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecintegrationsindexsecuritycontextseccompprofile">seccompProfile</a></b></td>
        <td>object</td>
        <td>
          The seccomp options to use by this container. If seccomp options are
provided at both the pod & container level, the container options
override the pod options.
Note that this field cannot be set when spec.os.name is windows.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecintegrationsindexsecuritycontextwindowsoptions">windowsOptions</a></b></td>
        <td>object</td>
        <td>
          The Windows specific settings applied to all containers.
If unspecified, the options from the PodSecurityContext will be used.
If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence.
Note that this field cannot be set when spec.os.name is linux.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.integrations[index].securityContext.capabilities
<sup><sup>[↩ Parent](#redroidtaskspecintegrationsindexsecuritycontext)</sup></sup>



The capabilities to add/drop when running containers.
Defaults to the default set of capabilities granted by the container runtime.
Note that this field cannot be set when spec.os.name is windows.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>add</b></td>
        <td>[]string</td>
        <td>
          Added capabilities<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>drop</b></td>
        <td>[]string</td>
        <td>
          Removed capabilities<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.integrations[index].securityContext.seLinuxOptions
<sup><sup>[↩ Parent](#redroidtaskspecintegrationsindexsecuritycontext)</sup></sup>



The SELinux context to be applied to the container.
If unspecified, the container runtime will allocate a random SELinux context for each
container.  May also be set in PodSecurityContext.  If set in both SecurityContext and
PodSecurityContext, the value specified in SecurityContext takes precedence.
Note that this field cannot be set when spec.os.name is windows.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>level</b></td>
        <td>string</td>
        <td>
          Level is SELinux level label that applies to the container.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          Role is a SELinux role label that applies to the container.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          Type is a SELinux type label that applies to the container.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>user</b></td>
        <td>string</td>
        <td>
          User is a SELinux user label that applies to the container.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.integrations[index].securityContext.seccompProfile
<sup><sup>[↩ Parent](#redroidtaskspecintegrationsindexsecuritycontext)</sup></sup>



The seccomp options to use by this container. If seccomp options are
provided at both the pod & container level, the container options
override the pod options.
Note that this field cannot be set when spec.os.name is windows.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type indicates which kind of seccomp profile will be applied.
Valid options are:

Localhost - a profile defined in a file on the node should be used.
RuntimeDefault - the container runtime default profile should be used.
Unconfined - no profile should be applied.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>localhostProfile</b></td>
        <td>string</td>
        <td>
          localhostProfile indicates a profile defined in a file on the node should be used.
The profile must be preconfigured on the node to work.
Must be a descending path, relative to the kubelet's configured seccomp profile location.
Must be set if type is "Localhost". Must NOT be set for any other type.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.integrations[index].securityContext.windowsOptions
<sup><sup>[↩ Parent](#redroidtaskspecintegrationsindexsecuritycontext)</sup></sup>



The Windows specific settings applied to all containers.
If unspecified, the options from the PodSecurityContext will be used.
If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence.
Note that this field cannot be set when spec.os.name is linux.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>gmsaCredentialSpec</b></td>
        <td>string</td>
        <td>
          GMSACredentialSpec is where the GMSA admission webhook
(https://github.com/kubernetes-sigs/windows-gmsa) inlines the contents of the
GMSA credential spec named by the GMSACredentialSpecName field.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>gmsaCredentialSpecName</b></td>
        <td>string</td>
        <td>
          GMSACredentialSpecName is the name of the GMSA credential spec to use.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>hostProcess</b></td>
        <td>boolean</td>
        <td>
          HostProcess determines if a container should be run as a 'Host Process' container.
All of a Pod's containers must have the same effective HostProcess value
(it is not allowed to have a mix of HostProcess containers and non-HostProcess containers).
In addition, if HostProcess is true then HostNetwork must also be set to true.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>runAsUserName</b></td>
        <td>string</td>
        <td>
          The UserName in Windows to run the entrypoint of the container process.
Defaults to the user specified in image metadata if unspecified.
May also be set in PodSecurityContext. If set in both SecurityContext and
PodSecurityContext, the value specified in SecurityContext takes precedence.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.integrations[index].volumeMounts[index]
<sup><sup>[↩ Parent](#redroidtaskspecintegrationsindex)</sup></sup>



VolumeMount describes a mounting of a Volume within a container.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>mountPath</b></td>
        <td>string</td>
        <td>
          Path within the container at which the volume should be mounted.  Must
not contain ':'.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          This must match the Name of a Volume.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>mountPropagation</b></td>
        <td>string</td>
        <td>
          mountPropagation determines how mounts are propagated from the host
to container and the other way around.
When not set, MountPropagationNone is used.
This field is beta in 1.10.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          Mounted read-only if true, read-write otherwise (false or unspecified).
Defaults to false.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>subPath</b></td>
        <td>string</td>
        <td>
          Path within the volume from which the container's volume should be mounted.
Defaults to "" (volume's root).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>subPathExpr</b></td>
        <td>string</td>
        <td>
          Expanded path within the volume from which the container's volume should be mounted.
Behaves similarly to SubPath but environment variable references $(VAR_NAME) are expanded using the container's environment.
Defaults to "" (volume's root).
SubPathExpr and SubPath are mutually exclusive.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.imagePullSecrets[index]
<sup><sup>[↩ Parent](#redroidtaskspec)</sup></sup>



LocalObjectReference contains enough information to let you locate the
referenced object inside the same namespace.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index]
<sup><sup>[↩ Parent](#redroidtaskspec)</sup></sup>



Volume represents a named volume in a pod that may be accessed by any container in the pod.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          name of the volume.
Must be a DNS_LABEL and unique within the pod.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexawselasticblockstore">awsElasticBlockStore</a></b></td>
        <td>object</td>
        <td>
          awsElasticBlockStore represents an AWS Disk resource that is attached to a
kubelet's host machine and then exposed to the pod.
More info: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexazuredisk">azureDisk</a></b></td>
        <td>object</td>
        <td>
          azureDisk represents an Azure Data Disk mount on the host and bind mount to the pod.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexazurefile">azureFile</a></b></td>
        <td>object</td>
        <td>
          azureFile represents an Azure File Service mount on the host and bind mount to the pod.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexcephfs">cephfs</a></b></td>
        <td>object</td>
        <td>
          cephFS represents a Ceph FS mount on the host that shares a pod's lifetime<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexcinder">cinder</a></b></td>
        <td>object</td>
        <td>
          cinder represents a cinder volume attached and mounted on kubelets host machine.
More info: https://examples.k8s.io/mysql-cinder-pd/README.md<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexconfigmap">configMap</a></b></td>
        <td>object</td>
        <td>
          configMap represents a configMap that should populate this volume<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexcsi">csi</a></b></td>
        <td>object</td>
        <td>
          csi (Container Storage Interface) represents ephemeral storage that is handled by certain external CSI drivers (Beta feature).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexdownwardapi">downwardAPI</a></b></td>
        <td>object</td>
        <td>
          downwardAPI represents downward API about the pod that should populate this volume<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexemptydir">emptyDir</a></b></td>
        <td>object</td>
        <td>
          emptyDir represents a temporary directory that shares a pod's lifetime.
More info: https://kubernetes.io/docs/concepts/storage/volumes#emptydir<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexephemeral">ephemeral</a></b></td>
        <td>object</td>
        <td>
          ephemeral represents a volume that is handled by a cluster storage driver.
The volume's lifecycle is tied to the pod that defines it - it will be created before the pod starts,
and deleted when the pod is removed.

Use this if:
a) the volume is only needed while the pod runs,
b) features of normal volumes like restoring from snapshot or capacity
   tracking are needed,
c) the storage driver is specified through a storage class, and
d) the storage driver supports dynamic volume provisioning through
   a PersistentVolumeClaim (see EphemeralVolumeSource for more
   information on the connection between this volume type
   and PersistentVolumeClaim).

Use PersistentVolumeClaim or one of the vendor-specific
APIs for volumes that persist for longer than the lifecycle
of an individual pod.

Use CSI for light-weight local ephemeral volumes if the CSI driver is meant to
be used that way - see the documentation of the driver for
more information.

A pod can use both types of ephemeral volumes and
persistent volumes at the same time.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexfc">fc</a></b></td>
        <td>object</td>
        <td>
          fc represents a Fibre Channel resource that is attached to a kubelet's host machine and then exposed to the pod.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexflexvolume">flexVolume</a></b></td>
        <td>object</td>
        <td>
          flexVolume represents a generic volume resource that is
provisioned/attached using an exec based plugin.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexflocker">flocker</a></b></td>
        <td>object</td>
        <td>
          flocker represents a Flocker volume attached to a kubelet's host machine. This depends on the Flocker control service being running<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexgcepersistentdisk">gcePersistentDisk</a></b></td>
        <td>object</td>
        <td>
          gcePersistentDisk represents a GCE Disk resource that is attached to a
kubelet's host machine and then exposed to the pod.
More info: https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexgitrepo">gitRepo</a></b></td>
        <td>object</td>
        <td>
          gitRepo represents a git repository at a particular revision.
DEPRECATED: GitRepo is deprecated. To provision a container with a git repo, mount an
EmptyDir into an InitContainer that clones the repo using git, then mount the EmptyDir
into the Pod's container.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexglusterfs">glusterfs</a></b></td>
        <td>object</td>
        <td>
          glusterfs represents a Glusterfs mount on the host that shares a pod's lifetime.
More info: https://examples.k8s.io/volumes/glusterfs/README.md<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexhostpath">hostPath</a></b></td>
        <td>object</td>
        <td>
          hostPath represents a pre-existing file or directory on the host
machine that is directly exposed to the container. This is generally
used for system agents or other privileged things that are allowed
to see the host machine. Most containers will NOT need this.
More info: https://kubernetes.io/docs/concepts/storage/volumes#hostpath<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexiscsi">iscsi</a></b></td>
        <td>object</td>
        <td>
          iscsi represents an ISCSI Disk resource that is attached to a
kubelet's host machine and then exposed to the pod.
More info: https://examples.k8s.io/volumes/iscsi/README.md<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexnfs">nfs</a></b></td>
        <td>object</td>
        <td>
          nfs represents an NFS mount on the host that shares a pod's lifetime
More info: https://kubernetes.io/docs/concepts/storage/volumes#nfs<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexpersistentvolumeclaim">persistentVolumeClaim</a></b></td>
        <td>object</td>
        <td>
          persistentVolumeClaimVolumeSource represents a reference to a
PersistentVolumeClaim in the same namespace.
More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexphotonpersistentdisk">photonPersistentDisk</a></b></td>
        <td>object</td>
        <td>
          photonPersistentDisk represents a PhotonController persistent disk attached and mounted on kubelets host machine<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexportworxvolume">portworxVolume</a></b></td>
        <td>object</td>
        <td>
          portworxVolume represents a portworx volume attached and mounted on kubelets host machine<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexprojected">projected</a></b></td>
        <td>object</td>
        <td>
          projected items for all in one resources secrets, configmaps, and downward API<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexquobyte">quobyte</a></b></td>
        <td>object</td>
        <td>
          quobyte represents a Quobyte mount on the host that shares a pod's lifetime<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexrbd">rbd</a></b></td>
        <td>object</td>
        <td>
          rbd represents a Rados Block Device mount on the host that shares a pod's lifetime.
More info: https://examples.k8s.io/volumes/rbd/README.md<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexscaleio">scaleIO</a></b></td>
        <td>object</td>
        <td>
          scaleIO represents a ScaleIO persistent volume attached and mounted on Kubernetes nodes.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexsecret">secret</a></b></td>
        <td>object</td>
        <td>
          secret represents a secret that should populate this volume.
More info: https://kubernetes.io/docs/concepts/storage/volumes#secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexstorageos">storageos</a></b></td>
        <td>object</td>
        <td>
          storageOS represents a StorageOS volume attached and mounted on Kubernetes nodes.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexvspherevolume">vsphereVolume</a></b></td>
        <td>object</td>
        <td>
          vsphereVolume represents a vSphere volume attached and mounted on kubelets host machine<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].awsElasticBlockStore
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



awsElasticBlockStore represents an AWS Disk resource that is attached to a
kubelet's host machine and then exposed to the pod.
More info: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>volumeID</b></td>
        <td>string</td>
        <td>
          volumeID is unique ID of the persistent disk resource in AWS (Amazon EBS volume).
More info: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fsType is the filesystem type of the volume that you want to mount.
Tip: Ensure that the filesystem type is supported by the host operating system.
Examples: "ext4", "xfs", "ntfs". Implicitly inferred to be "ext4" if unspecified.
More info: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>partition</b></td>
        <td>integer</td>
        <td>
          partition is the partition in the volume that you want to mount.
If omitted, the default is to mount by volume name.
Examples: For volume /dev/sda1, you specify the partition as "1".
Similarly, the volume partition for /dev/sda is "0" (or you can leave the property empty).<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly value true will force the readOnly setting in VolumeMounts.
More info: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].azureDisk
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



azureDisk represents an Azure Data Disk mount on the host and bind mount to the pod.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>diskName</b></td>
        <td>string</td>
        <td>
          diskName is the Name of the data disk in the blob storage<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>diskURI</b></td>
        <td>string</td>
        <td>
          diskURI is the URI of data disk in the blob storage<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>cachingMode</b></td>
        <td>string</td>
        <td>
          cachingMode is the Host Caching mode: None, Read Only, Read Write.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fsType is Filesystem type to mount.
Must be a filesystem type supported by the host operating system.
Ex. "ext4", "xfs", "ntfs". Implicitly inferred to be "ext4" if unspecified.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          kind expected values are Shared: multiple blob disks per storage account  Dedicated: single blob disk per storage account  Managed: azure managed data disk (only in managed availability set). defaults to shared<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly Defaults to false (read/write). ReadOnly here will force
the ReadOnly setting in VolumeMounts.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].azureFile
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



azureFile represents an Azure File Service mount on the host and bind mount to the pod.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>secretName</b></td>
        <td>string</td>
        <td>
          secretName is the  name of secret that contains Azure Storage Account Name and Key<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>shareName</b></td>
        <td>string</td>
        <td>
          shareName is the azure share Name<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly defaults to false (read/write). ReadOnly here will force
the ReadOnly setting in VolumeMounts.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].cephfs
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



cephFS represents a Ceph FS mount on the host that shares a pod's lifetime

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>monitors</b></td>
        <td>[]string</td>
        <td>
          monitors is Required: Monitors is a collection of Ceph monitors
More info: https://examples.k8s.io/volumes/cephfs/README.md#how-to-use-it<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          path is Optional: Used as the mounted root, rather than the full Ceph tree, default is /<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly is Optional: Defaults to false (read/write). ReadOnly here will force
the ReadOnly setting in VolumeMounts.
More info: https://examples.k8s.io/volumes/cephfs/README.md#how-to-use-it<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretFile</b></td>
        <td>string</td>
        <td>
          secretFile is Optional: SecretFile is the path to key ring for User, default is /etc/ceph/user.secret
More info: https://examples.k8s.io/volumes/cephfs/README.md#how-to-use-it<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexcephfssecretref">secretRef</a></b></td>
        <td>object</td>
        <td>
          secretRef is Optional: SecretRef is reference to the authentication secret for User, default is empty.
More info: https://examples.k8s.io/volumes/cephfs/README.md#how-to-use-it<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>user</b></td>
        <td>string</td>
        <td>
          user is optional: User is the rados user name, default is admin
More info: https://examples.k8s.io/volumes/cephfs/README.md#how-to-use-it<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].cephfs.secretRef
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexcephfs)</sup></sup>



secretRef is Optional: SecretRef is reference to the authentication secret for User, default is empty.
More info: https://examples.k8s.io/volumes/cephfs/README.md#how-to-use-it

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].cinder
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



cinder represents a cinder volume attached and mounted on kubelets host machine.
More info: https://examples.k8s.io/mysql-cinder-pd/README.md

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>volumeID</b></td>
        <td>string</td>
        <td>
          volumeID used to identify the volume in cinder.
More info: https://examples.k8s.io/mysql-cinder-pd/README.md<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fsType is the filesystem type to mount.
Must be a filesystem type supported by the host operating system.
Examples: "ext4", "xfs", "ntfs". Implicitly inferred to be "ext4" if unspecified.
More info: https://examples.k8s.io/mysql-cinder-pd/README.md<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly defaults to false (read/write). ReadOnly here will force
the ReadOnly setting in VolumeMounts.
More info: https://examples.k8s.io/mysql-cinder-pd/README.md<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexcindersecretref">secretRef</a></b></td>
        <td>object</td>
        <td>
          secretRef is optional: points to a secret object containing parameters used to connect
to OpenStack.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].cinder.secretRef
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexcinder)</sup></sup>



secretRef is optional: points to a secret object containing parameters used to connect
to OpenStack.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].configMap
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



configMap represents a configMap that should populate this volume

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>defaultMode</b></td>
        <td>integer</td>
        <td>
          defaultMode is optional: mode bits used to set permissions on created files by default.
Must be an octal value between 0000 and 0777 or a decimal value between 0 and 511.
YAML accepts both octal and decimal values, JSON requires decimal values for mode bits.
Defaults to 0644.
Directories within the path are not affected by this setting.
This might be in conflict with other options that affect the file
mode, like fsGroup, and the result can be other mode bits set.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexconfigmapitemsindex">items</a></b></td>
        <td>[]object</td>
        <td>
          items if unspecified, each key-value pair in the Data field of the referenced
ConfigMap will be projected into the volume as a file whose name is the
key and content is the value. If specified, the listed keys will be
projected into the specified paths, and unlisted keys will not be
present. If a key is specified which is not present in the ConfigMap,
the volume setup will error unless it is marked optional. Paths must be
relative and may not contain the '..' path or start with '..'.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          optional specify whether the ConfigMap or its keys must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].configMap.items[index]
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexconfigmap)</sup></sup>



Maps a string key to a path within a volume.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the key to project.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          path is the relative path of the file to map the key to.
May not be an absolute path.
May not contain the path element '..'.
May not start with the string '..'.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>mode</b></td>
        <td>integer</td>
        <td>
          mode is Optional: mode bits used to set permissions on this file.
Must be an octal value between 0000 and 0777 or a decimal value between 0 and 511.
YAML accepts both octal and decimal values, JSON requires decimal values for mode bits.
If not specified, the volume defaultMode will be used.
This might be in conflict with other options that affect the file
mode, like fsGroup, and the result can be other mode bits set.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].csi
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



csi (Container Storage Interface) represents ephemeral storage that is handled by certain external CSI drivers (Beta feature).

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>driver</b></td>
        <td>string</td>
        <td>
          driver is the name of the CSI driver that handles this volume.
Consult with your admin for the correct name as registered in the cluster.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fsType to mount. Ex. "ext4", "xfs", "ntfs".
If not provided, the empty value is passed to the associated CSI driver
which will determine the default filesystem to apply.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexcsinodepublishsecretref">nodePublishSecretRef</a></b></td>
        <td>object</td>
        <td>
          nodePublishSecretRef is a reference to the secret object containing
sensitive information to pass to the CSI driver to complete the CSI
NodePublishVolume and NodeUnpublishVolume calls.
This field is optional, and  may be empty if no secret is required. If the
secret object contains more than one secret, all secret references are passed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly specifies a read-only configuration for the volume.
Defaults to false (read/write).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>volumeAttributes</b></td>
        <td>map[string]string</td>
        <td>
          volumeAttributes stores driver-specific properties that are passed to the CSI
driver. Consult your driver's documentation for supported values.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].csi.nodePublishSecretRef
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexcsi)</sup></sup>



nodePublishSecretRef is a reference to the secret object containing
sensitive information to pass to the CSI driver to complete the CSI
NodePublishVolume and NodeUnpublishVolume calls.
This field is optional, and  may be empty if no secret is required. If the
secret object contains more than one secret, all secret references are passed.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].downwardAPI
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



downwardAPI represents downward API about the pod that should populate this volume

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>defaultMode</b></td>
        <td>integer</td>
        <td>
          Optional: mode bits to use on created files by default. Must be a
Optional: mode bits used to set permissions on created files by default.
Must be an octal value between 0000 and 0777 or a decimal value between 0 and 511.
YAML accepts both octal and decimal values, JSON requires decimal values for mode bits.
Defaults to 0644.
Directories within the path are not affected by this setting.
This might be in conflict with other options that affect the file
mode, like fsGroup, and the result can be other mode bits set.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexdownwardapiitemsindex">items</a></b></td>
        <td>[]object</td>
        <td>
          Items is a list of downward API volume file<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].downwardAPI.items[index]
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexdownwardapi)</sup></sup>



DownwardAPIVolumeFile represents information to create the file containing the pod field

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          Required: Path is  the relative path name of the file to be created. Must not be absolute or contain the '..' path. Must be utf-8 encoded. The first item of the relative path must not start with '..'<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexdownwardapiitemsindexfieldref">fieldRef</a></b></td>
        <td>object</td>
        <td>
          Required: Selects a field of the pod: only annotations, labels, name and namespace are supported.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>mode</b></td>
        <td>integer</td>
        <td>
          Optional: mode bits used to set permissions on this file, must be an octal value
between 0000 and 0777 or a decimal value between 0 and 511.
YAML accepts both octal and decimal values, JSON requires decimal values for mode bits.
If not specified, the volume defaultMode will be used.
This might be in conflict with other options that affect the file
mode, like fsGroup, and the result can be other mode bits set.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexdownwardapiitemsindexresourcefieldref">resourceFieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a resource of the container: only resources limits and requests
(limits.cpu, limits.memory, requests.cpu and requests.memory) are currently supported.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].downwardAPI.items[index].fieldRef
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexdownwardapiitemsindex)</sup></sup>



Required: Selects a field of the pod: only annotations, labels, name and namespace are supported.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>fieldPath</b></td>
        <td>string</td>
        <td>
          Path of the field to select in the specified API version.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          Version of the schema the FieldPath is written in terms of, defaults to "v1".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].downwardAPI.items[index].resourceFieldRef
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexdownwardapiitemsindex)</sup></sup>



Selects a resource of the container: only resources limits and requests
(limits.cpu, limits.memory, requests.cpu and requests.memory) are currently supported.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>resource</b></td>
        <td>string</td>
        <td>
          Required: resource to select<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>containerName</b></td>
        <td>string</td>
        <td>
          Container name: required for volumes, optional for env vars<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>divisor</b></td>
        <td>int or string</td>
        <td>
          Specifies the output format of the exposed resources, defaults to "1"<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].emptyDir
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



emptyDir represents a temporary directory that shares a pod's lifetime.
More info: https://kubernetes.io/docs/concepts/storage/volumes#emptydir

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>medium</b></td>
        <td>string</td>
        <td>
          medium represents what type of storage medium should back this directory.
The default is "" which means to use the node's default medium.
Must be an empty string (default) or Memory.
More info: https://kubernetes.io/docs/concepts/storage/volumes#emptydir<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sizeLimit</b></td>
        <td>int or string</td>
        <td>
          sizeLimit is the total amount of local storage required for this EmptyDir volume.
The size limit is also applicable for memory medium.
The maximum usage on memory medium EmptyDir would be the minimum value between
the SizeLimit specified here and the sum of memory limits of all containers in a pod.
The default is nil which means that the limit is undefined.
More info: https://kubernetes.io/docs/concepts/storage/volumes#emptydir<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].ephemeral
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



ephemeral represents a volume that is handled by a cluster storage driver.
The volume's lifecycle is tied to the pod that defines it - it will be created before the pod starts,
and deleted when the pod is removed.

Use this if:
a) the volume is only needed while the pod runs,
b) features of normal volumes like restoring from snapshot or capacity
   tracking are needed,
c) the storage driver is specified through a storage class, and
d) the storage driver supports dynamic volume provisioning through
   a PersistentVolumeClaim (see EphemeralVolumeSource for more
   information on the connection between this volume type
   and PersistentVolumeClaim).

Use PersistentVolumeClaim or one of the vendor-specific
APIs for volumes that persist for longer than the lifecycle
of an individual pod.

Use CSI for light-weight local ephemeral volumes if the CSI driver is meant to
be used that way - see the documentation of the driver for
more information.

A pod can use both types of ephemeral volumes and
persistent volumes at the same time.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidtaskspecvolumesindexephemeralvolumeclaimtemplate">volumeClaimTemplate</a></b></td>
        <td>object</td>
        <td>
          Will be used to create a stand-alone PVC to provision the volume.
The pod in which this EphemeralVolumeSource is embedded will be the
owner of the PVC, i.e. the PVC will be deleted together with the
pod.  The name of the PVC will be `<pod name>-<volume name>` where
`<volume name>` is the name from the `PodSpec.Volumes` array
entry. Pod validation will reject the pod if the concatenated name
is not valid for a PVC (for example, too long).

An existing PVC with that name that is not owned by the pod
will *not* be used for the pod to avoid using an unrelated
volume by mistake. Starting the pod is then blocked until
the unrelated PVC is removed. If such a pre-created PVC is
meant to be used by the pod, the PVC has to updated with an
owner reference to the pod once the pod exists. Normally
this should not be necessary, but it may be useful when
manually reconstructing a broken cluster.

This field is read-only and no changes will be made by Kubernetes
to the PVC after it has been created.

Required, must not be nil.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].ephemeral.volumeClaimTemplate
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexephemeral)</sup></sup>



Will be used to create a stand-alone PVC to provision the volume.
The pod in which this EphemeralVolumeSource is embedded will be the
owner of the PVC, i.e. the PVC will be deleted together with the
pod.  The name of the PVC will be `<pod name>-<volume name>` where
`<volume name>` is the name from the `PodSpec.Volumes` array
entry. Pod validation will reject the pod if the concatenated name
is not valid for a PVC (for example, too long).

An existing PVC with that name that is not owned by the pod
will *not* be used for the pod to avoid using an unrelated
volume by mistake. Starting the pod is then blocked until
the unrelated PVC is removed. If such a pre-created PVC is
meant to be used by the pod, the PVC has to updated with an
owner reference to the pod once the pod exists. Normally
this should not be necessary, but it may be useful when
manually reconstructing a broken cluster.

This field is read-only and no changes will be made by Kubernetes
to the PVC after it has been created.

Required, must not be nil.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidtaskspecvolumesindexephemeralvolumeclaimtemplatespec">spec</a></b></td>
        <td>object</td>
        <td>
          The specification for the PersistentVolumeClaim. The entire content is
copied unchanged into the PVC that gets created from this
template. The same fields as in a PersistentVolumeClaim
are also valid here.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexephemeralvolumeclaimtemplatemetadata">metadata</a></b></td>
        <td>object</td>
        <td>
          May contain labels and annotations that will be copied into the PVC
when creating it. No other fields are allowed and will be rejected during
validation.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].ephemeral.volumeClaimTemplate.spec
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexephemeralvolumeclaimtemplate)</sup></sup>



The specification for the PersistentVolumeClaim. The entire content is
copied unchanged into the PVC that gets created from this
template. The same fields as in a PersistentVolumeClaim
are also valid here.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>accessModes</b></td>
        <td>[]string</td>
        <td>
          accessModes contains the desired access modes the volume should have.
More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexephemeralvolumeclaimtemplatespecdatasource">dataSource</a></b></td>
        <td>object</td>
        <td>
          dataSource field can be used to specify either:
* An existing VolumeSnapshot object (snapshot.storage.k8s.io/VolumeSnapshot)
* An existing PVC (PersistentVolumeClaim)
If the provisioner or an external controller can support the specified data source,
it will create a new volume based on the contents of the specified data source.
When the AnyVolumeDataSource feature gate is enabled, dataSource contents will be copied to dataSourceRef,
and dataSourceRef contents will be copied to dataSource when dataSourceRef.namespace is not specified.
If the namespace is specified, then dataSourceRef will not be copied to dataSource.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexephemeralvolumeclaimtemplatespecdatasourceref">dataSourceRef</a></b></td>
        <td>object</td>
        <td>
          dataSourceRef specifies the object from which to populate the volume with data, if a non-empty
volume is desired. This may be any object from a non-empty API group (non
core object) or a PersistentVolumeClaim object.
When this field is specified, volume binding will only succeed if the type of
the specified object matches some installed volume populator or dynamic
provisioner.
This field will replace the functionality of the dataSource field and as such
if both fields are non-empty, they must have the same value. For backwards
compatibility, when namespace isn't specified in dataSourceRef,
both fields (dataSource and dataSourceRef) will be set to the same
value automatically if one of them is empty and the other is non-empty.
When namespace is specified in dataSourceRef,
dataSource isn't set to the same value and must be empty.
There are three important differences between dataSource and dataSourceRef:
* While dataSource only allows two specific types of objects, dataSourceRef
  allows any non-core object, as well as PersistentVolumeClaim objects.
* While dataSource ignores disallowed values (dropping them), dataSourceRef
  preserves all values, and generates an error if a disallowed value is
  specified.
* While dataSource only allows local objects, dataSourceRef allows objects
  in any namespaces.
(Beta) Using this field requires the AnyVolumeDataSource feature gate to be enabled.
(Alpha) Using the namespace field of dataSourceRef requires the CrossNamespaceVolumeDataSource feature gate to be enabled.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexephemeralvolumeclaimtemplatespecresources">resources</a></b></td>
        <td>object</td>
        <td>
          resources represents the minimum resources the volume should have.
If RecoverVolumeExpansionFailure feature is enabled users are allowed to specify resource requirements
that are lower than previous value but must still be higher than capacity recorded in the
status field of the claim.
More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexephemeralvolumeclaimtemplatespecselector">selector</a></b></td>
        <td>object</td>
        <td>
          selector is a label query over volumes to consider for binding.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>storageClassName</b></td>
        <td>string</td>
        <td>
          storageClassName is the name of the StorageClass required by the claim.
More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>volumeAttributesClassName</b></td>
        <td>string</td>
        <td>
          volumeAttributesClassName may be used to set the VolumeAttributesClass used by this claim.
If specified, the CSI driver will create or update the volume with the attributes defined
in the corresponding VolumeAttributesClass. This has a different purpose than storageClassName,
it can be changed after the claim is created. An empty string value means that no VolumeAttributesClass
will be applied to the claim but it's not allowed to reset this field to empty string once it is set.
If unspecified and the PersistentVolumeClaim is unbound, the default VolumeAttributesClass
will be set by the persistentvolume controller if it exists.
If the resource referred to by volumeAttributesClass does not exist, this PersistentVolumeClaim will be
set to a Pending state, as reflected by the modifyVolumeStatus field, until such as a resource
exists.
More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#volumeattributesclass
(Alpha) Using this field requires the VolumeAttributesClass feature gate to be enabled.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>volumeMode</b></td>
        <td>string</td>
        <td>
          volumeMode defines what type of volume is required by the claim.
Value of Filesystem is implied when not included in claim spec.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>volumeName</b></td>
        <td>string</td>
        <td>
          volumeName is the binding reference to the PersistentVolume backing this claim.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].ephemeral.volumeClaimTemplate.spec.dataSource
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexephemeralvolumeclaimtemplatespec)</sup></sup>



dataSource field can be used to specify either:
* An existing VolumeSnapshot object (snapshot.storage.k8s.io/VolumeSnapshot)
* An existing PVC (PersistentVolumeClaim)
If the provisioner or an external controller can support the specified data source,
it will create a new volume based on the contents of the specified data source.
When the AnyVolumeDataSource feature gate is enabled, dataSource contents will be copied to dataSourceRef,
and dataSourceRef contents will be copied to dataSource when dataSourceRef.namespace is not specified.
If the namespace is specified, then dataSourceRef will not be copied to dataSource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind is the type of resource being referenced<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of resource being referenced<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiGroup</b></td>
        <td>string</td>
        <td>
          APIGroup is the group for the resource being referenced.
If APIGroup is not specified, the specified Kind must be in the core API group.
For any other third-party types, APIGroup is required.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].ephemeral.volumeClaimTemplate.spec.dataSourceRef
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexephemeralvolumeclaimtemplatespec)</sup></sup>



dataSourceRef specifies the object from which to populate the volume with data, if a non-empty
volume is desired. This may be any object from a non-empty API group (non
core object) or a PersistentVolumeClaim object.
When this field is specified, volume binding will only succeed if the type of
the specified object matches some installed volume populator or dynamic
provisioner.
This field will replace the functionality of the dataSource field and as such
if both fields are non-empty, they must have the same value. For backwards
compatibility, when namespace isn't specified in dataSourceRef,
both fields (dataSource and dataSourceRef) will be set to the same
value automatically if one of them is empty and the other is non-empty.
When namespace is specified in dataSourceRef,
dataSource isn't set to the same value and must be empty.
There are three important differences between dataSource and dataSourceRef:
* While dataSource only allows two specific types of objects, dataSourceRef
  allows any non-core object, as well as PersistentVolumeClaim objects.
* While dataSource ignores disallowed values (dropping them), dataSourceRef
  preserves all values, and generates an error if a disallowed value is
  specified.
* While dataSource only allows local objects, dataSourceRef allows objects
  in any namespaces.
(Beta) Using this field requires the AnyVolumeDataSource feature gate to be enabled.
(Alpha) Using the namespace field of dataSourceRef requires the CrossNamespaceVolumeDataSource feature gate to be enabled.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind is the type of resource being referenced<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of resource being referenced<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiGroup</b></td>
        <td>string</td>
        <td>
          APIGroup is the group for the resource being referenced.
If APIGroup is not specified, the specified Kind must be in the core API group.
For any other third-party types, APIGroup is required.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace is the namespace of resource being referenced
Note that when a namespace is specified, a gateway.networking.k8s.io/ReferenceGrant object is required in the referent namespace to allow that namespace's owner to accept the reference. See the ReferenceGrant documentation for details.
(Alpha) This field requires the CrossNamespaceVolumeDataSource feature gate to be enabled.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].ephemeral.volumeClaimTemplate.spec.resources
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexephemeralvolumeclaimtemplatespec)</sup></sup>



resources represents the minimum resources the volume should have.
If RecoverVolumeExpansionFailure feature is enabled users are allowed to specify resource requirements
that are lower than previous value but must still be higher than capacity recorded in the
status field of the claim.
More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>limits</b></td>
        <td>map[string]int or string</td>
        <td>
          Limits describes the maximum amount of compute resources allowed.
More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>requests</b></td>
        <td>map[string]int or string</td>
        <td>
          Requests describes the minimum amount of compute resources required.
If Requests is omitted for a container, it defaults to Limits if that is explicitly specified,
otherwise to an implementation-defined value. Requests cannot exceed Limits.
More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].ephemeral.volumeClaimTemplate.spec.selector
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexephemeralvolumeclaimtemplatespec)</sup></sup>



selector is a label query over volumes to consider for binding.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidtaskspecvolumesindexephemeralvolumeclaimtemplatespecselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].ephemeral.volumeClaimTemplate.spec.selector.matchExpressions[index]
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexephemeralvolumeclaimtemplatespecselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].ephemeral.volumeClaimTemplate.metadata
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexephemeralvolumeclaimtemplate)</sup></sup>



May contain labels and annotations that will be copied into the PVC
when creating it. No other fields are allowed and will be rejected during
validation.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>annotations</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>finalizers</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>labels</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].fc
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



fc represents a Fibre Channel resource that is attached to a kubelet's host machine and then exposed to the pod.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fsType is the filesystem type to mount.
Must be a filesystem type supported by the host operating system.
Ex. "ext4", "xfs", "ntfs". Implicitly inferred to be "ext4" if unspecified.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>lun</b></td>
        <td>integer</td>
        <td>
          lun is Optional: FC target lun number<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly is Optional: Defaults to false (read/write). ReadOnly here will force
the ReadOnly setting in VolumeMounts.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>targetWWNs</b></td>
        <td>[]string</td>
        <td>
          targetWWNs is Optional: FC target worldwide names (WWNs)<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>wwids</b></td>
        <td>[]string</td>
        <td>
          wwids Optional: FC volume world wide identifiers (wwids)
Either wwids or combination of targetWWNs and lun must be set, but not both simultaneously.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].flexVolume
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



flexVolume represents a generic volume resource that is
provisioned/attached using an exec based plugin.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>driver</b></td>
        <td>string</td>
        <td>
          driver is the name of the driver to use for this volume.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fsType is the filesystem type to mount.
Must be a filesystem type supported by the host operating system.
Ex. "ext4", "xfs", "ntfs". The default filesystem depends on FlexVolume script.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>options</b></td>
        <td>map[string]string</td>
        <td>
          options is Optional: this field holds extra command options if any.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly is Optional: defaults to false (read/write). ReadOnly here will force
the ReadOnly setting in VolumeMounts.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexflexvolumesecretref">secretRef</a></b></td>
        <td>object</td>
        <td>
          secretRef is Optional: secretRef is reference to the secret object containing
sensitive information to pass to the plugin scripts. This may be
empty if no secret object is specified. If the secret object
contains more than one secret, all secrets are passed to the plugin
scripts.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].flexVolume.secretRef
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexflexvolume)</sup></sup>



secretRef is Optional: secretRef is reference to the secret object containing
sensitive information to pass to the plugin scripts. This may be
empty if no secret object is specified. If the secret object
contains more than one secret, all secrets are passed to the plugin
scripts.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].flocker
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



flocker represents a Flocker volume attached to a kubelet's host machine. This depends on the Flocker control service being running

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>datasetName</b></td>
        <td>string</td>
        <td>
          datasetName is Name of the dataset stored as metadata -> name on the dataset for Flocker
should be considered as deprecated<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>datasetUUID</b></td>
        <td>string</td>
        <td>
          datasetUUID is the UUID of the dataset. This is unique identifier of a Flocker dataset<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].gcePersistentDisk
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



gcePersistentDisk represents a GCE Disk resource that is attached to a
kubelet's host machine and then exposed to the pod.
More info: https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>pdName</b></td>
        <td>string</td>
        <td>
          pdName is unique name of the PD resource in GCE. Used to identify the disk in GCE.
More info: https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fsType is filesystem type of the volume that you want to mount.
Tip: Ensure that the filesystem type is supported by the host operating system.
Examples: "ext4", "xfs", "ntfs". Implicitly inferred to be "ext4" if unspecified.
More info: https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>partition</b></td>
        <td>integer</td>
        <td>
          partition is the partition in the volume that you want to mount.
If omitted, the default is to mount by volume name.
Examples: For volume /dev/sda1, you specify the partition as "1".
Similarly, the volume partition for /dev/sda is "0" (or you can leave the property empty).
More info: https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly here will force the ReadOnly setting in VolumeMounts.
Defaults to false.
More info: https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].gitRepo
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



gitRepo represents a git repository at a particular revision.
DEPRECATED: GitRepo is deprecated. To provision a container with a git repo, mount an
EmptyDir into an InitContainer that clones the repo using git, then mount the EmptyDir
into the Pod's container.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>repository</b></td>
        <td>string</td>
        <td>
          repository is the URL<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>directory</b></td>
        <td>string</td>
        <td>
          directory is the target directory name.
Must not contain or start with '..'.  If '.' is supplied, the volume directory will be the
git repository.  Otherwise, if specified, the volume will contain the git repository in
the subdirectory with the given name.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>revision</b></td>
        <td>string</td>
        <td>
          revision is the commit hash for the specified revision.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].glusterfs
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



glusterfs represents a Glusterfs mount on the host that shares a pod's lifetime.
More info: https://examples.k8s.io/volumes/glusterfs/README.md

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>endpoints</b></td>
        <td>string</td>
        <td>
          endpoints is the endpoint name that details Glusterfs topology.
More info: https://examples.k8s.io/volumes/glusterfs/README.md#create-a-pod<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          path is the Glusterfs volume path.
More info: https://examples.k8s.io/volumes/glusterfs/README.md#create-a-pod<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly here will force the Glusterfs volume to be mounted with read-only permissions.
Defaults to false.
More info: https://examples.k8s.io/volumes/glusterfs/README.md#create-a-pod<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].hostPath
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



hostPath represents a pre-existing file or directory on the host
machine that is directly exposed to the container. This is generally
used for system agents or other privileged things that are allowed
to see the host machine. Most containers will NOT need this.
More info: https://kubernetes.io/docs/concepts/storage/volumes#hostpath

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          path of the directory on the host.
If the path is a symlink, it will follow the link to the real path.
More info: https://kubernetes.io/docs/concepts/storage/volumes#hostpath<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type for HostPath Volume
Defaults to ""
More info: https://kubernetes.io/docs/concepts/storage/volumes#hostpath<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].iscsi
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



iscsi represents an ISCSI Disk resource that is attached to a
kubelet's host machine and then exposed to the pod.
More info: https://examples.k8s.io/volumes/iscsi/README.md

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>iqn</b></td>
        <td>string</td>
        <td>
          iqn is the target iSCSI Qualified Name.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>lun</b></td>
        <td>integer</td>
        <td>
          lun represents iSCSI Target Lun number.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>targetPortal</b></td>
        <td>string</td>
        <td>
          targetPortal is iSCSI Target Portal. The Portal is either an IP or ip_addr:port if the port
is other than default (typically TCP ports 860 and 3260).<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>chapAuthDiscovery</b></td>
        <td>boolean</td>
        <td>
          chapAuthDiscovery defines whether support iSCSI Discovery CHAP authentication<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>chapAuthSession</b></td>
        <td>boolean</td>
        <td>
          chapAuthSession defines whether support iSCSI Session CHAP authentication<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fsType is the filesystem type of the volume that you want to mount.
Tip: Ensure that the filesystem type is supported by the host operating system.
Examples: "ext4", "xfs", "ntfs". Implicitly inferred to be "ext4" if unspecified.
More info: https://kubernetes.io/docs/concepts/storage/volumes#iscsi<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>initiatorName</b></td>
        <td>string</td>
        <td>
          initiatorName is the custom iSCSI Initiator Name.
If initiatorName is specified with iscsiInterface simultaneously, new iSCSI interface
<target portal>:<volume name> will be created for the connection.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>iscsiInterface</b></td>
        <td>string</td>
        <td>
          iscsiInterface is the interface Name that uses an iSCSI transport.
Defaults to 'default' (tcp).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>portals</b></td>
        <td>[]string</td>
        <td>
          portals is the iSCSI Target Portal List. The portal is either an IP or ip_addr:port if the port
is other than default (typically TCP ports 860 and 3260).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly here will force the ReadOnly setting in VolumeMounts.
Defaults to false.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexiscsisecretref">secretRef</a></b></td>
        <td>object</td>
        <td>
          secretRef is the CHAP Secret for iSCSI target and initiator authentication<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].iscsi.secretRef
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexiscsi)</sup></sup>



secretRef is the CHAP Secret for iSCSI target and initiator authentication

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].nfs
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



nfs represents an NFS mount on the host that shares a pod's lifetime
More info: https://kubernetes.io/docs/concepts/storage/volumes#nfs

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          path that is exported by the NFS server.
More info: https://kubernetes.io/docs/concepts/storage/volumes#nfs<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>server</b></td>
        <td>string</td>
        <td>
          server is the hostname or IP address of the NFS server.
More info: https://kubernetes.io/docs/concepts/storage/volumes#nfs<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly here will force the NFS export to be mounted with read-only permissions.
Defaults to false.
More info: https://kubernetes.io/docs/concepts/storage/volumes#nfs<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].persistentVolumeClaim
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



persistentVolumeClaimVolumeSource represents a reference to a
PersistentVolumeClaim in the same namespace.
More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>claimName</b></td>
        <td>string</td>
        <td>
          claimName is the name of a PersistentVolumeClaim in the same namespace as the pod using this volume.
More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly Will force the ReadOnly setting in VolumeMounts.
Default false.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].photonPersistentDisk
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



photonPersistentDisk represents a PhotonController persistent disk attached and mounted on kubelets host machine

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>pdID</b></td>
        <td>string</td>
        <td>
          pdID is the ID that identifies Photon Controller persistent disk<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fsType is the filesystem type to mount.
Must be a filesystem type supported by the host operating system.
Ex. "ext4", "xfs", "ntfs". Implicitly inferred to be "ext4" if unspecified.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].portworxVolume
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



portworxVolume represents a portworx volume attached and mounted on kubelets host machine

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>volumeID</b></td>
        <td>string</td>
        <td>
          volumeID uniquely identifies a Portworx volume<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fSType represents the filesystem type to mount
Must be a filesystem type supported by the host operating system.
Ex. "ext4", "xfs". Implicitly inferred to be "ext4" if unspecified.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly defaults to false (read/write). ReadOnly here will force
the ReadOnly setting in VolumeMounts.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].projected
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



projected items for all in one resources secrets, configmaps, and downward API

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>defaultMode</b></td>
        <td>integer</td>
        <td>
          defaultMode are the mode bits used to set permissions on created files by default.
Must be an octal value between 0000 and 0777 or a decimal value between 0 and 511.
YAML accepts both octal and decimal values, JSON requires decimal values for mode bits.
Directories within the path are not affected by this setting.
This might be in conflict with other options that affect the file
mode, like fsGroup, and the result can be other mode bits set.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexprojectedsourcesindex">sources</a></b></td>
        <td>[]object</td>
        <td>
          sources is the list of volume projections<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].projected.sources[index]
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexprojected)</sup></sup>



Projection that may be projected along with other supported volume types

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidtaskspecvolumesindexprojectedsourcesindexclustertrustbundle">clusterTrustBundle</a></b></td>
        <td>object</td>
        <td>
          ClusterTrustBundle allows a pod to access the `.spec.trustBundle` field
of ClusterTrustBundle objects in an auto-updating file.

Alpha, gated by the ClusterTrustBundleProjection feature gate.

ClusterTrustBundle objects can either be selected by name, or by the
combination of signer name and a label selector.

Kubelet performs aggressive normalization of the PEM contents written
into the pod filesystem.  Esoteric PEM features such as inter-block
comments and block headers are stripped.  Certificates are deduplicated.
The ordering of certificates within the file is arbitrary, and Kubelet
may change the order over time.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexprojectedsourcesindexconfigmap">configMap</a></b></td>
        <td>object</td>
        <td>
          configMap information about the configMap data to project<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexprojectedsourcesindexdownwardapi">downwardAPI</a></b></td>
        <td>object</td>
        <td>
          downwardAPI information about the downwardAPI data to project<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexprojectedsourcesindexsecret">secret</a></b></td>
        <td>object</td>
        <td>
          secret information about the secret data to project<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexprojectedsourcesindexserviceaccounttoken">serviceAccountToken</a></b></td>
        <td>object</td>
        <td>
          serviceAccountToken is information about the serviceAccountToken data to project<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].projected.sources[index].clusterTrustBundle
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexprojectedsourcesindex)</sup></sup>



ClusterTrustBundle allows a pod to access the `.spec.trustBundle` field
of ClusterTrustBundle objects in an auto-updating file.

Alpha, gated by the ClusterTrustBundleProjection feature gate.

ClusterTrustBundle objects can either be selected by name, or by the
combination of signer name and a label selector.

Kubelet performs aggressive normalization of the PEM contents written
into the pod filesystem.  Esoteric PEM features such as inter-block
comments and block headers are stripped.  Certificates are deduplicated.
The ordering of certificates within the file is arbitrary, and Kubelet
may change the order over time.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          Relative path from the volume root to write the bundle.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexprojectedsourcesindexclustertrustbundlelabelselector">labelSelector</a></b></td>
        <td>object</td>
        <td>
          Select all ClusterTrustBundles that match this label selector.  Only has
effect if signerName is set.  Mutually-exclusive with name.  If unset,
interpreted as "match nothing".  If set but empty, interpreted as "match
everything".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Select a single ClusterTrustBundle by object name.  Mutually-exclusive
with signerName and labelSelector.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          If true, don't block pod startup if the referenced ClusterTrustBundle(s)
aren't available.  If using name, then the named ClusterTrustBundle is
allowed not to exist.  If using signerName, then the combination of
signerName and labelSelector is allowed to match zero
ClusterTrustBundles.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>signerName</b></td>
        <td>string</td>
        <td>
          Select all ClusterTrustBundles that match this signer name.
Mutually-exclusive with name.  The contents of all selected
ClusterTrustBundles will be unified and deduplicated.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].projected.sources[index].clusterTrustBundle.labelSelector
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexprojectedsourcesindexclustertrustbundle)</sup></sup>



Select all ClusterTrustBundles that match this label selector.  Only has
effect if signerName is set.  Mutually-exclusive with name.  If unset,
interpreted as "match nothing".  If set but empty, interpreted as "match
everything".

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidtaskspecvolumesindexprojectedsourcesindexclustertrustbundlelabelselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].projected.sources[index].clusterTrustBundle.labelSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexprojectedsourcesindexclustertrustbundlelabelselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].projected.sources[index].configMap
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexprojectedsourcesindex)</sup></sup>



configMap information about the configMap data to project

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidtaskspecvolumesindexprojectedsourcesindexconfigmapitemsindex">items</a></b></td>
        <td>[]object</td>
        <td>
          items if unspecified, each key-value pair in the Data field of the referenced
ConfigMap will be projected into the volume as a file whose name is the
key and content is the value. If specified, the listed keys will be
projected into the specified paths, and unlisted keys will not be
present. If a key is specified which is not present in the ConfigMap,
the volume setup will error unless it is marked optional. Paths must be
relative and may not contain the '..' path or start with '..'.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          optional specify whether the ConfigMap or its keys must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].projected.sources[index].configMap.items[index]
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexprojectedsourcesindexconfigmap)</sup></sup>



Maps a string key to a path within a volume.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the key to project.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          path is the relative path of the file to map the key to.
May not be an absolute path.
May not contain the path element '..'.
May not start with the string '..'.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>mode</b></td>
        <td>integer</td>
        <td>
          mode is Optional: mode bits used to set permissions on this file.
Must be an octal value between 0000 and 0777 or a decimal value between 0 and 511.
YAML accepts both octal and decimal values, JSON requires decimal values for mode bits.
If not specified, the volume defaultMode will be used.
This might be in conflict with other options that affect the file
mode, like fsGroup, and the result can be other mode bits set.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].projected.sources[index].downwardAPI
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexprojectedsourcesindex)</sup></sup>



downwardAPI information about the downwardAPI data to project

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidtaskspecvolumesindexprojectedsourcesindexdownwardapiitemsindex">items</a></b></td>
        <td>[]object</td>
        <td>
          Items is a list of DownwardAPIVolume file<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].projected.sources[index].downwardAPI.items[index]
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexprojectedsourcesindexdownwardapi)</sup></sup>



DownwardAPIVolumeFile represents information to create the file containing the pod field

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          Required: Path is  the relative path name of the file to be created. Must not be absolute or contain the '..' path. Must be utf-8 encoded. The first item of the relative path must not start with '..'<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexprojectedsourcesindexdownwardapiitemsindexfieldref">fieldRef</a></b></td>
        <td>object</td>
        <td>
          Required: Selects a field of the pod: only annotations, labels, name and namespace are supported.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>mode</b></td>
        <td>integer</td>
        <td>
          Optional: mode bits used to set permissions on this file, must be an octal value
between 0000 and 0777 or a decimal value between 0 and 511.
YAML accepts both octal and decimal values, JSON requires decimal values for mode bits.
If not specified, the volume defaultMode will be used.
This might be in conflict with other options that affect the file
mode, like fsGroup, and the result can be other mode bits set.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexprojectedsourcesindexdownwardapiitemsindexresourcefieldref">resourceFieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a resource of the container: only resources limits and requests
(limits.cpu, limits.memory, requests.cpu and requests.memory) are currently supported.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].projected.sources[index].downwardAPI.items[index].fieldRef
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexprojectedsourcesindexdownwardapiitemsindex)</sup></sup>



Required: Selects a field of the pod: only annotations, labels, name and namespace are supported.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>fieldPath</b></td>
        <td>string</td>
        <td>
          Path of the field to select in the specified API version.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          Version of the schema the FieldPath is written in terms of, defaults to "v1".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].projected.sources[index].downwardAPI.items[index].resourceFieldRef
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexprojectedsourcesindexdownwardapiitemsindex)</sup></sup>



Selects a resource of the container: only resources limits and requests
(limits.cpu, limits.memory, requests.cpu and requests.memory) are currently supported.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>resource</b></td>
        <td>string</td>
        <td>
          Required: resource to select<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>containerName</b></td>
        <td>string</td>
        <td>
          Container name: required for volumes, optional for env vars<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>divisor</b></td>
        <td>int or string</td>
        <td>
          Specifies the output format of the exposed resources, defaults to "1"<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].projected.sources[index].secret
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexprojectedsourcesindex)</sup></sup>



secret information about the secret data to project

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#redroidtaskspecvolumesindexprojectedsourcesindexsecretitemsindex">items</a></b></td>
        <td>[]object</td>
        <td>
          items if unspecified, each key-value pair in the Data field of the referenced
Secret will be projected into the volume as a file whose name is the
key and content is the value. If specified, the listed keys will be
projected into the specified paths, and unlisted keys will not be
present. If a key is specified which is not present in the Secret,
the volume setup will error unless it is marked optional. Paths must be
relative and may not contain the '..' path or start with '..'.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          optional field specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].projected.sources[index].secret.items[index]
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexprojectedsourcesindexsecret)</sup></sup>



Maps a string key to a path within a volume.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the key to project.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          path is the relative path of the file to map the key to.
May not be an absolute path.
May not contain the path element '..'.
May not start with the string '..'.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>mode</b></td>
        <td>integer</td>
        <td>
          mode is Optional: mode bits used to set permissions on this file.
Must be an octal value between 0000 and 0777 or a decimal value between 0 and 511.
YAML accepts both octal and decimal values, JSON requires decimal values for mode bits.
If not specified, the volume defaultMode will be used.
This might be in conflict with other options that affect the file
mode, like fsGroup, and the result can be other mode bits set.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].projected.sources[index].serviceAccountToken
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexprojectedsourcesindex)</sup></sup>



serviceAccountToken is information about the serviceAccountToken data to project

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          path is the path relative to the mount point of the file to project the
token into.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>audience</b></td>
        <td>string</td>
        <td>
          audience is the intended audience of the token. A recipient of a token
must identify itself with an identifier specified in the audience of the
token, and otherwise should reject the token. The audience defaults to the
identifier of the apiserver.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>expirationSeconds</b></td>
        <td>integer</td>
        <td>
          expirationSeconds is the requested duration of validity of the service
account token. As the token approaches expiration, the kubelet volume
plugin will proactively rotate the service account token. The kubelet will
start trying to rotate the token if the token is older than 80 percent of
its time to live or if the token is older than 24 hours.Defaults to 1 hour
and must be at least 10 minutes.<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].quobyte
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



quobyte represents a Quobyte mount on the host that shares a pod's lifetime

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>registry</b></td>
        <td>string</td>
        <td>
          registry represents a single or multiple Quobyte Registry services
specified as a string as host:port pair (multiple entries are separated with commas)
which acts as the central registry for volumes<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>volume</b></td>
        <td>string</td>
        <td>
          volume is a string that references an already created Quobyte volume by name.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>group</b></td>
        <td>string</td>
        <td>
          group to map volume access to
Default is no group<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly here will force the Quobyte volume to be mounted with read-only permissions.
Defaults to false.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>tenant</b></td>
        <td>string</td>
        <td>
          tenant owning the given Quobyte volume in the Backend
Used with dynamically provisioned Quobyte volumes, value is set by the plugin<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>user</b></td>
        <td>string</td>
        <td>
          user to map volume access to
Defaults to serivceaccount user<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].rbd
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



rbd represents a Rados Block Device mount on the host that shares a pod's lifetime.
More info: https://examples.k8s.io/volumes/rbd/README.md

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>image</b></td>
        <td>string</td>
        <td>
          image is the rados image name.
More info: https://examples.k8s.io/volumes/rbd/README.md#how-to-use-it<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>monitors</b></td>
        <td>[]string</td>
        <td>
          monitors is a collection of Ceph monitors.
More info: https://examples.k8s.io/volumes/rbd/README.md#how-to-use-it<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fsType is the filesystem type of the volume that you want to mount.
Tip: Ensure that the filesystem type is supported by the host operating system.
Examples: "ext4", "xfs", "ntfs". Implicitly inferred to be "ext4" if unspecified.
More info: https://kubernetes.io/docs/concepts/storage/volumes#rbd<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>keyring</b></td>
        <td>string</td>
        <td>
          keyring is the path to key ring for RBDUser.
Default is /etc/ceph/keyring.
More info: https://examples.k8s.io/volumes/rbd/README.md#how-to-use-it<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>pool</b></td>
        <td>string</td>
        <td>
          pool is the rados pool name.
Default is rbd.
More info: https://examples.k8s.io/volumes/rbd/README.md#how-to-use-it<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly here will force the ReadOnly setting in VolumeMounts.
Defaults to false.
More info: https://examples.k8s.io/volumes/rbd/README.md#how-to-use-it<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexrbdsecretref">secretRef</a></b></td>
        <td>object</td>
        <td>
          secretRef is name of the authentication secret for RBDUser. If provided
overrides keyring.
Default is nil.
More info: https://examples.k8s.io/volumes/rbd/README.md#how-to-use-it<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>user</b></td>
        <td>string</td>
        <td>
          user is the rados user name.
Default is admin.
More info: https://examples.k8s.io/volumes/rbd/README.md#how-to-use-it<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].rbd.secretRef
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexrbd)</sup></sup>



secretRef is name of the authentication secret for RBDUser. If provided
overrides keyring.
Default is nil.
More info: https://examples.k8s.io/volumes/rbd/README.md#how-to-use-it

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].scaleIO
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



scaleIO represents a ScaleIO persistent volume attached and mounted on Kubernetes nodes.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>gateway</b></td>
        <td>string</td>
        <td>
          gateway is the host address of the ScaleIO API Gateway.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexscaleiosecretref">secretRef</a></b></td>
        <td>object</td>
        <td>
          secretRef references to the secret for ScaleIO user and other
sensitive information. If this is not provided, Login operation will fail.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>system</b></td>
        <td>string</td>
        <td>
          system is the name of the storage system as configured in ScaleIO.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fsType is the filesystem type to mount.
Must be a filesystem type supported by the host operating system.
Ex. "ext4", "xfs", "ntfs".
Default is "xfs".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>protectionDomain</b></td>
        <td>string</td>
        <td>
          protectionDomain is the name of the ScaleIO Protection Domain for the configured storage.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly Defaults to false (read/write). ReadOnly here will force
the ReadOnly setting in VolumeMounts.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sslEnabled</b></td>
        <td>boolean</td>
        <td>
          sslEnabled Flag enable/disable SSL communication with Gateway, default false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>storageMode</b></td>
        <td>string</td>
        <td>
          storageMode indicates whether the storage for a volume should be ThickProvisioned or ThinProvisioned.
Default is ThinProvisioned.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>storagePool</b></td>
        <td>string</td>
        <td>
          storagePool is the ScaleIO Storage Pool associated with the protection domain.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>volumeName</b></td>
        <td>string</td>
        <td>
          volumeName is the name of a volume already created in the ScaleIO system
that is associated with this volume source.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].scaleIO.secretRef
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexscaleio)</sup></sup>



secretRef references to the secret for ScaleIO user and other
sensitive information. If this is not provided, Login operation will fail.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].secret
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



secret represents a secret that should populate this volume.
More info: https://kubernetes.io/docs/concepts/storage/volumes#secret

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>defaultMode</b></td>
        <td>integer</td>
        <td>
          defaultMode is Optional: mode bits used to set permissions on created files by default.
Must be an octal value between 0000 and 0777 or a decimal value between 0 and 511.
YAML accepts both octal and decimal values, JSON requires decimal values
for mode bits. Defaults to 0644.
Directories within the path are not affected by this setting.
This might be in conflict with other options that affect the file
mode, like fsGroup, and the result can be other mode bits set.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexsecretitemsindex">items</a></b></td>
        <td>[]object</td>
        <td>
          items If unspecified, each key-value pair in the Data field of the referenced
Secret will be projected into the volume as a file whose name is the
key and content is the value. If specified, the listed keys will be
projected into the specified paths, and unlisted keys will not be
present. If a key is specified which is not present in the Secret,
the volume setup will error unless it is marked optional. Paths must be
relative and may not contain the '..' path or start with '..'.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          optional field specify whether the Secret or its keys must be defined<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretName</b></td>
        <td>string</td>
        <td>
          secretName is the name of the secret in the pod's namespace to use.
More info: https://kubernetes.io/docs/concepts/storage/volumes#secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].secret.items[index]
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexsecret)</sup></sup>



Maps a string key to a path within a volume.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the key to project.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          path is the relative path of the file to map the key to.
May not be an absolute path.
May not contain the path element '..'.
May not start with the string '..'.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>mode</b></td>
        <td>integer</td>
        <td>
          mode is Optional: mode bits used to set permissions on this file.
Must be an octal value between 0000 and 0777 or a decimal value between 0 and 511.
YAML accepts both octal and decimal values, JSON requires decimal values for mode bits.
If not specified, the volume defaultMode will be used.
This might be in conflict with other options that affect the file
mode, like fsGroup, and the result can be other mode bits set.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].storageos
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



storageOS represents a StorageOS volume attached and mounted on Kubernetes nodes.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fsType is the filesystem type to mount.
Must be a filesystem type supported by the host operating system.
Ex. "ext4", "xfs", "ntfs". Implicitly inferred to be "ext4" if unspecified.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          readOnly defaults to false (read/write). ReadOnly here will force
the ReadOnly setting in VolumeMounts.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskspecvolumesindexstorageossecretref">secretRef</a></b></td>
        <td>object</td>
        <td>
          secretRef specifies the secret to use for obtaining the StorageOS API
credentials.  If not specified, default values will be attempted.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>volumeName</b></td>
        <td>string</td>
        <td>
          volumeName is the human-readable name of the StorageOS volume.  Volume
names are only unique within a namespace.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>volumeNamespace</b></td>
        <td>string</td>
        <td>
          volumeNamespace specifies the scope of the volume within StorageOS.  If no
namespace is specified then the Pod's namespace will be used.  This allows the
Kubernetes name scoping to be mirrored within StorageOS for tighter integration.
Set VolumeName to any name to override the default behaviour.
Set to "default" if you are not using namespaces within StorageOS.
Namespaces that do not pre-exist within StorageOS will be created.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].storageos.secretRef
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindexstorageos)</sup></sup>



secretRef specifies the secret to use for obtaining the StorageOS API
credentials.  If not specified, default values will be attempted.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.spec.volumes[index].vsphereVolume
<sup><sup>[↩ Parent](#redroidtaskspecvolumesindex)</sup></sup>



vsphereVolume represents a vSphere volume attached and mounted on kubelets host machine

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>volumePath</b></td>
        <td>string</td>
        <td>
          volumePath is the path that identifies vSphere volume vmdk<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          fsType is filesystem type to mount.
Must be a filesystem type supported by the host operating system.
Ex. "ext4", "xfs", "ntfs". Implicitly inferred to be "ext4" if unspecified.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>storagePolicyID</b></td>
        <td>string</td>
        <td>
          storagePolicyID is the storage Policy Based Management (SPBM) profile ID associated with the StoragePolicyName.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>storagePolicyName</b></td>
        <td>string</td>
        <td>
          storagePolicyName is the storage Policy Based Management (SPBM) profile name.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.status
<sup><sup>[↩ Parent](#redroidtask)</sup></sup>



RedroidTaskStatus defines the observed state of RedroidTask.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>activeJobs</b></td>
        <td>[]string</td>
        <td>
          ActiveJobs lists the names of currently running Jobs.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#redroidtaskstatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          Conditions holds detailed status conditions (Active, Complete, Failed).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>lastScheduleTime</b></td>
        <td>string</td>
        <td>
          LastScheduleTime is the last time the CronJob was scheduled.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>lastSuccessfulTime</b></td>
        <td>string</td>
        <td>
          LastSuccessfulTime is the last time a Job completed successfully.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          ObservedGeneration is the most recent generation observed by the controller.<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### RedroidTask.status.conditions[index]
<sup><sup>[↩ Parent](#redroidtaskstatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>
