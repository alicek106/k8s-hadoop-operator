# k8s-hadoop-operator
Kubernetes Hadoop operator for studying Operator SDK. it can be used for educational purpose (maybe :D)

# Korean Guide
https://blog.naver.com/alice_k106/221586279079

# How to use
1. Clone project and create operator 
```
$ git clone https://github.com/alicek106/k8s-hadoop-operator.git
'k8s-hadoop-operator'에 복제합니다...
remote: Enumerating objects: 80, done.
remote: Counting objects: 100% (80/80), done.
remote: Compressing objects: 100% (53/53), done.
remote: Total 80 (delta 22), reused 76 (delta 18), pack-reused 0
오브젝트 묶음 푸는 중: 100% (80/80), 완료.
 
$ cd k8s-hadoop-operator/temporary-gopath/src/hadoop-operator/
 
$ kubectl apply -f deploy/crds/alicek106_v1alpha1_hadoopservice_crd.yaml
customresourcedefinition.apiextensions.k8s.io/hadoopservices.alicek106.hadoop unchanged
 
$ kubectl apply -f deploy/
deployment.apps/hadoop-operator created
role.rbac.authorization.k8s.io/hadoop-operator configured
rolebinding.rbac.authorization.k8s.io/hadoop-operator unchanged
serviceaccount/hadoop-operator unchanged
```

2. Create hadoop custom resource
```
$ kubectl apply -f deploy/crds/alicek106_v1alpha1_hadoopservice_cr.yaml
hadoopservice.alicek106.hadoop/example-hadoopservice created

$ kubectl get pods
NAME                                    READY   STATUS        RESTARTS   AGE
example-hadoopservice-hadoop-master-0   0/1     Running       0          11s
example-hadoopservice-hadoop-slave-0    1/1     Running       0          11s
example-hadoopservice-hadoop-slave-1    1/1     Running       0          11s
example-hadoopservice-hadoop-slave-2    1/1     Running       0          11s
 
$ kubectl get hds # 원래 이름은 hadoopservice
NAME                    AGE
example-hadoopservice   31m
```

3. Check password for master ssh access (from operator log)
```
$ kubectl logs hadoop-operator-859db46995-f926p
...
{"level":"info","ts":1563175847.3946388,"logger":"controller_hadoopservice","msg":"Generated password is GSUJ"}
```

4. Access ssh to master (using nodePort. If you wish, change to LB :D)
```
hadoop-operator alice(k8s: aws-context) $ kubectl get svc
NAME                                               TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                                       AGE
example-hadoopservice-hadoop-master-svc            ClusterIP   None             <none>        <none>                                        74s
example-hadoopservice-hadoop-master-svc-external   NodePort    10.103.102.240   <none>        22:30085/TCP,8088:30988/TCP,50070:30530/TCP   74s
example-hadoopservice-hadoop-slave-svc             ClusterIP   None             <none>        22/TCP                                        74s
hadoop-operator                                    ClusterIP   10.106.162.110   <none>        8383/TCP                                      3m11s
kubernetes                                         ClusterIP   10.96.0.1        <none>        443/TCP                                       2d23h
 
 
hadoop-operator alice(k8s: aws-context) $ ssh root@MYIP -p 30085
root@13.125.34.239 password: # Operator 로그에서 출력된 비밀번호를 입력한다
Welcome to Ubuntu 14.04.3 LTS (GNU/Linux 4.4.0-1087-aws x86_64)
 
 * Documentation:  https://help.ubuntu.com/
 
root@example-hadoopservice-hadoop-master-0:~ hdfs dfsadmin -report # 하둡 동작 확인
Configured Capacity: 24777043968 (23.08 GB)
...
```

Happy Map-reduce!


# Limitations
- Embedded ssh key (it should be generated each provisioning)
- Not considered resource quota setting (QOS)
- etc
