TumCalProxy [![Build Status](https://travis-ci.org/TUM-Dev/CalendarProxy.svg?branch=master)](https://travis-ci.org/TUM-Dev/CalendarProxy)
===========

Proxy for the TUM iCal export to remove clutter from it and optimize the output

You can find more information on the about page: https://cal.bruck.me/

To run locally using docker:
 - run `docker run --rm -it -v /ABSOLUTE/PATH/TO/PROJECT/:/app composer install` to install dependencies
 - run `docker run --rm -it -v /ABSOLUTE/PATH/TO/PROJECT/:/app composer /app/vendor/bin/phpunit` to execute tests
 - run `docker run --rm -p 80:80 -v /ABSOLUTE/PATH/TO/PROJECT/:/app webdevops/php-nginx:alpine` to run the webserver on port 80
