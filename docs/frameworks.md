# Framework Guides

This guide covers how to use Arena v2 with each supported training framework. Each section references example YAML files in `examples/v2/` and explains the key configuration options.

For a complete reference of all YAML fields, see [yaml-schema.md](yaml-schema.md).

## PyTorch

Arena v2 supports standalone and distributed PyTorch training via the PyTorchJob CRD (`kubeflow.org/v1`). Arena generates the CRD from your task YAML, automatically wiring up the master/worker replica specs and the `nprocPerNode` field.

### Standalone Training

When you want to run PyTorch on a single node with no distributed coordination, define a `master` section and omit `worker`. Arena creates a single master replica and no worker pods — the job runs entirely within one pod.

Reference: [examples/v2/quickstart/pytorch-standalone.yaml](../examples/v2/quickstart/pytorch-standalone.yaml)

```yaml
framework:
  name: pytorch
  options:
    nproc_per_node: auto
master:
  resources:
    # nvidia.com/gpu: "1"  # Uncomment for GPU clusters
    cpu: "2"
    memory: "8Gi"
run: |
  python -c "
  import torch, torch.nn as nn
  model = nn.Linear(10, 1)
  optimizer = torch.optim.SGD(model.parameters(), lr=0.01)
  for epoch in range(3):
      x, y = torch.randn(64, 10), torch.randn(64, 1)
      loss = nn.MSELoss()(model(x), y)
      optimizer.zero_grad(); loss.backward(); optimizer.step()
      print(f'Epoch {epoch+1}/3 - loss: {loss.item():.4f}')
  print('Training complete.')
  "
```

The `nproc_per_node: auto` option lets Arena detect the number of GPUs available on the node and launch that many processes. You can also set it to a specific number (e.g. `2`) to control process count manually.

### Distributed Training

For multi-node training, define a `worker` section with the desired replica count. Arena always creates a Master replica (1 copy) alongside your workers — the master acts as the rendezvous point for the torchrun process group. If you do not explicitly configure `master`, it inherits the worker's resources and environment variables automatically.

Reference: [examples/v2/quickstart/pytorch-simple.yaml](../examples/v2/quickstart/pytorch-simple.yaml)

```yaml
framework:
  name: pytorch
  options:
    nproc_per_node: auto
worker:
  replicas: 2
  resources:
    # nvidia.com/gpu: "1"  # Uncomment for GPU clusters
    cpu: "2"
    memory: "8Gi"
run: |
  python -c "
  import torch, torch.distributed as dist, torch.nn as nn
  dist.init_process_group(backend='gloo')
  rank = dist.get_rank()
  model = nn.Linear(10, 1)
  optimizer = torch.optim.SGD(model.parameters(), lr=0.01)
  for epoch in range(3):
      x, y = torch.randn(64, 10), torch.randn(64, 1)
      loss = nn.MSELoss()(model(x), y)
      optimizer.zero_grad(); loss.backward(); optimizer.step()
      if rank == 0:
          print(f'Epoch {epoch+1}/3 - loss: {loss.item():.4f}')
  if rank == 0:
      print('Distributed training complete.')
  dist.destroy_process_group()
  "
```

Here 2 worker pods are created, each with 1 GPU (when uncommented). With `nproc_per_node: auto`, the total process count equals the number of GPUs across all workers.

If the master needs different resources or environment variables than the workers, define a `master` section explicitly. This overrides the inheritance behaviour so the master uses only the values you specify.

Reference: [examples/v2/reference/pytorch-mnist.yaml](../examples/v2/reference/pytorch-mnist.yaml)

```yaml
worker:
  replicas: 2
  resources:
    # nvidia.com/gpu: 1  # Uncomment for GPU clusters
  envs:
    NCCL_DEBUG: INFO

master:
  resources:
    # nvidia.com/gpu: 1  # Uncomment for GPU clusters
  envs:
    ROLE: master
```

In this example workers set `NCCL_DEBUG: INFO` for collective debugging, while the master sets `ROLE: master` to distinguish its behaviour in the training script.

### TensorBoard Integration

You can attach a TensorBoard instance to your training job to visualise logs in real time. Define storages for your data and log directories, then enable TensorBoard under `logging.tensorboard`. The `mounts` key lets you selectively mount only the storages the TensorBoard container needs — all defined storages become volumes, but only those listed under `mounts` are mounted into the TensorBoard pod.

Reference: [examples/v2/reference/pytorch-tensorboard-mounts.yaml](../examples/v2/reference/pytorch-tensorboard-mounts.yaml)

```yaml
storages:
  - name: data
    pvc: training-data-pvc
    mount_path: /data
  - name: logs
    pvc: tensorboard-logs-pvc
    mount_path: /logs
  - name: cache
    pvc: model-cache-pvc
    mount_path: /cache
logging:
  tensorboard:
    enabled: true
    logdir: /tb/logs
    mounts:
      - name: logs
        mount_path: /tb/logs
```

Here three PVCs are mounted into the training pods (`data`, `logs`, `cache`), but the TensorBoard pod only receives the `logs` storage — remounted at `/tb/logs` to match the `logdir` the training script writes to.

### Key Fields

- `framework.name: pytorch` — selects the PyTorchJob CRD provider.
- `framework.options.nproc_per_node` — `auto` (detect GPUs) or a number. Controls how many training processes run per node.
- `worker.replicas` — number of worker pods for distributed training.
- `master.replicas` — number of master pods (optional; defaults to 1, auto-created when workers are present).

## TensorFlow

Arena v2 supports distributed TensorFlow training via the TFJob CRD (`kubeflow.org/v1`). TensorFlow uses a role-based model where Worker is always required, and Chief, PS (parameter server), and Evaluator roles are optional.

### Simple Distributed Training

For straightforward parameter-server training, define only the `worker` section. Arena generates a TFJob with Worker replicas and handles the cluster spec.

Reference: [examples/v2/quickstart/tensorflow-simple.yaml](../examples/v2/quickstart/tensorflow-simple.yaml)

```yaml
framework:
  name: tensorflow
  options:
    ps_count: 1
worker:
  replicas: 2
  resources:
    # nvidia.com/gpu: "1"  # Uncomment for GPU clusters
run: |
  python -c "
  import tensorflow as tf
  strategy = tf.distribute.MultiWorkerMirroredStrategy()
  with strategy.scope():
      model = tf.keras.Sequential([tf.keras.layers.Dense(1, input_shape=(10,))])
      model.compile(optimizer='sgd', loss='mse')
  x = tf.random.normal((128, 10))
  y = tf.random.normal((128, 1))
  for epoch in range(3):
      history = model.fit(x, y, epochs=1, verbose=0)
      print(f'Epoch {epoch+1}/3 - loss: {history.history[\"loss\"][0]:.4f}')
  print('Training complete.')
  "
```

Here 2 worker pods handle training. The `ps_count` option indicates the expected number of parameter servers for the cluster configuration.

### Role-based Configuration (PS/Worker/Chief)

For full distributed TensorFlow, you can define Chief, PS, and Evaluator roles alongside Worker. Each role gets its own resources and replica count. The Chief role runs the chief replica (handles checkpoints and session coordination), PS runs parameter servers for variable storage, and Evaluator runs a separate evaluation pass.

Reference: [examples/v2/reference/tf-with-roles.yaml](../examples/v2/reference/tf-with-roles.yaml)

```yaml
worker:
  replicas: 4
  resources:
    # nvidia.com/gpu: 1  # Uncomment for GPU clusters

chief:
  resources:
    # nvidia.com/gpu: 2  # Uncomment for GPU clusters

ps:
  replicas: 2
  resources:
    cpu: "4"
    memory: 16Gi

evaluator:
  resources:
    cpu: "2"
    memory: 4Gi
```

In this example:
- **Worker** (4 replicas) performs gradient computation. Add GPUs when available.
- **Chief** (1 replica) coordinates training, saves checkpoints, and is the only role that writes summaries by default. Can use more GPUs than workers.
- **PS** (2 replicas, CPU-only) stores and synchronises model parameters — no GPU needed.
- **Evaluator** (1 replica, CPU-only) runs evaluation on a validation set in parallel with training.

Arena only includes a role in the CRD when its section is present in your YAML, except for Worker which is always required.

### Key Fields

- `framework.name: tensorflow` — selects the TFJob CRD provider.
- `worker.replicas` — number of worker pods (required).
- `chief` — optional; chief replica for checkpoint coordination (1 replica).
- `ps.replicas` — optional; number of parameter server pods.
- `evaluator` — optional; evaluation replica (1 replica).

## MPI (DeepSpeed/Horovod)

Arena v2 supports MPI-based training via the MPIJob CRD. The MPI provider always generates both a Launcher and a Worker replica spec. DeepSpeed and Horovod are framework aliases that use the same MPI provider internally — they produce identical MPIJob CRDs.

### Simple MPI

For basic MPI training, define only the `worker` section. Arena auto-creates a CPU-only launcher pod that runs your `mpirun` command and coordinates the worker pods.

Reference: [examples/v2/quickstart/mpi-simple.yaml](../examples/v2/quickstart/mpi-simple.yaml)

```yaml
framework:
  name: mpi
  options:
    slots_per_worker: 4
    mpi_implementation: OpenMPI
    launcher_creation_policy: AtStartup
worker:
  replicas: 4
  resources:
    cpu: "2"
    memory: "8Gi"
run: mpirun -n 16 python -c "print('Hello from rank', __import__('os').environ.get('OMPI_COMM_WORLD_RANK', '0'))"
```

Key options:
- `slots_per_worker` controls how many MPI ranks can run on each worker. Here 4 workers with 4 slots each gives 16 total ranks, matching `-n 16`.
- `mpi_implementation` selects the MPI flavour (`OpenMPI`, `Intel`, or `MPICH`).
- `launcher_creation_policy` determines when the launcher starts: `AtStartup` (immediately) or `WaitForWorkersReady` (after all workers are ready).

### MPI with Launcher

When the launcher needs specific resources (e.g. more CPU for orchestration, or GPUs for model initialization), define a `launcher` section explicitly. This overrides the default CPU-only behaviour.

Reference: [examples/v2/reference/mpi-horovod-mnist.yaml](../examples/v2/reference/mpi-horovod-mnist.yaml)

```yaml
worker:
  replicas: 2
  resources:
    # nvidia.com/gpu: 4  # Uncomment for GPU clusters
    cpu: "2"
    memory: "4Gi"

launcher:
  resources:
    cpu: "1"
    memory: "2Gi"
```

Here workers handle the training computation (add GPUs when available), while the launcher runs CPU-only with modest resources — it only orchestrates the `mpirun` command and does not participate in computation.

If you omit the `launcher` section, the launcher defaults to CPU-only. You can also set `framework.options.run_launcher_as_worker: true` to make the launcher inherit the worker's resources and environment variables instead.

### DeepSpeed and Horovod

DeepSpeed and Horovod are first-class framework names that map to the MPI provider. Setting `framework.name` to `deepspeed` or `horovod` generates an identical MPIJob CRD — the difference is purely semantic, helping you document which training engine the job uses.

```yaml
framework:
  name: deepspeed
worker:
  replicas: 2
  resources:
    nvidia.com/gpu: 1  # GPU required for DeepSpeed
```

See [examples/v2/pretrain/deepspeed-bert.yaml](../examples/v2/pretrain/deepspeed-bert.yaml) for a verified end-to-end example (tested on 2x Tesla T4).

```yaml
framework:
  name: horovod
worker:
  replicas: 2
  resources:
    # nvidia.com/gpu: 1  # Uncomment for GPU clusters
    cpu: "2"
    memory: "8Gi"
run: mpirun -np 2 python -c "print('Horovod rank', __import__('os').environ.get('OMPI_COMM_WORLD_RANK', '0'))"
```

Both produce an MPIJob with the same structure as `framework.name: mpi`. All MPI options (`slots_per_worker`, `mpi_implementation`, `launcher_creation_policy`, etc.) apply equally.

### Key Fields

- `framework.name: mpi` (or `deepspeed`, `horovod`) — selects the MPIJob CRD provider.
- `launcher` — optional; launcher pod resources and envs. Defaults to CPU-only if omitted.
- `worker.replicas` — number of worker pods.
- `framework.options.slots_per_worker` — MPI ranks per worker (default: 1).
- `framework.options.mpi_implementation` — `OpenMPI` (default), `Intel`, or `MPICH`.
- `framework.options.launcher_creation_policy` — `AtStartup` (default) or `WaitForWorkersReady`.
- `framework.options.run_launcher_as_worker` — when true and no launcher section is defined, the launcher inherits worker resources.
