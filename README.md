# TUM Calendar Proxy

This is a proxy service that simplifies and enhances the iCal export from TUM Online. It allows you to:

- Shorten long lesson names, such as 'Grundlagen Betriebssysteme und Systemsoftware' → 'GBS'
- Add locations that are recognized by Google / Apple Maps
- Filter out unwanted events, such as cancelled, duplicate or optional ones

You can use the proxy service by visiting <https://cal.tum.app/> and following the instructions there.

## Development
If you want to run the proxy service locally or contribute to the project, you will need:

- Go 1.19 or higher
- Docker (optional)

To run the service locally, follow these steps:

- Clone this repository: `git clone https://github.com/tum-calendar-proxy/tum-calendar-proxy.git`
- Navigate to the project directory: `cd tum-calendar-proxy`
- Run the proxy server: `go run cmd/proxy/proxy.go`
- The service will be available at <http://localhost:8081>

To build a production image using Docker, follow these steps:

- Build the image: `docker build -t tumcalproxy .`
- Run the container: `docker run -p 8081:8081 tumcalproxy`
- The service will be available at <http://localhost:8081>
