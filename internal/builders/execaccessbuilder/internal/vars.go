package podselection

import (
	corev1 "k8s.io/api/core/v1"
)

// PodPhaseRunning is exposed here so that we can reconfigure the search during
// tests to look for Pending pods.
var PodPhaseRunning = corev1.PodRunning
