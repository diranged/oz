[exec_access_request]: /API.md#execaccessrequest
[exec_access_template]: /API.md#execaccesstemplate
[pod_access_request]: /API.md#podaccessrequest
[pod_access_template]: /API.md#podaccesstemplate
[access_config]: /diranged/oz/blob/pod_watcher/API.md#accessconfig
[target_ref]: /diranged/oz/blob/pod_watcher/API.md#crossversionobjectreference
[runtime]: https://github.com/kubernetes-sigs/controller-runtime
[builders]: ./builders/README.md

# Controllers

The Controllers in this package leverage the [controller-runtime][runtime]
package to define controllers that handle our custom resources
([PodAccessRequest][pod_access_request],
[PodAccessTemplate][pod_access_template],
[ExecAccessRequest][exec_access_request],
[ExecAccessTemplate][exec_access_request]). There are also controllers in this
package that handle inbound webhooks via the [Admission
Controllers](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/)
system.

## Reconcilers

Our `Reconciler` controllers handle operating in a loop to ensure that our
Custom Resources are consistently in the desired state. These controllers all
implement a `reconcile()` function that is triggered by `Watch...` requests
against the Kubernetes API.

Generally speaking, we try to keep the `reconcile()` functions short and easy
to read/understand. The heavy lifting is actually done by our
[`Builder`][builders] structs.

## [`PodAccessTemplateReconciler`](pod_access_template_controller.go)

The [`PodAccessTemplateReconciler`](pod_access_template_controller.go) is a
very simple controller whose job is to make sure that the `PodAccessTemplate`
is valid and available for use. It primarily validates that the template has
valid [`AccessConfig`][access_config] settings, and a valid
[`TargetRef`][target_ref] pointing to a real Pod controller (Deployment, etc).

```mermaid
sequenceDiagram
  participant Kubernetes
  participant Oz
  participant PodAccessTemplateReconciler
  
  Note over Oz,Kubernetes: The Oz Controller begins to watch for resources
  Oz->>Kubernetes: Watch PodAccessTemplate{} Resources...
  Kubernetes->>Oz: New PodAccessTemplate{} Created
  
  loop Reconcile Loop...
    Note over Oz,PodAccessTemplateReconciler: Runtime calls Reconciler function
    Oz->>PodAccessTemplateReconciler: reconcile(...)
    
    Note over PodAccessTemplateReconciler: Verify Target Reference Exists
    PodAccessTemplateReconciler->>Kubernetes: Get Deployment{Name: foo}
    Kubernetes->>PodAccessTemplateReconciler: 
    
    Note over PodAccessTemplateReconciler: Verify Access Configurations Settings are Valid
    PodAccessTemplateReconciler-->PodAccessTemplateReconciler: api.VerifyMiscSettings()
    
    Note over PodAccessTemplateReconciler: Write ready state back into resource
    PodAccessTemplateReconciler->>Kubernetes: Update .Status.IsReady=True
  end
```

## PodAccess
 
```mermaid
sequenceDiagram
    participant Alice
    participant Ozctl
    participant Kubernetes
    participant Oz
    participant PodAccessRequest

    link PodAccessRequest: API @ [pod_access_request]
    
    Note over Alice,Ozctl: Alice requests access to a development Pod
    Alice->>Ozctl: ozctl create podaccessrequest
    
    Note over Ozctl,Kubernetes: CLI prepares a PodAccessRequest{} resource
    Ozctl->>Kubernetes: Create PodAccessRequest{}...

    Note over Kubernetes,Oz: Mutating Webhook called...
    Kubernetes->>Oz: /mutate-v1-pod...
    Oz-->Oz: Call Default(admission.Request)
    
    Note over Kubernetes,Oz: Mutated PodAccessRequest is returned
    Oz->>Kubernetes: User Info Context applied

    Note over Kubernetes,Oz: Validating Webhook called to record Alice's action
    Kubernetes->>Oz: /validate-v1-pod...
    
    Note over Kubernetes,Oz: Emit Log Event
    Oz-->Oz: Call ValidateCreate(...)
    Oz-->Oz: Call Log.Info("Alice ...")
    Oz->>Kubernetes: `Allowed=True`
    
    Note over Kubernetes,Ozctl: Cluster responds that the resource has been created
    Kubernetes->>Ozctl: PodAccessRequest{} created
    
    par
      loop Reconcile Loop...
      Note over Kubernetes,Oz: Initial trigger event from Kubernetes
        Kubernetes->>Oz: Reconcile(PodAccessRequest)

        Oz-->Oz: Verify Request Durations
        Oz-->Oz: Verify Access Still Valid
        Oz->>Kubernetes: Create Role, RoleBinding, Pod
        Kubernetes ->> Oz: Resources Created
        Oz-->Oz: Verify Pod is "Ready"
        Oz->>Kubernetes: Set Status.IsReady=True
      end
    and
      loop CLI Loop
        Ozctl->>Kubernetes: Is Status.IsReady?
        Kubernetes->>Ozctl: True
        Ozctl->>Alice: "You're ready... kubectl exec ..."
      end
    end
```