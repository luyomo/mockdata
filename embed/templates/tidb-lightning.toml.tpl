[lightning]
level = "info"
file = "tidb-lightning.log"

[tikv-importer]
backend = "local"
incremental-import = true
sorted-kv-dir = "/tmp/sorted-kv-dir"

[mydumper]
data-source-dir = "{{ .DataFolder }}"

filter = ['*.*', '!mysql.*', '!sys.*', '!INFORMATION_SCHEMA.*', '!PERFORMANCE_SCHEMA.*', '!METRICS_SCHEMA.*', '!INSPECTION_SCHEMA.*']

[mydumper.csv]
separator = ','
delimiter = '"'
terminator = ''
header = false
not-null = false
null = '\N'
backslash-escape = true
trim-last-separator = false
 
[tidb]
host = "{{ .TiDBHost }}"
port = {{ .TiDBPort }}
user = "{{ .TiDBUser }}"
password = "{{ .TiDBPassword }}"
status-port = 10080
pd-addr = "{{ .PDIP }}:2379"

[post-restore]
checksum = "off"
analyze = "off"
