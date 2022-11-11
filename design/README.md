# The Great and Powerful Oz

[kube_crd]: https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/
[kube_rbac]: https://kubernetes.io/docs/reference/access-authn-authz/rbac/
[exec_access_request]: ./crds/v1alpha1_exec_access_request.md
[access_request]: ./crds/v1alpha1_access_request.md

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

_TODO: Fill this out more_

## Command Line Interface

A CLI tool will provide a more user-friendly experience for creating [`AccessRequest`][access_request] CRs.

_TODO: Fill this out more_

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


