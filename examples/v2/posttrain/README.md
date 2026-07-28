# Post-training Examples

Post-training templates (SFT, LoRA, RLHF) are planned for future addition.

Planned templates:

- `sft.yaml` — Supervised fine-tuning (multi-PVC, `nproc_per_node: auto`, `torchrun`)
- `lora.yaml` — LoRA fine-tuning (single worker, `HF_HOME` env)
- `rlhf.yaml` — RLHF with PPO (4 workers, references SFT model and reward model paths)
