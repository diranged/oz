# Oz RBAC Controller - Development Guide

[kind]: https://sigs.k8s/kind

## Prerequisites

You’ll need a Kubernetes cluster to run against. You can use [KIND][kind] to
get a local cluster for testing, or run against a remote cluster.

**Note:** Your controller will automatically use the current context in your
kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### IDE: Recommend VSCode

The recommended IDE is [Visual Studio](https://code.visualstudio.com/) - though
any IDE will work, we have set up the [`./vscode`](./vscode). If you are using
Visual Studio, the [`./vscode/extensions.json`](./vscode/extensions.json) file
should provide the most common extensions that will make development easier.

## Build Environment

### Spin up your Kind Cluster

First, spin up an empty [KIND][kind] cluster in your development environment.
We recommend always creating a new KIND environment for every project you work
on.

```sh
$ kind create cluster
```

### Running on the cluster

2. Build the docker image, load it into your KIND environment, and
   install/upgrade the controller:

```sh
$ make docker-build docker-load manifests deploy
...
service/oz-controller-manager-metrics-service created
deployment.apps/oz-controller-manager created
kubectl -n oz-system rollout restart deployment -l app.kubernetes.io/component=manager
deployment.apps/oz-controller-manager restarted
```

3. Install some test resources:

The [`examples`](./examples) directory includes some test resources - a
`Deployment`, `AccessTemplate`, `AccessRequest`, `ExecAccessTemplate` and
`AccessTemplate`. These resources can be used to quickly test the controller
locally.

First, spin up the target workload - a [`Deployment`](./examples/deployment.yaml):

```sh
$ kubectl apply -f examples/deployment.yaml
deployment.apps/example created
$ kubectl apply -f k apply -f examples/access_template.yaml
accesstemplate.crds.wizardofoz.co/deployment-example created
```

Once they are installed, verify that the `AccessTemplate` is in a good healthy state:
```sh
$ kubectl describe accesstemplate deployment-example | tail -15
  Conditions:
    Last Transition Time:  2022-11-19T22:23:15Z
    Message:               Success
    Observed Generation:   1
    Reason:                Success
    Status:                True
    Type:                  TargetRefExists
    Last Transition Time:  2022-11-19T22:23:15Z
    Message:               spec.defaultDuration and spec.maxDuration valid
    Observed Generation:   1
    Reason:                Success
    Status:                True
    Type:                  AccessDurationsValid
  Ready:                   true
Events:                    <none>
```


### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller to the cluster:

```sh
make undeploy
```

## Integration Tests (E2E / End to End)

### Create a dedicated `kind` cluster

```sh
$ export KIND_CLUSTER_NAME=e2e
$ kind create cluster
```

### Run Tests with Make

```sh
$ export KIND_CLUSTER_NAME=e2e
$ make test-e2e
```



## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/)
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster

### Test It Out
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)
