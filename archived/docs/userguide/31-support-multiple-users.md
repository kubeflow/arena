
## Support Multiple Users

In some usage scenarios, you may want multiple users to use arena and these users have different permissions to operate the kubernetes cluster. This guide will tell you how to implement the goal. 

Now, assume that there is 3 users to use arena and their privileges are described as follow table:


| User Name | User Namespace | Quota | Additional Privileges |
| --------- | -------------- | ----- |---------------------- |
| alex      | workplace1     | -    |-|
| bob       | workplace2     |limits.cpu: "10",limits.memory: "20Gi",requests.cpu: "5",requests.memory: "10Gi" |list the jobs in the cluster scope|
| tom       | workplace3     |requests.nvidia.com/gpu: 20|list the jobs in the namespace scope|

the following steps describe how to generate the kubeconfig files of the users.

1.Prepare the user configuration file, you can refer the ~/charts/user/values.yaml or /charts/user/values.yaml to write your own user configuration file.

The user alex doesn't need to prepare a user configuration file,because it use the default configuration. 

The user bob's user configuration file is defined as: 

```
quota:
  limits.cpu: "10"
  requests.cpu: "5"
  requests.memory: "10Gi"
  limits.memory: "20Gi"

clusterRoles:
  - apiGroups:
    - batch
    resources:
    - jobs
    verbs:
    - list
```

and store it to /tmp/bob-config.yaml

The user tom's user configuration file is defined as: 

```
quota:
  requests.nvidia.com/gpu: 5

roles:
  - apiGroups:
    - batch
    resources:
    - jobs
    verbs:
    - list
```
and store it to /tmp/tom-config.yaml


2.Generate user kubeconfig, the script 'arena-gen-kubeconfig.sh' can help you:

```
$ arena-gen-kubeconfig.sh -h

Usage:

    arena-gen-kubeconfig.sh [OPTION1] [OPTION2] ...

Options:
    --user-name <USER_NAME>                    Specify the user name
    --user-namespace <USER_NAMESPACE>          Specify the user namespace
    --user-config <USER_CONFIG>                Specify the user config,refer the ~/charts/user/values.yaml or /charts/user/values.yaml
    --force                                    If the user has been existed,force to update the user
    --delete                                   Delete the user
    --output <KUBECONFIG|USER_MANIFEST_YAML>   Specify the output kubeconfig file or the user manifest yaml
    --admin-kubeconfig <ADMIN_KUBECONFIG>      Specify the Admin kubeconfig file
    --cluster-url <CLUSTER_URL>                Specify the Cluster URL,if not specified,the script will detect the cluster url
    --create-user-yaml                         Only generate the user manifest yaml,don't apply it and create kubeconfig file
```

Firstly, create the kubeconfig file of alex: 

```
$  arena-gen-kubeconfig.sh --user-name alex --user-namespace workplace1 --output /tmp/alex.kubeconfig --force

2021-02-08/11:38:44  DEBUG  found arena charts in /Users/yangjunfeng/charts
2021-02-08/11:38:44  DEBUG  the user configuration not set,use the default configuration file
resourcequota/arena-quota-alex created
serviceaccount/alex created
clusterrole.rbac.authorization.k8s.io/arena:workplace1:alex configured
clusterrolebinding.rbac.authorization.k8s.io/arena:workplace1:alex configured
role.rbac.authorization.k8s.io/arena:alex created
rolebinding.rbac.authorization.k8s.io/arena:alex created
configmap/arena-user-alex created
Cluster "https://192.168.1.42:6443" set.
User "alex" set.
Context "registry" created.
Switched to context "registry".
2021-02-08/11:38:48  DEBUG  kubeconfig written to file /tmp/alex.kubeconfig
```
As you see the kubeconfig file has been created(/tmp/alex.kubeconfig).

Secondly, create the kubeconfig file of user bob:

```
$ arena-gen-kubeconfig.sh --user-name bob --user-namespace workplace2 --user-config /tmp/bob.yaml --output /tmp/bob.kubeconfig --force
```
the kubeconfig file will store at /tmp/bob.kubeconfig 

Thirdly, create the kubeconfig file of user tom:

```
$ arena-gen-kubeconfig.sh --user-name tom --user-namespace workplace3 --user-config /tmp/tom.yaml --output /tmp/tom.kubeconfig --force
```
the kubeconfig file will store at /tmp/tom.kubeconfig 

3.Make the kubeconfig file is valid, you can set the env KUBECONFIG like:

```
$ export KUBECONFIG=/tmp/alex.kubeconfig
 
```

4.Now you can use arena to submit your training jobs.

5.If you want to delete the user,execute the command like:

```
$ arena-gen-kubeconfig.sh --user-name tom --user-namespace workplace3 --delete
```
