# Download the binary to linux
```
admin@workstation:/tmp/mockdata$ wget https://github.com/luyomo/mockdata/releases/download/v0.0.3/mockoracle-0.0.3.x86_64-linux.tar.gz
admin@workstation:/tmp/mockdata$ tar xvf mockoracle-0.0.3.x86_64-linux.tar.gz
```

# Oracle lib preparation
```
admin@workstation/tmp/mockdata$ wget https://download.oracle.com/otn_software/linux/instantclient/219000/instantclient-tools-linux.x64-21.9.0.0.0dbru.zip
admin@workstation/tmp/mockdata$ sudo unzip -d /opt/oracle instantclient-tools-linux.x64-21.9.0.0.0dbru.zip
admin@workstation/tmp/mockdata$ export PATH=/opt/oracle/instantclient_21_9:$PATH
admin@workstation/tmp/mockdata$ export LD_LIBRARY_PATH=/opt/oracle/instantclient_21_9:$LD_LIBRARY_PATH
```

# Config file preparation
```
admin@workstation:/tmp/mockdata$ more /tmp/mockdata/config.toml
[oracle]
username = "user"
password = "password"
host = "oracle host name"
port = 1521
service-name = "dev"
```

# Data generation from mockoracle
```
admin@workstation:/tmp/mockdata$ ./bin/mockoracle --config=/tmp/mockdata/config.toml --tables=admin.table01,admin.table02 --num-of-rows=20
20 rows have been instered into table(admin.table01) 
20 rows have been instered into table(admin.table02) 
```

# Others
* Tool does not support foreign key. If oracle table include foreign key, please remove it before test.
* Tool does not support primary key's uniqueness. The data is generated randomly, if it fails, please re-run it.
* The timestamp is not garanteed. Currently it only supports 23-JAN-23 10:11:23.000000 PM
