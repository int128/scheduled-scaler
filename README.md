# scheduled-scaler [![CircleCI](https://circleci.com/gh/int128/scheduled-scaler/tree/master.svg?style=shield)](https://circleci.com/gh/int128/scheduled-scaler/tree/master) [![Docker Repository on Quay](https://quay.io/repository/int128/scheduled-scaler/status "Docker Repository on Quay")](https://quay.io/repository/int128/scheduled-scaler)

This is a Kubernetes operator for scheduled scaling of deployments.

**Status:** Alpha. Specification may change.


## Getting Started

### Deploy

You can deploy the resources from [`deploy/scheduled-scaler.yaml`](deploy/scheduled-scaler.yaml).

```sh
kubectl apply -f https://raw.githubusercontent.com/int128/scheduled-scaler/master/deploy/scheduled-scaler.yaml
```


### Create a scaler

This tutorial shows how you can schedule a scaling of a deployment.

Deploy an echoserver by applying [echoserver.yaml](config/samples/echoserver.yaml).

```sh
kubectl apply -f echoserver.yaml
```

Create `echoserver-daytime.yaml` with the following content.

```yaml
apiVersion: scheduledscaling.int128.github.io/v1
kind: ScheduledPodScaler
metadata:
  name: echoserver-daytime
spec:
  scaleTarget:
    selectors:
      app: echoserver
  schedule:
    - daily:
        startTime: 21:00:00
        endTime: 07:00:00
      timezone: Asia/Tokyo
      spec:
        replicas: 0
  default:
    replicas: 1
```

Note that the timestamps are in UTC.

Apply the resource.

```sh
kubectl apply -f echoserver-daytime.yaml
```

Make sure the replicas of the deployment is the desired state.

```sh
kubectl -n echoserver get deployment
```


## Development

```sh
# Deploy the image to the kind cluster
make docker-load-kind
make deploy

# Undeploy from the kind cluster
make undeploy
```


## Contributions

This is an open source software.
Feel free to open issues and pull requests.
