#!/bin/bash

LOCAL=$(pwd)
docker run --rm -it -v $LOCAL:/app composer install
docker run --rm -e WEB_DOCUMENT_ROOT=/app/public -p 8081:80 -v $LOCAL:/app -v $LOCAL/src:/app/public webdevops/php-nginx:8.2-alpine