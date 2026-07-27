# Installation

Arena v2 is a command-line interface for running machine learning training jobs on Kubernetes. It relies on two Kubeflow operators that provide the custom resource definitions (CRDs) and controllers for distributed training workloads:

- **Kubeflow Training Operator** — provides the `PyTorchJob` and `TFJob` CRDs and their controller.
- **MPI Operator** — provides the `MPIJob` CRD and its controller.

Follow the steps below to install the operators, build the Arena binary, and verify the setup.

## Prerequisites

- A Kubernetes cluster (>= 1.21) with `kubectl` configured and connected.
- A kubeconfig file available at `~/.kube/config`, or the `KUBECONFIG` environment variable pointing to one.
- Go >= 1.25 (for building the Arena binary from source).

## 1. Install the Kubeflow Training Operator

The Kubeflow Training Operator installs the `PyTorchJob` and `TFJob` CRDs together with the controller that reconciles them. Arena v2 is built against the `v1.9.3` release.

Apply the standalone overlay with kustomize:

```shell
$ kubectl apply --server-side -k "github.com/kubeflow/trainer/manifests/overlays/standalone?ref=v1.9.3"
```

This creates the `kubeflow` namespace, the CRDs, the required RBAC, and the `training-operator` Deployment.

Verify that the controller is running:

```shell
$ kubectl get deploy -n kubeflow training-operator
NAME                READY   UP-TO-DATE   AVAILABLE   AGE
training-operator   1/1     1            1           2m
```

## 2. Install the MPI Operator

The MPI Operator installs the `MPIJob` CRD and the controller that reconciles MPI-based training jobs. Arena v2 is built against the `v0.8.0` release.

Apply the release manifest:

```shell
$ kubectl apply -f https://github.com/kubeflow/mpi-operator/releases/download/v0.8.0/mpi-operator.yaml
```

This creates the `mpi-operator` namespace, the `MPIJob` CRD, the required RBAC, and the `mpi-operator` Deployment.

Verify that the controller is running:

```shell
$ kubectl get deploy -n mpi-operator mpi-operator
NAME            READY   UP-TO-DATE   AVAILABLE   AGE
mpi-operator    1/1     1            1           1m
```

## 3. Install the Arena Binary

Arena v2 is built from source with Go.

Clone the repository and build the binary:

```shell
$ git clone https://github.com/kubeflow/arena.git
$ cd arena
$ make arena-v2
```

The `arena-v2` binary is placed in the `bin/` directory. Add it to your `PATH`:

```shell
$ export PATH="$(pwd)/bin:$PATH"
```

Alternatively, install it directly into your Go binary directory (`$(go env GOPATH)/bin`) with:

```shell
$ make v2-install
```

If you prefer the shorter command name `arena`, create a symlink:

```shell
$ ln -s "$(go env GOPATH)/bin/arena-v2" "$(go env GOPATH)/bin/arena"
```

## 4. Verify the Installation

Check that Arena can reach your cluster and that all required CRDs are installed:

```shell
$ arena-v2 check
✓ PyTorchJob: installed (expected: kubeflow.org/v1)
  versions: v1 (served, storage)
✓ TFJob: installed (expected: kubeflow.org/v1)
  versions: v1 (served, storage)
✓ MPIJob: installed (expected: kubeflow.org/v2beta1)
  versions: v2beta1 (served, storage), v1 (served)
  compatible: ✓ (storage version v2beta1 supported by arena)
```

If a CRD is missing, the output marks it with `✗` and the command exits with an error. Install the corresponding operator (see steps 1 and 2) and run `arena-v2 check` again.

Confirm the Arena version:

```shell
$ arena-v2 version
Arena v2
  Version:    0.15.4
  Git Commit: a1b2c3d
  Build Date: 2026-07-13T12:00:00Z
```

## Uninstall

Remove the operators from your cluster:

```shell
$ kubectl delete -k "github.com/kubeflow/trainer/manifests/overlays/standalone?ref=v1.9.3"
$ kubectl delete -f https://github.com/kubeflow/mpi-operator/releases/download/v0.8.0/mpi-operator.yaml
```

Remove the Arena binary:

```shell
$ rm -f "$(go env GOPATH)/bin/arena-v2"
```

If you created an `arena` symlink, remove that as well:

```shell
$ rm -f "$(go env GOPATH)/bin/arena"
```
