package internal

// PodPhaseRunning is exposed here so that we can reconfigure the search during
// tests to look for Pending pods.
var PodPhaseRunning = "Running"
