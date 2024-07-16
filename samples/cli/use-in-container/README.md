## Use arena CLI in container

This repository is a demo of using the arena CLI in container.

1. Build arena CLI image
```shell
docker build -t arena-demo:test -f Dockerfile .
```

2. Create rbac for arena
```shell
kubectl create -f rbac.yaml
```

3. Create arena CLI deployment
```shell
kubectl create -f deployment.yaml
```

4. Execute arena CLI in container
```shell
kubectl exec -it arena-demo-57f7b467f9-p48ph -- arena version
```
