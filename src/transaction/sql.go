package transaction

var slContSumCshsSql = `
	SELECT contact, SUM(amount) from cashes WHERE detail = $1 GROUP BY contact
`
