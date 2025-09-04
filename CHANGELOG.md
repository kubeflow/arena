# Changelog

## [v0.15.2](https://github.com/kubeflow/arena/tree/v0.15.2) (2025-09-03)

### Bug Fixes

- Fix: tfjob gets stuck in running state when succeeded pods are garbage collected ([#1370](https://github.com/kubeflow/arena/pull/1370) by [@ChenYi015](https://github.com/ChenYi015))

### Dependencies

- Bump helm.sh/helm/v3 from 3.16.3 to 3.18.4 ([#1350](https://github.com/kubeflow/arena/pull/1350) by [@ChenYi015](https://github.com/ChenYi015))
- chore(deps): bump golang.org/x/crypto from 0.39.0 to 0.40.0 ([#1351](https://github.com/kubeflow/arena/pull/1351) by [@dependabot[bot]](https://github.com/apps/dependabot))
- chore(deps): bump github.com/spf13/pflag from 1.0.6 to 1.0.7 ([#1352](https://github.com/kubeflow/arena/pull/1352) by [@dependabot[bot]](https://github.com/apps/dependabot))
- chore(deps): bump actions/download-artifact from 4 to 5 ([#1356](https://github.com/kubeflow/arena/pull/1356) by [@dependabot[bot]](https://github.com/apps/dependabot))
- chore(deps): bump actions/checkout from 4 to 5 ([#1359](https://github.com/kubeflow/arena/pull/1359) by [@dependabot[bot]](https://github.com/apps/dependabot))
- chore(deps): bump actions/setup-java from 4 to 5 ([#1366](https://github.com/kubeflow/arena/pull/1366) by [@dependabot[bot]](https://github.com/apps/dependabot))
- chore(deps): bump helm.sh/helm/v3 from 3.18.4 to 3.18.6 ([#1364](https://github.com/kubeflow/arena/pull/1364) by [@dependabot[bot]](https://github.com/apps/dependabot))
- chore(deps): bump github.com/onsi/ginkgo/v2 from 2.22.0 to 2.25.2 ([#1369](https://github.com/kubeflow/arena/pull/1369) by [@dependabot[bot]](https://github.com/apps/dependabot))

[Full Changelog](https://github.com/kubeflow/arena/compare/v0.15.1...v0.15.2)

## [v0.15.1](https://github.com/kubeflow/arena/tree/v0.15.1) (2025-06-25)

### Features

- Add support for configuring tolerations ([#1337](https://github.com/kubeflow/arena/pull/1337) by [@ChenYi015](https://github.com/ChenYi015))

### Misc

- Remove kubernetes artifacts ([#1329](https://github.com/kubeflow/arena/pull/1329) by [@ChenYi015](https://github.com/ChenYi015))
- [CI] Add CI workflow for releasing Arena images ([#1340](https://github.com/kubeflow/arena/pull/1340) by [@ChenYi015](https://github.com/ChenYi015))
- Update uninstall bash script ([#1335](https://github.com/kubeflow/arena/pull/1335) by [@ChenYi015](https://github.com/ChenYi015))
- Fix golangci-lint issues ([#1341](https://github.com/kubeflow/arena/pull/1341) by [@ChenYi015](https://github.com/ChenYi015))
- Bump golang version from 1.22.7 to 1.23.10 ([#1345](https://github.com/kubeflow/arena/pull/1345) by [@ChenYi015](https://github.com/ChenYi015))
- chore(deps): bump github.com/prometheus/common from 0.60.1 to 0.65.0 ([#1343](https://github.com/kubeflow/arena/pull/1343) by [@dependabot[bot]](https://github.com/apps/dependabot))
- chore(deps): bump golang.org/x/crypto from 0.38.0 to 0.39.0 ([#1334](https://github.com/kubeflow/arena/pull/1334) by [@dependabot[bot]](https://github.com/apps/dependabot))

[Full Changelog](https://github.com/kubeflow/arena/compare/v0.15.0...v0.15.1)

## [v0.15.0](https://github.com/kubeflow/arena/tree/v0.15.0) (2025-06-04)

### Features

- refactor: use helm lib instead of helm binary ([#1207](https://github.com/kubeflow/arena/pull/1207) by [@ChenYi015](https://github.com/ChenYi015))
- feat: add new value for using localtime in cron-operator ([#1296](https://github.com/kubeflow/arena/pull/1296) by [@ChenYi015](https://github.com/ChenYi015))
- Delete all services when the TFJob is terminated ([#1316](https://github.com/kubeflow/arena/pull/1316) by [@ChenYi015](https://github.com/ChenYi015))
- Make number of replicas of cron-operator deployment configurable ([#1325](https://github.com/kubeflow/arena/pull/1325) by [@ChenYi015](https://github.com/ChenYi015))
- Make number of replicas of tf-operator deployment configurable ([#1323](https://github.com/kubeflow/arena/pull/1323) by [@ChenYi015](https://github.com/ChenYi015))
- Add custom device support for kserve and kserving. ([#1315](https://github.com/kubeflow/arena/pull/1315) by [@Leoyzen](https://github.com/Leoyzen))
- Feat: support affinity policy for kserve and tfjob ([#1319](https://github.com/kubeflow/arena/pull/1319) by [@Syspretor](https://github.com/Syspretor))
- Feat: support separate affinity policy configuration for PS and worke… ([#1331](https://github.com/kubeflow/arena/pull/1331) by [@Syspretor](https://github.com/Syspretor))

### Bug Fixes

- fix: job status displays incorrectly ([#1289](https://github.com/kubeflow/arena/pull/1289) by [@ChenYi015](https://github.com/ChenYi015))
- fix: service account should use release namespace ([#1308](https://github.com/kubeflow/arena/pull/1308) by [@ChenYi015](https://github.com/ChenYi015))

### Misc

- Add basic e2e tests ([#1225](https://github.com/kubeflow/arena/pull/1225) by [@ChenYi015](https://github.com/ChenYi015))
- Bump github.com/containerd/containerd from 1.7.23 to 1.7.27 ([#1290](https://github.com/kubeflow/arena/pull/1290) by [@dependabot[bot]](https://github.com/apps/dependabot))
- Add stale bot to mark stale issues and PRs ([#1141](https://github.com/kubeflow/arena/pull/1141) by [@ChenYi015](https://github.com/ChenYi015))
- Fix typos in multiple files ([#1304](https://github.com/kubeflow/arena/pull/1304) by [@co63oc](https://github.com/co63oc))
- Fix typos in multiple files ([#1310](https://github.com/kubeflow/arena/pull/1310) by [@co63oc](https://github.com/co63oc))

[Full Changelog](https://github.com/kubeflow/arena/compare/v0.14.2...v0.15.0)

## [v0.14.2](https://github.com/kubeflow/arena/tree/v0.14.2) (2025-03-10)

### Misc

- Fix typos ([#1276](https://github.com/kubeflow/arena/pull/1276) by [@co63oc](https://github.com/co63oc))
- Update pytorch operator image ([#1281](https://github.com/kubeflow/arena/pull/1281) by [@ChenYi015](https://github.com/ChenYi015))

[Full Changelog](https://github.com/kubeflow/arena/compare/v0.14.1...v0.14.2)

## [v0.14.1](https://github.com/kubeflow/arena/tree/v0.14.1) (2025-02-24)

### Bug Fixes

- fix: device value does not support k8s resource quantity ([#1267](https://github.com/kubeflow/arena/pull/1267) by [@ChenYi015](https://github.com/ChenYi015))
- fix: pytorchjob does not support backoff limit ([#1272](https://github.com/kubeflow/arena/pull/1272) by [@ChenYi015](https://github.com/ChenYi015))
- unset env NVIDIA_VISIBLE_DEVICES when gpushare is enabled ([#1273](https://github.com/kubeflow/arena/pull/1273) by [@ChenYi015](https://github.com/ChenYi015))

### Misc

- docs: fixed typo ([#1257](https://github.com/kubeflow/arena/pull/1257) by [@DBMxrco](https://github.com/DBMxrco))
- Bump github.com/golang/glog from 1.2.3 to 1.2.4 ([#1263](https://github.com/kubeflow/arena/pull/1263) by [@dependabot[bot]](https://github.com/apps/dependabot))
- fix: format of tensorflow standalone training docs is messed up ([#1265](https://github.com/kubeflow/arena/pull/1265) by [@ChenYi015](https://github.com/ChenYi015))

[Full Changelog](https://github.com/kubeflow/arena/compare/v0.14.0...v0.14.1)

## [v0.14.0](https://github.com/kubeflow/arena/tree/v0.14.0) (2025-02-12)

### Features

- rename parameter ([#1262](https://github.com/kubeflow/arena/pull/1262) by [@gujingit](https://github.com/gujingit))

### Misc

- Add changelog for v0.13.1 ([#1248](https://github.com/kubeflow/arena/pull/1248) by [@ChenYi015](https://github.com/ChenYi015))
- Bump github.com/go-resty/resty/v2 from 2.16.0 to 2.16.5 ([#1254](https://github.com/kubeflow/arena/pull/1254) by [@dependabot[bot]](https://github.com/apps/dependabot))

[Full Changelog](https://github.com/kubeflow/arena/compare/v0.13.1...v0.14.0)

## [v0.13.1](https://github.com/kubeflow/arena/tree/v0.13.1) (2025-01-13)

### Misc

- feat: add linux/arm64 support for tf-operator image ([#1238](https://github.com/kubeflow/arena/pull/1238) by [@ChenYi015](https://github.com/ChenYi015))
- feat: add linux/arm64 support for mpi-operator image ([#1239](https://github.com/kubeflow/arena/pull/1239) by [@ChenYi015](https://github.com/ChenYi015))
- feat: add linux/arm64 support for cron-operator image ([#1240](https://github.com/kubeflow/arena/pull/1240) by [@ChenYi015](https://github.com/ChenYi015))
- feat: add linux/arm64 support for et-operator image ([#1241](https://github.com/kubeflow/arena/pull/1241) by [@ChenYi015](https://github.com/ChenYi015))
- Add PyTorch mnist example ([#1237](https://github.com/kubeflow/arena/pull/1237) by [@ChenYi015](https://github.com/ChenYi015))
- Update the version of elastic-job-supervisor in arena-artifacts ([#1247](https://github.com/kubeflow/arena/pull/1247) by [@AlanFokCo](https://github.com/AlanFokCo))

[Full Changelog](https://github.com/kubeflow/arena/compare/v0.13.0...v0.13.1)

## [v0.13.0](https://github.com/kubeflow/arena/tree/v0.13.0) (2024-12-23)

### New Features

- feat: add support for torchrun ([#1228](https://github.com/kubeflow/arena/pull/1228) by [@ChenYi015](https://github.com/ChenYi015))
- Update pytorch-operator image ([#1234](https://github.com/kubeflow/arena/pull/1234) by [@ChenYi015](https://github.com/ChenYi015))

### Bug Fix

- Avoid listing jobs and statefulsets when get pytorchjob ([#1229](https://github.com/kubeflow/arena/pull/1229) by [@ChenYi015](https://github.com/ChenYi015))

### Misc

- Update tfjob standalone training job doc ([#1222](https://github.com/kubeflow/arena/pull/1222) by [@ChenYi015](https://github.com/ChenYi015))
- Remove archived docs ([#1208](https://github.com/kubeflow/arena/pull/1208) by [@ChenYi015](https://github.com/ChenYi015))
- Add changelog for v0.12.1 ([#1224](https://github.com/kubeflow/arena/pull/1224) by [@ChenYi015](https://github.com/ChenYi015))
- Bump golang.org/x/crypto from 0.29.0 to 0.31.0 ([#1231](https://github.com/kubeflow/arena/pull/1231) by [@dependabot[bot]](https://github.com/apps/dependabot))
- Bump google.golang.org/protobuf from 1.35.1 to 1.36.0 ([#1227](https://github.com/kubeflow/arena/pull/1227) by [@dependabot[bot]](https://github.com/apps/dependabot))

[Full Changelog](https://github.com/kubeflow/arena/compare/v0.12.1...v0.13.0)

## [v0.12.1](https://github.com/kubeflow/arena/tree/v0.12.1) (2024-11-25)

### New Features

- Support MPI Job with generic devices ([#1209](https://github.com/kubeflow/arena/pull/1209) by [@cheyang](https://github.com/cheyang))

### Bug Fix

- Update tf-operator image to fix clean pod policy issues ([#1200](https://github.com/kubeflow/arena/pull/1200) by [@ChenYi015](https://github.com/ChenYi015))
- Fix etjob rendering error when using local logging dir ([#1203](https://github.com/kubeflow/arena/pull/1203) by [@TrafalgarZZZ](https://github.com/TrafalgarZZZ))
- Fix the functionality of generating kubeconfig (#1204) ([#1205](https://github.com/kubeflow/arena/pull/1205) by [@wqlparallel](https://github.com/wqlparallel))
- Update cron operator image ([#1214](https://github.com/kubeflow/arena/pull/1214) by [@ChenYi015](https://github.com/ChenYi015))

### Misc

- Add changelog for v0.12.0 ([#1199](https://github.com/kubeflow/arena/pull/1199) by [@ChenYi015](https://github.com/ChenYi015))
- Add go mod vendor check to integration test ([#1198](https://github.com/kubeflow/arena/pull/1198) by [@ChenYi015](https://github.com/ChenYi015))
- bump github.com/go-resty/resty/v2 from 2.15.3 to 2.16.0 ([#1202](https://github.com/kubeflow/arena/pull/1202) by [@dependabot[bot]](https://github.com/apps/dependabot))
- Publish releases only on master branch ([#1210](https://github.com/kubeflow/arena/pull/1210) by [@ChenYi015](https://github.com/ChenYi015))
- Add docs for releasing arena ([#1201](https://github.com/kubeflow/arena/pull/1201) by [@ChenYi015](https://github.com/ChenYi015))
- Bump golang.org/x/crypto from 0.28.0 to 0.29.0 ([#1206](https://github.com/kubeflow/arena/pull/1206) by [@dependabot[bot]](https://github.com/apps/dependabot))
- Release v0.12.1 ([#1215](https://github.com/kubeflow/arena/pull/1215) by [@ChenYi015](https://github.com/ChenYi015))

[Full Changelog](https://github.com/kubeflow/arena/compare/29b2d6d2...v0.12.1)

## [v0.12.0](https://github.com/kubeflow/arena/tree/v0.12.0) (2024-11-11)

### New Features

- Feat: add support for distributed serving type ([#1187](https://github.com/kubeflow/arena/pull/1187) by [@linnlh](https://github.com/linnlh))
- Support distributed serving with vendor update ([#1194](https://github.com/kubeflow/arena/pull/1194) by [@cheyang](https://github.com/cheyang))

### Misc

- Bump github.com/golang/glog from 1.2.2 to 1.2.3 ([#1189](https://github.com/kubeflow/arena/pull/1189) by [@dependabot[bot]](https://github.com/apps/dependabot))
- Bump github.com/prometheus/common from 0.60.0 to 0.60.1 ([#1182](https://github.com/kubeflow/arena/pull/1182) by [@dependabot[bot]](https://github.com/apps/dependabot))
- Bump mkdocs-material from 9.5.42 to 9.5.44 ([#1190](https://github.com/kubeflow/arena/pull/1190) by [@dependabot[bot]](https://github.com/apps/dependabot))
- Release v0.12.0 ([#1197](https://github.com/kubeflow/arena/pull/1197) by [@ChenYi015](https://github.com/ChenYi015))

[Full Changelog](https://github.com/kubeflow/arena/compare/46a795e3...v0.12.0)

## [v0.11.0](https://github.com/kubeflow/arena/tree/v0.11.0) (2024-10-24)

### New Features

- Support ray job  ([#1123](https://github.com/kubeflow/arena/pull/1123) by [@qile123](https://github.com/qile123))

### Misc

- Bump github.com/prometheus/client_golang from 1.20.4 to 1.20.5 ([#1176](https://github.com/kubeflow/arena/pull/1176) by [@dependabot[bot]](https://github.com/apps/dependabot))
- Bump mkdocs-material from 9.5.40 to 9.5.42 ([#1179](https://github.com/kubeflow/arena/pull/1179) by [@dependabot[bot]](https://github.com/apps/dependabot))

[Full Changelog](https://github.com/kubeflow/arena/compare/e15cb18...v0.11.0)

## [v0.10.1](https://github.com/kubeflow/arena/tree/v0.10.1) (2024-10-14)

### Bug Fixes

- fix: keep arena installer after installing the binary ([#1164](https://github.com/kubeflow/arena/pull/1164) by [@ChenYi015](https://github.com/ChenYi015))
- fix: unsupported success policy when success policy is not specified ([#1170](https://github.com/kubeflow/arena/pull/1170) by [@ChenYi015](https://github.com/ChenYi015))
- fix: failed to sync cache due to status subresouce missed in tfjob CRD ([#1173](https://github.com/kubeflow/arena/pull/1173) by [@ChenYi015](https://github.com/ChenYi015))

### Misc

- Bump github.com/prometheus/common from 0.59.1 to 0.60.0 ([#1160](https://github.com/kubeflow/arena/pull/1160) by [@dependabot[bot]](https://github.com/apps/dependabot))
- Bump golang.org/x/crypto from 0.27.0 to 0.28.0 ([#1162](https://github.com/kubeflow/arena/pull/1162) by [@dependabot[bot]](https://github.com/apps/dependabot))
- Migrate docker image to ACREE ([#1171](https://github.com/kubeflow/arena/pull/1171) by [@ChenYi015](https://github.com/ChenYi015))
- Bump mkdocs-material from 9.5.38 to 9.5.40 ([#1166](https://github.com/kubeflow/arena/pull/1166) by [@dependabot[bot]](https://github.com/apps/dependabot))
- Bump google.golang.org/protobuf from 1.34.2 to 1.35.1 ([#1163](https://github.com/kubeflow/arena/pull/1163) by [@dependabot[bot]](https://github.com/apps/dependabot))
- Remove redundant run_arena.sh file ([#1172](https://github.com/kubeflow/arena/pull/1172) by [@ChenYi015](https://github.com/ChenYi015))

[Full Changelog](https://github.com/kubeflow/arena/compare/v0.10.0...v0.10.1)

## [v0.10.0](https://github.com/kubeflow/arena/tree/v0.10.0) (2024-09-29)

### New Features

- Support multiple type devices ([#1122](https://github.com/kubeflow/arena/pull/1122) by [@lizhiboo](https://github.com/lizhiboo))
- Increase RSA key bit size from 1024 to 2048 ([#1130](https://github.com/kubeflow/arena/pull/1130) by [@ChenYi015](https://github.com/ChenYi015))
- Add success policy to TF training job ([#1148](https://github.com/kubeflow/arena/pull/1148) by [@ChenYi015](https://github.com/ChenYi015))

### Bug Fixes

- Fix submitting spark training jobs and update docs ([#1112](https://github.com/kubeflow/arena/pull/1112) by [@ChenYi015](https://github.com/ChenYi015))
- docs: fix broken links and add CI for checking document build status ([#1131](https://github.com/kubeflow/arena/pull/1131) by [@ChenYi015](https://github.com/ChenYi015))
- [Bugfix] Make PytorchJob devices format to key=value ([#1155](https://github.com/kubeflow/arena/pull/1155) by [@AlanFokCo](https://github.com/AlanFokCo))

### SDK

- Bump arena Java SDK version to 1.0.8 ([#1124](https://github.com/kubeflow/arena/pull/1124) by [@ChenYi015](https://github.com/ChenYi015))

### Misc

- Remove docker dependency ([#1113](https://github.com/kubeflow/arena/pull/1113) by [@Syulin7](https://github.com/Syulin7))
- Update Makefile and release workflow ([#1128](https://github.com/kubeflow/arena/pull/1128) by [@ChenYi015](https://github.com/ChenYi015))
- chore: remove travis and circle CI ([#1129](https://github.com/kubeflow/arena/pull/1129) by [@ChenYi015](https://github.com/ChenYi015))
- chore: add issue templates and update depenabot bot ([#1140](https://github.com/kubeflow/arena/pull/1140) by [@ChenYi015](https://github.com/ChenYi015))
- Bump github.com/golang/glog from 1.1.2 to 1.2.2 ([#1139](https://github.com/kubeflow/arena/pull/1139) by [@dependabot[bot]](https://github.com/apps/dependabot))
- Bump golang.org/x/crypto from 0.21.0 to 0.27.0 ([#1126](https://github.com/kubeflow/arena/pull/1126) by [@dependabot[bot]](https://github.com/apps/dependabot))
- Bump github.com/spf13/cobra from 1.8.0 to 1.8.1 ([#1137](https://github.com/kubeflow/arena/pull/1137) by [@dependabot[bot]](https://github.com/apps/dependabot))
- Bump github.com/go-resty/resty/v2 from 2.12.0 to 2.14.0 ([#1134](https://github.com/kubeflow/arena/pull/1134) by [@dependabot[bot]](https://github.com/apps/dependabot))
- Bump github.com/kserve/kserve from 0.13.0 to 0.13.1 ([#1135](https://github.com/kubeflow/arena/pull/1135) by [@dependabot[bot]](https://github.com/apps/dependabot))
- Bump github.com/prometheus/common from 0.45.0 to 0.59.1 ([#1138](https://github.com/kubeflow/arena/pull/1138) by [@dependabot[bot]](https://github.com/apps/dependabot))
- Bump client-java from 10.0.1 to 11.0.1 ([#1132](https://github.com/kubeflow/arena/pull/1132) by [@ChenYi015](https://github.com/ChenYi015))
- Bump github.com/prometheus/client_golang from 1.20.0 to 1.20.4 ([#1144](https://github.com/kubeflow/arena/pull/1144) by [@dependabot[bot]](https://github.com/apps/dependabot))
- Bump github.com/go-resty/resty/v2 from 2.14.0 to 2.15.0 ([#1143](https://github.com/kubeflow/arena/pull/1143) by [@dependabot[bot]](https://github.com/apps/dependabot))
- Bump mkdocs-material from 9.5.34 to 9.5.35 ([#1145](https://github.com/kubeflow/arena/pull/1145) by [@dependabot[bot]](https://github.com/apps/dependabot))
- Bump github.com/go-resty/resty/v2 from 2.15.0 to 2.15.1 ([#1147](https://github.com/kubeflow/arena/pull/1147) by [@dependabot[bot]](https://github.com/apps/dependabot))
- Bump github.com/go-resty/resty/v2 from 2.15.1 to 2.15.2 ([#1150](https://github.com/kubeflow/arena/pull/1150) by [@dependabot[bot]](https://github.com/apps/dependabot))
- Bump mkdocs-material from 9.5.35 to 9.5.36 ([#1151](https://github.com/kubeflow/arena/pull/1151) by [@dependabot[bot]](https://github.com/apps/dependabot))
- Bump golang from 1.21 to 1.22.7 ([#1142](https://github.com/kubeflow/arena/pull/1142) by [@ChenYi015](https://github.com/ChenYi015))
- Bump mkdocs-material from 9.5.36 to 9.5.38 ([#1153](https://github.com/kubeflow/arena/pull/1153) by [@dependabot[bot]](https://github.com/apps/dependabot))
- Bump github.com/go-resty/resty/v2 from 2.15.2 to 2.15.3 ([#1156](https://github.com/kubeflow/arena/pull/1156) by [@dependabot[bot]](https://github.com/apps/dependabot))
- Release v0.10.0 ([#1157](https://github.com/kubeflow/arena/pull/1157) by [@ChenYi015](https://github.com/ChenYi015))

[Full Changelog](https://github.com/kubeflow/arena/compare/v0.9.16...v0.10.0)
