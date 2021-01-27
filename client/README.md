# Client
The client binary provides an API for users to remotely control their group space node.
An important function of the client is to authenticate users, so only authorized people
can access the node. Note that users on the client do not correspond to members of a group
space, but rather who physically controls the node. Think of it like your account on a computer,
which is usually not very tightly linked to accounts on Google, Facebook, etc.

In the future, the client will allow users to start new group spaces, and access other critical
components of the group space API.


## Running
In the future, we are going to move this binary into a docker container and have it run automatically.
For now, just run `make run-client`, which starts the API on port 3000.

## TODO List
* Move to an embedded database rather than using files lol
* Add good unit tests (wait until after moving to a db)
* Dockerify
* RPC to communicate with coordinator container

## API

### Auth Routes
#### Login `POST /api/v1/login`
Login takes a username and password in the request body, and returns an auth token in the response if valid.
Auth tokens should be included in requests in the `Authorization` http header, for all requests except `bootstrap`, `version` and `login`.
```
Request Body:
{
    username: <string>,
    password: <string>
}

Request Response:
{
    auth_token: <string>
}   
```

#### Logout `POST /api/v1/logout
Logout just invalidates the current token a user is using. No body required, since the user id is determined
by the provided token in the `Authorization` header.


### User Routes
#### Bootstrap User: `POST /api/v1/bootstrap`
Creates the first admin user on the node. This can only be called when there are no users registered on the node. Body:

```
Request Body:
{
    username: <string>,
    password: <string>
}

Request Response:
HTTP response code
```

#### Create User: `POST /api/v1/users/`
Normal route for creating a user. This can only be called by users with admin permissions. 
TODO route for a user to change their password after they have been granted an account by an admin.
```
Request Body:
{
    username: <string>,
    password: <string>,
    permissions: {
        admin: <bool>
    }
}

Request Response:
HTTP response code
```

### Misc Routes
#### Version: `GET /api/v1/version`:
Just gets various version information about the node. No need for authentication, because it can help
users debug issues with their node if they are having trouble logging in.
```
Response Body:
{
    client_version: <string>
}
```
