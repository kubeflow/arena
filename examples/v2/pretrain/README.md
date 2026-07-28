# Pre-training Examples

| File                  | Framework | Description                                                          | Status            |
| --------------------- | --------- | -------------------------------------------------------------------- | ----------------- |
| `deepspeed-bert.yaml` | PyTorch   | DeepSpeed BERT pre-training with ZeRO-1 optimizer offloading (GPU).  | E2E verified (GPU) |

## deepspeed-bert.yaml

Verified end-to-end on a 2x Tesla T4 GPU cluster. Uses the `ghcr.io/kubeflow/training-v1/pytorch-deepspeed-demo:latest` image with DeepSpeed 0.7.2. Requires GPUs — DeepSpeed's CPUAdam optimizer depends on cuBLAS at runtime.

The `HF_ENDPOINT` and `PIP_INDEX_URL` environment variables are optional mirrors for clusters where HuggingFace Hub or PyPI are not directly accessible. Override or remove as needed.

## Planned templates

The following templates are planned for future addition:

- `llm-pretrain.yaml` — LLM pre-training with DeepSpeed (`deepspeed` framework, `shm` storage, multi-PVC)
- `deepspeed-zero3.yaml` — DeepSpeed ZeRO-3 with parameter partitioning
