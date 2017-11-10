FROM checkr/flagr-ci

VOLUME ["/data"]

WORKDIR /go/src/github.com/checkr/flagr
ADD . .
ADD ./buildscripts/demo_sqlite3.db /data/demo_sqlite3.db

ENV FLAGR_DB_DBCONNECTIONSTR=/data/demo_sqlite3.db
ENV FLAGR_RECORDER_ENABLED=false
ENV HOST=0.0.0.0
ENV PORT=18000

RUN cd ./browser/flagr-ui/ && yarn install && yarn run build
RUN make build

EXPOSE 18000
CMD ./flagr
