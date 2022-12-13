# Calendar-Proxy


Proxy for the TUM iCal export to remove clutter from it and optimize the output

You can find more information on the about page: https://cal.tum.sexy/

## Development
To run locally:
 - run `go run cmd/proxy/proxy.go` to install dependencies and run a local server for testing on port 8081

To build a production image:
- run `docker build -t tumcalproxy .`
- run `docker run -p 8081:8081 tumcalproxy`
- open http://127.0.0.1:8081 in your browser
