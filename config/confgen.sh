#!/bin/bash

confs=(
     #插件相关的信息
    '%PLUGIN_IP%%=127.0.0.1'
    #插件同步的ssh端口，一般为22号端口
    '%PLUGIN_PORT%%=22'
    #插件同步的ssh用户
    '%%PLUGIN_USER%%=root'
    #插件同步的ssh密码,如果配置为key的模式,密码可以为空
    '%%PLUGIN_PASSWD%%=root'
    # 插件的同步源地址
    '%%PLUGIN_PATH%%=/root'
    #插件的ssh同步私有key，如果用key回话，则必须要填写
    '%%SSH_PRIVATEKEY%%=' #/home/user/.ssh/rsa
    #监听端口，注意“0.0.0.0”
    #客户端监听ip地址，这个端口支持第三方数据，push到agent_http端口上
    '%%AGENT_HTTP%%=0.0.0.0:1988'
    #集群合并信息的http的地址
    '%%AGGREGATOR_HTTP%%=0.0.0.0:6055'
    #图的http地址
    '%%GRAPH_HTTP%%=0.0.0.0:6071'
    #图的rpc地址
    '%%GRAPH_RPC%%=0.0.0.0:6070'
    #心跳服务的http地址
    '%%HBS_HTTP%%=0.0.0.0:6031'
    #心跳的rpc服务地址
    '%%HBS_RPC%%=0.0.0.0:6030'
    #告警判断服务的http地址
    '%%JUDGE_HTTP%%=0.0.0.0:6081'
    #告警判断服务的rpc地址
    '%%JUDGE_RPC%%=0.0.0.0:6080'
    #空数据监控的http端口
    '%%NODATA_HTTP%%=0.0.0.0:6090'
    #发送组件的http地址
    '%%TRANSFER_HTTP%%=0.0.0.0:6060'
    #发送组件的rpc地址
    '%%TRANSFER_RPC%%=0.0.0.0:8433'

    #API接口的http地址
    '%%PLUS_API_HTTP%%=0.0.0.0:8080'
    #告警组件的http地址
    '%%ALARM_HTTP%%=0.0.0.0:9912'
    #配置转发的gateway地址

    '%%GATEWAY_HTTP%%=0.0.0.0:16060'
    '%%GATEWAY_RPC%%=0.0.0.0:18433'
    '%%GATEWAY_SOCKET%%=0.0.0.0:14444'
    #网络测速端口
    '%%NET_SPEED_PORT%%=10009'
    #配置地址类
    '%%REDIS%%=127.0.0.1:6379'
    '%%TRANSFER_IP%%=127.0.0.1'
    '%%MYSQL%%=root:@tcp(127.0.0.1:3306)'
    '%%PLUS_API_DEFAULT_TOKEN%%=default-token-used-in-server-side'
    '%%API_ADDR%%=127.0.0.1:8080'
    '%%DASHBOARD_HTTP%%=127.0.0.1:8081'

    #图的地址列表用逗号分隔，json的list格式，注意双引号
    '%%GRAPH_ADDRS%%=127.0.0.1:6070'

    #如果有多台agent，则有多个地址，建议只有一个agent提供服务，如果有多个agent，则建议根据资源，带宽，选择填写一台
    '%%PUSH_ADDR%%=127.0.0.1:1988'

    #tsdb地址，如果有发送到tsdb的需求，请修改transfer的配置文件，发送到tsdb，
     '%%TSDB_ADDR%%=127.0.0.1:8088'
 )

configurer() {
    for i in "${confs[@]}"
    do
        search="${i%%=*}"
        replace="${i##*=}"

        uname=`uname`
        if [ "$uname" == "Darwin" ] ; then
            # Note the "" and -e  after -i, needed in OS X
            find ./*.json -type f -exec sed -i .tpl -e "s/${search}/${replace}/g" {} \;
        else
            find ./*.json -type f -exec sed -i "s/${search}/${replace}/g" {} \;
        fi
    done
}
configurer
