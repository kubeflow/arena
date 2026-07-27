# Installation

Arena v2 is a command-line interface for running machine learning training jobs on Kubernetes. It relies on the Kubeflow Training Operator, which provides the custom resource definitions (CRDs) and controllers for distributed training workloads:

- **Kubeflow Training Operator** — provides the `PyTorchJob`, `TFJob`, and `MPIJob` CRDs and their controllers.

Follow the steps below to install the operator, build the Arena binary, and verify the setup.

## Prerequisites

- A Kubernetes cluster (>= 1.21) with `kubectl` configured and connected.
- A kubeconfig file available at `~/.kube/config`, or the `KUBECONFIG` environment variable pointing to one.
- Go >= 1.26.5 (for building the Arena binary from source).

## 1. Install the Kubeflow Training Operator

The Kubeflow Training Operator installs the `PyTorchJob`, `TFJob`, and `MPIJob` CRDs together with the controller that reconciles them. Arena v2 is built against the `v1.9.3` release.

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

## 2. Using Mirror Registries (Restricted Networks)

The manifest in step 1 pulls container images from Docker Hub
(`kubeflow/training-operator`, `alpine`, etc.).
If your cluster cannot reach Docker Hub — common in mainland China or
air-gapped environments — replace the images with a mirror registry endpoint
after applying the manifests.

### Patching the Training Operator

The `training-operator` Deployment references three images. The table below
shows the default Docker Hub values alongside the Alibaba Cloud ACK mirror
endpoints used in the examples that follow.

| Image | Default (Docker Hub) | ACK mirror (example) | How it is configured |
|-------|----------------------|----------------------|----------------------|
| Operator main container | `kubeflow/training-operator:v1.9.3` | `registry-cn-shanghai.ack.aliyuncs.com/acs/training-operator:8f8a791-aliyun` | `spec.template.spec.containers[*].image` |
| PyTorch init container | `alpine:3.22` | `registry-cn-shanghai.ack.aliyuncs.com/acs/alpine:3.6` | `--pytorch-init-container-image` arg |
| MPI kubectl delivery | `mpioperator/kubectl-delivery` | `registry-cn-shanghai.ack.aliyuncs.com/acs/kubectl-delivery:50b5d5b-aliyun` | `--mpi-kubectl-delivery-image` arg |

Patch the main container image:

```shell
$ kubectl set image deploy/training-operator -n kubeflow \
  training-operator=registry-cn-shanghai.ack.aliyuncs.com/acs/training-operator:8f8a791-aliyun
```

Patch the init container and kubectl-delivery images (passed as CLI args).
Check the current args first, then apply a JSON patch:

```shell
# View current args to confirm flag positions
$ kubectl get deploy training-operator -n kubeflow \
  -o jsonpath='{.spec.template.spec.containers[0].args}'

# Patch the PyTorch init container image (alpine >= 3.6 is sufficient)
$ kubectl patch deploy training-operator -n kubeflow --type=json \
  -p='[{"op":"replace","path":"/spec/template/spec/containers/0/args/0","value":"--pytorch-init-container-image=registry-cn-shanghai.ack.aliyuncs.com/acs/alpine:3.6"}]'

# Patch the MPI kubectl-delivery image
$ kubectl patch deploy training-operator -n kubeflow --type=json \
  -p='[{"op":"replace","path":"/spec/template/spec/containers/0/args/1","value":"--mpi-kubectl-delivery-image=registry-cn-shanghai.ack.aliyuncs.com/acs/kubectl-delivery:50b5d5b-aliyun"}]'
```

> **Note:** The ACK mirror endpoints above are examples for clusters in
> mainland China. If you are using a different mirror registry (e.g. a private
> harbor or another cloud provider), replace the registry host and tag with
> your own mirrored image. The `alpine` image only needs to be >= 3.6.

> **Tip:** If the args indices differ between operator versions, use
> `kubectl edit deploy training-operator -n kubeflow` and replace the image
> strings directly.

After patching, verify that the operator pods restart successfully:

```shell
$ kubectl get pods -n kubeflow -l app.kubernetes.io/name=training-operator
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
✓ MPIJob: installed (expected: kubeflow.org/v1)
  versions: v1 (served, storage)
  compatible: ✓ (storage version v1 supported by arena)
```

If a CRD is missing, the output marks it with `✗` and the command exits with an error. Ensure the Training Operator is installed correctly (see step 1) and run `arena-v2 check` again.

Confirm the Arena version:

```shell
$ arena-v2 version
Arena v2
  Version:    0.15.4
  Git Commit: a1b2c3d
  Build Date: 2026-07-13T12:00:00Z
```

## Uninstall

Remove the operator from your cluster:

```shell
$ kubectl delete -k "github.com/kubeflow/trainer/manifests/overlays/standalone?ref=v1.9.3"
```

Remove the Arena binary:

```shell
$ rm -f "$(go env GOPATH)/bin/arena-v2"
```

If you created an `arena` symlink, remove that as well:

```shell
$ rm -f "$(go env GOPATH)/bin/arena"
```
