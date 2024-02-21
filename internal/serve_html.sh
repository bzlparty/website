#!/usr/bin/env bash

TAR_OPTS="xf"
HTTP_OPTS="--silent"

if [ "$3" == "-v" ]; then
  TAR_OPTS="xvf"
  HTTP_OPTS=
fi

tar "$TAR_OPTS" "$1"

http-server "$2" "$HTTP_OPTS"
