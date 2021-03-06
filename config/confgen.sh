#!/bin/bash
#特别说明，上部分为监听端口，如果sever主机部署，请修改对应的下面的addr的信息，
# 例如tarsfer的监听为8433，则下面地址要对应修改
# ssh 的原路径，目标路径，需要为非“连接的路径，即软连接（ln）”，这点很重要，不然不能同步
confs=(
     #插件相关的信息
    '%%PLUGIN_IP%%=127.0.0.1'
    #插件同步的ssh端口，一般为22号端口
    '%%PLUGIN_PORT%%=22'
    #插件同步的ssh用户
    '%%PLUGIN_USER%%=root'
    #插件同步的ssh密码,如果配置为key的模式,密码可以为空
    '%%PLUGIN_PASSWD%%=root'
    # 插件的同步源地址
    '%%PLUGIN_PATH%%=/root'
    #插件的ssh同步私有key，如果用key回话，则必须要填写
    '%%SSH_PRIVATEKEY%%=' #/home/user/.ssh/rsa
    #监听端口，注意“0.0.0.0”
    #客户端监听ip地址，这个端口支持第三方数据，push到agent_http端口上,这个地址在前台写死的，建议不要修改，
    # 如果要修改，就需要修改前台的端口号
    '%%AGENT_HTTP%%=0.0.0.0:1988'
    #集群合并信息的http的监听地址
    '%%AGGREGATOR_HTTP%%=0.0.0.0:6055'
    #图的http监听地址
    '%%GRAPH_HTTP%%=0.0.0.0:6071'
    #图的rpc监听地址
    '%%GRAPH_RPC%%=0.0.0.0:6070'
    #心跳服务的http监听地址
    '%%HBS_HTTP%%=0.0.0.0:6031'
    #心跳的rpc服务监听地址
    '%%HBS_RPC%%=0.0.0.0:6030'
    #告警判断服务的http监听地址
    '%%JUDGE_HTTP%%=0.0.0.0:6081'
    #告警判断服务的rpc监听地址
    '%%JUDGE_RPC%%=0.0.0.0:6080'
    #空数据监控的http监听端口
    '%%NODATA_HTTP%%=0.0.0.0:6090'
    #发送组件的http监听地址
    '%%TRANSFER_HTTP%%=0.0.0.0:6060'
    #发送组件的rpc监听地址
    '%%TRANSFER_RPC%%=0.0.0.0:8433'

    #API接口的http监听地址
    '%%PLUS_API_HTTP%%=0.0.0.0:8080'
    #告警组件的http监听地址
    '%%ALARM_HTTP%%=0.0.0.0:9912'
    #配置转发的gateway地址

    '%%GATEWAY_HTTP%%=0.0.0.0:16060'
    '%%GATEWAY_RPC%%=0.0.0.0:18433'
    '%%GATEWAY_SOCKET%%=0.0.0.0:14444'
    #网络测速端口
    '%%NET_SPEED_PORT%%=10009'
    #配置地址类
    '%%HBS_ADDR%%=127.0.0.1:6030'
    '%%REDIS%%=127.0.0.1:6379'
    '%%TRANSFER_ADDR%%=127.0.0.1:8433'
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
            find ./*.json -type f -exec sed -i "s#${search}#${replace}#g" {} \;
        fi
    done
}
init_db(){
  #数据库初始化
  # Falcon+
  mysql_cmd=`which mysql`
  if [ ${#mysql_cmd}  -lt 3 ]
  then
     echo "没有发现mysql命令，请手动执行数据库初始化"
     return
   else
    ip="127.0.0.1"
    port=3306
    user="root"
    passwd=""

    for i in "${confs[@]}"
     do
        search="${i%%=*}"
        replace="${i##*=}"
        if [ ${search} == '%%MYSQL%%' ]
        then
            echo "mysql 信息如下: "+${replace}
            user=`echo ${replace}|awk -F':' '{print $1}'`
            passwd=`echo ${replace}|awk -F'@' '{print $1}'|awk -F':' '{print $2}'`
            ip=`echo ${replace}|awk -F'@' '{print $2}'|sed -e 's/tcp(//' -e 's/)//' |awk -F':' '{print $1}'`
            port=`echo ${replace}|awk -F'@' '{print $2}'|sed -e 's/tcp(//' -e 's/)//' |awk -F':' '{print $2}'`
            break
         else
            continue
        fi
     done
      mysql -h ${ip} -u ${user} -P ${port} -p ${passwd} < ../mysq/db_schema/1_uic-db-schema.sql
      mysql -h ${ip} -u ${user} -P ${port} -p ${passwd} < ../mysq/db_schema/2_portal-db-schema.sql
      mysql -h ${ip} -u ${user} -P ${port} -p ${passwd} < ../mysq/db_schema/3_dashboard-db-schema.sql
      mysql -h ${ip} -u ${user} -P ${port} -p ${passwd} < ../mysq/db_schema/4_graph-db-schema.sql
      mysql -h ${ip} -u ${user} -P ${port} -p ${passwd} < ../mysq/db_schema/5_alarms-db-schema.sql
  fi
}
configurer
init_db