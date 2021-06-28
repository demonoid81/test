Бекап Базы

docker run -v $(pwd)/data:/data cockroachdb/cockroach:v20.2.5 sql --host=192.168.10.244 --insecure  -e "BACKUP TO 's3://backup/db?AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE&AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY&AWS_REGION=us-west-1&AWS_ENDPOINT=http://192.168.10.244:9000'"


Воссановлние базы

docker run cockroachdb/cockroach:v20.2.5 sql --host=192.168.10.244:26258 --insecure  -e "RESTORE FROM 's3://backup/db?AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE&AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY&AWS_REGION=us-west-1&AWS_ENDPOINT=http://192.168.10.244:9000'"