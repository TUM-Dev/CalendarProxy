#!/bin/bash

LOCAL=$(pwd)
docker run --rm -it -v $LOCAL:/app composer test