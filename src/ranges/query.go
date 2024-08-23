package ranges

var sClientsQ = `
	SELECT name FROM clients WHERE api_key = $1
`

var iRangesQ = `
	INSERT INTO ranges (uuid, created_at, updated_at, client, detail, note)
		VALUES ($1, $2, $3, $4, $5, $6)
`

var sRangesByLimQ = `
	SELECT r.uuid, r.created_at, r.updated_at, r.client, r.detail, r.note FROM ranges r ORDER BY r.created_at DESC OFFSET $1 LIMIT $2;
`

var sRangeByClientQ = `
	SELECT r.created_at, r.client FROM ranges r  WHERE created_at < $1 AND client=$2 ORDER BY created_at DESC limit 1
`

var sSumCashesByClientQ = `
	SELECT SUM(amount) FROM cashes where created_at >= $1 AND created_at <= $2 AND client=$3
`

var sSumCashByGroupQ = `
	SELECT SUM(amount),COUNT(amount) FROM cashes where created_at >= $1 AND created_at <= $2 AND client=$3 AND amount = $4 GROUP BY amount
`

var sCountRangesQ = `
	SELECT COUNT(*) FROM ranges
`

var sAllRangesQ = `
	
WITH rp AS (
    SELECT 
        created_at, 
        client,
        ROW_NUMBER() OVER (PARTITION BY client ORDER BY created_at DESC) AS rn
    FROM ranges
    -- TODO if add condition on created_at, speed decreases half.
),lst as(
    SELECT 
        created_at,
        client
    FROM rp
    WHERE rn <= 2
    ORDER BY client, created_at DESC
), js as (
    select 
        json_agg(
            created_at
        ) dates,
        client
    from lst
    group by client    
), c as (
    select 
        amount,
        count(amount),
        js.client
    from js
    inner join cashes ch on ch.client = js.client
    where ch.created_at < (js.dates ->> 0)::timestamp and ch.created_at > (js.dates ->> 1)::timestamp
    group by js.client, amount
    order by amount desc
), rd as (
    select
        json_agg(
            json_build_object(
                'valuta', amount,
                'count', count
            )
        ) as amount,
        sum(amount*count) as summa,
        client
    from c
    group by client
) 
select 
    rd.amount,
    rd.summa,
    js.dates,
    js.client
from js
    inner join rd on rd.client = js.client;

`
