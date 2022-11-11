# Kind: `AccessRequestApproval`, Group: `wizardoz.io/v1alpha1`

**Phase: Two**

An `AccessRequestApproval` resource will be used to provide conditional approvals for certain types of `AccessRequest` and `ExecAccessRequest` resources. An `AccessRequestApproval` resource informs "Oz" that a pending `AccessRequest` is approved and can move forward, or is denied.

```yaml
apiVersion: wizardoz.io/v1alpha
kind: AccessRequestApproval
metadata:
  name: <dynamically generated name>
  namespace: <target namespace>
spec:
  accessRequestRef:
    apiVersion: wizardoz.io/v1alpha1
    kind: ExecAccessRquest
    name: accessRequestName
  approval: <true/false>
```
