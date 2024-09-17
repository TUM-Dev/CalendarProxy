# TUM Calendar Proxy

![Alt text](image.png)

This is a proxy service that simplifies and enhances the iCal export from TUM Online. It allows you to:

- Shorten long lesson names, such as 'Grundlagen Betriebssysteme und Systemsoftware' → 'GBS'
- Add locations that are recognized by Google / Apple Maps
- Filter out unwanted events, such as cancelled, duplicate or optional ones

You can use the proxy service by visiting <https://cal.tum.app/> and following the instructions there.

## Development
If you want to run the proxy service locally or contribute to the project, you will need:

- Go 1.22 or higher
- Docker (optional)

To run the service locally, follow these steps:

- Clone this repository
  ```sh
  git clone https://github.com/tum-calendar-proxy/tum-calendar-proxy.git
  ```
- Navigate to the project directory: 
  ```sh
  cd tum-calendar-proxy
  ```
- Run the proxy server:
  ```sh
  go run cmd/proxy/proxy.go
  ```
- The service will be available at <http://localhost:4321>

To build an image using Docker, follow these steps:

- ```sh
  docker compose -f docker-compose.local.yaml up --build
  ```
- The service will be available at <http://localhost:4321>
