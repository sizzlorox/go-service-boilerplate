db.createUser({
	user: 'gs_user',
	pwd: 'gs_pwd',
	roles: [
		{
			role: 'readWrite',
			db: 'gs_service',
		}
	]
})
