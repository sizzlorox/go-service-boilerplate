mongo -- $SERVICE_NAME <<EOF
	use admin
	db.createUser({
		user: '$DB_USERNAME',
		pwd: '$DB_PWD',
		roles: [
			{
				role: 'readWrite',
				db: '$SERVICE_NAME',
			}
		]
	})
EOF
