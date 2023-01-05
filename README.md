# watchfiletail

This small Go app that just watches one folder for specific files and `tail` them on the standard output

Use Case: run it as a sidecar in a Kubernetes Pod, to watch for the main container's logs and stream them so you can access them with `kubectl logs <your pod> -c <watchfiletail container name>`
