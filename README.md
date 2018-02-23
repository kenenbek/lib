To start simulation:
* Set environment variable:

```
PacketSize=65e3
ServerAmount=4
```
Where `ServerAmount` is amount of working servers.
* Run in a shell:

```
go run main.go topology/platform.xml topology/deployment.xml
```

To try out anomaly situations:
```
PacketSize=65e3
ServerAmount=1 or (2, 3)
```
Run in a shell:
```
go run main.go topology/platform.xml topology/anomalies/deployment${ServerAmount}.xml
```