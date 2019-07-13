#!/bin/bash
sed -i -e "s/<MASTER_ENDPOINT>/$MASTER_ENDPOINT/g" /hadoop/etc/hadoop/yarn-site.xml
sed -i -e "s/master/$MASTER_ENDPOINT/g" /hadoop/etc/hadoop/core-site.xml
sed -i -e "s/master/$MASTER_ENDPOINT/g" /hadoop/etc/hadoop/mapred-site.xml

/usr/sbin/sshd
tail -f /dev/null
