# LazyLogger 

Lazylogger is a small app to watch log files from different hosts in one place. Using the TUI, it is very easy to switch between log files.
You can even split, horizontaly or verticaly, the current window and add more logs on the same page.

![Gif](/docs/demo.gif)

## Configuration

The configuration is based on a YAML file with the following structure:

```yaml
defaultChunkSize: 40096
services:
    - 
        name: admin-local 
        host:
            address: 192.168.1.1
            username: foo 
            password: bar 
        file: /home/foo/file-to-watch.log 
    - 
        name: tomcat 
        jumpHost:
            address: aws-jump-host-address 
            username: ec2-user
            key: /home/foo/.aws/key.pem
        host:
            address: 172.1.1.1
            username: ec2-user
            key: /home/foo/.aws/key.pem
        file: /home/ec2-user/apache-tomcat-9.0.30/logs/catalina.out 
```
Each entry in `services` represent a log service. 

Credentials for ssh are set in `host` node. You can use password or key to connect to ssh. 
> Lazylogger will connect to port 22 only. 

For cases when a jump host is required (e.g. `aws`), you can add a `jumpHost` with the same structre as `host`.


To use a configuration file, the following command must be executed:

`lazylogger --config config.yml`

> Lazylogger will try to connect only when a new logger is created.

