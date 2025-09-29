db_login:
	psql ${DATABASE_URL}

db_create_migration:
	migrate create -ext sql -dir migrations -seq $(name)
# 	name=init_schema make db_create_migration