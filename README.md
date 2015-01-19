Go short!
=========
Shorten url service written in GO.

API Endpoints
-------------
### Create a short code [/api/v1/urls]
Shorten your input url.
```
curl -X POST -H "Content-Type: application/json" \
     -d '{"url": "http://yulun.me"}' \
     'http://localhost:3000/api/v1/urls'

	{
		"id":1,
		"url": "http://yulun.me",
		"code": "1pBnf",
		"hits": 0,
		"created_at": "2015-01-19T01:00:24+08:00"
	}
```

### Query url by short code [/api/v1/urls/:code]
Get short code's information.
```
curl 'http://localhost:3000/api/v1/urls/1pBnf'

	{
		"id":1,
		"url":"http://yulun.me",
		"code":"1pBnf",
		"hits":1,
		"created_at":"2015-01-19T01:00:24+08:00"
	}
```

### Redirect [/r/:code]
Redirect to url by assigned short code with status 301.

**Short code chars**: `abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ`


Demo
----
[Demo site](http://eit-shorten.herokuapp.com/) powered by Heroku

[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy?template=https://github.com/imZack/go-short)


License
-------
[MIT](http://yulun.mit-license.org/)