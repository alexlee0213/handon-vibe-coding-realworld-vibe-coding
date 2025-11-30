# API Documentation

RealWorld Conduit API follows the [RealWorld API Spec](https://realworld-docs.netlify.app/specifications/backend/endpoints/).

## Base URL

- **Development**: `http://localhost:8080/api`
- **Production**: `http://<ALB_DNS>/api`

## Authentication

The API uses JWT tokens for authentication. Include the token in the `Authorization` header:

```
Authorization: Token jwt.token.here
```

## Common Response Formats

### Success Response

```json
{
  "user": { ... },
  "article": { ... },
  "articles": [ ... ],
  "profile": { ... },
  "comments": [ ... ],
  "tags": [ ... ]
}
```

### Error Response

```json
{
  "errors": {
    "body": ["error message 1", "error message 2"]
  }
}
```

## Endpoints

### Health Check

#### GET /health

Check API health status.

**Response**: `200 OK`
```json
{
  "status": "ok"
}
```

---

### Authentication

#### POST /api/users

Register a new user.

**Request Body**:
```json
{
  "user": {
    "username": "jacob",
    "email": "jake@example.com",
    "password": "jakejake"
  }
}
```

**Response**: `201 Created`
```json
{
  "user": {
    "email": "jake@example.com",
    "token": "jwt.token.here",
    "username": "jacob",
    "bio": "",
    "image": null
  }
}
```

#### POST /api/users/login

Login with email and password.

**Request Body**:
```json
{
  "user": {
    "email": "jake@example.com",
    "password": "jakejake"
  }
}
```

**Response**: `200 OK`
```json
{
  "user": {
    "email": "jake@example.com",
    "token": "jwt.token.here",
    "username": "jacob",
    "bio": "I like to code",
    "image": "https://example.com/image.jpg"
  }
}
```

---

### User

#### GET /api/user

Get current user. **Authentication required**.

**Response**: `200 OK`
```json
{
  "user": {
    "email": "jake@example.com",
    "token": "jwt.token.here",
    "username": "jacob",
    "bio": "I like to code",
    "image": "https://example.com/image.jpg"
  }
}
```

#### PUT /api/user

Update current user. **Authentication required**.

**Request Body** (all fields optional):
```json
{
  "user": {
    "email": "newemail@example.com",
    "username": "newusername",
    "password": "newpassword",
    "bio": "New bio",
    "image": "https://example.com/newimage.jpg"
  }
}
```

**Response**: `200 OK`
```json
{
  "user": {
    "email": "newemail@example.com",
    "token": "jwt.token.here",
    "username": "newusername",
    "bio": "New bio",
    "image": "https://example.com/newimage.jpg"
  }
}
```

---

### Profiles

#### GET /api/profiles/:username

Get user profile. **Authentication optional**.

**Response**: `200 OK`
```json
{
  "profile": {
    "username": "jacob",
    "bio": "I like to code",
    "image": "https://example.com/image.jpg",
    "following": false
  }
}
```

#### POST /api/profiles/:username/follow

Follow a user. **Authentication required**.

**Response**: `200 OK`
```json
{
  "profile": {
    "username": "jacob",
    "bio": "I like to code",
    "image": "https://example.com/image.jpg",
    "following": true
  }
}
```

#### DELETE /api/profiles/:username/follow

Unfollow a user. **Authentication required**.

**Response**: `200 OK`
```json
{
  "profile": {
    "username": "jacob",
    "bio": "I like to code",
    "image": "https://example.com/image.jpg",
    "following": false
  }
}
```

---

### Articles

#### GET /api/articles

List articles. **Authentication optional**.

**Query Parameters**:
- `tag` - Filter by tag
- `author` - Filter by author username
- `favorited` - Filter by favorited by username
- `limit` - Limit (default: 20)
- `offset` - Offset (default: 0)

**Response**: `200 OK`
```json
{
  "articles": [
    {
      "slug": "how-to-train-your-dragon",
      "title": "How to train your dragon",
      "description": "Ever wonder how?",
      "body": "You have to believe",
      "tagList": ["dragons", "training"],
      "createdAt": "2024-01-01T12:00:00.000Z",
      "updatedAt": "2024-01-01T12:00:00.000Z",
      "favorited": false,
      "favoritesCount": 0,
      "author": {
        "username": "jacob",
        "bio": "I like to code",
        "image": "https://example.com/image.jpg",
        "following": false
      }
    }
  ],
  "articlesCount": 1
}
```

#### GET /api/articles/feed

Get articles from followed users. **Authentication required**.

**Query Parameters**:
- `limit` - Limit (default: 20)
- `offset` - Offset (default: 0)

**Response**: Same as GET /api/articles

#### POST /api/articles

Create an article. **Authentication required**.

**Request Body**:
```json
{
  "article": {
    "title": "How to train your dragon",
    "description": "Ever wonder how?",
    "body": "You have to believe",
    "tagList": ["dragons", "training"]
  }
}
```

**Response**: `201 Created`
```json
{
  "article": {
    "slug": "how-to-train-your-dragon",
    "title": "How to train your dragon",
    "description": "Ever wonder how?",
    "body": "You have to believe",
    "tagList": ["dragons", "training"],
    "createdAt": "2024-01-01T12:00:00.000Z",
    "updatedAt": "2024-01-01T12:00:00.000Z",
    "favorited": false,
    "favoritesCount": 0,
    "author": {
      "username": "jacob",
      "bio": "I like to code",
      "image": "https://example.com/image.jpg",
      "following": false
    }
  }
}
```

#### GET /api/articles/:slug

Get an article. **Authentication optional**.

**Response**: `200 OK`
```json
{
  "article": { ... }
}
```

#### PUT /api/articles/:slug

Update an article. **Authentication required** (author only).

**Request Body** (all fields optional):
```json
{
  "article": {
    "title": "Updated title",
    "description": "Updated description",
    "body": "Updated body"
  }
}
```

**Response**: `200 OK`
```json
{
  "article": { ... }
}
```

#### DELETE /api/articles/:slug

Delete an article. **Authentication required** (author only).

**Response**: `204 No Content`

---

### Favorites

#### POST /api/articles/:slug/favorite

Favorite an article. **Authentication required**.

**Response**: `200 OK`
```json
{
  "article": {
    ...
    "favorited": true,
    "favoritesCount": 1
  }
}
```

#### DELETE /api/articles/:slug/favorite

Unfavorite an article. **Authentication required**.

**Response**: `200 OK`
```json
{
  "article": {
    ...
    "favorited": false,
    "favoritesCount": 0
  }
}
```

---

### Comments

#### GET /api/articles/:slug/comments

Get comments for an article. **Authentication optional**.

**Response**: `200 OK`
```json
{
  "comments": [
    {
      "id": 1,
      "createdAt": "2024-01-01T12:00:00.000Z",
      "updatedAt": "2024-01-01T12:00:00.000Z",
      "body": "This is a comment",
      "author": {
        "username": "jacob",
        "bio": "I like to code",
        "image": "https://example.com/image.jpg",
        "following": false
      }
    }
  ]
}
```

#### POST /api/articles/:slug/comments

Add a comment to an article. **Authentication required**.

**Request Body**:
```json
{
  "comment": {
    "body": "This is a comment"
  }
}
```

**Response**: `201 Created`
```json
{
  "comment": {
    "id": 1,
    "createdAt": "2024-01-01T12:00:00.000Z",
    "updatedAt": "2024-01-01T12:00:00.000Z",
    "body": "This is a comment",
    "author": { ... }
  }
}
```

#### DELETE /api/articles/:slug/comments/:id

Delete a comment. **Authentication required** (author only).

**Response**: `204 No Content`

---

### Tags

#### GET /api/tags

Get all tags.

**Response**: `200 OK`
```json
{
  "tags": ["dragons", "training", "coding"]
}
```

---

## Error Codes

| Status Code | Description |
|-------------|-------------|
| 200 | OK - Request succeeded |
| 201 | Created - Resource created |
| 204 | No Content - Deleted successfully |
| 400 | Bad Request - Invalid request body |
| 401 | Unauthorized - Missing or invalid token |
| 403 | Forbidden - Not authorized for this action |
| 404 | Not Found - Resource not found |
| 422 | Unprocessable Entity - Validation errors |
| 500 | Internal Server Error |

## Rate Limiting

Currently no rate limiting is implemented. Consider adding rate limiting for production use.

## CORS

The API supports CORS with the following configuration:
- Allowed Origins: `*` (configurable)
- Allowed Methods: `GET, POST, PUT, DELETE, OPTIONS`
- Allowed Headers: `Authorization, Content-Type`
