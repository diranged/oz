# The Great and Powerful Oz

[kube_crd]: https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/
[kube_rbac]: https://kubernetes.io/docs/reference/access-authn-authz/rbac/
[exec_access_request]: ./crds/v1alpha1_exec_access_request.md
[access_request]: ./crds/v1alpha1_access_request.md
[access_template]: ./crds/v1alpha1_access_template.md
[exec_access_request]: ./crds/v1alpha1_exec_access_request.md
[exec_access_template]: ./crds/v1alpha1_exec_access_template.md
[access_approval]: ./crds/v1alpha1_access_request_approval.md

"Oz" aims to provde developers with seamless, authorized access into their application containers while allowing the cluster operators to define the parameters by which access is granted. Oz acts as a Kubernetes Operator - meaning it is a controller that runs within the cluster that acts on [Custom Resource Definitions][kube_crd] to dynamically provision `Pods` along with corresponding `Roles` and `RoleBindings`.

### Problems "Oz" will solve

#### Interactive "Shell" access into Production-Like Pod

When a developer needs to run a manual command from within a Production or Staging environment - such as a database migration, or investigation into the a code bug that is occuring, etc - it is not a good practice for that shell command to execute within a live container that is actively taking traffic.

_Why is this not possible today with RBAC?_

The Kubernetes [RBAC][kube_rbac] system does not currently allow for wildcards (`*`) in the `resource` key. This means that cluster operators must provide `exec` access into either _all_ pods in a Namespace, or _none_ of them.

_How will Oz solve this?_

"Oz" will provide an automated method for launching a dedicated shell Pod based on the `PodSpec` from the desired service. This pod will have all of the same environment variables, volume mounts and other specifications - but "Oz" will override the `spec.container[].command`, `spec.container[].args`, and `metadata.labels` fields so that the core application does not start up, and the Pod is not discovered by any corresponding `Services`. Access for the developer will be granted through dynamically provisioned `Role` and `RoleBindings`, ensuring narrowly scoped access to just the Pod that has been created.

#### Approved "Exec" Access into Specific Pods

Certain workloads (`DaemonSets` and `StatefulSets` mostly) have different charactaristics where launching a copy of a Pod may not be enough to provide operational access to the applications. In some cases you do need to have native `kubectl exec ...` or `kubectl debug ...` access for specific pods. 

_Why is this not possible today with RBAC?_
For the same reason as above (the resource wildcard issue), you may only grant access to _all_ or _none_ of the Pods in a namespace via the `kubectl exec` and `kubectl debug` commands.

Authentication in Kubernetes is pretty great - the [Kubernetes RBAC system][kube_rbac] provides the ability to narrowly scope permissions to individual Users, Groups and ServiceAccounts for almost API call and resource in the cluster.

## Kubernetes-Native Design

The entire design of the "Oz" system will be kubernetes-native ... it will operate on a number of [CRDs][kube_crd] that provide both configuration, access request and approval models. Through these CRDs, a developer (or any other tool) can request access with native Kubernetes tooling.

### [`AccessTemplate`][access_template]

In order to configure "Oz" to understand how to create a particular "workload" pod, an [`AccessTemplate`][access_template] resource first must be created. This resource defines a configuration for a `Pod` that is either based on an existing controller (`Deployment`, `StatefulSet`, `DaemonSet`), or may just be its own `PodSpec` entirely on its own.

In addition to the Pod configuration, the `AccessTemplate` defines _who_ can request one of these pods, and _how_ they request it.

**Who - A list of Groups**

This tool is designed to be used by humans, so the only focus here is on `Users` and `Groups` within Kubernetes. `Groups` are identified by whatever authentication system a user is using - commonly an OIDC login system. A list of allowed groups is included in the `AccessTemplate`.

When an `AccessRequest` is created, a `ValidatingWebhookConfiguration` verifies with "Oz" whether or not the requesting user has approval to create that `AccessRequest` against that particular `AccessTemplate`.

**How - Future plumbing for Approvals**

In a future phase of the project, [`AccessApproval`](#accessapproval-phase-two) resources will be implemented. When these are in place, the `AccessTemplate` will optionally define a number of "approvals" required for the request to go through. This system will allow for elevated privileges in certain cases, given a manual approval by another human operator.

### [`AccessRequest`][access_request]

An `AccessRequest` resource is used by a user to request a dedicated workload pod to be created. The user identifies which [`AccessTemplate`](#accesstemplate) that they are targeting, which will define the configuration of their pod.

The `AccessRequest` CR is deliberately left as simple as possible to reduce user configuration errors, as well as to ensure that the cluster/application operators who define the [`AccessTemplates`](#accesstemplate) are the ones who define the shapes (memory, CPU, service account, custom environment variables, etc) of the pods.

### [`ExecAccessTemplate`][exec_access_template]

Similar to the [`AccessTemplate`](#accesstemplate), but greatly simplified. This template primarily defines permissions around who can request `exec` access into a running Pod. Because an `exec` session does not mutate an existing pod or container, there are relatively few options here.

_TODO: Decide if `debug` access should be its own `DebugAccessTemplate` and `DebugAccessRequest` resource_

### [`ExecAccessRequest`][exec_access_request]

An `ExecAccessRequest` resource is used by the users similar to the [`AccessRequest`](#accessrequest) resource - but for requesting permissions to `kubectl exec` into a running Pod and Container.

### [`AccessApproval`][access_approval] (_Phase Two_)

Stubbing this out as a future resource. In the future we will provide the ability to require "approvals" for an access request. When an `AccessRequest` is created for an `AccessTemplate` that requires approvals, the request will go into a `PendingApprovals` state. It will then require that _a different human_ creates the appropriate `AccessApproval` resource (or multiple, if required).

## Command Line Interface

A CLI tool will provide a more user-friendly experience for creating [`AccessRequest`](#accessrequest), [`ExecAccessRequest`](#execaccessrequest) and [`AccessApproval`](#accessapproval-phase-two) CRs. The tool is primarily here to provide a clean interface for developers, but it is not strictly required as everything is done through standard Kubernetes resources.

Eg:
```bash
$ oz access-request <existingAccessRequestTemplate> -m 128Mi -c 2
Generating `access-request-fdcdac1` CR...
Waiting for approval...
Waiting for Pod...
Pod Launched: oz-fdcac1
Opening shell...
---
$ 
```