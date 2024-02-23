#!/usr/bin/env bash

mkdir -p public

bazel --bazelrc=.bazelrc --bazelrc=.ci.bazelrc \
  build //:website

tar xvf bazel-bin/website.tar \
  -C ./public \
  --wildcards bzlparty_website/dist/** \
  --strip-components=2
