# Group: `wizardoz.io/v1alpha1`
# Kind: `ExecAccessTemplate`

[kubernetes_group]: https://kubernetes.io/docs/reference/access-authn-authz/rbac/#role-binding-examples

**Phase: One**

An `ExecAccessTemplate` resource is used to pre-define allowed access rules into a particular workload. Because `ExecAccessRequests` do not create new `Pods` or modify them in any way, this template mostly serves as an authorization configuruation to ensure the right teams have access to this sensitive operation.

```yaml
apiVersion: wizardoz.io/v1alpha
kind: ExecAccessTemplate
metadata:
  name: <some predictable name>
  namespace: <target namespace>
spec:
  # Identifies the target workload that is having access granted.
  targetRef:
    apiVersion: apps/v1
    kind: DaemonSet
    name: targetApp

  # A list of Kubernetes Groups that are allowed to request access through this template.
  allowedGroups:
    - admins
    - devs
```
