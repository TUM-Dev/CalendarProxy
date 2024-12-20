# TUM Calendar Proxy

![Alt text](image.png)

This is a fork of the proxy service that simplifies and enhances the iCal export from TUM Online. It allows you to:

- Shorten long lesson names, such as 'Grundlagen Betriebssysteme und Systemsoftware' → 'GBS'
- Add locations that are recognized by Google / Apple Maps
- Filter out unwanted events, such as cancelled, duplicate or optional ones

Additionally, I've implemented filtering the calendar into Vorlesungen, Prüfungen and other. This way, it is possible to colour-code by category in calendar apps. To do this, add the query string filter, which can be "vo" for Vorlesungen, "pr" for Prüfungen, or "ot" for other (that is, everything that isn't a Vorlesung or Prüfung). The following formats should work:

```
.../?pStud=ABCDEF&pToken=XYZ             <-- calendar with everything
.../?pStud=ABCDEF&pToken=XYZ&filter=vo   <-- only Vorlesungen
.../?pStud=ABCDEF&pToken=XYZ&filter=pr   <-- only Prüfungen
.../?pStud=ABCDEF&pToken=XYZ&filter=ot   <-- everything ex. Vorlesungen/Prüfungen
```

Additional features:
 - Any trailing space-comma-space is removed.
 - location detection is reworked and is now based entirely off of location tags (for example, `(5508.02.801)`). It'll also get the location from the description, if necessary.
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
