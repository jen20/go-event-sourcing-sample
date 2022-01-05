# To test against an event store DB instance 

docker run -it -p 2113:2113 -p 1113:1113 \ 
    eventstore/eventstore:latest --insecure --run-projections=All \
    --enable-external-tcp --enable-atom-pub-over-http --mem-db
