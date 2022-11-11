# Kind: `ExecAccessRequest`, Group: `wizardoz.io/v1alpha1`

**Phase: One**

The `ExecAccessRequest` CRD is used to dynamically request `exec` or `debug` access into an existing Pod. This differes from the `AccessRequest` CRD in that we do not launch a new pod but instead only manage the creation of a short-lived `Role` and `RoleBinding` for the duration of the access.

```yaml
apiVersion: wizardoz.io/v1alpha
kind: ExecAccessRequest
metadata:
  name: <dynamically generated name>
  namespace: <target namespace>
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: targetApp
```
