# gae_snapshot_disk
Snapshot GCE Instance Disks with GAE Cron Job and Delete Expired Snapashots

Instead of deploy this gae application, now you can [Creating scheduled snapshots for persistent disk](Creating scheduled snapshots for persistent disk) on GCP Cloud Console.

## Get Source Code
```shell
$ git clone git@github.com:CloudMile/gae_snapshot_disk.git
```

## Install gcloud SDK
Follow [here](https://cloud.google.com/sdk/install) to install gcloud SDK.

## Creat/Use a GAE Project on GCP
Follow [here](https://console.cloud.google.com/projectselector/appengine/create?lang=go&st=true) to select/create a GCP project.

## Setup
### A Easy Way
```shell
$ make set PROJECT=<YOUR_PROJECT_ID>
```
where `<YOUR_PROJECT_ID>` is your project id

```shell
$ make yaml SERVICE=<YOUR_SERVICE_NAME>
```
where `<YOUR_SERVICE_NAME>` is sericve name you want, HOWEVER, the first GAE service name MUST be `default`

### More Info
Go to the project ditrctory.
```shell
$ cd ./gae_snapshot_disk
```

You have to setup below configurations first (in folder `app`):
- `app.yaml`: main application deployment setting
- `cron.yaml`: cron job setting
- `queue.yaml`: job queue setting

Edit `app.yaml`

```shell
$ vim ./app/app.yaml
```

```yaml
service: snapshot
env_variables:
  PROJECT_ID: "<YOUR_PROJECT_ID>" # only work for localhost test
  DAYS_AGO: '7' # delete snapshots 7 days ago
  STORAGE_LOCATION: 'none'
```
- service, if this is your first GAE service, please replcase `snapshot` to `default`
- PROJECT_ID, GCP project, it's only work on local
- DAYS_AGO, how many days ago, the snapshots will be deleted.
- STORAGE_LOCATION, Storage location, either regional (`asia-east1`) or multi-regional (`asia`), where snapshot content is to be stored. If set `none`, a nearby regional or multi-regional location is chosen automatically. Also if you don't want to store out-side from TW, please change to `asia-east1`

Edit `cron.yaml`

```shell
$ vim ./app/cron.yaml
```

```yaml
target: snapshot
```
- target, if this is your first GAE service, please change replace `snapshot` to `default`

You can add more cron case you need.
You can add query string to filter (`?filter=labels.<LABEL_KEY>%3A<LABEL_VALUE>`) for disk which you do want to snapshot. also the disk which doesn't set label will not be snapshoted.

For the disk which set the label `expired_can_delete:yes`; this service will delete the snapshot after days you set `DAYS_AGO`.

Edit `queue.yaml`
```shell
$ vim ./app/queue.yaml
```

```yaml
target: snapshot
```
- target, if this is your first GAE service, please change replace `snapshot` to `default`

## Deploy

```shell
$ gcloud config set project <YOUR_PROJECT_ID>
```

Replace `<YOUR_PROJECT_ID>` to your project id

```shell
$ gcloud app deploy app/app.yaml app/cron.yaml app/queue.yaml
```

## Test
Go to GCP console GCE disk page, you can add lables for disk which you do want to snapshot.
![image](step/step1.png)

Go to GAE page -> Task queues, click the `run now` for test.
![image](step/step2.png)

GO to GCE snapshot page, the snapshot will be created.
![image](step/step3.png)
