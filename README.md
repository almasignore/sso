# sso
docker build -t sso:latest .

docker volume create sso_storage

docker run --rm --entrypoint ./migrator -v sso_storage:/app/storage sso:latest --storage-path=./storage/sso.db --migrations-path=./migrations