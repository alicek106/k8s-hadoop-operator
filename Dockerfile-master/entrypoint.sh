#!/bin/bash
for (( i=0; i<=$SLAVES_COUNT; i++ ))
do
  echo "$SLAVES_SS_NAME-$i.$SLAVES_SVC_NAME.$NAMESPACE.svc.cluster.local" >> /hadoop/etc/hadoop/slaves
done

for (( i=0; i<=$SLAVES_COUNT; i++ ))
do
	while :
	do
		ip=$(dig +short "$SLAVES_SS_NAME-$i.$SLAVES_SVC_NAME.$NAMESPACE.svc.cluster.local")
		if [ "$ip" != "" ]; then
			echo "$ip $SLAVES_SS_NAME-$i.$SLAVES_SVC_NAME.$NAMESPACE.svc.cluster.local" >> /etc/hosts
			echo "$SLAVES_SS_NAME-$i.$SLAVES_SVC_NAME.$NAMESPACE.svc.cluster.local is running"
			break
		else
			echo "$SLAVES_SS_NAME-$i.$SLAVES_SVC_NAME.$NAMESPACE.svc.cluster.local is not running"
			sleep 1
		fi
	done
done

sed -i -e "s/<MASTER_ENDPOINT>/$MASTER_ENDPOINT/g" /hadoop/etc/hadoop/yarn-site.xml
sed -i -e "s/master/$MASTER_ENDPOINT/g" /hadoop/etc/hadoop/core-site.xml
sed -i -e "s/master/$MASTER_ENDPOINT/g" /hadoop/etc/hadoop/mapred-site.xml

sed -i -e "s/without-password/yes/g" /etc/ssh/sshd_config
password=$(cat /etc/rootpwd/password)
echo "root:$password" | chpasswd

/usr/sbin/sshd

start-all.sh

tail -f /dev/null
