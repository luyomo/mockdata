# Download the binary to linux
```
admin@workstation:/tmp/mockdata$ wget https://github.com/luyomo/mockdata/releases/download/v0.0.3/mockoracle-0.0.3.x86_64-linux.tar.gz
admin@workstation:/tmp/mockdata$ tar xvf mockoracle-0.0.3.x86_64-linux.tar.gz
```

# Oracle lib preparation
```
admin@workstation/tmp/mockdata$ wget https://download.oracle.com/otn_software/linux/instantclient/214000/instantclient-basic-linux.x64-21.4.0.0.0dbru.zip
admin@workstation/tmp/mockdata$ sudo unzip -d /opt/oracle instantclient-basic-linux.x64-21.4.0.0.0dbru.zip
admin@workstation/tmp/mockdata$ export LD_LIBRARY_PATH=/opt/oracle/instantclient_21_4:$LD_LIBRARY_PATH
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

# TODO
## Check table whether has foreign key
Add reference key into config
  table name | 
```
SELECT a.table_name, a.column_name, a.constraint_name, c.owner, 
       c.r_owner, c_pk.table_name r_table_name, c_pk.constraint_name r_pk
  FROM all_cons_columns a
  JOIN all_constraints c 
    ON a.owner = c.owner
   AND a.constraint_name = c.constraint_name
  JOIN all_constraints c_pk
    ON c.r_owner = c_pk.owner
   AND c.r_constraint_name = c_pk.constraint_name
 WHERE c.constraint_type = 'R'
   AND a.owner = 'ADMIN'
   AND a.table_name = 'ADMIN_USER'
```
