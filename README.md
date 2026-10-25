# sso
docker build -t sso:latest .

docker volume create sso_storage

docker run --rm --entrypoint ./migrator -v sso_storage:/app/storage sso:latest --storage-path=/app/storage/sso.db --migrations-path=/app/migrations

docker run -d --name sso-app -p 44044:44044 -p 8080:8080 -v sso_storage:/app/storage sso:latest