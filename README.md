# TumCalProxy [![Build Status](https://travis-ci.org/TUM-Dev/CalendarProxy.svg?branch=master)](https://travis-ci.org/TUM-Dev/CalendarProxy)


Proxy for the TUM iCal export to remove clutter from it and optimize the output

You can find more information on the about page: https://cal.tum.sexy/

## Development
To run locally using docker:
 - run `./runLocal.sh` to install dependencies and run a local server for testing on port 8081
 - run `./testLocal.sh` to execute tests

To build a production image:
- run `docker build -t tumcalproxy -f Dockerfile .`
- run `docker run -p 80:80 tumcalproxy`
- open http://127.0.0.1 in your browser