# 2019 路线图

本文档给出了 Arena 开发的路线图。

### 2019

#### 核心用户场景

目标: 方便与外部系统集成

* 集成工作
	* 提供Java, Python and C++ SDK
	* 和Kubeflow Pipelines项目集成，并且提供Standalone Job, MPI Job, Estimator Job支持

目标: 扩展能力范围，提供更多具体任务类型的提交和管理以及模型预测的能力

* High Level Interfaces:

	* 支持更多类型的数据处理和机器学习任务Spark, Flink, [XDL](https://github.com/alibaba/x-deeplearning), PyTorch, MXNet
	* 支持 Model Serving的全生命周期管理，这会依赖于[KF Serving的实现](https://github.com/kubeflow/kfserving)


目标: 支持同一个后端Operator的不同API版本，避免因为后端API版本升级影响用户对于现有任务的使用

* 适配不同版本:
	* v1aphla2 和 v1 TFJob
	* v1alpha1 和 v1aphla2 MPIJob

目标: 重构代码并且提升自动化测试能力进而能实现快速迭代

* 重构源码
	* 将`cmd`包中的通用逻辑迁移到`pkg`包，比如Trainer的接口和实现

* 自动化测试能力: 
	* 单元测试
	* 集成测试
