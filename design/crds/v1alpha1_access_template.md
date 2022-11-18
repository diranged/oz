# Group: `wizardoz.io/v1alpha1`
# Kind: `AccessTemplate`

**Phase: One**

An `AccessTemplate` resource is used to pre-define allowed access rules into a particular workload. This resource can be launched either by the application owners as part of their release process, or could be pre-created by cluster administrators.

`AccessTemplate` resources define the "shape" of the access provided into a particular workload. For example, an `AccessTemplate` can be used to define the shell that should be entered by default, or define the maximum amount of time that a subsequent `AccessRequest` pod may live.

```yaml
apiVersion: wizardoz.io/v1alpha
kind: AccessTemplate
metadata:
  name: <some predictable name>
  namespace: <target namespace>
spec:
  # Identifies the target controller who's `PodSpec` should be used
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: targetApp

  # A list of Kubernetes Groups that are allowed to request access through this template.
  allowedGroups:
    - admins
    - devs

  # Overrides the default container's `command` - used to prevent the core application from starting
  # up (for example, if it is a background task processing application instead of a web service).
  command: [/bin/sleep, 9999]

  # Overrides the resource requests of the default container in the PodSpec.
  resources:
    limits:
      cpu: 4
      memory: 256Mi
    requests:
      cpu: 2
      memory: 128Mi

  # Provides a maximum ceiling for how much memory can be requested.
  maxMemory: 1Gi

  # Provides a maximum ceiling for how many CPU cores can be requested.
  maxCpu: 4

  # Provides a maximum ceiling for how much ephemeral storage can be requested
  maxStorage: 10Gi
```
