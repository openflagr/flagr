#!/usr/bin/env bash
set -e
git checkout master
git pull
f_tag=$(git describe --tags)
git checkout $f_tag
docker build -t checkr/flagr:latest .
docker tag checkr/flagr:latest checkr/flagr:$f_tag
docker tag checkr/flagr:latest registry.heroku.com/try-flagr/web && docker push registry.heroku.com/try-flagr/web && heroku container:release -a try-flagr web
docker push checkr/flagr:$f_tag
docker push checkr/flagr:latest
git checkout master
