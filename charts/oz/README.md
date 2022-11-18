# oz

![Version: 0.0.2](https://img.shields.io/badge/Version-0.0.2-informational?style=flat-square) ![AppVersion: 0.0.0-rc1](https://img.shields.io/badge/AppVersion-0.0.0--rc1-informational?style=flat-square)

Installation for the Oz RBAC Controller

**Homepage:** <https://github.com/diranged/oz>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| diranged |  | <https://github.com/diranged> |

## Source Code

* <https://github.com/diranged/oz>

## Requirements

Kubernetes: `>=1.22.0-0`

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| controllerManager.kubeRbacProxy.image.repository | string | `"gcr.io/kubebuilder/kube-rbac-proxy"` |  |
| controllerManager.kubeRbacProxy.image.tag | string | `"v0.13.0"` |  |
| controllerManager.kubeRbacProxy.resources.limits.cpu | string | `"500m"` |  |
| controllerManager.kubeRbacProxy.resources.limits.memory | string | `"128Mi"` |  |
| controllerManager.kubeRbacProxy.resources.requests.cpu | string | `"5m"` |  |
| controllerManager.kubeRbacProxy.resources.requests.memory | string | `"64Mi"` |  |
| controllerManager.manager.image.repository | `string` | `"ghcr.io/diranged/oz"` | Docker Image repository and name to use for the controller. |
| controllerManager.manager.image.tag | `string` | `nil` | If set, overrides the .Chart.AppVersion field to set the target image version for the Oz controller. |
| controllerManager.manager.resources.limits.cpu | string | `"500m"` |  |
| controllerManager.manager.resources.limits.memory | string | `"128Mi"` |  |
| controllerManager.manager.resources.requests.cpu | string | `"10m"` |  |
| controllerManager.manager.resources.requests.memory | string | `"64Mi"` |  |
| controllerManager.replicas | `int` | `1` | Number of Oz Controllers to run. If more than one is used, leader-election is used to ensure only one controller is operating at a time. |
| kubernetesClusterDomain | `string` | `"cluster.local"` | Configures the KUBERNETES_CLUSTER_DOMAIN environment variable. |
| metricsService.ports[0].name | string | `"https"` |  |
| metricsService.ports[0].port | int | `8443` |  |
| metricsService.ports[0].protocol | string | `"TCP"` |  |
| metricsService.ports[0].targetPort | string | `"https"` |  |
| metricsService.type | string | `"ClusterIP"` |  |

