### climgr

climgr用于管理平时常用/实用，但是又容易忘记的一些资源信息（命令行工具使用模板，服务器ip、用户名密码等）. climgr提供命令行交互式操作界面，
方便进行资源选取和预览，并且支持设置选中动作（打印/执行。。）

#### Installation
```bash
go install github.com/jiandahao/climgr
```

#### Easy to use

![Alt Text](./assets/example.gif)

默认情况下，配置文件存放在`~/.climgr/config.d/`， 可以将自定义配置放置到该目录下，也可以通过环境变量`CLIMGR_CONFIG_PATH`指定存放目录。


资源配置格式如下：
```yaml
- name: 网络抓包 # <资源名称>
  description: "使用tcpdump等工具完成网络抓包" # <资源秒速>
  children: # 定义子资源
    - name: "tcpdump:查看请求详细数据包" # <子资源名称>
      description: "tcpdump -i eth0 -n -nn host 10.xx.xx.35" # <子资源描述>
      selectedAction: print # 选中后执行的动作，print - 直接终端打印，run - 终端运行
      command: "tcpdump -i eth0 -n -nn host 10.xx.xx.35" # 命令，当selectedAction为run时，选中后将执行这边定义的命令
```
配置文件需要命名为以`.yml`或`yaml`为后缀。更多例子见：[resource.yaml](./_config.d/resource.yaml)

TODO:
- 支持add命令，实现通过命令行交互方式增加资源配置
- 完善search功能，实现快速查找需要的资源
- 选择条目，如果存在命令，则支持是否拷贝到剪贴板