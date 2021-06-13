# Depremon (Deprecated k8s API Monitor)

Depremon is an open source utility to help users easily find deprecated Kubernetes API versions in a live Openshift cluster.

**Known Limitation:** Currently, it only works on Openshift since it is using Openshift service-ca-operator to sign certificate. We will evolute it to work on all the k8s platforms

## Background

As the Kubernetes API evolves, APIs are periodically reorganized or upgraded. When APIs evolve, the old API is deprecated and eventually removed.

There are some existing tools designed to check if project yaml files has deprecated Kubernetes API before the release, like [pluto](https://github.com/FairwindsOps/pluto) and kubepug(https://github.com/rikatz/kubepug), but they don't work well with [operator deployments](https://github.com/operator-framework) or other services basing on Kubernetes controllers.

## How Depremon works

Depremon is a Kubernetes Operator deploying a Kubernetes webhook to record resources with Kubernetes API which are going to be removed and save the result into a configmap.

Users can use Depremon custom resource to customize the configurations.
