# Arena

A command-line tool of managing the machine learning training jobs.

---

## **Overview**

Arena is a command-line interface for the data scientists to run and monitor the machine learning training jobs and check their results in an easy way. Currently it supports solo/distributed TensorFlow training. In the backend, it is based on Kubernetes, helm and Kubeflow. But the data scientists can have very little knowledge about kubernetes.

Meanwhile, the end users require GPU resource and node management. Arena also provides top command to check available GPU resources in the Kubernetes cluster.

In one word, Arena's goal is to make the data scientists feel like to work on a single machine but with the Power of GPU clusters indeed.

## **Host on Linux and MacOS**

Arena supports running on Linux and MacOS systems, please choose installation packages for different platforms to install.

## **Easy to use**

It is easy to use arena to manage your training jobs only needs you to run some commands.

## **Supports multiple types of training jobs**

You can use arena to submit multiple types of training jobs such as Tensorflow,MPI,Pytorch,Spark,Volcano,Xgboost...

## **Supports multiple users**

Arena depends on the kubeconfig file to submit the training jobs to the kubernetes cluster,so if you want to grant certain users different permissions to use arena, you can generate different kubeconfig files. 
