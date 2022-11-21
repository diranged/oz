# Group: `wizardoz.io/v1alpha1`
# Kind: `PodAccessRequest`

**Phase: One**

The `PodAccessRequest` CRD is the primary resouce that developers will interact with. An `PodAccessRequest` resource is used to provision a new "copy" of an existing workload `Pod` along with temporary `Role` and `RoleBinding` resources to provide access into the pod.

```yaml
apiVersion: wizardoz.io/v1alpha
kind: PodAccessRequest
metadata:
  name: <dynamically generated name>
  namespace: <target namespace>
spec:
  accessTemplateRef:
    apiVersion: wizardoz.io/v1alpha1
    kind: PodAccessTemplate
    name: targetTemplateName
```
