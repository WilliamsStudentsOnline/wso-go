# Host setup for systemd jobs

Schedules: `jobs/registry`, applied on deploy via `wso-jobctl gen-units`.

## One-time

```bash
sudo cp prod_files/polkit/50-wso-jobs.rules /etc/polkit-1/rules.d/
sudo cp prod_files/journald/wso-jobs.conf /etc/systemd/journald.conf.d/
sudo cp prod_files/tmpfiles.d/wso-jobs.conf /etc/tmpfiles.d/
sudo cp prod_files/systemd/wso-jobs-dispatch.* /etc/systemd/system/
sudo systemd-tmpfiles --create /etc/tmpfiles.d/wso-jobs.conf
sudo usermod -aG systemd-journal wso
sudo systemctl restart polkit systemd-journald
sudo systemctl daemon-reload
sudo systemctl enable --now wso-jobs-dispatch.timer
```

`/opt/wso-go/config.yaml` and job-specific files (e.g. `vendor_info.yaml`) must exist on the host.

## Retire crontab

After timers are enabled, remove old WSO cron lines (legacy paths under `/home/wso/wso/wso-backend/jobs/`):

```bash
crontab -l
crontab -e
```

Do not run the same job from cron and systemd.
