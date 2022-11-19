[design]: ./design/README.md
[exec_access_request]: ./design/crds/v1alpha1_exec_access_request.md
[exec_access_template]: ./design/crds/v1alpha1_exec_access_template.md
[kube_crd]: https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/
[kube_rbac]: https://kubernetes.io/docs/reference/access-authn-authz/rbac/

# Oz RBAC Controller

[![ci](https://github.com/diranged/oz/actions/workflows/main.yaml/badge.svg?branch=main)](https://github.com/diranged/oz/actions/workflows/main.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/diranged/oz)](https://goreportcard.com/report/github.com/diranged/oz)

_The Wizard of Oz_: The "Great and Powerfull Oz", or also known as the "man
behind the curtain."

**"Oz RBAC Controller"** is a Kubernetes operator that provides short-term
customized RBAC privileges to end-users through the native Kubernetes
[RBAC][kube_rbac] system. It aims to be the "man behind the curtain" -
carefully creating `Roles`, `RoleBindings` and `Pods` on-demand that enable
developers to quickly get their jobs done, and administartors to ensure that
the principal of least privilege is honored.

**Oz** primarly works with two resource constructs - **Access Requests** and
**Access Templates**.

**Access Templates** are defined by the cluster operators or application owners
to create a "template" for how a particular type of short-term access can be
granted. For example, the [`ExecAccessTemplate`][exec_access_template] defines
a particular target (`DaemonSet` for example) that a user can request access
to, the `Groups` that are allowed to have that access, and rules around the
maximum duration that the request can remain active.

**Access Requests** are created by end-users when they need access to a
resource. RBAC privileges must be granted to users to even create the resource -
but once that is done, actual access to the final target pod is controlled
natively through the Kubernetes RBAC system, ensuring that we are not
bypassing any standard internal RBAC controls.

An example is the [`ExecAccessRequest`][exec_access_request] resource which
points to a particular [`ExecAccessTemplate`][exec_access_template]. When the
request is created, the **Oz** controller will dynamically create a `Role` and
`RoleBinding` granting access into the desired target pod to the specific
groups defined in the template itself.

## Example Use Cases

### Use Case: Interactive "Shell" access into Production-Like Pod

When a developer needs to run a manual command from within a Production or
Staging environment - such as a database migration, or investigation into the a
code bug that is occuring, etc - it is not a good practice for that shell
command to execute within a live container that is actively taking traffic.

_Why is this not possible today with RBAC?_

The Kubernetes [RBAC][kube_rbac] system does not currently allow for wildcards
(`*`) in the `resource` key. This means that cluster operators must provide
`exec` access into either _all_ pods in a Namespace, or _none_ of them.

_How will Oz solve this?_

"Oz" will provide an automated method for launching a dedicated shell Pod based
on the `PodSpec` from the desired service. This pod will have all of the same
environment variables, volume mounts and other specifications - but "Oz" will
override the `spec.container[].command`, `spec.container[].args`, and
`metadata.labels` fields so that the core application does not start up, and
the Pod is not discovered by any corresponding `Services`. Access for the
developer will be granted through dynamically provisioned `Role` and
`RoleBindings`, ensuring narrowly scoped access to just the Pod that has been
created.

### Use Case: "Exec" Access into Specific Live Pods

Certain workloads (`DaemonSets` and `StatefulSets` mostly) have different
charactaristics where launching a copy of a Pod may not be enough to provide
operational access to the applications. In some cases you do need to have
native `kubectl exec ...` or `kubectl debug ...` access for specific pods.

_Why is this not possible today with RBAC?_

For the same reason as above (the resource wildcard issue), you may only grant
access to _all_ or _none_ of the Pods in a namespace via the `kubectl exec` and
`kubectl debug` commands.

Authentication in Kubernetes is pretty great - the [Kubernetes RBAC
system][kube_rbac] provides the ability to narrowly scope permissions to
individual Users, Groups and ServiceAccounts for almost API call and resource
in the cluster.

## Installation

### Helm-Installation of the Controller

[helm_chart]: https://github.com/diranged/oz-charts/tree/main/charts/oz
[releases]: https://github.com/diranged/oz/releases

The controller can be installed today through a [helm chart][helm_chart]. The
chart is hosted by Github and can be easily installed like this:

```bash
$ helm repo add oz-charts https://diranged.github.io/oz-charts
$ helm repo update
$ helm search repo oz-charts
NAME        	CHART VERSION	APP VERSION	DESCRIPTION
oz-charts/oz	0.0.6        	0.0.0-rc1  	Installation for the Oz RBAC Controller
```

### Installation of the CLI tool

An [`ozctl` CLI tool](./ozctl) is provided primarily for the end users of the
`AccessRequest` objects. This tool simplifies the process of quickly creating
an access request, waiting for it to be processed, and then reporting
instructions on how to make use of the resources.

The `ozctl` binaries are available through the ["releases"][releases] page and
are built for OSX in both Intel and Arm variants.

### Creation of an `ExecAccessTemplate`

_TODO_...

## License

Copyright 2022 Matt Wise.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

