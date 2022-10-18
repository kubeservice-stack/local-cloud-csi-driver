# local-cloud-csi-driver

## 背景

LVM CSI 插件可用于帮助简化存储管理. 使用 csi 配置创建 pv，并且 pvc、pod 像往常一样定义。

Lvm 是基于主机的，而不是您的应用程序的高可用性解决方案。使用 lvm 应该接受主机关闭的效果。

## 配置要求

具有所需 RBAC 权限的服务帐户

## 功能状态

编译打包 `local.csi.ecloud.cmss.com` 可以编译成容器的形式。

构建容器：

```bash
$ docker build -f hack/local/Dockerfile .
```
## 用法

### 先决条件
使用localdisk 或者 挂载clouddisk方式，挂载或生成 `lvm pvcreate` 或 `lvm vgcreate` 

```bash
$ fdisk -l
Disk /dev/sdc: 68.7 GB, 68719476736 bytes, 134217728 sectors
Units = sectors of 1 * 512 = 512 bytes
Sector size (logical/physical): 512 bytes / 512 bytes
I/O size (minimum/optimal): 512 bytes / 512 bytes

$ fdisk /dev/sdc
Command (m for help): p

Disk /dev/sdc: 68.7 GB, 68719476736 bytes, 134217728 sectors
Units = sectors of 1 * 512 = 512 bytes
Sector size (logical/physical): 512 bytes / 512 bytes
I/O size (minimum/optimal): 512 bytes / 512 bytes
Disk label type: dos
Disk identifier: 0x72a4d30c

   Device Boot      Start         End      Blocks   Id  System

Command (m for help): n
Partition type:
   p   primary (0 primary, 0 extended, 4 free)
   e   extended
Select (default p): p
Partition number (1-4, default 1): 1
First sector (2048-134217727, default 2048): 
Using default value 2048
Last sector, +sectors or +size{K,M,G} (2048-134217727, default 134217727): 64217727        
Partition 1 of type Linux and of size 30.6 GiB is set

Command (m for help): t
Selected partition 1
Hex code (type L to list all codes): 8e
Changed type of partition 'Linux' to 'Linux LVM'

Command (m for help): w
The partition table has been altered!

$ lsblk 
NAME   MAJ:MIN RM  SIZE RO TYPE MOUNTPOINT
sdb      8:16   0  100G  0 disk 
|-sdb4   8:20   0   25G  0 part 
`-sdb2   8:18   0   75G  0 part 
sr0     11:0    1  506K  0 rom  
sdc      8:32   0   64G  0 disk 
`-sdc1   8:33   0 30.6G  0 part 
sda      8:0    0  100G  0 disk 
`-sda1   8:1    0  100G  0 part /

$ pvcreate /dev/sdc1 
  Physical volume "/dev/sdc1" successfully created.
$ vgcreate volumegroup1 /dev/sdc1
  Volume group "volumegroup1" successfully created
$ vgdisplay 
  --- Volume group ---
  VG Name               volumegroup1
  System ID             
  Format                lvm2
  Metadata Areas        1
  Metadata Sequence No  2
  VG Access             read/write
  VG Status             resizable
  MAX LV                0
  Cur LV                1
  Open LV               1
  Max PV                0
  Cur PV                1
  Act PV                1
  VG Size               <30.62 GiB
  PE Size               4.00 MiB
  Total PE              7838
  Alloc PE / Size       512 / 2.00 GiB
  Free  PE / Size       7326 / <28.62 GiB
  VG UUID               V6TVTh-AcIi-hLmR-bozc-9QeA-EBnU-Mhhd6y
```

### 执行步骤
第 1 步：创建 `CSI Provisioner`
```bash
$ kubectl create -f ./deploy/local/provisioner.yaml
```
第 2 步：创建 `CSI` 插件
```bash
$ kubectl create -f ./deploy/local/plugin.yaml
```

第 3 步：创建存储类
```bash
$ kubectl create -f ./examples/storageclass.yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
   name: csi-lvm
provisioner: local.csi.ecloud.cmss.com
parameters:
    vgName: volumegroup1
    fsType: ext4
    pvType: localdisk
    nodeAffinity: "false"
    readIOPS: "2000"
    writeIOPS: "1000"
    readBPS: "10000"
    writeBPS: "5000"
reclaimPolicy: Delete
```
用法：

vgName：定义存储类的卷组名；

fsType：默认为ext4，定义lvm文件系统类型，支持ext4、ext3、xfs；

pvType：可选，默认为云盘。定义使用的物理磁盘类型，支持clouddisk、localdisk；

nodeAffinity：可选，默认为 true。决定是否在 PV 中添加 nodeAffinity。
----> true：默认，使用 nodeAffinity 配置创建 PV；
----> false：不配置nodeAffinity创建PV，pod可以调度到任意节点

volumeBindingMode：支持 Immediate/WaitForFirstConsumer 
----> Immediate：表示将在创建 pvc 时配置卷，在此配置中 nodeAffinity 将可用；
----> WaitForFirstConsumer：表示在相关的pod创建之前不会创建volume；在配置中，nodeAffinity 将不可用；

第 4 步：使用 `lvm` 创建 `nginx` 部署
```bash
$ kubectl create -f ./examples/pvc.yaml
$ kubectl create -f ./examples/deploy.yaml
```

第 5 步：检查 PVC/PV 的状态
```bash
$ kubectl get pvc
NAME      STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
lvm-pvc   Bound    lvm-29def33c-8dae-482f-8d64-c45e741facd9   2Gi        RWO            csi-lvm        3h37m

$ kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM             STORAGECLASS   REASON   AGE
lvm-29def33c-8dae-482f-8d64-c45e741facd9   2Gi        RWO            Delete           Bound    default/lvm-pvc   csi-lvm                 3h38m
```

第 6 步：检查 pod 的状态
检查 pod 中的目录

```bash
$ kubectl get pod | grep deployment-lvm
deployment-lvm-57bc9bcd64-j7r9x   1/1     Running   0          77s

$ kubectl exec -ti deployment-lvm-57bc9bcd64-j7r9x   sh
kubectl exec [POD] [COMMAND] is DEPRECATED and will be removed in a future version. Use kubectl exec [POD] -- [COMMAND] instead.
# df -h | grep data
/dev/mapper/volumegroup1-lvm--9e30e658--5f85--4ec6--ada2--c4ff308b506e  2.0G  6.0M  1.8G   1% /data
```

检查主机中的目录：
```bash
$ kubectl describe pod deployment-lvm-57bc9bcd64-j7r9x | grep Node:
Node:         kcs-cpu-test-m-8mzmj/172.16.0.67

$ ifconfig | grep 172.16.0.67
        inet 172.16.0.67  netmask 255.255.0.0  broadcast 172.16.255.255
  
$ mount | grep volumegroup
/dev/mapper/volumegroup1-lvm--9e30e658--5f85--4ec6--ada2--c4ff308b506e on /var/lib/kubelet/pods/c06d5521-3d9c-4517-bdc2-e6df34b9e8f1/volumes/kubernetes.io~csi/lvm-9e30e658-5f85-4ec6-ada2-c4ff308b506e/mount type ext4 (rw,relatime,data=ordered)
/dev/mapper/volumegroup1-lvm--9e30e658--5f85--4ec6--ada2--c4ff308b506e on /var/lib/paascontainer/kubelet/pods/c06d5521-3d9c-4517-bdc2-e6df34b9e8f1/volumes/kubernetes.io~csi/lvm-9e30e658-5f85-4ec6-ada2-c4ff308b506e/mount type ext4 (rw,relatime,data=ordered)
```

检查pod disk iops和bps设置，是否生效：
```bash
$ pwd
/sys/fs/cgroup/blkio/kubepods.slice/kubepods-besteffort.slice/kubepods-besteffort-podc06d5521_3d9c_4517_bdc2_e6df34b9e8f1.slice

$ cat blkio.throttle.read_bps_device 
253:1 10000
$ cat blkio.throttle.write_bps_device 
253:1 5000
$ cat blkio.throttle.write_iops_device 
253:1 1000
$ cat blkio.throttle.read_iops_device 
253:1 2000
```
