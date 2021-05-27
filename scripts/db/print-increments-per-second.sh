#!/bin/bash

count(){
    PGPASSWORD=$POSTGRES_PASSWORD psql -h localhost -U "$POSTGRES_USER" -t -c "select count(*) from increments"
}

prev=$(count)
sleep 1
while true
do
    current=$(count)    
    echo -ne "$(($current-$prev)) \\r"
    prev=$current
    sleep 1
done

