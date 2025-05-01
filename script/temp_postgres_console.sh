POSTGRES_ADMIN_PASSWORD=$(kubectl get secret --namespace postgresql postgresql -o jsonpath="{.data.postgres-password}" | base64 -d)
kubectl run postgresql-client --rm --tty -i --restart='Never' --namespace postgresql --image docker.io/bitnami/postgresql:17.4.0-debian-12-r17 --env="PGPASSWORD=$POSTGRES_ADMIN_PASSWORD" --command -- psql --host postgresql -U postgres -d tasklist -p 5432
