
## 环境信息

OS：CentOS7.9

> 注意: 要刚安装的系统，配置好IP地址

## 提前准备

```bash
# 关闭防火墙
systemctl disable firewalld --now
# 关闭selinux
setenforce 0
# 修复CentOS7软件源
mkdir -p /etc/yum.repos.d/backup
mv /etc/yum.repos.d/*.repo /etc/yum.repos.d/backup/
tee /etc/yum.repos.d/CentOS-Vault.repo <<-'EOF'
[base]
name=CentOS-$releasever - Base
baseurl=https://mirrors.tuna.tsinghua.edu.cn/centos-vault/7.9.2009/os/$basearch/
gpgcheck=1
gpgkey=https://mirrors.tuna.tsinghua.edu.cn/centos-vault/RPM-GPG-KEY-CentOS-7

[updates]
name=CentOS-$releasever - Updates
baseurl=https://mirrors.tuna.tsinghua.edu.cn/centos-vault/7.9.2009/updates/$basearch/
gpgcheck=1
gpgkey=https://mirrors.tuna.tsinghua.edu.cn/centos-vault/RPM-GPG-KEY-CentOS-7

[extras]
name=CentOS-$releasever - Extras
baseurl=https://mirrors.tuna.tsinghua.edu.cn/centos-vault/7.9.2009/extras/$basearch/
gpgcheck=1
gpgkey=https://mirrors.tuna.tsinghua.edu.cn/centos-vault/RPM-GPG-KEY-CentOS-7

[centosplus]
name=CentOS-$releasever - Plus
baseurl=https://mirrors.tuna.tsinghua.edu.cn/centos-vault/7.9.2009/centosplus/$basearch/
gpgcheck=1
enabled=0
gpgkey=https://mirrors.tuna.tsinghua.edu.cn/centos-vault/RPM-GPG-KEY-CentOS-7
EOF
curl -o /etc/yum.repos.d/epel.repo http://mirrors.aliyun.com/repo/epel-7.repo
curl -o /etc/yum.repos.d/docker-ce.repo https://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
yum clean all && yum makecache


# 安装基础软件
yum install -y git wget docker-ce
systemctl restart docker && systemctl enable docker
```

## clone项目（需要魔法）

```bash
# docker-compose
wget -P /usr/local/src/ https://github.com/docker/compose/releases/download/v2.29.2/docker-compose-linux-x86_64
cp /usr/local/src/docker-compose-linux-x86_64 /usr/local/bin/docker-compose && chmod +x /usr/local/bin/docker-compose
# 检查docker-compose
docker-compose -v

mkdir -pv /opt/ecmdb_project && cd /opt/ecmdb_project
git clone https://github.com/Duke1616/ecmdb.git
git clone https://github.com/Duke1616/ecmdb-web.git
```

## 执行安装脚本（需要魔法）

```bash
cd /opt/ecmdb_project/ecmdb/deploy
bash -x install_centos7.sh
```