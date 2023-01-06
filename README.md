# watchfiletail

This small Go app that just watches one folder for specific files and `tail` them on the standard output

Use Case: run it as a sidecar in a Kubernetes Pod, to watch for the main container's logs and stream them so you can access them with `kubectl logs <your pod> -c <watchfiletail container name>`


## Usage

Here is an application that is logging some informations in a log file, here `/var/log/logging-app/logging-app.log`:

```yaml
---
apiVersion: v1
kind: Pod
metadata:
  name: logging-app-pod-host-log
spec:
  containers:
  - name: logging-app
    image: alpine:latest
    command: ["/bin/sh", "-c"]
    args:
      - |
        echo "starting"
        while true
        do
          echo $(date) >> /var/log/logging-app/logging-app.log
          sleep 5
        done
        echo "done"
    volumeMounts:
      - name: logging-app-log-dir
        mountPath: /var/log/logging-app
  volumes:
    - name: logging-app-log-dir
      emptyDir: {}
  terminationGracePeriodSeconds: 1
```

When using `kubectl logs logging-app-pod-host-log`, or better, [Stern](https://github.com/stern/stern), you only see a `starting` log message.

The logStreamer is here to `tail -f` the log files and print them on `stdout` so Kubernetes can grab them.

You can add the log streamer like so:

```yaml
---
apiVersion: v1
kind: Pod
metadata:
  name: logging-app-pod-host-log
spec:
  containers:
    - name: logging-app
      image: alpine:latest
      command: ["/bin/sh", "-c"]
      args:
        - |
          echo "starting"
          while true
          do
            echo $(date) >> /var/log/logging-app/logging-app.log
            sleep 5
          done
          echo "done"
      volumeMounts:
        - name: logging-app-log-dir
          mountPath: /var/log/logging-app
    - name: logstreamer
      image: ghcr.io/prune998/watchfiletail:v0.0.3
      env:
        - name: FOLDERPATH
          value: "/var/log/logging-app"
        - name: LOGLEVEL
          value: info
        - name: FILEMATCH
          value: "^logging-app.*"
      resources:
        limits:
          cpu: "100m"
          memory: 128Mi
        requests:
          cpu: "100m"
          memory: 128Mi
      volumeMounts:
        - name: logging-app-log-dir
          mountPath: /var/log/logging-app
  volumes:
    - name: logging-app-log-dir
      emptyDir: {}
  terminationGracePeriodSeconds: 1
```

## TODO

Add a mecanisme to delete old log files...