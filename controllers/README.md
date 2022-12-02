[exec_access_request]: /API.md#execaccessrequest
[exec_access_template]: /API.md#execaccesstemplate
[pod_access_request]: /API.md#podaccessrequest
[pod_access_template]: /API.md#podaccesstemplate
[access_config]: /API.md#accessconfig
[target_ref]: /API.md#crossversionobjectreference
[builders]: ./builders/README.md
[runtime]: https://github.com/kubernetes-sigs/controller-runtime

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

## [`ExecAccessTemplateReconciler`](exec_access_template_controller.go)

The [`ExecAccessTemplateReconciler`](exec_access_template_controller.go) is a
very simple controller whose job is to make sure that the `ExecAccessTemplate`
is valid and available for use. It primarily validates that the template has
valid [`AccessConfig`][access_config] settings, and a valid
[`TargetRef`][target_ref] pointing to a real Pod controller (Deployment, etc).

```mermaid
sequenceDiagram
  participant Kubernetes
  participant Oz
  participant ExecAccessTemplateReconciler
  
  Note over Oz,Kubernetes: The Oz Controller begins to watch for resources
  Oz->>Kubernetes: Watch ExecAccessTemplate{} Resources...
  Kubernetes->>Oz: New ExecAccessTemplate{} Created
  
  loop Reconcile Loop...
    Note over Oz,ExecAccessTemplateReconciler: Runtime calls Reconciler function
    Oz->>ExecAccessTemplateReconciler: reconcile(...)
    
    Note over ExecAccessTemplateReconciler: Verify Target Reference Exists
    ExecAccessTemplateReconciler->>Kubernetes: Get Deployment{Name: foo}
    Kubernetes->>ExecAccessTemplateReconciler: 
    
    Note over ExecAccessTemplateReconciler: Verify Access Configurations Settings are Valid
    ExecAccessTemplateReconciler-->ExecAccessTemplateReconciler: api.VerifyMiscSettings()
    
    Note over ExecAccessTemplateReconciler: Write ready state back into resource
    ExecAccessTemplateReconciler->>Kubernetes: Update .Status.IsReady=True
  end
```

## [`ExecAccessRequestReconciler`](exec_access_request_controller.go)

The [`ExecAccessRequestReconciler`](exec_access_request_controller.go) handles
creating a `Role` and `RoleBinding` that grant an engineer `kubectl exec ...`
access into an already existing Pod for a particular target deploymnt.

The reconciler logic itself is fairly simple, and most of the heavy lifting is
actually handled by a [`ExecAccessBuilder`](builders/exec_access_builder.go).

```mermaid
sequenceDiagram
  participant Kubernetes
  participant Oz
  participant ExecAccessRequestReconciler
  participant ExecAccessBuilder
  participant ExecAccessTemplate

  Oz->>Kubernetes: Watch ExecAccessRequest{} Resources...
  Kubernetes->>Oz: New ExecAccessRequest{} Created

  loop Reconcile Loop...
    Note over Oz,ExecAccessRequestReconciler: Runtime calls Reconciler function
    Oz-->>ExecAccessRequestReconciler: reconcile(...)

    Note over ExecAccessRequestReconciler: Verify `ExecAccessTemplate` Exists
    ExecAccessRequestReconciler->>Kubernetes: Get ExecAccessTemplate{Name: foo}
    Kubernetes->>ExecAccessRequestReconciler: 

    Note over ExecAccessRequestReconciler: Verify AccessConfiguration Settings are Valid
    ExecAccessRequestReconciler-->>ExecAccessRequestReconciler: verifyDuration()
    ExecAccessRequestReconciler-->>ExecAccessRequestReconciler: isAccessExpired()

    Note over ExecAccessRequestReconciler,ExecAccessBuilder: Begin Building Access Resources
    ExecAccessRequestReconciler-->>ExecAccessBuilder: verifyAccessResourcesBuilt()

    ExecAccessBuilder->>Kubernetes: Get Deployment{Name: foo..}
    Kubernetes->>ExecAccessBuilder: 

    Note over ExecAccessBuilder: Create the Resources
    ExecAccessBuilder->>Kubernetes: Create Role{Name: foo...}
    ExecAccessBuilder->>Kubernetes: Create RoleBinding{Name: foo...}

    Note over ExecAccessRequestReconciler: Write ready state back into resource
    ExecAccessRequestReconciler->>Kubernetes: Update .Status.IsReady=True
  end
```

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

## [`PodAccessRequestReconciler`](pod_access_request_controller.go)

The [`PodAccessRequestReconciler`](pod_access_request_controller.go) handles
the creation of a dedicated workload `Pod` for an engineer on-demand based on
the configuration of a [`PodAccessTemplate`](#podaccesstemplatereconciler). The
reconciler logic itself is fairly simple, and most of the heavy lifting is
actually handled by a [`PodAccessBuilder`](builders/pod_access_builder.go).

```mermaid
sequenceDiagram
  participant Kubernetes
  participant Oz
  participant PodAccessRequestReconciler
  participant PodAccessBuilder
  participant PodAccessTemplate
  
  Oz->>Kubernetes: Watch PodAccessRequest{} Resources...
  Kubernetes->>Oz: New PodAccessRequest{} Created

  loop Reconcile Loop...
    Note over Oz,PodAccessRequestReconciler: Runtime calls Reconciler function
    Oz-->>PodAccessRequestReconciler: reconcile(...)
    
    Note over PodAccessRequestReconciler: Verify `PodAccessTemplate` Exists
    PodAccessRequestReconciler->>Kubernetes: Get PodAccessTemplate{Name: foo}
    Kubernetes->>PodAccessRequestReconciler: 
    
    Note over PodAccessRequestReconciler: Verify AccessConfiguration Settings are Valid
    PodAccessRequestReconciler-->>PodAccessRequestReconciler: verifyDuration()
    PodAccessRequestReconciler-->>PodAccessRequestReconciler: isAccessExpired()
    
    Note over PodAccessRequestReconciler,PodAccessBuilder: Begin Building Access Resources
    PodAccessRequestReconciler-->>PodAccessBuilder: verifyAccessResourcesBuilt()
    
    PodAccessBuilder->>Kubernetes: Get Deployment{Name: foo..}
    Kubernetes->>PodAccessBuilder: 
    PodAccessBuilder-->>PodAccessTemplate: GenerateMutatedPodSpec(Deployment{}...)

    Note over PodAccessBuilder: Create the Resources
    PodAccessBuilder->>Kubernetes: Create Pod{Name: foo...}
    PodAccessBuilder->>Kubernetes: Create Role{Name: foo...}
    PodAccessBuilder->>Kubernetes: Create RoleBinding{Name: foo...}
    

    Note over PodAccessBuilder: Verify Resources Ready
    PodAccessRequestReconciler-->>PodAccessBuilder: verifyAccessResourcesReady()
    
    PodAccessBuilder->>Kubernetes: Get Pod{}.Status.Ready
    Kubernetes->>PodAccessBuilder: Pod{}.Status.Ready=True
    PodAccessBuilder-->>PodAccessRequestReconciler: Pod Is Ready

    Note over PodAccessRequestReconciler: Write ready state back into resource
    PodAccessRequestReconciler->>Kubernetes: Update .Status.IsReady=True
  end
```