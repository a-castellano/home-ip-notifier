#!/bin/sh

echo "### NOT starting on installation, please execute the following statements to configure windmaker-home-ip-notifier to start automatically using systemd"
echo "### Check /etc/default/windmaker-home-ip-notifier and make required changes"
echo " sudo /bin/systemctl daemon-reload"
echo "### Enable service with the following command"
echo " sudo /bin/systemctl enable windmaker-home-ip-notifier.service"
