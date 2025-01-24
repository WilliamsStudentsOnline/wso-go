#!/bin/bash -e
##### WSO-Backend WSO 2.0 #####
# Copy Grafana files to the final correct destinations.
# Use this only if you installed Grafana with an RPM/DEB file,
# otherwise use Grafana's command line switches.
if ! [ $(id -u) = 0 ]; then
   echo "The script need to be run as root." >&2
   exit 1
fi
if [ $SUDO_USER ]; then
    real_user=$SUDO_USER
else
    real_user=$(whoami)
fi
sudo -u $real_user echo "Copying files"
mv -v grafana.db /var/lib/grafana/grafana.db
mv -v grafana-config.ini /etc/grafana/grafana.ini
sudo -u $real_user echo "Files successfully copied"
