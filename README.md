# dp-docker-stats

#### 启动方式
docker run -it -d -v /var/run/docker.sock:/var/run/docker.sock \
             -e "INFLUX_HOST=http://172.17.0.1:8086" \
             -e "INFLUX_DBNAME=dp_docker_stats" \
             -e "INFLUX_USERNAME=root" \
             -e "INFLUX_PASSWORD=123456" \
             -e "INFLUX_TABLE_SUFFIX=mx-dev" \
             dp-docker-stats
             
             
#### 环境变量
- *INFLUX_HOST:  influxdb 地址 http://ip:port
- *INFLUX_DBNAME:  influxdb database name,必须是已经存在的
- *INFLUX_USERNAME:  influxdb 用户名
- *INFLUX_PASSWORD:  influxdb 密码
- INFLUX_TABLE_SUFFIX: influxdb 的后缀,未设置为:dp-docker-stats  
``` * 为必须选项 ```
