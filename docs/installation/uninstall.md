# Uninstall Arena

Sometimes you may wish to delete arena, this document will help you.

## What will be deleted?

The following resources will be deleted: 

* Custom Resource Defines(CRDs) created by arena
* the 'arena-system' namespace
* the directories include ~/charts and /charts on you computer
* the arena binary file which is hosted on /usr/local/bin/arena  

## Arena Version >= 0.8.4

If your arena version >= 0.8.4,the `arena-uninstall` already exists on your computer and you can run `arena-uninstall -h` to get the usage.

```
$  arena-uninstall -h

Usage of arena-uninstall:
  -force
    	force delete the Custom Resource Instances
  -manifest-dir string
    	specify the kubernetes-artifacts directory
  -quiet
    	quiet for all choices
```

1\. There is an question about that whether you force delete the Running Custom Resource Instances(eg: tfjobs,mpijobs,pytorchjobs), like: 

```
$ arena-uninstall
Please confirm whether to delete the running Custom Resource Instances(eg: tfjob,mpijob)[Y/N]:
``` 

2\. If you want to force delete the running Custom Resource Instances, you can add option '--force':


```
$ arena-uninstall --force
```

3\. If you don't want the question when run the arena-uninstall,you can:

```
$ arena-uninstall --quiet
```

## Arena Version < 0.8.4

Firstly, you should download the arena package from [releases](https://github.com/kubeflow/arena/releases) and its' version must large than 0.8.4

Then,execute the following commands to delete arena:

```
$ tar -xf arena-installer-xxxx-xxxxx-linux-amd64.tar.gz

$ cd arena-installer

$ bin/arena-uninstall
```
