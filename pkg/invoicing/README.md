# Invoicing

Provides REST services to deliver invoices and billing via [checkbook.io](http://checkbook.io).

# API

## `POST /v1/invoicing/send`

Sends an invoice to the recipient email specified in the payload.

### Content-Type
`application/json`

### Request Payload
```
{
	"first_name": <string>,
	"last_name": <string>,
	"email": <string>,
	"phone": <string>,
	"amount": <float>,

	// true if using checkbook.io, false for wire. If false,
 	// ignores amount and invoice_desc fields and only stores user
 	// data. defaults to false
	"is_checkbook": <bool>,

	// true if payment occurred with third party (i.e. Token Foundry)
	// If true, no invoices sent to checkbook.io and no wire transfer email
	// is sent internally, but user data is recorded and referral email is sent.
	// Mainly used to send referral emails from third party ETH or USD payers.
	"is_third_party": <bool>,

	// referred by code, none if not referred
	"referred_by": <string>
}
```

### Response Payload

`200` - Invoice was sent.

```
{
	"status": "ok"
}
```

OR

`400/500` - Something broke or invalid request

```
{
	"status": "<error message>",
	"code": <int code for error>
}
```


## `POST /v1/invoicing/cb`

Handles the webhook for `checkbook.io`.  It is defined in the API docs for checkbook.io here -> [https://checkbook.io/docs/api/overview/#document-webhooks](https://checkbook.io/docs/api/overview/#document-webhooks)

### Request Payload
```
{
	"id": <string>,
	"status": <string>
}

```


### Response Payload

`200` - Invoice was sent.

```
{
	"status": "ok"
}
```

OR

`400/500` - Something broke or invalid request

```
{
	"status": "<error message>",
	"code": <int code for error>
}
```

