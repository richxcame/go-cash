package auth

var sUserSql = `
	select username, password from users where username = $1
`
