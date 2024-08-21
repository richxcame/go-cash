package cash

var slClsSql = `
	SELECT name FROM clients WHERE api_key = $1
`

var insCshSql = `
	INSERT INTO cashes (uuid, created_at, updated_at, client, contact, amount, detail, note)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`

var slSumCshsSql = `
	SELECT c.contact, c.detail, SUM(c.amount), MAX(created_at) FROM cashes c
`

var slCntCshsSql = `
	SELECT COUNT(*) FROM (SELECT DISTINCT detail FROM cashes 
`
var slCshsSql = `
	SELECT uuid, created_at, client, contact, amount, detail, note FROM cashes where detail = $1
`

var slDatesCashSql = `
	select
        sum(amount)
    from cashes c
    where c.created_at > $1 and c.created_at < $2 ;
`

var slDatesTolegSql = `
	with tr as (
		select 
			'toleg' as service,
			sum(request_amount::int/100.0),
			reason
		from transactions
		where reason != '' 
			and created_at > $1 
			and created_at < date_add($2, '1 day')
			and status = 'SUCCESS' 
		group by reason
	)
		select 
			json_agg(
				json_build_object(
					'reason', reason,
					'sum', sum
				)
			) as reasons,
			sum(sum) as total_sum
		from tr
		group by tr.service;
`

// select
// 	sum(request_amount::int/100.0)
// from transactions
// where reason != '' and created_at > $1 and created_at < date_add($2, '1 day') and status = 'SUCCESS';
