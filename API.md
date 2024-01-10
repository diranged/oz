<p>Packages:</p>
<ul>
<li>
<a href="#crds.wizardofoz.co%2fv1alpha1">crds.wizardofoz.co/v1alpha1</a>
</li>
</ul>
<h2 id="crds.wizardofoz.co/v1alpha1">crds.wizardofoz.co/v1alpha1</h2>
<div>
<p>Package v1alpha1 contains API Schema definitions for the templates v1alpha1 API group</p>
</div>
Resource Types:
<ul></ul>
<h3 id="crds.wizardofoz.co/v1alpha1.AccessConfig">AccessConfig
</h3>
<p>
(<em>Appears on:</em><a href="#crds.wizardofoz.co/v1alpha1.ExecAccessTemplateSpec">ExecAccessTemplateSpec</a>, <a href="#crds.wizardofoz.co/v1alpha1.PodAccessTemplateSpec">PodAccessTemplateSpec</a>)
</p>
<div>
<p>AccessConfig provides a common interface for our Template structs (which implement
ITemplateResource) for defining which entities are being granted access to a resource, and for
how long they are granted that access.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>allowedGroups</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>AllowedGroups lists out the groups (in string name form) that will be allowed to Exec into
the target pod.</p>
</td>
</tr>
<tr>
<td>
<code>defaultDuration</code><br/>
<em>
string
</em>
</td>
<td>
<p>DefaultDuration sets the default time that an access request resource will live. Must
be set below MaxDuration.</p>
<p>Valid time units are &ldquo;ns&rdquo;, &ldquo;us&rdquo; (or &ldquo;µs&rdquo;), &ldquo;ms&rdquo;, &ldquo;s&rdquo;, &ldquo;m&rdquo;, &ldquo;h&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>maxDuration</code><br/>
<em>
string
</em>
</td>
<td>
<p>MaxDuration sets the maximum duration that an access request resource can request to
stick around.</p>
<p>Valid time units are &ldquo;ns&rdquo;, &ldquo;us&rdquo; (or &ldquo;µs&rdquo;), &ldquo;ms&rdquo;, &ldquo;s&rdquo;, &ldquo;m&rdquo;, &ldquo;h&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>accessCommand</code><br/>
<em>
string
</em>
</td>
<td>
<p>AccessCommand is used to describe to the user how they can make use of their temporary access.
The AccessCommand can reference data from a Pod ObjectMeta.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="crds.wizardofoz.co/v1alpha1.ControllerKind">ControllerKind
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#crds.wizardofoz.co/v1alpha1.CrossVersionObjectReference">CrossVersionObjectReference</a>)
</p>
<div>
<p>ControllerKind is a string that represents an Apps/V1 known controller kind that this codebase
supports. This is used to limit the inputs on the AccessTemplate and ExecAccessTemplate CRDs.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;DaemonSet&#34;</p></td>
<td><p>DaemonSetController maps to APIVersion: apps/v1, Kind: DaemonSet</p>
</td>
</tr><tr><td><p>&#34;Deployment&#34;</p></td>
<td><p>DeploymentController maps to APIVersion: apps/v1, Kind: Deployment</p>
</td>
</tr><tr><td><p>&#34;StatefulSet&#34;</p></td>
<td><p>StatefulSetController maps to APIVersion: apps/v1, Kind: StatfulSet</p>
</td>
</tr></tbody>
</table>
<h3 id="crds.wizardofoz.co/v1alpha1.CoreStatus">CoreStatus
</h3>
<p>
(<em>Appears on:</em><a href="#crds.wizardofoz.co/v1alpha1.ExecAccessRequestStatus">ExecAccessRequestStatus</a>, <a href="#crds.wizardofoz.co/v1alpha1.ExecAccessTemplateStatus">ExecAccessTemplateStatus</a>, <a href="#crds.wizardofoz.co/v1alpha1.PodAccessRequestStatus">PodAccessRequestStatus</a>, <a href="#crds.wizardofoz.co/v1alpha1.PodAccessTemplateStatus">PodAccessTemplateStatus</a>)
</p>
<div>
<p>CoreStatus provides a common set of .Status fields and functions. The goal is to
conform to the interfaces.OzResource interface commonly across all of our core CRDs.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>conditions</code><br/>
<em>
<a href="https://v1-18.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#condition-v1-meta">
[]Kubernetes meta/v1.Condition
</a>
</em>
</td>
<td>
<p>Current status of the Access Template</p>
</td>
</tr>
<tr>
<td>
<code>ready</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Simple boolean to let us know if the resource is ready for use or not</p>
</td>
</tr>
<tr>
<td>
<code>accessMessage</code><br/>
<em>
string
</em>
</td>
<td>
<p>AccessMessage is used to describe to the user how they can make use of their temporary access
request. Eg, for a PodAccessTemplate the value set here would be something like:</p>
<p>&ldquo;Access Granted, connect to your pod with: kubectl exec -ti -n namespace pod-xyz &ndash; /bin/bash&rdquo;</p>
</td>
</tr>
</tbody>
</table>
<h3 id="crds.wizardofoz.co/v1alpha1.CrossVersionObjectReference">CrossVersionObjectReference
</h3>
<p>
(<em>Appears on:</em><a href="#crds.wizardofoz.co/v1alpha1.ExecAccessTemplateSpec">ExecAccessTemplateSpec</a>, <a href="#crds.wizardofoz.co/v1alpha1.PodAccessTemplateSpec">PodAccessTemplateSpec</a>)
</p>
<div>
<p>CrossVersionObjectReference provides us a generic way to define a reference to an APIGroup, Kind
and Name of a particular resource. Primarily used for the AccessTemplate and ExecAccessTemplate,
but generic enough to be used in other resources down the road.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code><br/>
<em>
string
</em>
</td>
<td>
<p>Defines the &ldquo;APIVersion&rdquo; of the resource being referred to. Eg, &ldquo;apps/v1&rdquo;.</p>
<p>TODO: Figure out how to regex validate that it has a &ldquo;/&rdquo; in it</p>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
<em>
<a href="#crds.wizardofoz.co/v1alpha1.ControllerKind">
ControllerKind
</a>
</em>
</td>
<td>
<p>Defines the &ldquo;Kind&rdquo; of resource being referred to.</p>
</td>
</tr>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<p>Defines the &ldquo;metadata.Name&rdquo; of the target resource.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="crds.wizardofoz.co/v1alpha1.ExecAccessRequest">ExecAccessRequest
</h3>
<div>
<p>ExecAccessRequest is the Schema for the execaccessrequests API</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://v1-18.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#crds.wizardofoz.co/v1alpha1.ExecAccessRequestSpec">
ExecAccessRequestSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>templateName</code><br/>
<em>
string
</em>
</td>
<td>
<p>Defines the name of the <code>ExecAcessTemplate</code> that should be used to grant access to the target
resource.</p>
</td>
</tr>
<tr>
<td>
<code>targetPod</code><br/>
<em>
string
</em>
</td>
<td>
<p>TargetPod is used to explicitly define the target pod that the Exec privilges should be
granted to. If not supplied, then a random pod is chosen.</p>
</td>
</tr>
<tr>
<td>
<code>duration</code><br/>
<em>
string
</em>
</td>
<td>
<p>Duration sets the length of time from the <code>spec.creationTimestamp</code> that this object will live. After the
time has expired, the resouce will be automatically deleted on the next reconcilliation loop.</p>
<p>If omitted, the spec.defautlDuration from the ExecAccessTemplate is used.</p>
<p>Valid time units are &ldquo;ns&rdquo;, &ldquo;us&rdquo; (or &ldquo;µs&rdquo;), &ldquo;ms&rdquo;, &ldquo;s&rdquo;, &ldquo;m&rdquo;, &ldquo;h&rdquo;.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#crds.wizardofoz.co/v1alpha1.ExecAccessRequestStatus">
ExecAccessRequestStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="crds.wizardofoz.co/v1alpha1.ExecAccessRequestSpec">ExecAccessRequestSpec
</h3>
<p>
(<em>Appears on:</em><a href="#crds.wizardofoz.co/v1alpha1.ExecAccessRequest">ExecAccessRequest</a>)
</p>
<div>
<p>ExecAccessRequestSpec defines the desired state of ExecAccessRequest</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>templateName</code><br/>
<em>
string
</em>
</td>
<td>
<p>Defines the name of the <code>ExecAcessTemplate</code> that should be used to grant access to the target
resource.</p>
</td>
</tr>
<tr>
<td>
<code>targetPod</code><br/>
<em>
string
</em>
</td>
<td>
<p>TargetPod is used to explicitly define the target pod that the Exec privilges should be
granted to. If not supplied, then a random pod is chosen.</p>
</td>
</tr>
<tr>
<td>
<code>duration</code><br/>
<em>
string
</em>
</td>
<td>
<p>Duration sets the length of time from the <code>spec.creationTimestamp</code> that this object will live. After the
time has expired, the resouce will be automatically deleted on the next reconcilliation loop.</p>
<p>If omitted, the spec.defautlDuration from the ExecAccessTemplate is used.</p>
<p>Valid time units are &ldquo;ns&rdquo;, &ldquo;us&rdquo; (or &ldquo;µs&rdquo;), &ldquo;ms&rdquo;, &ldquo;s&rdquo;, &ldquo;m&rdquo;, &ldquo;h&rdquo;.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="crds.wizardofoz.co/v1alpha1.ExecAccessRequestStatus">ExecAccessRequestStatus
</h3>
<p>
(<em>Appears on:</em><a href="#crds.wizardofoz.co/v1alpha1.ExecAccessRequest">ExecAccessRequest</a>)
</p>
<div>
<p>ExecAccessRequestStatus defines the observed state of ExecAccessRequest</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>CoreStatus</code><br/>
<em>
<a href="#crds.wizardofoz.co/v1alpha1.CoreStatus">
CoreStatus
</a>
</em>
</td>
<td>
<p>
(Members of <code>CoreStatus</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>podName</code><br/>
<em>
string
</em>
</td>
<td>
<p>The Target Pod Name where access has been granted</p>
</td>
</tr>
</tbody>
</table>
<h3 id="crds.wizardofoz.co/v1alpha1.ExecAccessTemplate">ExecAccessTemplate
</h3>
<div>
<p>ExecAccessTemplate is the Schema for the execaccesstemplates API</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://v1-18.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#crds.wizardofoz.co/v1alpha1.ExecAccessTemplateSpec">
ExecAccessTemplateSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>accessConfig</code><br/>
<em>
<a href="#crds.wizardofoz.co/v1alpha1.AccessConfig">
AccessConfig
</a>
</em>
</td>
<td>
<p>AccessConfig provides a common struct for defining who has access to the resources this
template controls, how long they have access, etc.</p>
</td>
</tr>
<tr>
<td>
<code>controllerTargetRef</code><br/>
<em>
<a href="#crds.wizardofoz.co/v1alpha1.CrossVersionObjectReference">
CrossVersionObjectReference
</a>
</em>
</td>
<td>
<p>ControllerTargetRef provides a pattern for referencing objects from another API in a generic way.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#crds.wizardofoz.co/v1alpha1.ExecAccessTemplateStatus">
ExecAccessTemplateStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="crds.wizardofoz.co/v1alpha1.ExecAccessTemplateSpec">ExecAccessTemplateSpec
</h3>
<p>
(<em>Appears on:</em><a href="#crds.wizardofoz.co/v1alpha1.ExecAccessTemplate">ExecAccessTemplate</a>)
</p>
<div>
<p>ExecAccessTemplateSpec defines the desired state of ExecAccessTemplate</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>accessConfig</code><br/>
<em>
<a href="#crds.wizardofoz.co/v1alpha1.AccessConfig">
AccessConfig
</a>
</em>
</td>
<td>
<p>AccessConfig provides a common struct for defining who has access to the resources this
template controls, how long they have access, etc.</p>
</td>
</tr>
<tr>
<td>
<code>controllerTargetRef</code><br/>
<em>
<a href="#crds.wizardofoz.co/v1alpha1.CrossVersionObjectReference">
CrossVersionObjectReference
</a>
</em>
</td>
<td>
<p>ControllerTargetRef provides a pattern for referencing objects from another API in a generic way.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="crds.wizardofoz.co/v1alpha1.ExecAccessTemplateStatus">ExecAccessTemplateStatus
</h3>
<p>
(<em>Appears on:</em><a href="#crds.wizardofoz.co/v1alpha1.ExecAccessTemplate">ExecAccessTemplate</a>)
</p>
<div>
<p>ExecAccessTemplateStatus is the core set of status fields that we expect to be in each and every one of
our template (AccessTemplate, ExecAccessTemplate, etc) resources.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>CoreStatus</code><br/>
<em>
<a href="#crds.wizardofoz.co/v1alpha1.CoreStatus">
CoreStatus
</a>
</em>
</td>
<td>
<p>
(Members of <code>CoreStatus</code> are embedded into this type.)
</p>
</td>
</tr>
</tbody>
</table>
<h3 id="crds.wizardofoz.co/v1alpha1.IConditionType">IConditionType
</h3>
<div>
<p>IConditionType provides an interface for accepting any condition string that
has a String() function. This simplifies the
controllers/internal/status/update_status.go code to have a single
UpdateStatus() function.</p>
</div>
<h3 id="crds.wizardofoz.co/v1alpha1.ICoreResource">ICoreResource
</h3>
<div>
<p>The ICoreResource interface wraps a standard client.Object resource (metav1.Object + runtime.Object)
with a few additional requirements for common methods that we use throughout our reconciliation process.</p>
</div>
<h3 id="crds.wizardofoz.co/v1alpha1.ICoreStatus">ICoreStatus
</h3>
<div>
<p>ICoreStatus is used to define the core common status functions that all Status structs in this
API must adhere to. These common functions simplify the reconciler() functions so that they can
easily get/set status on the resources in a common way.</p>
</div>
<h3 id="crds.wizardofoz.co/v1alpha1.IPodRequestResource">IPodRequestResource
</h3>
<div>
<p>IPodRequestResource is a Pod-access specific request interface that exposes a few more functions
for storing references to specific Pods that the requestor is being granted access to.</p>
</div>
<h3 id="crds.wizardofoz.co/v1alpha1.IRequestResource">IRequestResource
</h3>
<div>
<p>IRequestResource represents a common &ldquo;AccesRequest&rdquo; resource for the Oz Controller. These requests
have a common set of required methods that are used by the OzRequestReconciler.</p>
</div>
<h3 id="crds.wizardofoz.co/v1alpha1.IRequestStatus">IRequestStatus
</h3>
<div>
<p>IRequestStatus is a more specific Status interface that enables getting and
setting access instruction methods.</p>
</div>
<h3 id="crds.wizardofoz.co/v1alpha1.ITemplateResource">ITemplateResource
</h3>
<div>
<p>ITemplateResource represents a common &ldquo;AccessTemplate&rdquo; resource for the Oz Controller. These
templates provide different types of access into resources (eg, &ldquo;Exec&rdquo; vs &ldquo;Debug&rdquo; vs &ldquo;launch me a
dedicated pod&rdquo;). A set of common methods are required though that are used by the
OzTemplateReconciler.</p>
</div>
<h3 id="crds.wizardofoz.co/v1alpha1.ITemplateStatus">ITemplateStatus
</h3>
<div>
<p>ITemplateStatus provides a more specific Status interface for Access
Templates. Functionality to come in the future.</p>
</div>
<h3 id="crds.wizardofoz.co/v1alpha1.PodAccessRequest">PodAccessRequest
</h3>
<div>
<p>PodAccessRequest is the Schema for the accessrequests API</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://v1-18.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#crds.wizardofoz.co/v1alpha1.PodAccessRequestSpec">
PodAccessRequestSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>templateName</code><br/>
<em>
string
</em>
</td>
<td>
<p>Defines the name of the <code>ExecAcessTemplate</code> that should be used to grant access to the target
resource.</p>
</td>
</tr>
<tr>
<td>
<code>duration</code><br/>
<em>
string
</em>
</td>
<td>
<p>Duration sets the length of time from the <code>spec.creationTimestamp</code> that this object will live. After the
time has expired, the resouce will be automatically deleted on the next reconcilliation loop.</p>
<p>If omitted, the spec.defautlDuration from the ExecAccessTemplate is used.</p>
<p>Valid time units are &ldquo;s&rdquo;, &ldquo;m&rdquo;, &ldquo;h&rdquo;.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#crds.wizardofoz.co/v1alpha1.PodAccessRequestStatus">
PodAccessRequestStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="crds.wizardofoz.co/v1alpha1.PodAccessRequestSpec">PodAccessRequestSpec
</h3>
<p>
(<em>Appears on:</em><a href="#crds.wizardofoz.co/v1alpha1.PodAccessRequest">PodAccessRequest</a>)
</p>
<div>
<p>PodAccessRequestSpec defines the desired state of AccessRequest</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>templateName</code><br/>
<em>
string
</em>
</td>
<td>
<p>Defines the name of the <code>ExecAcessTemplate</code> that should be used to grant access to the target
resource.</p>
</td>
</tr>
<tr>
<td>
<code>duration</code><br/>
<em>
string
</em>
</td>
<td>
<p>Duration sets the length of time from the <code>spec.creationTimestamp</code> that this object will live. After the
time has expired, the resouce will be automatically deleted on the next reconcilliation loop.</p>
<p>If omitted, the spec.defautlDuration from the ExecAccessTemplate is used.</p>
<p>Valid time units are &ldquo;s&rdquo;, &ldquo;m&rdquo;, &ldquo;h&rdquo;.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="crds.wizardofoz.co/v1alpha1.PodAccessRequestStatus">PodAccessRequestStatus
</h3>
<p>
(<em>Appears on:</em><a href="#crds.wizardofoz.co/v1alpha1.PodAccessRequest">PodAccessRequest</a>)
</p>
<div>
<p>PodAccessRequestStatus defines the observed state of AccessRequest</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>CoreStatus</code><br/>
<em>
<a href="#crds.wizardofoz.co/v1alpha1.CoreStatus">
CoreStatus
</a>
</em>
</td>
<td>
<p>
(Members of <code>CoreStatus</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>podName</code><br/>
<em>
string
</em>
</td>
<td>
<p>The Target Pod Name where access has been granted</p>
</td>
</tr>
</tbody>
</table>
<h3 id="crds.wizardofoz.co/v1alpha1.PodAccessTemplate">PodAccessTemplate
</h3>
<div>
<p>PodAccessTemplate is the Schema for the accesstemplates API</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://v1-18.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#crds.wizardofoz.co/v1alpha1.PodAccessTemplateSpec">
PodAccessTemplateSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>accessConfig</code><br/>
<em>
<a href="#crds.wizardofoz.co/v1alpha1.AccessConfig">
AccessConfig
</a>
</em>
</td>
<td>
<p>AccessConfig provides a common struct for defining who has access to the resources this
template controls, how long they have access, etc.</p>
</td>
</tr>
<tr>
<td>
<code>controllerTargetRef</code><br/>
<em>
<a href="#crds.wizardofoz.co/v1alpha1.CrossVersionObjectReference">
CrossVersionObjectReference
</a>
</em>
</td>
<td>
<p>ControllerTargetRef provides a pattern for referencing objects from another API in a generic way.</p>
</td>
</tr>
<tr>
<td>
<code>controllerTargetMutationConfig</code><br/>
<em>
<a href="#crds.wizardofoz.co/v1alpha1.PodTemplateSpecMutationConfig">
PodTemplateSpecMutationConfig
</a>
</em>
</td>
<td>
<p>ControllerTargetMutationConfig contains parameters that allow for customizing the copy of a
controller-sourced PodSpec. This setting is only valid if controllerTargetRef is set.</p>
</td>
</tr>
<tr>
<td>
<code>podSpec</code><br/>
<em>
<a href="https://v1-18.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#podspec-v1-core">
Kubernetes core/v1.PodSpec
</a>
</em>
</td>
<td>
<p>PodSpec &hellip;</p>
</td>
</tr>
<tr>
<td>
<code>maxStorage</code><br/>
<em>
k8s.io/apimachinery/pkg/api/resource.Quantity
</em>
</td>
<td>
<p>Upper bound of the ephemeral storage that an AccessRequest can make against this template for
the primary container.</p>
</td>
</tr>
<tr>
<td>
<code>maxCpu</code><br/>
<em>
k8s.io/apimachinery/pkg/api/resource.Quantity
</em>
</td>
<td>
<p>Upper bound of the CPU that an AccessRequest can make against this tmemplate for the primary container.</p>
</td>
</tr>
<tr>
<td>
<code>maxMemory</code><br/>
<em>
k8s.io/apimachinery/pkg/api/resource.Quantity
</em>
</td>
<td>
<p>Upper bound of the memory that an AccessRequest can make against this template for the primary container.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#crds.wizardofoz.co/v1alpha1.PodAccessTemplateStatus">
PodAccessTemplateStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="crds.wizardofoz.co/v1alpha1.PodAccessTemplateSpec">PodAccessTemplateSpec
</h3>
<p>
(<em>Appears on:</em><a href="#crds.wizardofoz.co/v1alpha1.PodAccessTemplate">PodAccessTemplate</a>)
</p>
<div>
<p>PodAccessTemplateSpec defines the desired state of AccessTemplate</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>accessConfig</code><br/>
<em>
<a href="#crds.wizardofoz.co/v1alpha1.AccessConfig">
AccessConfig
</a>
</em>
</td>
<td>
<p>AccessConfig provides a common struct for defining who has access to the resources this
template controls, how long they have access, etc.</p>
</td>
</tr>
<tr>
<td>
<code>controllerTargetRef</code><br/>
<em>
<a href="#crds.wizardofoz.co/v1alpha1.CrossVersionObjectReference">
CrossVersionObjectReference
</a>
</em>
</td>
<td>
<p>ControllerTargetRef provides a pattern for referencing objects from another API in a generic way.</p>
</td>
</tr>
<tr>
<td>
<code>controllerTargetMutationConfig</code><br/>
<em>
<a href="#crds.wizardofoz.co/v1alpha1.PodTemplateSpecMutationConfig">
PodTemplateSpecMutationConfig
</a>
</em>
</td>
<td>
<p>ControllerTargetMutationConfig contains parameters that allow for customizing the copy of a
controller-sourced PodSpec. This setting is only valid if controllerTargetRef is set.</p>
</td>
</tr>
<tr>
<td>
<code>podSpec</code><br/>
<em>
<a href="https://v1-18.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#podspec-v1-core">
Kubernetes core/v1.PodSpec
</a>
</em>
</td>
<td>
<p>PodSpec &hellip;</p>
</td>
</tr>
<tr>
<td>
<code>maxStorage</code><br/>
<em>
k8s.io/apimachinery/pkg/api/resource.Quantity
</em>
</td>
<td>
<p>Upper bound of the ephemeral storage that an AccessRequest can make against this template for
the primary container.</p>
</td>
</tr>
<tr>
<td>
<code>maxCpu</code><br/>
<em>
k8s.io/apimachinery/pkg/api/resource.Quantity
</em>
</td>
<td>
<p>Upper bound of the CPU that an AccessRequest can make against this tmemplate for the primary container.</p>
</td>
</tr>
<tr>
<td>
<code>maxMemory</code><br/>
<em>
k8s.io/apimachinery/pkg/api/resource.Quantity
</em>
</td>
<td>
<p>Upper bound of the memory that an AccessRequest can make against this template for the primary container.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="crds.wizardofoz.co/v1alpha1.PodAccessTemplateStatus">PodAccessTemplateStatus
</h3>
<p>
(<em>Appears on:</em><a href="#crds.wizardofoz.co/v1alpha1.PodAccessTemplate">PodAccessTemplate</a>)
</p>
<div>
<p>PodAccessTemplateStatus defines the observed state of PodAccessTemplate</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>CoreStatus</code><br/>
<em>
<a href="#crds.wizardofoz.co/v1alpha1.CoreStatus">
CoreStatus
</a>
</em>
</td>
<td>
<p>
(Members of <code>CoreStatus</code> are embedded into this type.)
</p>
</td>
</tr>
</tbody>
</table>
<h3 id="crds.wizardofoz.co/v1alpha1.PodTemplateSpecMutationConfig">PodTemplateSpecMutationConfig
</h3>
<p>
(<em>Appears on:</em><a href="#crds.wizardofoz.co/v1alpha1.PodAccessTemplateSpec">PodAccessTemplateSpec</a>)
</p>
<div>
<p>PodTemplateSpecMutationConfig provides a common pattern for describing mutations to an existing PodSpec
that should be applied. The primary use case is in the PodAccessTemplate, where an existing
controller (Deployment, DaemonSet, StatefulSet) can be used as the reference for the PodSpec
that is launched for the user. However, the operator may want to make modifications to the
PodSpec at launch time (eg, change the entrypoint command or arguments).</p>
<p>TODO: Add affinity</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>defaultContainerName</code><br/>
<em>
string
</em>
</td>
<td>
<p>DefaultContainerName allows the operator to define which container is considered the default
container, and that is the container that this mutation configuration applies to. If not set,
then the first container defined in the spec.containers[] list is patched.</p>
</td>
</tr>
<tr>
<td>
<code>command</code><br/>
<em>
string
</em>
</td>
<td>
<p>Command is used to override the .Spec.containers[0].command field for the target Pod and
Container. This can be handy in ensuring that the default application does not start up and
do any work. If set, this overrides the Spec.conatiners[0].args property as well.</p>
</td>
</tr>
<tr>
<td>
<code>args</code><br/>
<em>
string
</em>
</td>
<td>
<p>Args will override the Spec.containers[0].args property.</p>
</td>
</tr>
<tr>
<td>
<code>env</code><br/>
<em>
<a href="https://v1-18.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#envvar-v1-core">
[]Kubernetes core/v1.EnvVar
</a>
</em>
</td>
<td>
<p>Env allows overriding specific environment variables (or adding new ones). Note, we do not
purge the original environmnt variables.</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br/>
<em>
<a href="https://v1-18.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<p>If supplied these resource requirements will override the default .Spec.containers[0].resource requested for the
the pod. Note though that we do not override all of the resource requests in the Pod because there may be many
containers.</p>
</td>
</tr>
<tr>
<td>
<code>podAnnotations</code><br/>
<em>
string
</em>
</td>
<td>
<p>If supplied, these
<a href="https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/">annotations</a>
are applied to the target
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#podtemplatespec-v1-core"><code>PodTemplateSpec</code></a>.
These are merged into the final Annotations. If you want to <em>replace</em>
the annotations, make sure to set the <code>purgeAnnotations</code> flag to <code>true</code>.</p>
</td>
</tr>
<tr>
<td>
<code>podLabels</code><br/>
<em>
string
</em>
</td>
<td>
<p>If supplied, Oz will insert these
<a href="https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/">labels</a>
into the target
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#podtemplatespec-v1-core"><code>PodTemplateSpec</code></a>.
By default Oz purges all Labels from pods (to prevent the new Pod from
having traffic routed to it), so this is effectively a new set of labels
applied to the Pod.</p>
</td>
</tr>
<tr>
<td>
<code>purgeAnnotations</code><br/>
<em>
bool
</em>
</td>
<td>
<p>By default, Oz keeps the original
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#podtemplatespec-v1-core"><code>PodTemplateSpec</code></a>
<code>metadata.annotations</code> field. If you want to purge this, set this flag
to <code>true.</code></p>
</td>
</tr>
<tr>
<td>
<code>patchSpecOperations</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>PatchSpecOperations contains a list of JSON patch operations to apply to the PodSpec.
<a href="https://www.rfc-editor.org/rfc/rfc6902.html"><code>JSONPatch</code></a></p>
</td>
</tr>
<tr>
<td>
<code>keepTerminationGracePeriod</code><br/>
<em>
bool
</em>
</td>
<td>
<p>By default, Oz wipes out the PodSpec
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#podspec-v1-core"><code>terminationGracePeriodSeconds</code></a>
setting on Pods to ensure that they can be killed as soon as the
AccessRequest expires. This flag overrides that behavior.</p>
</td>
</tr>
<tr>
<td>
<code>keepLivenessProbe</code><br/>
<em>
bool
</em>
</td>
<td>
<p>By default, Oz wipes out the PodSpec
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#podspec-v1-core"><code>livenessProbe</code></a>
configuration for the default container so that the container does not
get terminated if the main application is not running or passing checks.
This setting overrides that behavior.</p>
</td>
</tr>
<tr>
<td>
<code>keepReadinessProbe</code><br/>
<em>
bool
</em>
</td>
<td>
<p>By default, Oz wipes out the PodSpec
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#podspec-v1-core"><code>readinessProbe</code></a>
configuration for the default container so that the container does not
get terminated if the main application is not running or passing checks.
This setting overrides that behavior.</p>
</td>
</tr>
<tr>
<td>
<code>keepStartupProbe</code><br/>
<em>
bool
</em>
</td>
<td>
<p>By default, Oz wipes out the PodSpec
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#podspec-v1-core"><code>startupProbe</code></a>
configuration for the default container so that the container does not
get terminated if the main application is not running or passing checks.
This setting overrides that behavior.</p>
</td>
</tr>
<tr>
<td>
<code>keepTopologySpreadConstraints</code><br/>
<em>
bool
</em>
</td>
<td>
<p>By default, Oz wipes out the PodSpec
<a href="https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#scheduling"><code>topologySpreadConstraints</code></a>
configuration for the Pod because these access pods are not part of the
same group of pods that are passing traffic. This setting overrides that behavior.</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code><br/>
<em>
string
</em>
</td>
<td>
<p>If supplied, Oz will insert these
<a href="https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#scheduling">nodeSelector</a>
into the target
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#podtemplatespec-v1-core"><code>PodTemplateSpec</code></a>.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="crds.wizardofoz.co/v1alpha1.RequestConditionTypes">RequestConditionTypes
(<code>string</code> alias)</h3>
<div>
<p>RequestConditionTypes defines a set of known Status.Condition[].ConditionType fields that are
used throughout the AccessRequest and AccessTemplate reconcilers.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;AccessMessage&#34;</p></td>
<td><p>ConditionAccessMessage is used to record</p>
</td>
</tr><tr><td><p>&#34;AccessResourcesCreated&#34;</p></td>
<td><p>ConditionAccessResourcesCreated indicates whether or not the target
access request resources have been properly created.</p>
</td>
</tr><tr><td><p>&#34;AccessResourcesReady&#34;</p></td>
<td><p>ConditionAccessResourcesReady indicates that all of the &ldquo;access
resources&rdquo; (eg, a Pod) are up and in the ready state.</p>
</td>
</tr><tr><td><p>&#34;AccessStillValid&#34;</p></td>
<td><p>ConditionAccessStillValid is continaully updated based on whether or not
the Access Request has timed out.</p>
</td>
</tr><tr><td><p>&#34;AccessDurationsValid&#34;</p></td>
<td><p>ConditionRequestDurationsValid is used by both AccessTemplate and
AccessRequest resources. It indicates whether or not the various
duration fields are valid.</p>
</td>
</tr><tr><td><p>&#34;TargetTemplateExists&#34;</p></td>
<td><p>ConditionTargetTemplateExists indicates that the Access Request is
pointing to a valid Access Template.</p>
</td>
</tr></tbody>
</table>
<h3 id="crds.wizardofoz.co/v1alpha1.TemplateConditionTypes">TemplateConditionTypes
(<code>string</code> alias)</h3>
<div>
<p>TemplateConditionTypes defines a set of known Status.Condition[].ConditionType fields that are
used throughout the AccessTemplate reconcilers and written to the ITemplateResource resources.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;TargetRefExists&#34;</p></td>
<td><p>ConditionTargetRefExists indicates whether or not an AccessTemplate is
pointing to a valid Controller.</p>
</td>
</tr><tr><td><p>&#34;TemplateDurationsValid&#34;</p></td>
<td><p>ConditionTemplateDurationsValid is used by both AccessTemplate and
AccessRequest resources. It indicates whether or not the various
duration fields are valid.</p>
</td>
</tr></tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
.
</em></p>
