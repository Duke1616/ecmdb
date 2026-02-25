#!/bin/bash

# 获取os类型
determine_os() {
    local os_type
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        if [[ "$ID" == "ubuntu" ]]; then
            os_type="Ubuntu"
        elif [[ "$ID" == "centos" ]]; then
            os_type="CentOS"
        else
            os_type="$ID"
        fi
    elif [ -f /etc/centos-release ]; then
        os_type="CentOS"
    else
        os_type="Unknown"
    fi

    # If os_type is empty, set it to "unknown"
    if [ -z "$os_type" ]; then
        os_type="unknown"
    fi

    echo "$os_type"
}

yum_repo_list(){
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
}

init_package(){
    local os_type
    os_type=$(determine_os)
    if [[ "$os_type" == "CentOS" ]]; then
        # 修复yum源
        yum_repo_list
        # 安装基础软件
        yum install -y wget docker-ce
        # docker-compose
        wget -P /usr/local/src/ https://github.com/docker/compose/releases/download/v2.29.2/docker-compose-linux-x86_64
        cp /usr/local/src/docker-compose-linux-x86_64 /usr/local/bin/docker-compose && chmod +x /usr/local/bin/docker-compose
        if docker-compose -v 2>/dev/null; then
            echo "docker-compose install success"
        else
            echo "docker-compose install error"
            exit 1
        fi
        # 重启docker
        systemctl restart docker
        systemctl enable docker
    else
        echo "Unsupported system"
        exit 1
    fi
}

# 等待MySQL运行的函数
wait_for_mysql() {
    # 设置最大等待时间（秒）
    TIMEOUT=60
    START_TIME=$(date +%s)

    while true; do
        # 获取当前时间
        CURRENT_TIME=$(date +%s)
        # 计算已等待的时间
        ELAPSED_TIME=$((CURRENT_TIME - START_TIME))

        # 如果超时，退出循环
        if [ $ELAPSED_TIME -ge $TIMEOUT ]; then
            echo "Timeout reached. MySQL did not start in time."
            exit 1
        fi

        # 检查MySQL是否运行
        if docker exec ecmdb-mysql mysqladmin ping -h"127.0.0.1" --silent; then
            echo "MySQL is up and running!"
            break
        else
            echo "Waiting for MySQL to start... ($ELAPSED_TIME seconds elapsed)"
            sleep 2
        fi
    done
}

install_backend(){
    local workdir
    project_path="/opt/ecmdb_project/ecmdb"
    if [ -d "${project_path}" ];then
        echo "后端项目目录存在"
        cd ${project_path}/deploy/ || exit 1
    else
        echo "后端项目不存在，请克隆项目后重试"
        exit 1
    fi
    # 取上级目录
    workdir=$(dirname "$(pwd)")

    docker network create ecmdb
    docker-compose up -d
    # 等待服务启动
    sleep 10
    # 创建用户
    curl -L 'http://127.0.0.1:8666/api/user/register' \
    -H 'Content-Type: application/json' \
    -d '{
        "username": "admin",
        "password": "123456",
        "re_password": "123456",
        "display_name": "系统管理员"
    }'

    # 同步权限数据
    docker cp $workdir/init/menu.tar.gz ecmdb-mongo:/mnt
    docker cp $workdir/init/role.tar.gz ecmdb-mongo:/mnt
    docker exec ecmdb-mongo mongorestore --uri="mongodb://ecmdb:123456@127.0.0.1:27017/ecmdb?authSource=admin" --gzip  --collection c_menu --archive=/mnt/menu.tar.gz
    docker exec ecmdb-mongo mongorestore --uri="mongodb://ecmdb:123456@127.0.0.1:27017/ecmdb?authSource=admin" --gzip  --collection c_role --archive=/mnt/role.tar.gz

    # 修正 ID 自增值
    docker exec ecmdb-mongo mongosh "mongodb://ecmdb:123456@127.0.0.1:27017/ecmdb" --eval 'db.c_id_generator.insertOne({ name: "c_role", next_id: NumberLong("6") })'
    docker exec ecmdb-mongo mongosh "mongodb://ecmdb:123456@127.0.0.1:27017/ecmdb" --eval 'db.c_id_generator.insertOne( { name: "c_menu", next_id:  NumberLong("171") } )'

    wait_for_mysql
    # 导入 Casbin 权限数据
    docker exec -i ecmdb-mysql mysql -u ecmdb -p123456 ecmdb < $workdir/init/casbin_rule.sql

    # 用户添加权限
    docker exec ecmdb-mongo mongosh "mongodb://ecmdb:123456@127.0.0.1:27017/ecmdb?authSource=admin" --eval 'db.c_user.updateOne( { username: "admin" }, { $set: { role_codes: ["admin"] } } )'

    # 重启后端服务，加载策略
    docker restart ecmdb
}

install_frontend(){
    local_ip=$(ip route get 1.1.1.1 | awk '{print $7}')
    cd /opt/ecmdb_project/ecmdb-web || exit
    sed -i 's#npm install -g pnpm#npm install -g pnpm --registry=https://registry.npmmirror.com#g' deploy/Dockerfile
    docker build -t duke1616/ecmdb-web:deploy-v1.0.0 -f deploy/Dockerfile .
    cd deploy/ || exit
    sed -i 's/sre/ecmdb/g' docker-compose.yaml
    sed -i 's/8888:8000/80:80/g' docker-compose.yaml
    docker-compose up -d
    sleep 5
    echo "Install Done."
    echo "Web: http://$local_ip Username: admin Password: 123456"
}

# 检查防火墙状态 (适用于 systemd)
check_firewall() {
    firewall_status=$(systemctl is-active firewalld 2>/dev/null)

    if [ "$firewall_status" == "active" ]; then
        echo "防火墙正在运行，请关闭后重试"
        exit 1
    else
        echo "防火墙已关闭"
    fi
}

# 检查 SELinux 状态
check_selinux() {
    selinux_status=$(getenforce 2>/dev/null)

    if [ "$selinux_status" = "Disabled" ] || [ "$selinux_status" = "Permissive" ]; then
        echo "SELinux 已关闭或处于宽容模式"
    else
        echo "SELinux 正在运行"
        exit 1
    fi
}

pre_check(){
    # 执行检查
    check_firewall
    check_selinux
}


main(){
    pre_check
#    init_package
    install_backend
    install_frontend
}

main