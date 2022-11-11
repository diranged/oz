# Group: `wizardoz.io/v1alpha1`
# Kind: `AccessRequest`

**Phase: One**

The `AccessRequest` CRD is the primary resouce that developers will interact with. An `AccessRequest` resource is used to provision a new "copy" of an existing workload `Pod` along with temporary `Role` and `RoleBinding` resources to provide access into the pod.

```yaml
apiVersion: wizardoz.io/v1alpha
kind: AccessRequest
metadata:
  name: <dynamically generated name>
  namespace: <target namespace>
spec:
  accessTemplateRef:
    apiVersion: wizardoz.io/v1alpha1
    kind: AccessTemplate
    name: targetTemplateName
```
