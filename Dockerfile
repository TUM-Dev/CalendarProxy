FROM composer:1.6.3 as composer

WORKDIR /app
COPY ./composer.json /app
COPY ./composer.lock /app
COPY ./src /app

RUN composer install --no-dev

FROM webdevops/php-nginx:alpine
COPY --from=composer /app /app