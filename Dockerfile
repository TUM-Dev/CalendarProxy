FROM composer:2.3.5 as composer

WORKDIR /app
COPY ./composer.json /app
COPY ./composer.lock /app
COPY ./src /app

RUN composer install --no-dev

FROM webdevops/php-nginx:8.1-alpine
COPY --from=composer /app /app