#!/bin/bash
echo Wait for servers to be up
sleep 10

HOSTPARAMS="--host=roach --insecure"
SQL="cockroach sql $HOSTPARAMS"

$SQL -e "CREATE DATABASE wfmt;"
$SQL -d wfmt -e "CREATE TABLE articles(lock_id int8);"