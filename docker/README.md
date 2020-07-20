
docker编译部署环境
------


## docker 安装

docker安装 [http://get.daocloud.io/#install-docker-for-mac-windows](http://get.daocloud.io/#install-docker-for-mac-windows)


docker-hub国内源设置 [https://www.daocloud.io/mirror#accelerator-doc](https://www.daocloud.io/mirror#accelerator-doc)

## 构建docker镜像

```shell
cd PlatONE-Go/docker

docker build -t platone:dev .
```

## 启动容器

```shell
export PathToPlatONE=/home/gexin/PlatONE-Go
docker run -itd -p 6791:6791 16791:16791 26791:26791 -v ${PathToPlatONE}:/PlatONE-Go --name platone platone:dev /bin/bash
```

## 进入容器


```shell
docker exec -it platone /bin/bash
```

## 编译或者搭链

```
cd /PlatONE-Go/
make clean & make all

cd release/linux/script/
./platonectl.sh one
```