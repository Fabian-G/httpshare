# httpshare

httpshare utilizes golangs file server implementation to allow sharing of files via http by spinning up an http server locally. 
You just have to make sure that there is no firewall between the server and the clients.

To share a file just run.

```bash
httpshare myFile.mp4
```

After that the file will be available at http://yourIp:8080/randomId. The resulting URL is printed for each file. The IP in the URL is the first non loopback adresse it finds on a local interface. Optionally httpshare can get your public IP Adresse for you if you are behind a NAT.

## Install

```bash
go get -u github.com/Fabian-G/httpshare
```

## Usage

```bash
httpshare [OPTION...] [FILE...]
```

Option | Description
-------|-------------
-i     | If the served content should be marked as inline content (Displayed directly in browser instead of opening a download dialog whenever possible).
-l n   | Limits the number of requests to *n* per file. Any subsequent request will receive an unauthorized error
-p port | Starts the server on port *port*
-e      | Enables TLS Encryption by generating a self signed certificate on startup. The SHA-1 sum of that certificate is printed for sharing with the clients.
-r      | If set the generated URLs will contain your public IP Addresse. For that another server will be queried.
