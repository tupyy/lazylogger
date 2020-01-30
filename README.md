# OpenWis Logger


OpenWis Logger can log any file from remote hosts. It reacts to changes in size for growning logs like Apache, Tomcat or JBoss.

## Configuration

The configuration is based on a JSON file with the following structure:

```json
{
    "defaultchunksize": 4096,
    "services": [
            {
                "host" : "192.168.3.100:22",
                "username" :"cosmin",
                "password": "cosmin",
                "file": "/home/cosmin/apache-tomcat-8.5.46/logs/catalina.out"
            },
            {
                "host" : "192.168.3.100:22",
                "username" :"cosmin",
                "password": "cosmin",
                "file": "/home/cosmin/catalina.out"
            }
    ]
}
```
Each entry in `services` describe a configure a logger. 

To use a configuration file, the following command must be executed:

`openwislogger --config config.json.file`

>At start-up time, OpenWisLogger will try to connect and read the size for each file in the configuration file.
>If the host is unreachable or cannot connect to it or the filepath is incorrent, OpenWisLogger will ignore this service.


