***

##Channels API’s
###*List all defined channels*
**request**

```
GET /channels
```
**response**

```
200 OK
[
  {
    "id": 1,
    "name": "yang's channel",
    "description": "this is yang's first channel."
  },
  "..."
]
```

###*Show details of defined channel*
**request**

```
GET /channels/1
```
**response**

```
200 OK
{
  "id": 1,
  "name": "yang's channel",
  "description": "this is yang's first channel.",
  "tags": ["ip","device_id","product","brand","part_num"],
  "fields": {
    "temp": "float",
    "count": "int",
    "status": "string",
    "on_off": "boolean"
  }
}
```
Or, if not found

```
404 Not Found
```

###*Create a channel*
**request**

```
POST /channels
{
  "name": "yang's channel",
  "description": "this is yang's first channel that is .",
  "tags": ["ip", "device_id", "product", "brand", "part_num"],
  "fields": {
    "temp": "float",
    "count": "int",
    "status": "string",
    "on_off": "boolean"
  }
}
```

`Some validations need to be done here:`

`1. Fields such as name and description can't be empty. At least one field needs to be defined.`

`2. All tags and field names should only contain letters, numbers and underscores.`

`3. Only four data types are allowed in fields: float, int, string, boolean.`

`4. Tag names or field names should be unique.`

**response**

on success

```
200 OK
```
on failure

```
400 Bad Request
{
  "errors": {
    "description": "empty description."
  }
}
```
`Response returns detailed error messages on failure.`


###*Update existing channel*
**request**

```
PUT /channels/1
{
  "name": "yang's channel **updated**",
  "description": "",
  "tags": ["ip", "device_id", "product", "brand", "part_num", "**new_tag**"]
}
```

`Front end should show warning as anyone tries to update the existing channel. Reasons being:`

`1. Removing a tag will result in that tag will no longer be searchable, although the historical data might still exist. Removing a tag will also result that new data coming in to lose that tag from going onwards.`

`2. Removing the fields has the similar issues with tags.`

`3. Changing the type of the field is not allowed.`

`So to easier address these issues, what we can do is to allow only following updates:`

`1. Updating the name or description.`

`2. Adding new fields or tags.`

**response**

on success

```
200 OK
```
on failure

```
400 Bad Request
{
  "errors": {
    "description": "empty description."
  }
}
```

###*Delete an existing channel*
**request**

```
DELETE /channels/1
```

**response**

on success

```
200 OK
```

on failure

```
500 Internal Service Error
{
  "errors": {
    "database": "database error"
  }
}
```

***

##Dashboards API’s
###*List all defined dashboards*
**request**

```
GET /dashboards
```

**response**

```
200 OK
[
  {
    "id": 1,
    "name": "yang's dashboard",
    "description": "this is yang's first dashboard."
  },
  "..."
]
```

###*Show details of defined dashboard*
**request**

```
GET /dashboards/1
```

response

```
200 OK
{
  "id": 1,
  "name": "yang's dashboard",
  "description": "this is yang's first dashboard",
  "definition": "..."
}
```
Or, if not found

```
404 Not Found
```

`Dashboard definition currently is a big stringified json.`

###*Create a dashboard*
**request**

```
POST /dashboards
{
  "name": "yang's dashboard updated",
  "description": "this is yang's first dashboard.",
  "definition": "..."
}
```
`Some validations need to be done here:`

`1. Fields such as name, description and definition can't be empty.`

`2. Some other checks need to be done on front end to make sense of the dashboards.`

**response**

on success

```
201 Created
```

on failure

```
400 Bad Request
{
  "errors": {
    "name": "name is empty."
  }
}
```

###*Update an existing dashboard*
**request**

```
PUT /dashboards/1
{
  "name": "yang's dashboard updated"
}
```
`You can update one field, such as name, description, etc. Or`

```
{
  "definition": "..."
}
```
`If you want to update the dashboard definition, you need to send along the entire stringified json.`

**response**

on success

```
200 OK
```

on failure

```
404 Bad Request
{
  "errors": {
    "name": "name is empty"
  }
}
```

###*Delete an existing dashboard*
**request**

```
DELETE /dashboards/1
```
**response**

on success

```
200 OK
```

on failure

```
500 Internal Service Error
{
  "errors": {
    "database": "connection error"
  }
}
```
